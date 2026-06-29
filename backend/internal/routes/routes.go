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
		// Campaign — the single staff claim action (register + redeem).
		api.POST("/claim", h.Claim)
		api.GET("/campaign", h.ActiveCampaignInfo)
		// Generic primitives (kept for the future membership product; not used by
		// the campaign-only FE).
		api.POST("/register", h.RegisterMember)
		api.POST("/lookup", h.Lookup)
		api.POST("/redeem", h.Redeem)

		// Admin — members + campaigns + analytics (auth to be added; none in dev).
		admin := api.Group("/admin")
		{
			admin.GET("/members", h.ListMembers)
			admin.GET("/menu-stats", h.MenuStats)
			admin.GET("/campaigns", h.ListCampaigns)
			admin.POST("/campaigns", h.CreateCampaign)
			// Production-house queue board (POS-style ticket flow).
			admin.GET("/queue", h.ProductionQueue)
			admin.POST("/queue/:id/ready", h.TicketReady)
			admin.POST("/queue/:id/done", h.TicketDone)
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
