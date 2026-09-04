package economy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"

	sqlc "github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/shopspring/decimal"
)

// ErrInsufficientFunds is returned when a transfer attempts to debit more than the source account holds.
// This should ideally be defined in internal/economy/ledger.go, but is placed here for immediate compilation safety.
var ErrInsufficientFunds = errors.New("insufficient funds")

// BankLayoutData defines the data structure for the bank confirmation image compositor.
// STRICT RULE: No float64 is used for financial values.
type BankLayoutData struct {
	Username  string
	Action    string // "DEPOSIT" or "WITHDRAW"
	Amount    decimal.Decimal
	NewWallet decimal.Decimal
	NewBank   decimal.Decimal
}

// HandleBank processes the /bank command, routing to the appropriate action.
func (s *Service) HandleBank(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	log := logger.FromContext(ctx, "commands.economy")

	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please specify an action (deposit, withdraw, transfer, history).",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	action := options[0].StringValue()

	switch action {
	case "deposit":
		return s.handleDeposit(ctx, sess, i)
	case "withdraw":
		return s.handleWithdraw(ctx, sess, i)
	case "transfer":
		return s.handleTransfer(ctx, sess, i)
	case "history":
		return s.handleHistory(ctx, sess, i)
	default:
		log.Warn().Str("action", action).Msg("unknown bank action")
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid banking action.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
}

func (s *Service) handleDeposit(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.handleBankTransfer(ctx, sess, i, "DEPOSIT")
}

func (s *Service) handleWithdraw(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.handleBankTransfer(ctx, sess, i, "WITHDRAW")
}

func (s *Service) handleBankTransfer(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate, action string) error {
	log := logger.FromContext(ctx, "commands.economy")

	playerIDStr, ok := ctx.Value(commands.PlayerIDKey).(string)
	if !ok || playerIDStr == "" {
		log.Error().Msg("player_id missing from context in bank handler")
		return fmt.Errorf("player_id missing from context")
	}

	var amountStr string
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == "amount" {
			if opt.Type == discordgo.ApplicationCommandOptionInteger {
				amountStr = strconv.FormatInt(opt.IntValue(), 10)
			} else if opt.Type == discordgo.ApplicationCommandOptionString {
				amountStr = opt.StringValue()
			}
			break
		}
	}

	amount, err := decimal.NewFromString(amountStr)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide a valid positive amount to transfer.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	player, err := s.Queries.GetPlayerByID(ctx, playerIDStr)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch player: %w", err)).Str("player_id", playerIDStr).Msg("bank transfer failed")
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	accounts, err := s.Queries.GetAccountsByPlayer(ctx, player.ID)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch accounts: %w", err)).Str("player_id", playerIDStr).Msg("bank transfer failed")
		return fmt.Errorf("failed to fetch accounts: %w", err)
	}

	var walletID, bankID string
	for _, acc := range accounts {
		switch acc.AccountType {
		case sqlc.AccountTypeWALLET:
			walletID = acc.ID
		case sqlc.AccountTypeBANK:
			bankID = acc.ID
		}
	}

	if walletID == "" || bankID == "" {
		log.Error().Str("player_id", playerIDStr).Msg("missing wallet or bank account")
		return fmt.Errorf("missing required accounts for player")
	}

	var sourceID, destID string
	var txType string
	if action == "DEPOSIT" {
		sourceID, destID = walletID, bankID
		txType = "DEPOSIT"
	} else {
		sourceID, destID = bankID, walletID
		txType = "WITHDRAW"
	}

	// Execute the double-entry transfer via the Ledger
	err = s.Ledger.Transfer(ctx, sourceID, destID, amount, txType, player.ID, fmt.Sprintf("Bank %s", action))
	if err != nil {
		if errors.Is(err, ErrInsufficientFunds) {
			return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Insufficient funds in the source account.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
		log.Error().Err(fmt.Errorf("ledger transfer failed: %w", err)).Str("player_id", playerIDStr).Msg("bank transfer failed")
		return fmt.Errorf("ledger transfer failed: %w", err)
	}

	// Fetch updated balances
	walletBal, _ := s.getAccountMetrics(ctx, walletID)
	bankBal, _ := s.getAccountMetrics(ctx, bankID)

	data := BankLayoutData{
		Username:  player.Username,
		Action:    action,
		Amount:    amount,
		NewWallet: walletBal,
		NewBank:   bankBal,
	}

	imgBytes, err := s.Composer.Compose("bank", data)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to compose bank image: %w", err)).Str("player_id", playerIDStr).Msg("image composition failed")
		return s.sendBankTextFallback(ctx, sess, i, player.Username, action, amount, walletBal, bankBal)
	}

	err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Bank Transfer Confirmation",
					Description: fmt.Sprintf("Successfully processed **%s** of **⊄%s**.", action, amount.String()),
					Color:       0x4CAF76, // Green for successful transaction
					Image:       &discordgo.MessageEmbedImage{URL: "attachment://bank.png"},
					Footer:      &discordgo.MessageEmbedFooter{Text: "Simlife Federal Reserve"},
				},
			},
			Files: []*discordgo.File{
				{Name: "bank.png", ContentType: "image/png", Reader: bytes.NewReader(imgBytes)},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		log.Error().Err(fmt.Errorf("failed to respond to interaction: %w", err)).Str("player_id", playerIDStr).Msg("discord response failed")
		return fmt.Errorf("failed to respond to interaction: %w", err)
	}

	log.Info().
		Str("player_id", playerIDStr).
		Str("action", action).
		Str("amount", amount.String()).
		Msg("bank transfer completed successfully")

	return nil
}

