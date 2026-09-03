// bot/internal/bot/middleware.go
package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/simlife/bot/internal/commands"
	"github.com/simlife/bot/internal/config"
	"github.com/simlife/bot/internal/logger"
)

// LoggingMiddleware logs the start and end of command execution with duration.
func LoggingMiddleware() commands.Middleware {
	return func(next commands.Handler) commands.Handler {
		return func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
			start := time.Now()
			cmdName := i.ApplicationCommandData().Name
			
			log := logger.Package("middleware").With().
				Str("command", cmdName).
				Logger()

			log.Debug().Msg("command execution initiated")

			err := next(ctx, s, i)

			duration := time.Since(start)
			if err != nil {
				log.Error().Err(err).Dur("duration", duration).Msg("command execution failed")
			} else {
				log.Info().Dur("duration", duration).Msg("command execution completed successfully")
			}

			return err
		}
	}
}

// AuthMiddleware verifies the player exists (creating a new record if not) and injects ID.
func AuthMiddleware() commands.Middleware {
	return func(next commands.Handler) commands.Handler {
		return func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
			var userID string
			if i.Member != nil && i.Member.User != nil {
				userID = i.Member.User.ID
			} else if i.User != nil {
				userID = i.User.ID
			} else {
				return fmt.Errorf("unable to determine user ID from interaction payload")
			}

			// TODO: Query PostgreSQL via ledger/db package to confirm player existence.
			// Create new player ledger accounts if they do not exist.
			
			ctx = context.WithValue(ctx, commands.PlayerIDKey, userID)
			return next(ctx, s, i)
		}
	}
}

// RateLimitMiddleware applies Redis-based rate limiting per player and per command.
func RateLimitMiddleware() commands.Middleware {
	return func(next commands.Handler) commands.Handler {
		return func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
			// TODO: Implement Redis sliding window rate limiting logic here.
			return next(ctx, s, i)
		}
	}
}

// FeatureFlagMiddleware checks whether the command's layer is enabled in config.
func FeatureFlagMiddleware(cfg *config.Config) commands.Middleware {
	return func(next commands.Handler) commands.Handler {
		return func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
			cmdName := i.ApplicationCommandData().Name
			
			var layer string
			for _, cmd := range commands.Registry {
				if cmd.Command.Name == cmdName {
					layer = cmd.Layer
					break
				}
			}

			if !isLayerEnabled(layer, cfg) {
				_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "This feature is currently undergoing maintenance. Please check back later.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return fmt.Errorf("feature layer '%s' is disabled via feature flags", layer)
			}

			return next(ctx, s, i)
		}
	}
}

func isLayerEnabled(layer string, cfg *config.Config) bool {
	switch layer {
	case "market":
		return cfg.Features.MarketEnabled
	case "property":
		return cfg.Features.PropertyEnabled
	case "business":
		return cfg.Features.BusinessEnabled
	default:
		return true // Core features are always enabled
	}
}