package admin

import (
	"context"
	"fmt"
	"strconv"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/economy"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/shopspring/decimal"
)

// HandleGrant processes the /admin grant command.
// This command creates new currency from the city treasury and deposits it into a player's wallet.
// It is used for economic stimulus, compensation, or developer testing.
func HandleGrant(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate, queries *sqlc.Queries, ledger *economy.Ledger) error {
	log := logger.FromContext(ctx, "commands.admin")

	adminIDStr, ok := ctx.Value(commands.PlayerIDKey).(string)
	if !ok || adminIDStr == "" {
		log.Error().Msg("player_id missing from context in admin grant handler")
		return fmt.Errorf("player_id missing from context")
	}

	// Verify admin privileges
	isAdmin, err := queries.IsPlayerAdmin(ctx, adminIDStr)
	if err != nil || !isAdmin {
		log.Warn().Str("admin_id", adminIDStr).Msg("unauthorized admin grant attempt")
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ You do not have permission to execute admin commands.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	var targetUser *discordgo.User
	var amountStr string

	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == "target" && opt.Type == discordgo.ApplicationCommandOptionUser {
			targetUser = opt.UserValue()
		}
		if opt.Name == "amount" {
			if opt.Type == discordgo.ApplicationCommandOptionInteger {
				amountStr = strconv.FormatInt(opt.IntValue(), 10)
			} else if opt.Type == discordgo.ApplicationCommandOptionString {
				amountStr = opt.StringValue()
			}
		}
	}

	if targetUser == nil {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please specify a valid target user.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	amount, err := decimal.NewFromString(amountStr)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide a valid positive amount to grant.",
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

	// Fetch target's wallet
	wallet, err := queries.GetAccountByType(ctx, sqlc.GetAccountByTypeParams{
		PlayerID:    targetPlayer.ID,
		AccountType: sqlc.AccountTypeWALLET,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch target wallet: %w", err)
	}

	// Fetch treasury
	treasury, err := queries.GetTreasuryAccount(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch treasury: %w", err)
	}

	// Execute grant via Ledger (Treasury -> Wallet)
	err = ledger.Transfer(ctx, treasury.ID, wallet.ID, amount, "ADMIN_GRANT", targetPlayer.ID, fmt.Sprintf("Admin grant by %s", adminIDStr))
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to execute grant: %w", err)).Str("target_id", targetPlayer.ID).Msg("admin grant failed")
		return fmt.Errorf("failed to execute grant: %w", err)
	}

	// Log to admin_audit_log
	err = queries.LogAdminAction(ctx, sqlc.LogAdminActionParams{
		AdminPlayerID:  adminIDStr,
		TargetPlayerID: targetPlayer.ID,
		ActionType:     "GRANT",
		Parameters:     fmt.Sprintf("Granted ⊄%s", amount.String()),
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to log admin action: %w", err)).Msg("admin audit log failed")
	}

	log.Info().
		Str("admin_id", adminIDStr).
		Str("target_id", targetPlayer.ID).
		Str("amount", amount.String()).
		Msg("admin grant executed successfully")

	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("✅ Successfully granted **⊄%s** to **%s**'s wallet.", amount.String(), targetPlayer.Username),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}