func (s *Service) handleTransfer(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	log := logger.FromContext(ctx, "commands.economy")
	playerIDStr, _ := ctx.Value(commands.PlayerIDKey).(string)

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

	if targetUser == nil || targetUser.ID == playerIDStr {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please specify a valid target user other than yourself.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	amount, err := decimal.NewFromString(amountStr)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide a valid positive amount to transfer.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	targetPlayer, err := s.Queries.GetPlayerByDiscordID(ctx, targetUser.ID)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch target player: %w", err)).Str("target_id", targetUser.ID).Msg("transfer failed")
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Target user is not registered in Aether City.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	senderAccounts, err := s.Queries.GetAccountsByPlayer(ctx, playerIDStr)
	if err != nil {
		return fmt.Errorf("failed to fetch sender accounts: %w", err)
	}
	var senderWalletID string
	for _, acc := range senderAccounts {
		if acc.AccountType == sqlc.AccountTypeWALLET {
			senderWalletID = acc.ID
			break
		}
	}

	receiverAccounts, err := s.Queries.GetAccountsByPlayer(ctx, targetPlayer.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch receiver accounts: %w", err)
	}
	var receiverWalletID string
	for _, acc := range receiverAccounts {
		if acc.AccountType == sqlc.AccountTypeWALLET {
			receiverWalletID = acc.ID
			break
		}
	}

	// Apply 2% transaction fee (money sink)
	feeRate := decimal.NewFromFloat(0.02)
	fee := amount.Mul(feeRate).Truncate(0) // Round down to whole number
	netAmount := amount.Sub(fee)

	err = s.Ledger.Transfer(ctx, senderWalletID, receiverWalletID, netAmount, "P2P_TRANSFER", playerIDStr, fmt.Sprintf("Transfer to %s", targetPlayer.Username))
	if err != nil {
		if errors.Is(err, ErrInsufficientFunds) {
			return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Insufficient funds in your wallet.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
		return fmt.Errorf("ledger transfer failed: %w", err)
	}

	// Transfer fee to Treasury if > 0
	if fee.GreaterThan(decimal.Zero) {
		treasury, err := s.Queries.GetTreasuryAccount(ctx)
		if err == nil && treasury.ID != "" {
			_ = s.Ledger.Transfer(ctx, senderWalletID, treasury.ID, fee, "FEE", playerIDStr, "P2P transfer fee")
		}
	}

	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Successfully transferred **⊄%s** to **%s** (Fee: ⊄%s).", netAmount.String(), targetPlayer.Username, fee.String()),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (s *Service) handleHistory(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	log := logger.FromContext(ctx, "commands.economy")
	playerIDStr, _ := ctx.Value(commands.PlayerIDKey).(string)

	player, err := s.Queries.GetPlayerByID(ctx, playerIDStr)
	if err != nil {
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	accounts, err := s.Queries.GetAccountsByPlayer(ctx, player.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch accounts: %w", err)
	}

	var walletID string
	for _, acc := range accounts {
		if acc.AccountType == sqlc.AccountTypeWALLET {
			walletID = acc.ID
			break
		}
	}

	// Fetch last 5 transactions
	txs, err := s.Queries.GetTransactionHistory(ctx, sqlc.GetTransactionHistoryParams{
		AccountID: walletID,
		Limit:     5,
		Offset:    0,
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch history: %w", err)).Str("player_id", playerIDStr).Msg("history fetch failed")
		return fmt.Errorf("failed to fetch history: %w", err)
	}

	if len(txs) == 0 {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You have no recent transactions.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	var historyText string
	for _, tx := range txs {
		sign := "-"
		if string(tx.EntryType) == "CREDIT" {
			sign = "+"
		}
		amt := numericToDecimal(tx.Amount)
		historyText += fmt.Sprintf("%s **⊄%s** | %s | %s\n", sign, amt.String(), tx.TransactionType, tx.Description)
	}

	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Transaction History",
					Description: fmt.Sprintf("Recent activity for **%s**:\n\n%s", player.Username, historyText),
					Color:       0x4A90D9, // Blue for info
					Footer:      &discordgo.MessageEmbedFooter{Text: "Simlife Federal Reserve"},
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

// sendBankTextFallback provides a graceful degradation if the imaging compositor fails,
// ensuring the player still receives their transfer confirmation without a bot crash.
func (s *Service) sendBankTextFallback(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate, username, action string, amount, newWallet, newBank decimal.Decimal) error {
	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Bank Transfer Confirmation",
					Description: fmt.Sprintf("Successfully processed **%s** of **⊄%s**.\n\n*New Wallet:* `⊄%s`\n*New Bank:* `⊄%s`\n*(Image generation temporarily unavailable)*", action, amount.String(), newWallet.String(), newBank.String()),
					Color:       0x4CAF76,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}