package business

import (
	"context"
	"fmt"

	sqlc "github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/shopspring/decimal"
)

// HandleHire processes the /business hire command.
func HandleHire(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate, queries *sqlc.Queries) error {
	log := logger.FromContext(ctx, "commands.business")

	ownerIDStr, ok := ctx.Value(commands.PlayerIDKey).(string)
	if !ok || ownerIDStr == "" {
		log.Error().Msg("player_id missing from context")
		return fmt.Errorf("player_id missing from context")
	}

	var targetUser *discordgo.User
	var wageStr, businessID string

	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == "business" {
			businessID = opt.StringValue()
		}
		if opt.Name == "target" && opt.Type == discordgo.ApplicationCommandOptionUser {
			targetUser = opt.UserValue()
		}
		if opt.Name == "wage" {
			wageStr = opt.StringValue()
		}
	}

	if targetUser == nil || wageStr == "" || businessID == "" {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide a valid business ID, target user, and wage rate.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	wage, err := decimal.NewFromString(wageStr)
	if err != nil || wage.LessThanOrEqual(decimal.Zero) {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide a valid positive wage rate.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	biz, err := queries.GetBusinessByID(ctx, businessID)
	if err != nil {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Business not found.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	if biz.OwnerID != ownerIDStr {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You do not own this business.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	targetPlayer, err := queries.GetPlayerByDiscordID(ctx, targetUser.ID)
	if err != nil {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Target user is not registered in Aether City.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	_, err = queries.GetActiveEmployment(ctx, targetPlayer.ID)
	if err == nil {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "This player is already employed elsewhere.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	_, err = queries.CreateEmployment(ctx, sqlc.CreateEmploymentParams{
		BusinessID:    biz.ID,
		EmployeeID:    targetPlayer.ID,
		WageRate:      wage,
		MinDailyHours: decimal.NewFromInt(8),
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to create employment: %w", err)).Str("business_id", businessID).Msg("hire failed")
		return fmt.Errorf("failed to create employment record: %w", err)
	}

	log.Info().
		Str("business_id", businessID).
		Str("employee_id", targetPlayer.ID).
		Str("wage", wage.String()).
		Msg("player hired successfully")

	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Successfully hired **%s** at **⊄%s**/day.", targetPlayer.Username, wage.String()),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}