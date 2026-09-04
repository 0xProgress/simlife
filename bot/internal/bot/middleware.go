package bot

import (
	"context"
	"fmt"
	"time"

	sqlc "github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/config"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
)

// rateLimitScript is a Lua script for atomic sliding window rate limiting.
// It removes expired entries, counts current entries, and adds a new entry if under the limit.
const rateLimitScript = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window_ms = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])

-- Remove entries outside the current window
redis.call('ZREMRANGEBYSCORE', key, 0, now - window_ms)

-- Count current entries in the window
local count = redis.call('ZCARD', key)

if count < limit then
    -- Add new entry with a unique member (timestamp + random to prevent collisions)
    redis.call('ZADD', key, now, now .. '-' .. math.random(100000))
    -- Set expiration to window + 1 second to allow natural cleanup
    redis.call('EXPIRE', key, math.ceil(window_ms / 1000) + 1)
    return 1
else
    return 0
end
`

// LoggingMiddleware logs the start and end of command execution with duration and player context.
func LoggingMiddleware() commands.Middleware {
	return func(next commands.Handler) commands.Handler {
		return func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
			start := time.Now()
			cmdName := i.ApplicationCommandData().Name

			log := logger.FromContext(ctx, "middleware")
			log.Debug().Str("command", cmdName).Msg("command execution initiated")

			err := next(ctx, s, i)

			duration := time.Since(start)
			if err != nil {
				log.Error().Err(fmt.Errorf("command execution failed: %w", err)).Dur("duration_ms", duration.Milliseconds()).Msg("command execution failed")
			} else {
				log.Info().Dur("duration_ms", duration.Milliseconds()).Msg("command execution completed successfully")
			}

			return err
		}
	}
}

// AuthMiddleware verifies the player exists (creating a new record and accounts if not) and injects the player ID.
func AuthMiddleware(dbPool *sqlc.DBTX) commands.Middleware {
	return func(next commands.Handler) commands.Handler {
		return func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
			var userID string
			var username string

			if i.Member != nil && i.Member.User != nil {
				userID = i.Member.User.ID
				username = i.Member.User.Username
			} else if i.User != nil {
				userID = i.User.ID
				username = i.User.Username
			} else {
				return fmt.Errorf("unable to determine user ID from interaction payload")
			}

			log := logger.FromContext(ctx, "middleware")
			queries := sqlc.New(dbPool)

			// 1. Get or create the player record
			player, err := queries.GetOrCreatePlayer(ctx, sqlc.GetOrCreatePlayerParams{
				DiscordID: userID,
				Username:  username,
			})
			if err != nil {
				log.Error().Err(fmt.Errorf("failed to get or create player: %w", err)).Msg("player auth failed")
				_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "An error occurred while initializing your player profile. Please try again.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return fmt.Errorf("player auth failed: %w", err)
			}

			// 2. Ensure the 3 core ledger accounts exist for this player
			err = queries.EnsurePlayerAccounts(ctx, sqlc.EnsurePlayerAccountsParams{
				PlayerID: player.ID,
			})
			if err != nil {
				log.Error().Err(fmt.Errorf("failed to ensure player accounts: %w", err)).Msg("player account initialization failed")
				// We don't fail the request here, as the accounts might already exist and the error could be transient,
				// but we log it for operator awareness. The ledger will catch missing accounts on first transaction.
			}

			// 3. Inject player ID into context for downstream handlers
			ctx = context.WithValue(ctx, PlayerIDKey, player.ID)

			// 4. Enrich the logger context with the resolved player ID for all subsequent logs in this request
			reqLogger := log.With().Str("player_id", player.ID).Logger()
			ctx = reqLogger.WithContext(ctx)

			return next(ctx, s, i)
		}
	}
}

// RateLimitMiddleware applies Redis-based sliding window rate limiting per player and per command.
func RateLimitMiddleware(redisClient *redis.Client, cfg *config.Config) commands.Middleware {
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

			cmdName := i.ApplicationCommandData().Name
			log := logger.FromContext(ctx, "middleware")

			// Default limits: 1 execution per 5 seconds
			limit := int64(1)
			window := 5 * time.Second

			// Special case for /work command based on config
			if cmdName == "work" {
				window = time.Duration(cfg.WorkCooldownMinutes) * time.Minute
			}

			key := fmt.Sprintf("ratelimit:%s:%s", userID, cmdName)
			now := time.Now().UnixMilli()
			windowMs := window.Milliseconds()

			result, err := redisClient.Eval(ctx, rateLimitScript, []string{key}, now, windowMs, limit).Result()
			if err != nil {
				log.Error().Err(fmt.Errorf("redis rate limit eval failed: %w", err)).Msg("rate limit check failed")
				// Fail closed to prevent spam/abuse if Redis is acting up
				return fmt.Errorf("rate limit check failed: %w", err)
			}

			allowed, ok := result.(int64)
			if !ok || allowed == 0 {
				_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("You are doing that too fast! Please wait %.0f seconds before using this command again.", window.Seconds()),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return fmt.Errorf("rate limit exceeded for user %s on command %s", userID, cmdName)
			}

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
						Content: "This feature is currently undergoing maintenance or is not yet available. Please check back later.",
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
	case "government":
		return cfg.Features.GovernmentEnabled
	case "crime":
		return cfg.Features.CrimeEnabled
	default:
		return true // Core features (Layer 1-2) are always enabled
	}
}
