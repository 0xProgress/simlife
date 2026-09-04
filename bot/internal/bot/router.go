package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/config"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	// SessionKey is the context key for the discordgo.Session.
	SessionKey contextKey = "discord_session"
	// InteractionKey is the context key for the discordgo.InteractionCreate.
	InteractionKey contextKey = "discord_interaction"
)

// Router manages the dispatch of Discord interactions to command handlers.
type Router struct {
	cfg         *config.Config
	dbPool      *pgxpool.Pool
	redisClient *redis.Client
	natsConn    *nats.Conn
	handlers    map[string]commands.Handler
}

// NewRouter initializes the command router and applies the global middleware chain.
func NewRouter(cfg *config.Config, dbPool *pgxpool.Pool, redisClient *redis.Client, natsConn *nats.Conn) *Router {
	r := &Router{
		cfg:         cfg,
		dbPool:      dbPool,
		redisClient: redisClient,
		natsConn:    natsConn,
		handlers:    make(map[string]commands.Handler),
	}

	// Apply the standard middleware chain to every registered command.
	// Order is critical:
	// 1. Logging: Captures the attempt and duration.
	// 2. RateLimit: Rejects spam before hitting the database.
	// 3. FeatureFlag: Rejects commands for layers not yet enabled.
	// 4. PlayerAuth: Ensures the player record exists and injects it into the context.
	middlewares := []commands.Middleware{
		LoggingMiddleware(),
		RateLimitMiddleware(redisClient, cfg),
		FeatureFlagMiddleware(cfg),
		PlayerAuthMiddleware(dbPool),
	}

	for _, cmd := range commands.Registry {
		r.handlers[cmd.Command.Name] = commands.Chain(cmd.Handler, middlewares...)
	}

	return r
}

// HandleInteraction processes incoming Discord interactions and routes them safely.
func (r *Router) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()
	cmdName := data.Name

	handler, ok := r.handlers[cmdName]
	if !ok {
		log := logger.Package("router")
		log.Error().
			Str("command", cmdName).
			Str("user_id", i.Member.User.ID).
			Msg("unknown command received from Discord")

		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "This command is currently unavailable or does not exist.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Discord requires an initial response within 3 seconds. 
	// We set a context timeout slightly shorter to ensure handlers fail fast 
	// and trigger our fallback response rather than leaving the user hanging.
	ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	defer cancel()

	// Inject Discord session and interaction into context for handlers/middleware that need them
	ctx = context.WithValue(ctx, SessionKey, s)
	ctx = context.WithValue(ctx, InteractionKey, i)

	// Execute the fully wrapped middleware chain
	err := handler(ctx, s, i)
	if err != nil {
		log := logger.Package("router")
		log.Error().
			Err(fmt.Errorf("command execution failed: %w", err)).
			Str("command", cmdName).
			Str("user_id", i.Member.User.ID).
			Msg("unhandled error bubbled up from command execution")

		// Fallback response: If the handler failed and didn't respond to the user,
		// we provide a generic error message. InteractionRespond will safely return 
		// an error if the handler already responded, which we ignore.
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "An unexpected error occurred while processing your request. The developers have been notified.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
}