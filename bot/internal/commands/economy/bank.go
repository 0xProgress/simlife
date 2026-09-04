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

// BankLayoutData defines the data structure for the bank confirmation image compositor.
type BankLayoutData struct {
	Username  string
	Action    string // "DEPOSIT" or "WITHDRAW"
	Amount    float64
	NewWallet float64
	NewBank   float64
}

// HandleDeposit processes the /deposit command.
func (s *Service) HandleDeposit(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.handleBankTransfer(ctx, sess, i, "DEPOSIT")
}

// HandleWithdraw processes the /withdraw command.
func (s *Service) HandleWithdraw(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.handleBankTransfer(ctx, sess, i, "WITHDRAW")
}

func (s *Service) handleBankTransfer(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate, action string) error {
	playerIDStr := ctx.Value(commands.PlayerIDKey).(string)
	
	amount, ok := getFloatOption(i, "amount")
	if !ok || amount <= 0 {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide a valid positive amount to transfer.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	player, err := s.Queries.GetPlayerByDiscordID(ctx, playerIDStr)
	if err != nil {
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	pgPlayerID := pgtype.UUID{Bytes: player.ID, Valid: true}
	accounts, err := s.Queries.GetAccountsByPlayer(ctx, pgPlayerID)
	if err != nil {
		return fmt.Errorf("failed to fetch accounts: %w", err)
	}

	var walletID, bankID uuid.UUID
	for _, acc := range accounts {
		switch acc.Type {
		case db.AccountTypeWALLET:
			walletID = acc.ID
		case db.AccountTypeBANK:
			bankID = acc.ID
		}
	}

	var sourceID, destID uuid.UUID
	if action == "DEPOSIT" {
		sourceID, destID = walletID, bankID
	} else {
		sourceID, destID = bankID, walletID
	}

	_, err = s.Ledger.PostTransaction(ctx, economy.PostTransactionParams{
		SourceAccountID: sourceID,
		DestAccountID:   destID,
		Amount:          amount,
		Type:            db.TransactionTypePLAYERTRANSFER,
		Description:     fmt.Sprintf("Bank %s", action),
	})
	if err != nil {
		if err == economy.ErrInsufficientFunds {
			return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Insufficient funds in the source account.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
		return fmt.Errorf("ledger transaction failed: %w", err)
	}

	newWallet, _ := s.getBalance(ctx, walletID)
	newBank, _ := s.getBalance(ctx, bankID)

	data := BankLayoutData{
		Username:  player.Username,
		Action:    action,
		Amount:    amount,
		NewWallet: newWallet,
		NewBank:   newBank,
	}

	imgBytes, err := s.Composer.Compose("bank", data)
	if err != nil {
		return fmt.Errorf("failed to compose bank image: %w", err)
	}

	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Bank Transfer Confirmation",
					Description: fmt.Sprintf("Successfully processed %s of $%.2f.", action, amount),
					Color:       0xFFD700,
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
}