package business

import (
	"context"
	"fmt"

	sqlc "github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
)

// HandleFire processes the /business fire command.
func HandleFire(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate, queries *sqlc.Queries) error {
	log := logger.FromContext(ctx, "commands.business")

	ownerIDStr, ok := ctx.Value(commands.PlayerIDKey).(string)
	if !ok || ownerIDStr == "" {
		log.Error().Msg("player_id missing from context")
		return fmt.Errorf("player_id missing from context")
	}

	var targetUser *discordgo.User
	var businessID string

	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == "business" {
			businessID = opt.StringValue()
		}
		if opt.Name == "target" && opt.Type == discordgo.ApplicationCommandOptionUser {
			targetUser = opt.UserValue()
		}
	}

	if targetUser == nil || businessID == "" {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide a valid business ID and target user.",
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

	employment, err := queries.GetActiveEmployment(ctx, targetPlayer.ID)
	if err != nil {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "This player is not currently employed by you.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	if employment.BusinessID != businessID {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "This player is employed by a different business.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	err = queries.TerminateEmployment(ctx, employment.ID)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to terminate employment: %w", err)).Str("employment_id", employment.ID).Msg("fire failed")
		return fmt.Errorf("failed to terminate employment: %w", err)
	}

	log.Info().
		Str("business_id", businessID).
		Str("employee_id", targetPlayer.ID).
		Msg("player fired successfully")

	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Successfully terminated **%s**'s employment.", targetPlayer.Username),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}