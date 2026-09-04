package economy

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	db "github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/config"
	"github.com/0xProgress/simlife/bot/internal/economy"
	"github.com/0xProgress/simlife/bot/internal/imaging"
)

// Service holds the dependencies for all economy command handlers.
type Service struct {
	Queries  *db.Queries
	Ledger   *economy.Ledger
	Cfg      *config.Config
	Log      zerolog.Logger
	Composer *imaging.Composer
	Redis    *redis.Client
}

// NewService initializes the economy command service.
func NewService(q *db.Queries, l *economy.Ledger, cfg *config.Config, log zerolog.Logger, c *imaging.Composer, r *redis.Client) *Service {
	return &Service{
		Queries:  q,
		Ledger:   l,
		Cfg:      cfg,
		Log:      log.With().Str("package", "commands.economy").Logger(),
		Composer: c,
		Redis:    r,
	}
}

// BalanceLayoutData defines the data structure for the balance image compositor.
type BalanceLayoutData struct {
	Username    string
	Wallet      float64
	Bank        float64
	Escrow      float64
	NetWorth    float64
	Change24h   float64
	RankPercent int
}

// HandleBalance processes the /balance command.
func (s *Service) HandleBalance(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	playerIDStr := ctx.Value(commands.PlayerIDKey).(string)
	
	player, err := s.Queries.GetPlayerByDiscordID(ctx, playerIDStr)
	if err != nil {
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	// sqlc prefers pgtype.UUID for parameters mapped from UUID columns
	pgPlayerID := pgtype.UUID{Bytes: player.ID, Valid: true}
	accounts, err := s.Queries.GetAccountsByPlayer(ctx, pgPlayerID)
	if err != nil {
		return fmt.Errorf("failed to fetch accounts: %w", err)
	}

	var walletID, bankID, escrowID uuid.UUID
	for _, acc := range accounts {
		switch acc.Type {
		case db.AccountTypeWALLET:
			walletID = acc.ID
		case db.AccountTypeBANK:
			bankID = acc.ID
		case db.AccountTypeESCROW:
			escrowID = acc.ID
		}
	}

	walletBal, _ := s.getBalance(ctx, walletID)
	bankBal, _ := s.getBalance(ctx, bankID)
	escrowBal, _ := s.getBalance(ctx, escrowID)
	netWorth := walletBal + bankBal + escrowBal

	data := BalanceLayoutData{
		Username:    player.Username,
		Wallet:      walletBal,
		Bank:        bankBal,
		Escrow:      escrowBal,
		NetWorth:    netWorth,
		Change24h:   0.0, // TODO: Calculate from transaction history
		RankPercent: 50,  // TODO: Calculate from global wealth percentiles
	}

	imgBytes, err := s.Composer.Compose("balance", data)
	if err != nil {
		return fmt.Errorf("failed to compose balance image: %w", err)
	}

	embedColor := 0xFFD700 // Gold for economy commands
	if data.Change24h < 0 {
		embedColor = 0xFF4500 // Red-tinted for loss
	} else if data.Change24h == 0 {
		embedColor = 0x808080 // Muted for flat
	}

	err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Financial Ledger",
					Description: "Your current economic standing.",
					Color:       embedColor,
					Image:       &discordgo.MessageEmbedImage{URL: "attachment://balance.png"},
					Footer:      &discordgo.MessageEmbedFooter{Text: "Day 142 | Next settlement in 4h 20m"},
				},
			},
			Files: []*discordgo.File{
				{Name: "balance.png", ContentType: "image/png", Reader: bytes.NewReader(imgBytes)},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	
	return err
}

// getBalance safely extracts a float64 balance from the sqlc generated return type.
func (s *Service) getBalance(ctx context.Context, accountID uuid.UUID) (float64, error) {
	if accountID == uuid.Nil {
		return 0, nil
	}
	bal, err := s.Queries.GetAccountBalance(ctx, accountID)
	if err != nil {
		return 0, err
	}
	return toFloat64(bal), nil
}

func toFloat64(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		return 0 
	default:
		return 0
	}
}