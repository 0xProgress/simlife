package admin

import (
	"context"
	"fmt"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/economy"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/shopspring/decimal"
)

// HandleReset processes the /admin reset command.
// This command completely resets a player's economic state, wiping all balances, 
// transactions, and employment records. It is a destructive operation reserved for 
// emergency exploit remediation or developer testing.
func HandleReset(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate, queries *sqlc.Queries, ledger *economy.Ledger) error {
	log := logger.FromContext(ctx, "commands.admin")

	adminIDStr, ok := ctx.Value(commands.PlayerIDKey).(string)
	if !ok || adminIDStr == "" {
		log.Error().Msg("player_id missing from context in admin reset handler")
		return fmt.Errorf("player_id missing from context")
	}

	// Verify admin privileges
	isAdmin, err := queries.IsPlayerAdmin(ctx, adminIDStr)
	if err != nil || !isAdmin {
		log.Warn().Str("admin_id", adminIDStr).Msg("unauthorized admin reset attempt")
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ You do not have permission to execute admin commands.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	var targetUser *discordgo.User
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == "target" && opt.Type == discordgo.ApplicationCommandOptionUser {
			targetUser = opt.UserValue()
			break
		}
	}

	if targetUser == nil {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please specify a valid target user to reset.",
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

	// Execute the reset via the Ledger (confiscates all funds to treasury, then wipes records)
	treasury, err := queries.GetTreasuryAccount(ctx)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch treasury: %w", err)).Msg("admin reset failed")
		return fmt.Errorf("failed to fetch treasury: %w", err)
	}

	accounts, err := queries.GetAccountsByPlayer(ctx, targetPlayer.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch target accounts: %w", err)
	}

	// Confiscate all funds from all accounts to treasury
	for _, acc := range accounts {
		bal, err := queries.GetAccountBalance(ctx, acc.ID)
		if err != nil {
			continue
		}
		balDecimal := numericToDecimal(bal)
		if balDecimal.GreaterThan(decimal.Zero) {
			err = ledger.Transfer(ctx, acc.ID, treasury.ID, balDecimal, "ADMIN_CONFISCATION", targetPlayer.ID, fmt.Sprintf("Admin reset confiscation by %s", adminIDStr))
			if err != nil {
				log.Error().Err(fmt.Errorf("failed to confiscate funds: %w", err)).Str("account_id", acc.ID).Msg("admin reset partial failure")
			}
		}
	}

	// Wipe employment, business ownership, and property records
	err = queries.ResetPlayerEconomicState(ctx, targetPlayer.ID)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to reset economic state: %w", err)).Str("target_id", targetPlayer.ID).Msg("admin reset failed")
		return fmt.Errorf("failed to reset economic state: %w", err)
	}

	// Log to admin_audit_log
	err = queries.LogAdminAction(ctx, sqlc.LogAdminActionParams{
		AdminPlayerID: adminIDStr,
		TargetPlayerID: targetPlayer.ID,
		ActionType: "RESET",
		Parameters: fmt.Sprintf("Full economic reset executed"),
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to log admin action: %w", err)).Msg("admin audit log failed")
	}

	log.Info().
		Str("admin_id", adminIDStr).
		Str("target_id", targetPlayer.ID).
		Msg("admin reset executed successfully")

	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("✅ Successfully reset **%s**'s economic state. All funds confiscated to treasury, all employment and property records cleared.", targetPlayer.Username),
			Flags:   discordgo.MessageFlagsEphemeral,
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