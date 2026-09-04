package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration for the Simlife bot.
// It is the single source of truth for how the bot behaves in a given deployment environment.
type Config struct {
	// Discord configuration
	DiscordToken     string
	DiscordAppID     string
	DiscordPublicKey string

	// Infrastructure configuration
	PostgresDSN string
	RedisAddr   string
	NatsURL     string

	// API and Security configuration
	JWTSecret string
	HTTPPort  string

	// Game and Feature configuration
	SettlementCron      string
	ActivityClientID    string
	WorkCooldownMinutes int

	// Logging configuration
	LogLevel  string
	LogFormat string

	// Feature flags for progressive layer rollout
	Features FeatureFlags

	AnalyticsSecret     string `env:"ANALYTICS_SECRET"`
	DiscordClientSecret string `env:"DISCORD_CLIENT_SECRET"`
}

// FeatureFlags controls the enablement of specific game layers, allowing for
// safe, progressive rollout without requiring code changes between environments.
type FeatureFlags struct {
	MarketEnabled     bool
	PropertyEnabled   bool
	BusinessEnabled   bool
	GovernmentEnabled bool
	CrimeEnabled      bool
}

// Load reads all runtime configuration from environment variables, applies defaults,
// and validates that all required values are present.
func Load() (*Config, error) {
	cfg := &Config{
		DiscordToken:     strings.TrimSpace(os.Getenv("DISCORD_BOT_TOKEN")),
		DiscordAppID:     strings.TrimSpace(os.Getenv("DISCORD_APP_ID")),
		DiscordPublicKey: strings.TrimSpace(os.Getenv("DISCORD_PUBLIC_KEY")),
		PostgresDSN:      strings.TrimSpace(os.Getenv("POSTGRES_DSN")),
		RedisAddr:        strings.TrimSpace(os.Getenv("REDIS_ADDR")),
		NatsURL:          strings.TrimSpace(os.Getenv("NATS_URL")),
		JWTSecret:        strings.TrimSpace(os.Getenv("JWT_SECRET")),
		SettlementCron:   strings.TrimSpace(os.Getenv("SETTLEMENT_CRON")),
		ActivityClientID: strings.TrimSpace(os.Getenv("ACTIVITY_CLIENT_ID")),

		LogLevel:            getEnvDefault("LOG_LEVEL", "info"),
		LogFormat:           getEnvDefault("LOG_FORMAT", "json"),
		HTTPPort:            getEnvDefault("HTTP_PORT", "8080"),
		WorkCooldownMinutes: getIntEnvDefault("WORK_COOLDOWN_MINUTES", 60),
	}

	cfg.parseFeatureFlags()

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	var missing []string

	if c.DiscordToken == "" {
		missing = append(missing, "DISCORD_BOT_TOKEN")
	}
	if c.DiscordAppID == "" {
		missing = append(missing, "DISCORD_APP_ID")
	}
	if c.DiscordPublicKey == "" {
		missing = append(missing, "DISCORD_PUBLIC_KEY")
	}
	if c.PostgresDSN == "" {
		missing = append(missing, "POSTGRES_DSN")
	}
	if c.RedisAddr == "" {
		missing = append(missing, "REDIS_ADDR")
	}
	if c.NatsURL == "" {
		missing = append(missing, "NATS_URL")
	}
	if c.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if c.SettlementCron == "" {
		missing = append(missing, "SETTLEMENT_CRON")
	}
	if c.ActivityClientID == "" {
		missing = append(missing, "ACTIVITY_CLIENT_ID")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: [%s]", strings.Join(missing, ", "))
	}

	if c.WorkCooldownMinutes <= 0 {
		return fmt.Errorf("WORK_COOLDOWN_MINUTES must be greater than 0, got %d", c.WorkCooldownMinutes)
	}

	if c.LogFormat != "json" && c.LogFormat != "pretty" {
		return fmt.Errorf("LOG_FORMAT must be 'json' or 'pretty', got '%s'", c.LogFormat)
	}

	if c.AnalyticsSecret == "" {
		missing = append(missing, "ANALYTICS_SECRET")
	}
	if c.DiscordClientSecret == "" {
		missing = append(missing, "DISCORD_CLIENT_SECRET")
	}

	return nil
}

func (c *Config) parseFeatureFlags() {
	// Layer 2+
	c.Features.MarketEnabled = parseBoolEnv("FEATURE_MARKET_ENABLED", true)
	// Layer 3+ (Default false to enforce progressive rollout)
	c.Features.PropertyEnabled = parseBoolEnv("FEATURE_PROPERTY_ENABLED", false)
	// Layer 4+
	c.Features.BusinessEnabled = parseBoolEnv("FEATURE_BUSINESS_ENABLED", false)
	// Layer 7+
	c.Features.GovernmentEnabled = parseBoolEnv("FEATURE_GOVERNMENT_ENABLED", false)
	// Layer 9+
	c.Features.CrimeEnabled = parseBoolEnv("FEATURE_CRIME_ENABLED", false)
}

func getEnvDefault(key, defaultVal string) string {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		return val
	}
	return defaultVal
}

func getIntEnvDefault(key string, defaultVal int) int {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		// Falling back to default is the safest action for non-critical
		// integer configs to prevent startup crashes on malformed env vars.
		return defaultVal
	}
	return i
}

func parseBoolEnv(key string, defaultVal bool) bool {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultVal
	}
	return b
}
