// bot/internal/commands/registry.go
package commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

type contextKey string

// PlayerIDKey is the context key for the authenticated player's ID.
const PlayerIDKey contextKey = "playerID"

// Handler is the signature for all command execution functions.
type Handler func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error

// Middleware defines a function that wraps a Handler to provide layered functionality.
type Middleware func(Handler) Handler

// CommandDef wraps a Discord application command and its handler mapping.
type CommandDef struct {
	Command      *discordgo.ApplicationCommand
	Handler      Handler
	Layer        string // e.g., "market", "property", "business", "core"
	RequiresAuth bool
}

// Chain applies a slice of middleware to a handler in the correct execution order.
func Chain(h Handler, middlewares ...Middleware) Handler {
	if len(middlewares) == 0 {
		return h
	}
	return middlewares[0](Chain(h, middlewares[1:]...))
}

// Registry holds all registered slash commands. Adding a new command requires adding it here.
var Registry = []CommandDef{
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "balance",
			Description: "Check your current bank balance and net worth",
		},
		Handler:      handleBalance,
		Layer:        "core",
		RequiresAuth: true,
	},
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "market",
			Description: "View global market prices and trade volumes",
		},
		Handler:      handleMarket,
		Layer:        "market",
		RequiresAuth: false,
	},
}

// handleBalance is a placeholder handler for the balance command.
func handleBalance(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Your balance is $0.00 (Financial Ledger integration pending)",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	return nil
}

// handleMarket is a placeholder handler for the market command.
func handleMarket(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Market data is currently static. (Market Engine integration pending)",
		},
	})
	return nil
}
