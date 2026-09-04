package admin

import (
	"bytes"
	"context"
	"fmt"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/imaging"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

// AdminInspectData defines the data for the admin inspection summary image.
type AdminInspectData struct {
	Username         string
	WalletBalance    decimal.Decimal
	BankBalance      decimal.Decimal
	EscrowBalance    decimal.Decimal
	NetWorth         decimal.Decimal
	TransactionCount int
	EmploymentStatus string
	BusinessCount    int
	PropertyCount    int
}

// HandleInspect processes the /admin inspect command.
// This command provides a comprehensive view of a player's economic state, including 
// all account balances, recent transaction history, employment status, business ownership, 
// and property holdings. It returns a composited summary image and a detailed text embed.
func HandleInspect(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate, queries *sqlc.Queries, composer *imaging.Composer) error {
	log := logger.FromContext(ctx, "commands.admin")

	adminIDStr, ok := ctx.Value(commands.PlayerIDKey).(string)
	if !ok || adminIDStr == "" {
		log.Error().Msg("player_id missing from context in admin inspect handler")
		return fmt.Errorf("player_id missing from context")
	}

	// Verify admin privileges
	isAdmin, err := queries.IsPlayerAdmin(ctx, adminIDStr)
	if err != nil || !isAdmin {
		log.Warn().Str("admin_id", adminIDStr).Msg("unauthorized admin inspect attempt")
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
				Content: "Please specify a valid target user to inspect.",
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

	// Fetch all account balances
	accounts, err := queries.GetAccountsByPlayer(ctx, targetPlayer.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch target accounts: %w", err)
	}

	var walletBal, bankBal, escrowBal decimal.Decimal
	for _, acc := range accounts {
		bal, err := queries.GetAccountBalance(ctx, acc.ID)
		if err != nil {
			continue
		}
		balDecimal := numericToDecimal(bal)
		switch acc.AccountType {
		case sqlc.AccountTypeWALLET:
			walletBal = balDecimal
		case sqlc.AccountTypeBANK:
			bankBal = balDecimal
		case sqlc.AccountTypeESCROW:
			escrowBal = balDecimal
		}
	}

	netWorth := walletBal.Add(bankBal).Add(escrowBal)

	// Fetch transaction count
	txCount, err := queries.GetPlayerTransactionCount(ctx, targetPlayer.ID)
	if err != nil {
		log.Warn().Err(fmt.Errorf("failed to fetch transaction count: %w", err)).Msg("inspect partial failure")
		txCount = 0
	}

	// Fetch employment status
	employment, err := queries.GetActiveEmployment(ctx, targetPlayer.ID)
	employmentStatus := "Unemployed"
	if err == nil && employment.ID != "" {
		employmentStatus = fmt.Sprintf("Employed at %s", employment.BusinessName)
	}

	// Fetch business count
	businessCount, err := queries.GetPlayerBusinessCount(ctx, targetPlayer.ID)
	if err != nil {
		log.Warn().Err(fmt.Errorf("failed to fetch business count: %w", err)).Msg("inspect partial failure")
		businessCount = 0
	}

	// Fetch property count
	propertyCount, err := queries.GetPlayerPropertyCount(ctx, targetPlayer.ID)
	if err != nil {
		log.Warn().Err(fmt.Errorf("failed to fetch property count: %w", err)).Msg("inspect partial failure")
		propertyCount = 0
	}

	// Fetch recent transactions (last 10)
	walletID := ""
	for _, acc := range accounts {
		if acc.AccountType == sqlc.AccountTypeWALLET {
			walletID = acc.ID
			break
		}
	}

	var recentTxs []sqlc.Transaction
	if walletID != "" {
		recentTxs, err = queries.GetTransactionHistory(ctx, sqlc.GetTransactionHistoryParams{
			AccountID: walletID,
			Limit:     10,
			Offset:    0,
		})
		if err != nil {
			log.Warn().Err(fmt.Errorf("failed to fetch transaction history: %w", err)).Msg("inspect partial failure")
		}
	}

	// Build inspection data
	data := AdminInspectData{
		Username:         targetPlayer.Username,
		WalletBalance:    walletBal,
		BankBalance:      bankBal,
		EscrowBalance:    escrowBal,
		NetWorth:         netWorth,
		TransactionCount: int(txCount),
		EmploymentStatus: employmentStatus,
		BusinessCount:    int(businessCount),
		PropertyCount:    int(propertyCount),
	}

	// Compose summary image
	imgBytes, err := composer.Compose("admin_inspect", data)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to compose inspect image: %w", err)).Msg("image composition failed")
		return sendInspectTextFallback(sess, i, data, recentTxs)
	}

	// Build transaction history text
	var txHistoryText string
	for _, tx := range recentTxs {
		sign := "-"
		if string(tx.EntryType) == "CREDIT" {
			sign = "+"
		}
		amt := numericToDecimal(tx.Amount)
		txHistoryText += fmt.Sprintf("%s **⊄%s** | %s | %s\n", sign, amt.String(), tx.TransactionType, tx.Description)
	}

	// Log admin action
	err = queries.LogAdminAction(ctx, sqlc.LogAdminActionParams{
		AdminPlayerID:  adminIDStr,
		TargetPlayerID: targetPlayer.ID,
		ActionType:     "INSPECT",
		Parameters:     "Full economic inspection",
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to log admin action: %w", err)).Msg("admin audit log failed")
	}

	log.Info().
		Str("admin_id", adminIDStr).
		Str("target_id", targetPlayer.ID).
		Msg("admin inspect executed successfully")

	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       fmt.Sprintf("🔍 Admin Inspection: %s", targetPlayer.Username),
					Description: fmt.Sprintf("**Net Worth:** ⊄%s\n**Transactions:** %d\n**Employment:** %s\n**Businesses:** %d\n**Properties:** %d\n\n**Recent Transactions:**\n%s", netWorth.String(), txCount, employmentStatus, businessCount, propertyCount, txHistoryText),
					Color:       0x8a6fd8, // Purple for admin
					Image:       &discordgo.MessageEmbedImage{URL: "attachment://admin_inspect.png"},
					Footer:      &discordgo.MessageEmbedFooter{Text: "Aether City Administrative Oversight"},
				},
			},
			Files: []*discordgo.File{
				{Name: "admin_inspect.png", ContentType: "image/png", Reader: bytes.NewReader(imgBytes)},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func sendInspectTextFallback(sess *discordgo.Session, i *discordgo.InteractionCreate, data AdminInspectData, txs []sqlc.Transaction) error {
	var txHistoryText string
	for _, tx := range txs {
		sign := "-"
		if string(tx.EntryType) == "CREDIT" {
			sign = "+"
		}
		amt := numericToDecimal(tx.Amount)
		txHistoryText += fmt.Sprintf("%s ⊄%s | %s | %s\n", sign, amt.String(), tx.TransactionType, tx.Description)
	}

	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       fmt.Sprintf("🔍 Admin Inspection: %s", data.Username),
					Description: fmt.Sprintf("**Net Worth:** ⊄%s\n**Transactions:** %d\n**Employment:** %s\n**Businesses:** %d\n**Properties:** %d\n\n**Recent Transactions:**\n%s\n*(Image generation temporarily unavailable)*", data.NetWorth.String(), data.TransactionCount, data.EmploymentStatus, data.BusinessCount, data.PropertyCount, txHistoryText),
					Color:       0x8a6fd8,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}