package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DiscordToken        string
	DiscordAppID        string
	DiscordPublicKey    string
	PostgresDSN         string
	RedisAddr           string
	NatsURL             string
	JWTSecret           string
	SettlementCron      string
	ActivityClientID    string
	LogLevel            string
	LogFormat           string
	HTTPPort            string
	WorkCooldownMinutes int // Added for /work command cooldown
	Features            FeatureFlags
}

type FeatureFlags struct {
	MarketEnabled   bool
	PropertyEnabled bool
	BusinessEnabled bool
}

func Load() (*Config, error) {
	cfg := &Config{
		DiscordToken:        os.Getenv("DISCORD_BOT_TOKEN"),
		DiscordAppID:        os.Getenv("DISCORD_APP_ID"),
		DiscordPublicKey:    os.Getenv("DISCORD_PUBLIC_KEY"),
		PostgresDSN:         os.Getenv("POSTGRES_DSN"),
		RedisAddr:           os.Getenv("REDIS_ADDR"),
		NatsURL:             os.Getenv("NATS_URL"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		SettlementCron:      os.Getenv("SETTLEMENT_CRON"),
		ActivityClientID:    os.Getenv("ACTIVITY_CLIENT_ID"),
		LogLevel:            getEnvDefault("LOG_LEVEL", "info"),
		LogFormat:           getEnvDefault("LOG_FORMAT", "json"),
		HTTPPort:            getEnvDefault("HTTP_PORT", "8080"),
		WorkCooldownMinutes: getIntEnvDefault("WORK_COOLDOWN_MINUTES", 60),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	cfg.parseFeatureFlags()
	return cfg, nil
}

func (c *Config) validate() error {
	required := map[string]string{
		"DISCORD_BOT_TOKEN":  c.DiscordToken,
		"DISCORD_APP_ID":     c.DiscordAppID,
		"DISCORD_PUBLIC_KEY": c.DiscordPublicKey,
		"POSTGRES_DSN":       c.PostgresDSN,
		"REDIS_ADDR":         c.RedisAddr,
		"NATS_URL":           c.NatsURL,
		"JWT_SECRET":         c.JWTSecret,
		"SETTLEMENT_CRON":    c.SettlementCron,
		"ACTIVITY_CLIENT_ID": c.ActivityClientID,
	}

	for name, val := range required {
		if strings.TrimSpace(val) == "" {
			return fmt.Errorf("required environment variable %s is missing or empty", name)
		}
	}
	return nil
}

func (c *Config) parseFeatureFlags() {
	c.Features.MarketEnabled = parseBoolEnv("FEATURE_MARKET_ENABLED", true)
	c.Features.PropertyEnabled = parseBoolEnv("FEATURE_PROPERTY_ENABLED", true)
	c.Features.BusinessEnabled = parseBoolEnv("FEATURE_BUSINESS_ENABLED", true)
}

func getEnvDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getIntEnvDefault(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}

func parseBoolEnv(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultVal
	}
	return b
}