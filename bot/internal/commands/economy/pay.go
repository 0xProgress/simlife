package economy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/shopspring/decimal"
)

// PaySenderLayoutData defines the data for the sender's confirmation image.
// STRICT RULE: No float64 is used for financial values.
type PaySenderLayoutData struct {
	SenderName    string
	RecipientName string
	Amount        decimal.Decimal // Net amount sent
	Fee           decimal.Decimal
	NewBalance    decimal.Decimal
}

// PayRecipientLayoutData defines the data for the recipient's notification image.
type PayRecipientLayoutData struct {
	RecipientName string
	SenderName    string
	Amount        decimal.Decimal // Net amount received
	NewBalance    decimal.Decimal
}

// HandlePay processes the /pay command.
func (s *Service) HandlePay(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	log := logger.FromContext(ctx, "commands.economy")

	senderIDStr, ok := ctx.Value(commands.PlayerIDKey).(string)
	if !ok || senderIDStr == "" {
		log.Error().Msg("player_id missing from context in pay handler")
		return fmt.Errorf("player_id missing from context")
	}

	// 1. Parse Discord Interaction Options
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
				Content: "Please mention a valid user to pay.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	recipientIDStr := targetUser.ID

	amount, err := decimal.NewFromString(amountStr)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide a valid positive amount to pay.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	if senderIDStr == recipientIDStr {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You cannot pay yourself.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// 2. Fetch Player Records
	sender, err := s.Queries.GetPlayerByID(ctx, senderIDStr)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch sender: %w", err)).Str("player_id", senderIDStr).Msg("pay command failed")
		return fmt.Errorf("failed to fetch sender: %w", err)
	}

	recipient, err := s.Queries.GetPlayerByDiscordID(ctx, recipientIDStr)
	if err != nil {
		log.Warn().Str("target_id", recipientIDStr).Msg("recipient not found in economy")
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "The specified recipient has not yet arrived in Aether City.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// 3. Fetch Wallet Accounts
	senderWallet, err := s.Queries.GetAccountByType(ctx, sqlc.GetAccountByTypeParams{
		PlayerID:    sender.ID,
		AccountType: sqlc.AccountTypeWALLET,
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch sender wallet: %w", err)).Str("player_id", senderIDStr).Msg("pay command failed")
		return fmt.Errorf("failed to fetch sender wallet: %w", err)
	}

	recipientWallet, err := s.Queries.GetAccountByType(ctx, sqlc.GetAccountByTypeParams{
		PlayerID:    recipient.ID,
		AccountType: sqlc.AccountTypeWALLET,
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch recipient wallet: %w", err)).Str("target_id", recipientIDStr).Msg("pay command failed")
		return fmt.Errorf("failed to fetch recipient wallet: %w", err)
	}

	// 4. Calculate 2% Transaction Fee (Money Sink)
	feeRate := decimal.NewFromFloat(0.02)
	fee := amount.Mul(feeRate).Truncate(0) // Round down to whole number
	netAmount := amount.Sub(fee)

	if netAmount.LessThanOrEqual(decimal.Zero) {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "The transfer amount is too small to cover the 2% transaction fee.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// 5. Execute the double-entry transfer via the Ledger (Sender -> Recipient)
	err = s.Ledger.Transfer(ctx, senderWallet.ID, recipientWallet.ID, netAmount, "P2P_TRANSFER", sender.ID, fmt.Sprintf("Payment to %s", recipient.Username))
	if err != nil {
		if errors.Is(err, ErrInsufficientFunds) {
			return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You do not have enough funds in your wallet to cover the amount and the 2% transaction fee.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
		log.Error().Err(fmt.Errorf("ledger transfer failed: %w", err)).Str("player_id", senderIDStr).Msg("pay command failed")
		return fmt.Errorf("ledger transfer failed: %w", err)
	}

	// 6. Transfer fee to Treasury if > 0
	if fee.GreaterThan(decimal.Zero) {
		treasury, err := s.Queries.GetTreasuryAccount(ctx)
		if err == nil && treasury.ID != "" {
			// We log errors here but do not fail the primary interaction, 
			// as the main transfer has already succeeded atomically.
			if err := s.Ledger.Transfer(ctx, senderWallet.ID, treasury.ID, fee, "FEE", sender.ID, "P2P transfer fee"); err != nil {
				log.Error().Err(fmt.Errorf("failed to transfer fee to treasury: %w", err)).Str("player_id", senderIDStr).Msg("fee transfer failed")
			}
		}
	}

	newSenderBal, _ := s.getAccountMetrics(ctx, senderWallet.ID)
	newRecipientBal, _ := s.getAccountMetrics(ctx, recipientWallet.ID)

	// 7. Respond to Sender (Public)
	senderData := PaySenderLayoutData{
		SenderName:    sender.Username,
		RecipientName: recipient.Username,
		Amount:        netAmount,
		Fee:           fee,
		NewBalance:    newSenderBal,
	}
	
	senderImg, err := s.Composer.Compose("pay_sender", senderData)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to compose sender image: %w", err)).Msg("image composition failed")
		// Fallback to text to ensure the user gets a response
		_ = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Transfer of **⊄%s** sent to **%s** (Fee: ⊄%s). Your new balance: **⊄%s**.", netAmount.String(), recipient.Username, fee.String(), newSenderBal.String()),
			},
		})
		return nil
	}

	err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("**%s** paid **%s**.", sender.Username, recipient.Username),
			Embeds: []*discordgo.MessageEmbed{
				{
					Title: "Transfer Confirmation",
					Color: 0x4CAF76, // Green
					Image: &discordgo.MessageEmbedImage{URL: "attachment://pay_sender.png"},
					Footer: &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Net: ⊄%s | Fee: ⊄%s", netAmount.String(), fee.String())},
				},
			},
			Files: []*discordgo.File{
				{Name: "pay_sender.png", ContentType: "image/png", Reader: bytes.NewReader(senderImg)},
			},
		},
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to respond to interaction: %w", err)).Msg("discord response failed")
	}

	// 8. Notify Recipient (Public follow-up message in channel)
	recipientData := PayRecipientLayoutData{
		RecipientName: recipient.Username,
		SenderName:    sender.Username,
		Amount:        netAmount,
		NewBalance:    newRecipientBal,
	}
	
	recipientImg, err := s.Composer.Compose("pay_recipient", recipientData)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to compose recipient image: %w", err)).Msg("image composition failed")
		_, _ = sess.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("💰 **Incoming Transfer:** %s received **⊄%s** from %s.", recipient.Username, netAmount.String(), sender.Username),
		})
		return nil
	}

	_, err = sess.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf("💰 **Incoming Transfer:** %s received **⊄%s** from %s.", recipient.Username, netAmount.String(), sender.Username),
		Embeds: []*discordgo.MessageEmbed{
			{
				Title: "Incoming Transfer",
				Color: 0x4CAF76, // Green
				Image: &discordgo.MessageEmbedImage{URL: "attachment://pay_recipient.png"},
			},
		},
		Files: []*discordgo.File{
			{Name: "pay_recipient.png", ContentType: "image/png", Reader: bytes.NewReader(recipientImg)},
		},
	})
	
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to send followup message: %w", err)).Msg("discord followup failed")
	}

	log.Info().
		Str("sender_id", senderIDStr).
		Str("recipient_id", recipientIDStr).
		Str("amount", amount.String()).
		Str("fee", fee.String()).
		Msg("pay command completed successfully")

	return nil
}