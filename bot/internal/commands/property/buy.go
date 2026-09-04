package property

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/economy"
	"github.com/0xProgress/simlife/bot/internal/imaging"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

// Service holds the dependencies for all property command handlers.
type Service struct {
	Queries  *sqlc.Queries
	Ledger   *economy.Ledger
	Composer *imaging.Composer
}

// NewService initializes the property command service.
func NewService(q *sqlc.Queries, l *economy.Ledger, c *imaging.Composer) *Service {
	return &Service{
		Queries:  q,
		Ledger:   l,
		Composer: c,
	}
}

// PropertyBuyData defines the data for the property purchase confirmation image.
type PropertyBuyData struct {
	PropertyName   string
	ZoneType       string
	DevelopmentLvl int
	PurchasePrice  decimal.Decimal
	NewBalance     decimal.Decimal
}

// HandleBuy processes the /property buy command.
func (s *Service) HandleBuy(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	log := logger.FromContext(ctx, "commands.property")

	playerIDStr, ok := ctx.Value(commands.PlayerIDKey).(string)
	if !ok || playerIDStr == "" {
		log.Error().Msg("player_id missing from context in property buy handler")
		return fmt.Errorf("player_id missing from context")
	}

	var propertyID string
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == "property_id" {
			propertyID = opt.StringValue()
			break
		}
	}

	if propertyID == "" {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide a valid property ID.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	player, err := s.Queries.GetPlayerByID(ctx, playerIDStr)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch player: %w", err)).Str("player_id", playerIDStr).Msg("property buy failed")
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	prop, err := s.Queries.GetPropertyByID(ctx, propertyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "This property does not exist.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
		return fmt.Errorf("failed to fetch property: %w", err)
	}

	if prop.OwnerPlayerID.Valid {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "This property is already owned.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	purchasePrice := numericToDecimal(prop.AssessedValue)

	wallet, err := s.Queries.GetAccountByType(ctx, sqlc.GetAccountByTypeParams{
		PlayerID:    player.ID,
		AccountType: sqlc.AccountTypeWALLET,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch wallet: %w", err)
	}

	treasury, err := s.Queries.GetTreasuryAccount(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch treasury: %w", err)
	}

	err = s.Ledger.Transfer(ctx, wallet.ID, treasury.ID, purchasePrice, "PROPERTY_PURCHASE", player.ID, fmt.Sprintf("Property purchase: %s", propertyID))
	if err != nil {
		if errors.Is(err, economy.ErrInsufficientFunds) {
			return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Insufficient funds. This property costs ⊄%s.", purchasePrice.String()),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
		return fmt.Errorf("ledger transfer failed: %w", err)
	}

	err = s.Queries.TransferPropertyOwnership(ctx, sqlc.TransferPropertyOwnershipParams{
		ID:             propertyID,
		OwnerPlayerID:  player.ID,
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to transfer ownership: %w", err)).Str("property_id", propertyID).Msg("property buy failed after payment")
		return fmt.Errorf("failed to transfer property ownership: %w", err)
	}

	newWalletBal, _ := s.getAccountMetrics(ctx, wallet.ID)

	data := PropertyBuyData{
		PropertyName:   fmt.Sprintf("Plot %s", propertyID[:8]),
		ZoneType:       string(prop.ZoneType),
		DevelopmentLvl: int(prop.DevelopmentLevel),
		PurchasePrice:  purchasePrice,
		NewBalance:     newWalletBal,
	}

	imgBytes, err := s.Composer.Compose("property", data)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to compose property image: %w", err)).Msg("image composition failed")
		return s.sendBuyTextFallback(sess, i, data)
	}

	err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Property Purchased",
					Description: fmt.Sprintf("Successfully purchased **%s** in the **%s** district for **⊄%s**.", data.PropertyName, data.ZoneType, purchasePrice.String()),
					Color:       0x4CAF76,
					Image:       &discordgo.MessageEmbedImage{URL: "attachment://property.png"},
					Footer:      &discordgo.MessageEmbedFooter{Text: "Aether City Registry of Deeds"},
				},
			},
			Files: []*discordgo.File{
				{Name: "property.png", ContentType: "image/png", Reader: bytes.NewReader(imgBytes)},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		log.Error().Err(fmt.Errorf("failed to respond to interaction: %w", err)).Msg("discord response failed")
	}

	log.Info().
		Str("player_id", playerIDStr).
		Str("property_id", propertyID).
		Str("price", purchasePrice.String()).
		Msg("property purchased successfully")

	return nil
}

func (s *Service) sendBuyTextFallback(sess *discordgo.Session, i *discordgo.InteractionCreate, data PropertyBuyData) error {
	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Property Purchased",
					Description: fmt.Sprintf("Successfully purchased **%s** in the **%s** district.\n\nPurchase Price: `⊄%s`\nNew Wallet Balance: `⊄%s`\n*(Image generation temporarily unavailable)*", data.PropertyName, data.ZoneType, data.PurchasePrice.String(), data.NewBalance.String()),
					Color:       0x4CAF76,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

// Helper to convert pgtype.Numeric to decimal.Decimal
func numericToDecimal(n pgtype.Numeric) decimal.Decimal {
	if !n.Valid {
		return decimal.Zero
	}
	d, err := decimal.NewFromString(n.String())
	if err != nil {
		return decimal.Zero
	}
	return d
}

// Helper to get account balance
func (s *Service) getAccountMetrics(ctx context.Context, accountID string) (decimal.Decimal, error) {
	bal, err := s.Queries.GetAccountBalance(ctx, accountID)
	if err != nil {
		return decimal.Zero, err
	}
	return numericToDecimal(bal), nil
}