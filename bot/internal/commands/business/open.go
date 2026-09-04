package business

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	sqlc "github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/economy"
	"github.com/0xProgress/simlife/bot/internal/imaging"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/shopspring/decimal"
)

// OpenLayoutData defines the data for the business opening confirmation image.
type OpenLayoutData struct {
	BusinessName    string
	BusinessType    string
	OwnerName       string
	RegistrationFee decimal.Decimal
}

// HandleOpen processes the /business open command.
func HandleOpen(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate, queries *sqlc.Queries, ledger *economy.Ledger, composer *imaging.Composer) error {
	log := logger.FromContext(ctx, "commands.business")

	playerIDStr, ok := ctx.Value(commands.PlayerIDKey).(string)
	if !ok || playerIDStr == "" {
		log.Error().Msg("player_id missing from context")
		return fmt.Errorf("player_id missing from context")
	}

	player, err := queries.GetPlayerByID(ctx, playerIDStr)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch player: %w", err)).Str("player_id", playerIDStr).Msg("business open failed")
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	var businessName, businessType string
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == "name" {
			businessName = opt.StringValue()
		}
		if opt.Name == "type" {
			businessType = opt.StringValue()
		}
	}

	if businessName == "" || businessType == "" {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide both a business name and type.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	regFee := decimal.NewFromInt(5000)

	wallet, err := queries.GetAccountByType(ctx, sqlc.GetAccountByTypeParams{
		PlayerID:    player.ID,
		AccountType: sqlc.AccountTypeWALLET,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch wallet: %w", err)
	}

	treasury, err := queries.GetTreasuryAccount(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch treasury: %w", err)
	}

	err = ledger.Transfer(ctx, wallet.ID, treasury.ID, regFee, "BUSINESS_REGISTRATION", player.ID, fmt.Sprintf("Business Registration: %s", businessName))
	if err != nil {
		if errors.Is(err, economy.ErrInsufficientFunds) {
			return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Insufficient funds. Business registration requires ⊄%s.", regFee.String()),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
		return fmt.Errorf("ledger transfer failed: %w", err)
	}

	biz, err := queries.CreateBusiness(ctx, sqlc.CreateBusinessParams{
		OwnerID:      player.ID,
		BusinessType: businessType,
		Name:         businessName,
		CityPlotID:   "", // Layer 5 feature, empty for Layer 4
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to create business: %w", err)).Str("player_id", playerIDStr).Msg("business creation failed after fee deduction")
		return fmt.Errorf("failed to create business record: %w", err)
	}

	data := OpenLayoutData{
		BusinessName:    biz.Name,
		BusinessType:    biz.BusinessType,
		OwnerName:       player.Username,
		RegistrationFee: regFee,
	}

	imgBytes, err := composer.Compose("business", data)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to compose image: %w", err)).Msg("image composition failed")
		return sendOpenTextFallback(sess, i, biz.Name, biz.BusinessType, regFee)
	}

	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Business Registered",
					Description: fmt.Sprintf("**%s** is now open for business!", biz.Name),
					Color:       0x4CAF76,
					Image:       &discordgo.MessageEmbedImage{URL: "attachment://business_open.png"},
					Footer:      &discordgo.MessageEmbedFooter{Text: "Aether City Chamber of Commerce"},
				},
			},
			Files: []*discordgo.File{
				{Name: "business_open.png", ContentType: "image/png", Reader: bytes.NewReader(imgBytes)},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func sendOpenTextFallback(sess *discordgo.Session, i *discordgo.InteractionCreate, name, bType string, fee decimal.Decimal) error {
	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Business Registered",
					Description: fmt.Sprintf("**%s** (%s) is now open for business!\nRegistration fee: `⊄%s`\n*(Image generation temporarily unavailable)*", name, bType, fee.String()),
					Color:       0x4CAF76,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}