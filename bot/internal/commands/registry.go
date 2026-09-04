package commands

import (
	"context"

	"github.com/0xProgress/simlife/bot/internal/commands/admin"
	"github.com/0xProgress/simlife/bot/internal/commands/business"
	"github.com/0xProgress/simlife/bot/internal/commands/economy"
	"github.com/0xProgress/simlife/bot/internal/commands/market"
	"github.com/0xProgress/simlife/bot/internal/commands/property"
	"github.com/bwmarrin/discordgo"
)

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

// PlayerIDKey is the context key for the authenticated player's database ID.
// The AuthMiddleware injects this into the context.
const PlayerIDKey ContextKey = "player_id"

// Handler is the signature for all command execution functions.
type Handler func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error

// Middleware defines a function that wraps a Handler to provide layered functionality.
type Middleware func(Handler) Handler

// CommandDef wraps a Discord application command and its handler mapping.
type CommandDef struct {
	Command      *discordgo.ApplicationCommand
	Handler      Handler
	Layer        string // e.g., "core", "market", "property", "business", "admin"
	RequiresAuth bool
}

// Chain applies a slice of middleware to a handler in the correct execution order.
// The first middleware in the slice is the outermost wrapper (executed first).
func Chain(h Handler, middlewares ...Middleware) Handler {
	if len(middlewares) == 0 {
		return h
	}
	// Recursively wrap the handler. 
	// Chain(h, m1, m2) -> m1(Chain(h, m2)) -> m1(m2(h))
	return middlewares[0](Chain(h, middlewares[1:]...))
}

