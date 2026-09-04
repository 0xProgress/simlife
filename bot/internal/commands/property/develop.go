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

// PropertyDevelopData defines the data for the property development confirmation image.
type PropertyDevelopData struct {
	PropertyName     string
	ZoneType         string
	DevelopmentLvl   int
	DevelopmentCost  decimal.Decimal
	NewPropertyValue decimal.Decimal
	NewBalance       decimal.Decimal
}

// HandleDevelop processes the /property develop command.
func (s *Service) HandleDevelop(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	log := logger.FromContext(ctx, "commands.property")

	playerIDStr, ok := ctx.Value(commands.PlayerIDKey).(string)
	if !ok || playerIDStr == "" {
		log.Error().Msg("player_id missing from context in property develop handler")
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

	if prop.DevelopmentLevel >= 5 {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "This property is already at maximum development level.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	currentValue := numericToDecimal(prop.AssessedValue)
	developmentCost := currentValue.Mul(decimal.NewFromFloat(0.25)) // 25% of current value

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

	err = s.Ledger.Transfer(ctx, wallet.ID, treasury.ID, developmentCost, "PROPERTY_DEVELOPMENT", player.ID, fmt.Sprintf("Property development: %s", propertyID))
	if err != nil {
		if errors.Is(err, economy.ErrInsufficientFunds) {
			return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Insufficient funds. Development costs ⊄%s.", developmentCost.String()),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
		return fmt.Errorf("ledger transfer failed: %w", err)
	}

	err = s.Queries.UpgradePropertyDevelopment(ctx, sqlc.UpgradePropertyDevelopmentParams{
		ID:            propertyID,
		OwnerPlayerID: player.ID,
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to upgrade development: %w", err)).Str("property_id", propertyID).Msg("property development failed after payment")
		return fmt.Errorf("failed to upgrade property development: %w", err)
	}

	newPropertyValue := currentValue.Mul(decimal.NewFromFloat(1.25))
	newWalletBal, _ := s.getAccountMetrics(ctx, wallet.ID)

	data := PropertyDevelopData{
		PropertyName:     fmt.Sprintf("Plot %s", propertyID[:8]),
		ZoneType:         string(prop.ZoneType),
		DevelopmentLvl:   int(prop.DevelopmentLevel) + 1,
		DevelopmentCost:  developmentCost,
		NewPropertyValue: newPropertyValue,
		NewBalance:       newWalletBal,
	}

	imgBytes, err := s.Composer.Compose("property", data)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to compose property image: %w", err)).Msg("image composition failed")
		return s.sendDevelopTextFallback(sess, i, data)
	}

	err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Property Developed",
					Description: fmt.Sprintf("Successfully upgraded **%s** to development level **%d**.", data.PropertyName, data.DevelopmentLvl),
					Color:       0x8a6fd8, // Purple for premium/development
					Image:       &discordgo.MessageEmbedImage{URL: "attachment://property.png"},
					Footer:      &discordgo.MessageEmbedFooter{Text: "Aether City Planning Commission"},
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
		Int("new_level", data.DevelopmentLvl).
		Str("cost", developmentCost.String()).
		Msg("property developed successfully")

	return nil
}

func (s *Service) sendDevelopTextFallback(sess *discordgo.Session, i *discordgo.InteractionCreate, data PropertyDevelopData) error {
	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Property Developed",
					Description: fmt.Sprintf("Successfully upgraded **%s** to development level **%d**.\n\nDevelopment Cost: `⊄%s`\nNew Property Value: `⊄%s`\nNew Wallet Balance: `⊄%s`\n*(Image generation temporarily unavailable)*", data.PropertyName, data.DevelopmentLvl, data.DevelopmentCost.String(), data.NewPropertyValue.String(), data.NewBalance.String()),
					Color:       0x8a6fd8,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}