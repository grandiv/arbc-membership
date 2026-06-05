// arbc-membership BFF — wires the KreaZcy engines (KonsumZcy, PromoZcy,
// NotifikaZcy, AgregaZcy) behind a single API the tana arabica frontend talks to.
package main

import (
	"log/slog"
	"net/http"
	"os"

	kzcydash "github.com/KreaZcy/kzcy-dashboard"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/KreaZcy/arbc-membership-backend/internal/clients"
	"github.com/KreaZcy/arbc-membership-backend/internal/config"
	"github.com/KreaZcy/arbc-membership-backend/internal/routes"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "arbc-membership"))
	gin.SetMode(cfg.GinMode)

	h := &routes.Handlers{
		Cfg:    cfg,
		Konsum: clients.NewKonsumZcy(cfg.KonsumZcyURL),
		Promo:  clients.NewPromoZcy(cfg.PromoZcyURL),
		Notify: clients.NewNotifikaZcy(cfg.NotifikaZcyURL, cfg.NotifyEnabled),
		Agrega: clients.NewAgregaZcy(cfg.AgregaZcyURL, cfg.DefaultCampaign),
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsAllowAll())

	r.GET("/health", gin.WrapH(kzcydash.HealthHandler("arbc-membership")))
	h.Register(r)

	slog.Info("arbc-membership BFF listening", "port", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
		slog.Error("server stopped", "err", err)
	}
}

// corsAllowAll is a permissive dev CORS middleware. TODO: replace with
// kzcy-middleware/cors (net/http-style; needs a gin adapter) for prod parity.
func corsAllowAll() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Request-ID")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
