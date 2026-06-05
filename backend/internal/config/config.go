package config

import "os"

// Config holds all BFF settings, loaded from env with local-dev defaults.
type Config struct {
	Port    string
	GinMode string

	// Upstream KreaZcy engines (internal Docker network in prod, localhost in dev).
	KonsumZcyURL  string
	PromoZcyURL   string
	NotifikaZcyURL string
	AgregaZcyURL  string

	// Brand / campaign config (vocabulary lives in the BFF, never in the engines).
	DefaultCampaign string // tags every AgregaZcy event for attribution
	CoffeePrice     string // string so it's trivially env-overridable; parsed where needed

	// NotifikaZcy delivery toggle — when false, the voucher code is returned in the
	// API response instead of emailed (lets us ship before Resend is wired).
	NotifyEnabled bool
}

// Load returns config with sensible local-dev defaults, overridden by env vars.
func Load() *Config {
	cfg := &Config{
		Port:            getenv("PORT", "8080"),
		GinMode:         getenv("GIN_MODE", "debug"),
		KonsumZcyURL:    getenv("KONSUMZCY_URL", "http://localhost:8084"),
		PromoZcyURL:     getenv("PROMOZCY_URL", "http://localhost:8082"),
		NotifikaZcyURL:  getenv("NOTIFIKAZCY_URL", "http://localhost:8086"),
		AgregaZcyURL:    getenv("AGREGAZCY_URL", "http://localhost:5900"),
		DefaultCampaign: getenv("DEFAULT_CAMPAIGN", "launch-free-coffee"),
		CoffeePrice:     getenv("COFFEE_PRICE", "20000"),
		NotifyEnabled:   getenv("NOTIFY_ENABLED", "false") == "true",
	}
	return cfg
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
