// Package routes holds the BFF's client-facing endpoints. This is where all the
// tana arabica brand vocabulary ("member", "kopi gratis", campaigns) lives — the
// KreaZcy engines stay generic behind the clients.
package routes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/KreaZcy/arbc-membership-backend/internal/clients"
	"github.com/KreaZcy/arbc-membership-backend/internal/config"
)

// Handlers wires the engine clients + config into the route methods.
type Handlers struct {
	Cfg     *config.Config
	Konsum  *clients.KonsumZcy
	Promo   *clients.PromoZcy
	Notify  *clients.NotifikaZcy
	Agrega  *clients.AgregaZcy
}

// Register mounts all BFF routes on the gin engine.
func (h *Handlers) Register(r *gin.Engine) {
	api := r.Group("/api")
	{
		// Public — the data-for-value loop + barista flow.
		api.POST("/register", h.RegisterMember)
		api.POST("/lookup", h.Lookup)
		api.POST("/redeem", h.Redeem)

		// Admin — members + campaigns + analytics (auth to be added; none in dev).
		admin := api.Group("/admin")
		{
			admin.GET("/members", h.ListMembers)
			admin.GET("/campaigns", h.ListCampaigns)
			admin.POST("/campaigns", h.CreateCampaign)
		}
	}
}

// coffeePrice parses the configured price; defaults to 0 if unset/garbage.
func (h *Handlers) coffeePrice() float64 {
	v, _ := strconv.ParseFloat(h.Cfg.CoffeePrice, 64)
	return v
}

// fail maps an upstream/engine error to a sanitized client response (no leaks).
func fail(c *gin.Context, err error) {
	var ue *clients.UpstreamError
	if errors.As(err, &ue) {
		// Pass through 4xx (client's fault, e.g. invalid code); mask 5xx as 502.
		if ue.Status >= 400 && ue.Status < 500 {
			c.JSON(ue.Status, gin.H{"code": "UPSTREAM_REJECTED", "message": ue.Error()})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"code": "UPSTREAM_ERROR", "message": "a downstream service is unavailable"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL", "message": "unexpected error"})
}
