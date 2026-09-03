package bot

import (
	"context"

	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/config"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
)

// Router manages the dispatch of Discord interactions to command handlers.
type Router struct {
	cfg      *config.Config
	handlers map[string]commands.Handler
}

// NewRouter initializes the command router and applies the global middleware chain.
func NewRouter(cfg *config.Config) *Router {
	r := &Router{
		cfg:      cfg,
		handlers: make(map[string]commands.Handler),
	}

	// Apply the standard middleware chain to every registered command
	middlewares := []commands.Middleware{
		LoggingMiddleware(),
		AuthMiddleware(),
		RateLimitMiddleware(),
		FeatureFlagMiddleware(cfg),
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

	cmdName := i.ApplicationCommandData().Name
	handler, ok := r.handlers[cmdName]
	if !ok {
		log := logger.Package("router")
		log.Error().
			Str("command", cmdName).
			Msg("unknown command received from Discord")
		return
	}

	ctx := context.Background()

	// Execute the fully wrapped middleware chain
	err := handler(ctx, s, i)
	if err != nil {
		// Middleware handles its own logging for failures, but we catch unhandled panics/errors here
		log := logger.Package("router")
		log.Error().
			Err(err).
			Str("command", cmdName).
			Msg("unhandled error bubbled up from command execution")
	}
}
