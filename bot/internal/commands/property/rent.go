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

// PropertyRentData defines the data for the property rental confirmation image.
type PropertyRentData struct {
	PropertyName  string
	ZoneType      string
	MonthlyRent   decimal.Decimal
	NewBalance    decimal.Decimal
	IsLandlord    bool
}

// HandleRent processes the /property rent command.
func (s *Service) HandleRent(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	log := logger.FromContext(ctx, "commands.property")

	playerIDStr, ok := ctx.Value(commands.PlayerIDKey).(string)
	if !ok || playerIDStr == "" {
		log.Error().Msg("player_id missing from context in property rent handler")
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

	if !prop.OwnerPlayerID.Valid {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "This property has no owner. You must purchase it first.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	if prop.OwnerPlayerID.String == player.ID {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You cannot rent a property you own.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	propertyValue := numericToDecimal(prop.AssessedValue)
	monthlyRent := propertyValue.Mul(decimal.NewFromFloat(0.01)) // 1% of property value per month

	tenantWallet, err := s.Queries.GetAccountByType(ctx, sqlc.GetAccountByTypeParams{
		PlayerID:    player.ID,
		AccountType: sqlc.AccountTypeWALLET,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch tenant wallet: %w", err)
	}

	landlordWallet, err := s.Queries.GetAccountByType(ctx, sqlc.GetAccountByTypeParams{
		PlayerID:    prop.OwnerPlayerID.String,
		AccountType: sqlc.AccountTypeWALLET,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch landlord wallet: %w", err)
	}

	err = s.Ledger.Transfer(ctx, tenantWallet.ID, landlordWallet.ID, monthlyRent, "RENT_PAYMENT", player.ID, fmt.Sprintf("Rent payment: %s", propertyID))
	if err != nil {
		if errors.Is(err, economy.ErrInsufficientFunds) {
			return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Insufficient funds. Monthly rent is ⊄%s.", monthlyRent.String()),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
		return fmt.Errorf("ledger transfer failed: %w", err)
	}

	newTenantBal, _ := s.getAccountMetrics(ctx, tenantWallet.ID)

	data := PropertyRentData{
		PropertyName: fmt.Sprintf("Plot %s", propertyID[:8]),
		ZoneType:     string(prop.ZoneType),
		MonthlyRent:  monthlyRent,
		NewBalance:   newTenantBal,
		IsLandlord:   false,
	}

	imgBytes, err := s.Composer.Compose("property", data)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to compose property image: %w", err)).Msg("image composition failed")
		return s.sendRentTextFallback(sess, i, data)
	}

	err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Rent Paid",
					Description: fmt.Sprintf("Successfully paid **⊄%s** monthly rent for **%s**.", monthlyRent.String(), data.PropertyName),
					Color:       0x4CAF76,
					Image:       &discordgo.MessageEmbedImage{URL: "attachment://property.png"},
					Footer:      &discordgo.MessageEmbedFooter{Text: "Aether City Tenancy Board"},
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
		Str("tenant_id", playerIDStr).
		Str("landlord_id", prop.OwnerPlayerID.String).
		Str("property_id", propertyID).
		Str("rent", monthlyRent.String()).
		Msg("rent paid successfully")

	return nil
}

func (s *Service) sendRentTextFallback(sess *discordgo.Session, i *discordgo.InteractionCreate, data PropertyRentData) error {
	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Rent Paid",
					Description: fmt.Sprintf("Successfully paid **⊄%s** monthly rent for **%s**.\n\nNew Wallet Balance: `⊄%s`\n*(Image generation temporarily unavailable)*", data.MonthlyRent.String(), data.PropertyName, data.NewBalance.String()),
					Color:       0x4CAF76,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}