// Registry holds all registered slash commands. 
// Adding a new command requires adding its definition here and implementing 
// the corresponding exported handler in the appropriate subpackage.
var Registry = []CommandDef{
	// ========================================================================
	// CORE ECONOMY (Layer 1-2)
	// ========================================================================
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "balance",
			Description: "View your wallet, bank, net worth, and 24h change",
		},
		Handler:      economy.HandleBalance,
		Layer:        "core",
		RequiresAuth: true,
	},
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "work",
			Description: "Work your current job or find a new one to earn wages",
		},
		Handler:      economy.HandleWork,
		Layer:        "core",
		RequiresAuth: true,
	},
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "bank",
			Description: "Deposit, withdraw, or transfer funds, and view history",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "action",
					Description: "The banking action to perform",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "deposit", Value: "deposit"},
						{Name: "withdraw", Value: "withdraw"},
						{Name: "transfer", Value: "transfer"},
						{Name: "history", Value: "history"},
					},
				},
				{
					Name:        "amount",
					Description: "The amount to deposit, withdraw, or transfer",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    false,
				},
				{
					Name:        "target",
					Description: "The user to transfer funds to",
					Type:        discordgo.ApplicationCommandOptionUser,
					Required:    false,
				},
			},
		},
		Handler:      economy.HandleBank,
		Layer:        "core",
		RequiresAuth: true,
	},
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "pay",
			Description: "Quickly transfer money directly to another player",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "target",
					Description: "The user to pay",
					Type:        discordgo.ApplicationCommandOptionUser,
					Required:    true,
				},
				{
					Name:        "amount",
					Description: "The amount to pay",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
				},
			},
		},
		Handler:      economy.HandlePay,
		Layer:        "core",
		RequiresAuth: true,
	},
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "shop",
			Description: "Browse and purchase basic needs from the NPC shop",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "item",
					Description: "The item to purchase or view",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "housing", Value: "housing"},
						{Name: "food", Value: "food"},
					},
				},
			},
		},
		Handler:      economy.HandleShop,
		Layer:        "core",
		RequiresAuth: true,
	},

	// ========================================================================
	// MARKET (Layer 2)
	// ========================================================================
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "market-list",
			Description: "List an item for sale on the global market",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "item",
					Description: "The item to list",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "quantity",
					Description: "Quantity to list",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
				},
				{
					Name:        "price",
					Description: "Price per unit",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
				},
			},
		},
		Handler:      market.HandleList,
		Layer:        "market",
		RequiresAuth: true,
	},
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "market-buy",
			Description: "Buy an item from an active market listing",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "listing_id",
					Description: "The ID of the listing to buy",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "quantity",
					Description: "Quantity to buy (defaults to 1)",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    false,
				},
			},
		},
		Handler:      market.HandleBuy,
		Layer:        "market",
		RequiresAuth: true,
	},
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "market-bid",
			Description: "Place a bid on an auction-style market listing",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "listing_id",
					Description: "The ID of the listing",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "amount",
					Description: "Your bid amount",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
				},
			},
		},
		Handler:      market.HandleBid,
		Layer:        "market",
		RequiresAuth: true,
	},
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "market-history",
			Description: "View the 30-day price history of a market item",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "item",
					Description: "The item to check",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
		Handler:      market.HandleHistory,
		Layer:        "market",
		RequiresAuth: false,
	},

	// ========================================================================
	// BUSINESS (Layer 4)
	// ========================================================================
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "business-open",
			Description: "Register and open a new player-owned business",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "name",
					Description: "The name of your business",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "type",
					Description: "The type of business (e.g., service, production)",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
		Handler:      business.HandleOpen,
		Layer:        "business",
		RequiresAuth: true,
	},
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "business-hire",
			Description: "Hire a player to work at your business",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "business",
					Description: "The ID of your business",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "target",
					Description: "The player to hire",
					Type:        discordgo.ApplicationCommandOptionUser,
					Required:    true,
				},
				{
					Name:        "wage",
					Description: "The daily wage rate to offer",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
				},
			},
		},
		Handler:      business.HandleHire,
		Layer:        "business",
		RequiresAuth: true,
	},
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "business-fire",
			Description: "Terminate a player's employment at your business",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "business",
					Description: "The ID of your business",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "target",
					Description: "The player to fire",
					Type:        discordgo.ApplicationCommandOptionUser,
					Required:    true,
				},
			},
		},
		Handler:      business.HandleFire,
		Layer:        "business",
		RequiresAuth: true,
	},
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "business-produce",
			Description: "Open your business for production today",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "business",
					Description: "The ID of your business",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
		Handler:      business.HandleProduce,
		Layer:        "business",
		RequiresAuth: true,
	},
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "business-status",
			Description: "View the operational status and metrics of your business",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "business",
					Description: "The ID of your business",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
		Handler:      business.HandleStatus,
		Layer:        "business",
		RequiresAuth: true,
	},

	// ========================================================================
	// PROPERTY (Layer 5)
	// ========================================================================
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "property-buy",
			Description: "Purchase an available plot of land or property",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "property_id",
					Description: "The ID of the property to buy",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
		Handler:      property.HandleBuy,
		Layer:        "property",
		RequiresAuth: true,
	},
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "property-sell",
			Description: "Sell a property you currently own",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "property_id",
					Description: "The ID of the property to sell",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
		Handler:      property.HandleSell,
		Layer:        "property",
		RequiresAuth: true,
	},
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "property-develop",
			Description: "Upgrade the development level of a property you own",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "property_id",
					Description: "The ID of the property to develop",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
		Handler:      property.HandleDevelop,
		Layer:        "property",
		RequiresAuth: true,
	},
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "property-rent",
			Description: "Pay monthly rent for a property you are leasing",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "property_id",
					Description: "The ID of the rented property",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
		Handler:      property.HandleRent,
		Layer:        "property",
		RequiresAuth: true,
	},

	// ========================================================================
	// ADMIN (Layer 0 / Ops)
	// ========================================================================
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "admin-reset",
			Description: "[ADMIN] Completely reset a player's economic state",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "target",
					Description: "The player to reset",
					Type:        discordgo.ApplicationCommandOptionUser,
					Required:    true,
				},
			},
		},
		Handler:      admin.HandleReset,
		Layer:        "admin",
		RequiresAuth: true,
	},
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "admin-grant",
			Description: "[ADMIN] Grant currency to a player from the treasury",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "target",
					Description: "The player to grant funds to",
					Type:        discordgo.ApplicationCommandOptionUser,
					Required:    true,
				},
				{
					Name:        "amount",
					Description: "The amount to grant",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
				},
			},
		},
		Handler:      admin.HandleGrant,
		Layer:        "admin",
		RequiresAuth: true,
	},
	{
		Command: &discordgo.ApplicationCommand{
			Name:        "admin-inspect",
			Description: "[ADMIN] View a comprehensive economic summary of a player",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "target",
					Description: "The player to inspect",
					Type:        discordgo.ApplicationCommandOptionUser,
					Required:    true,
				},
			},
		},
		Handler:      admin.HandleInspect,
		Layer:        "admin",
		RequiresAuth: true,
	},
}