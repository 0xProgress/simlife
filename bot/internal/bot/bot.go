package bot

import (
	"fmt"

	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/config"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

// Bot manages the DiscordGo session, lifecycle events, and command routing.
type Bot struct {
	cfg         *config.Config
	session     *discordgo.Session
	router      *Router
	dbPool      *pgxpool.Pool
	redisClient *redis.Client
	natsConn    *nats.Conn
}

// NewBot initializes the Discord session, applies intents, and wires the router.
func NewBot(cfg *config.Config, dbPool *pgxpool.Pool, redisClient *redis.Client, natsConn *nats.Conn) (*Bot, error) {
	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	// Set required intents for slash commands and guild presence
	session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages

	// Initialize the router with necessary dependencies
	router := NewRouter(cfg, dbPool, redisClient, natsConn)

	b := &Bot{
		cfg:         cfg,
		session:     session,
		router:      router,
		dbPool:      dbPool,
		redisClient: redisClient,
		natsConn:    natsConn,
	}

	// Route all interactions through the central command router
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		b.router.HandleInteraction(s, i)
	})

	// Lifecycle hooks for operational logging
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log := logger.Package("bot")
		log.Info().
			Str("username", s.State.User.Username).
			Str("bot_id", s.State.User.ID).
			Msg("bot successfully connected to Discord gateway and is ready")
	})

	session.AddHandler(func(s *discordgo.Session, c *discordgo.Connect) {
		log := logger.Package("bot")
		log.Info().Msg("websocket connection established")
	})

	session.AddHandler(func(s *discordgo.Session, d *discordgo.Disconnect) {
		log := logger.Package("bot")
		log.Warn().Msg("websocket disconnected, attempting automatic reconnection")
	})

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Resumed) {
		log := logger.Package("bot")
		log.Info().Msg("websocket session resumed")
	})

	return b, nil
}

// RegisterCommands syncs the command registry with Discord's API globally.
// It uses BulkOverwrite to ensure the remote state exactly matches the local registry,
// preventing orphaned commands and avoiding per-command rate limits.
func (b *Bot) RegisterCommands() error {
	log := logger.Package("bot")

	// Extract the ApplicationCommand definitions from the registry
	var appCommands []*discordgo.ApplicationCommand
	for _, cmd := range commands.Registry {
		appCommands = append(appCommands, cmd.Command)
	}

	log.Info().Int("count", len(appCommands)).Msg("syncing slash commands with Discord API")

	_, err := b.session.ApplicationCommandBulkOverwrite(b.cfg.DiscordAppID, "", appCommands)
	if err != nil {
		return fmt.Errorf("failed to bulk overwrite commands: %w", err)
	}

	log.Info().Msg("slash commands synced successfully")
	return nil
}

// Start opens the Discord gateway connection. It is non-blocking, allowing 
// the main function to proceed with starting the HTTP server and scheduler.
func (b *Bot) Start() error {
	log := logger.Package("bot")
	log.Info().Msg("opening Discord gateway connection")

	err := b.session.Open()
	if err != nil {
		return fmt.Errorf("failed to open Discord gateway: %w", err)
	}

	// Set initial economic status presence
	b.UpdateEconomicStatus("Initializing economy...")

	log.Info().Msg("Discord gateway connection opened successfully")
	return nil
}

// Close gracefully closes the Discord gateway connection, stopping all 
// incoming interaction handling and cleaning up websocket resources.
func (b *Bot) Close() error {
	log := logger.Package("bot")
	log.Info().Msg("closing Discord gateway connection")
	return b.session.Close()
}

// UpdateEconomicStatus updates the bot's presence message to reflect the 
// current economic health indicator. This is called by the settlement engine 
// after each daily close to provide a live, visible heartbeat of the world state.
func (b *Bot) UpdateEconomicStatus(status string) {
	err := b.session.UpdateGameStatus(0, status)
	if err != nil {
		log := logger.Package("bot")
		log.Error().Err(fmt.Errorf("presence update failed: %w", err)).Msg("failed to update presence status")
	}
}