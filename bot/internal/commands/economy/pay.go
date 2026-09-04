package economy

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/economy"
)

// PaySenderLayoutData defines the data for the sender's confirmation image.
type PaySenderLayoutData struct {
	SenderName    string
	RecipientName string
	Amount        float64
	NewBalance    float64
}

// PayRecipientLayoutData defines the data for the recipient's notification image.
type PayRecipientLayoutData struct {
	RecipientName string
	SenderName    string
	Amount        float64
	NewBalance    float64
}

// HandlePay processes the /pay command.
func (s *Service) HandlePay(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	senderIDStr := ctx.Value(commands.PlayerIDKey).(string)

	amount, ok := getFloatOption(i, "amount")
	if !ok || amount <= 0 {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide a valid positive amount to pay.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	recipientIDStr, ok := getUserIDOption(i, "user")
	if !ok {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please mention a valid user to pay.",
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

	sender, err := s.Queries.GetPlayerByDiscordID(ctx, senderIDStr)
	if err != nil {
		return fmt.Errorf("failed to fetch sender: %w", err)
	}

	recipient, err := s.Queries.GetPlayerByDiscordID(ctx, recipientIDStr)
	if err != nil {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "The specified recipient does not exist in the economy.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	senderWallet, err := s.getAccountByType(ctx, sender.ID, db.AccountTypeWALLET)
	if err != nil {
		return fmt.Errorf("failed to fetch sender wallet: %w", err)
	}

	recipientWallet, err := s.getAccountByType(ctx, recipient.ID, db.AccountTypeWALLET)
	if err != nil {
		return fmt.Errorf("failed to fetch recipient wallet: %w", err)
	}

	_, err = s.Ledger.PostTransaction(ctx, economy.PostTransactionParams{
		SourceAccountID: senderWallet,
		DestAccountID:   recipientWallet,
		Amount:          amount,
		Type:            db.TransactionTypePLAYERTRANSFER,
		Description:     fmt.Sprintf("P2P Payment from %s", sender.Username),
	})
	if err != nil {
		if err == economy.ErrInsufficientFunds {
			return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You do not have enough funds in your wallet.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
		return fmt.Errorf("ledger transaction failed: %w", err)
	}

	newSenderBal, _ := s.getBalance(ctx, senderWallet)
	newRecipientBal, _ := s.getBalance(ctx, recipientWallet)

	// 1. Respond to Sender (Public)
	senderData := PaySenderLayoutData{
		SenderName:    sender.Username,
		RecipientName: recipient.Username,
		Amount:        amount,
		NewBalance:    newSenderBal,
	}
	senderImg, _ := s.Composer.Compose("pay_sender", senderData)

	err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Transfer of $%.2f sent to %s.", amount, recipient.Username),
			Embeds: []*discordgo.MessageEmbed{
				{
					Title: "Transfer Confirmation",
					Color: 0xFFD700,
					Image: &discordgo.MessageEmbedImage{URL: "attachment://pay_sender.png"},
				},
			},
			Files: []*discordgo.File{
				{Name: "pay_sender.png", ContentType: "image/png", Reader: bytes.NewReader(senderImg)},
			},
		},
	})
	if err != nil {
		return err
	}

	// 2. Notify Recipient (Public follow-up message in channel)
	recipientData := PayRecipientLayoutData{
		RecipientName: recipient.Username,
		SenderName:    sender.Username,
		Amount:        amount,
		NewBalance:    newRecipientBal,
	}
	recipientImg, _ := s.Composer.Compose("pay_recipient", recipientData)

	_, err = sess.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Incoming Transfer: %s received $%.2f from %s.", recipient.Username, amount, sender.Username),
		Embeds: []*discordgo.MessageEmbed{
			{
				Title: "Incoming Transfer",
				Color: 0x00FF00,
				Image: &discordgo.MessageEmbedImage{URL: "attachment://pay_recipient.png"},
			},
		},
		Files: []*discordgo.File{
			{Name: "pay_recipient.png", ContentType: "image/png", Reader: bytes.NewReader(recipientImg)},
		},
	})

	return err
}

// getAccountByType fetches the account ID for a specific player and account type.
func (s *Service) getAccountByType(ctx context.Context, playerID uuid.UUID, accType db.AccountType) (uuid.UUID, error) {
	pgPlayerID := pgtype.UUID{Bytes: playerID, Valid: true}
	accounts, err := s.Queries.GetAccountsByPlayer(ctx, pgPlayerID)
	if err != nil {
		return uuid.Nil, err
	}
	for _, acc := range accounts {
		if acc.Type == accType {
			return acc.ID, nil
		}
	}
	return uuid.Nil, fmt.Errorf("account of type %s not found for player", accType)
}
