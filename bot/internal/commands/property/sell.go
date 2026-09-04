package property

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/economy"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// PropertySellData defines the data for the property sale confirmation image.
type PropertySellData struct {
	PropertyName  string
	ZoneType      string
	SalePrice     decimal.Decimal
	NewBalance    decimal.Decimal
}

// HandleSell processes the /property sell command.
func (s *Service) HandleSell(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	log := logger.FromContext(ctx, "commands.property")

	playerIDStr, ok := ctx.Value(commands.PlayerIDKey).(string)
	if !ok || playerIDStr == "" {
		log.Error().Msg("player_id missing from context in property sell handler")
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

	if !prop.OwnerPlayerID.Valid || prop.OwnerPlayerID.String != player.ID {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You do not own this property.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	salePrice := numericToDecimal(prop.AssessedValue)
	salePrice = salePrice.Mul(decimal.NewFromFloat(0.9)) // 10% depreciation on resale

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

	err = s.Ledger.Transfer(ctx, treasury.ID, wallet.ID, salePrice, "PROPERTY_SALE", player.ID, fmt.Sprintf("Property sale: %s", propertyID))
	if err != nil {
		return fmt.Errorf("ledger transfer failed: %w", err)
	}

	err = s.Queries.TransferPropertyOwnership(ctx, sqlc.TransferPropertyOwnershipParams{
		ID:             propertyID,
		OwnerPlayerID:  "",
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to clear ownership: %w", err)).Str("property_id", propertyID).Msg("property sell failed after payment")
		return fmt.Errorf("failed to clear property ownership: %w", err)
	}

	newWalletBal, _ := s.getAccountMetrics(ctx, wallet.ID)

	data := PropertySellData{
		PropertyName: fmt.Sprintf("Plot %s", propertyID[:8]),
		ZoneType:     string(prop.ZoneType),
		SalePrice:    salePrice,
		NewBalance:   newWalletBal,
	}

	imgBytes, err := s.Composer.Compose("property", data)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to compose property image: %w", err)).Msg("image composition failed")
		return s.sendSellTextFallback(sess, i, data)
	}

	err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Property Sold",
					Description: fmt.Sprintf("Successfully sold **%s** in the **%s** district for **⊄%s**.", data.PropertyName, data.ZoneType, salePrice.String()),
					Color:       0x4A90D9,
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
		Str("sale_price", salePrice.String()).
		Msg("property sold successfully")

	return nil
}

func (s *Service) sendSellTextFallback(sess *discordgo.Session, i *discordgo.InteractionCreate, data PropertySellData) error {
	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Property Sold",
					Description: fmt.Sprintf("Successfully sold **%s** in the **%s** district.\n\nSale Price: `⊄%s`\nNew Wallet Balance: `⊄%s`\n*(Image generation temporarily unavailable)*", data.PropertyName, data.ZoneType, data.SalePrice.String(), data.NewBalance.String()),
					Color:       0x4A90D9,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}