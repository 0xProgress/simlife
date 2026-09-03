package bot

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/config"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
)

// Bot manages the DiscordGo session and lifecycle events.
type Bot struct {
	cfg     *config.Config
	session *discordgo.Session
	router  *Router
}

// NewBot initializes the Discord session, applies intents, and wires the router.
func NewBot(cfg *config.Config) (*Bot, error) {
	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	// Set required intents for slash commands and guild presence
	session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages

	router := NewRouter(cfg)

	b := &Bot{
		cfg:     cfg,
		session: session,
		router:  router,
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
			Str("discriminator", s.State.User.Discriminator).
			Msg("bot successfully connected to Discord gateway")
	})

	session.AddHandler(func(s *discordgo.Session, c *discordgo.Connect) {
		log := logger.Package("bot")
		log.Info().Msg("websocket connection established")
	})

	session.AddHandler(func(s *discordgo.Session, d *discordgo.Disconnect) {
		log := logger.Package("bot")
		log.Warn().Msg("websocket disconnected, attempting automatic reconnection")
	})

	// Fix: The correct event struct for a resumed session is *discordgo.Resumed
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Resumed) {
		log := logger.Package("bot")
		log.Info().Msg("websocket session resumed")
	})

	return b, nil
}

// RegisterCommands syncs the command registry with Discord's API globally.
func (b *Bot) RegisterCommands() error {
	log := logger.Package("bot")
	log.Info().Int("count", len(commands.Registry)).Msg("syncing slash commands with Discord API")

	for _, cmd := range commands.Registry {
		_, err := b.session.ApplicationCommandCreate(b.cfg.DiscordAppID, "", cmd.Command)
		if err != nil {
			return fmt.Errorf("failed to register command %s: %w", cmd.Command.Name, err)
		}
		log.Debug().Str("command", cmd.Command.Name).Msg("command registered")
	}

	return nil
}

// Start opens the Discord gateway connection and blocks until a shutdown signal is received.
func (b *Bot) Start() error {
	log := logger.Package("bot")
	log.Info().Msg("opening Discord gateway connection")

	err := b.session.Open()
	if err != nil {
		return fmt.Errorf("failed to open Discord gateway: %w", err)
	}

	// Set initial economic status presence
	b.updateStatus("Initializing economy...")

	log.Info().Msg("bot is now running. Awaiting shutdown signal (CTRL+C)")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Info().Msg("closing Discord gateway connection")
	return b.session.Close()
}

// updateStatus updates the bot's presence message to reflect economic health.
func (b *Bot) updateStatus(status string) {
	err := b.session.UpdateGameStatus(0, status)
	if err != nil {
		log := logger.Package("bot")
		log.Error().Err(err).Msg("failed to update presence status")
	}
}
