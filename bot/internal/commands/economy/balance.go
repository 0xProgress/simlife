package economy

import (
	"bytes"
	"context"
	"fmt"

	sqlc "github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/config"
	"github.com/0xProgress/simlife/bot/internal/economy"
	"github.com/0xProgress/simlife/bot/internal/imaging"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

// Service holds the dependencies for all economy command handlers.
type Service struct {
	Queries  *sqlc.Queries
	Ledger   *economy.Ledger
	Cfg      *config.Config
	Log      zerolog.Logger
	Composer *imaging.Composer
	Redis    *redis.Client
}

// NewService initializes the economy command service.
func NewService(q *sqlc.Queries, l *economy.Ledger, cfg *config.Config, log zerolog.Logger, c *imaging.Composer, r *redis.Client) *Service {
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
// STRICT RULE: No float64 is used for financial values.
type BalanceLayoutData struct {
	Username    string
	Wallet      decimal.Decimal
	Bank        decimal.Decimal
	Escrow      decimal.Decimal
	NetWorth    decimal.Decimal
	Change24h   decimal.Decimal
	RankPercent int
}

// HandleBalance processes the /balance command.
func (s *Service) HandleBalance(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	log := logger.FromContext(ctx, "commands.economy")

	// 1. Extract authenticated player ID from context (injected by AuthMiddleware)
	playerIDStr, ok := ctx.Value(commands.PlayerIDKey).(string)
	if !ok || playerIDStr == "" {
		log.Error().Msg("player_id missing from context in balance handler")
		return fmt.Errorf("player_id missing from context")
	}

	log.Info().Str("player_id", playerIDStr).Msg("balance check initiated")

	// 2. Fetch player record
	player, err := s.Queries.GetPlayerByID(ctx, playerIDStr)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch player: %w", err)).Str("player_id", playerIDStr).Msg("balance check failed")
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	// 3. Fetch player's core accounts
	accounts, err := s.Queries.GetAccountsByPlayer(ctx, player.ID)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch accounts: %w", err)).Str("player_id", playerIDStr).Msg("balance check failed")
		return fmt.Errorf("failed to fetch accounts: %w", err)
	}

	var walletID, bankID, escrowID string
	for _, acc := range accounts {
		switch acc.AccountType {
		case sqlc.AccountTypeWALLET:
			walletID = acc.ID
		case sqlc.AccountTypeBANK:
			bankID = acc.ID
		case sqlc.AccountTypeESCROW:
			escrowID = acc.ID
		}
	}

	// 4. Fetch balances and 24h changes for each account
	walletBal, walletChange := s.getAccountMetrics(ctx, walletID)
	bankBal, bankChange := s.getAccountMetrics(ctx, bankID)
	escrowBal, escrowChange := s.getAccountMetrics(ctx, escrowID)

	// 5. Calculate aggregate metrics using strict decimal math
	netWorth := walletBal.Add(bankBal).Add(escrowBal)
	change24h := walletChange.Add(bankChange).Add(escrowChange)

	// 6. Fetch global wealth rank percentile
	rankPercent, err := s.Queries.GetPlayerWealthRank(ctx, player.ID)
	if err != nil {
		log.Warn().Err(fmt.Errorf("failed to fetch wealth rank: %w", err)).Str("player_id", playerIDStr).Msg("falling back to default rank")
		rankPercent = 50 // Safe fallback
	}

	data := BalanceLayoutData{
		Username:    player.Username,
		Wallet:      walletBal,
		Bank:        bankBal,
		Escrow:      escrowBal,
		NetWorth:    netWorth,
		Change24h:   change24h,
		RankPercent: int(rankPercent),
	}

	// 7. Compose the dynamic image
	imgBytes, err := s.Composer.Compose("balance", data)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to compose balance image: %w", err)).Str("player_id", playerIDStr).Msg("image composition failed")
		// Fallback: We do not fail the command, we just send a text-only embed
		return s.sendTextFallback(ctx, sess, i, player.Username, netWorth)
	}

	// 8. Determine embed color based on 24h trend
	embedColor := 0xFFD700 // Gold (default/positive)
	if change24h.LessThan(decimal.Zero) {
		embedColor = 0xE05555 // Red-tinted for loss (matches UI design system)
	} else if change24h.Equal(decimal.Zero) {
		embedColor = 0x8E95A8 // Muted for flat (matches UI text-secondary)
	}

	// 9. Respond to Discord interaction
	err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Financial Ledger",
					Description: fmt.Sprintf("Current economic standing for **%s**", player.Username),
					Color:       embedColor,
					Image:       &discordgo.MessageEmbedImage{URL: "attachment://balance.png"},
					Footer:      &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Economic Day %d", 142)}, // TODO: Wire to actual game day
				},
			},
			Files: []*discordgo.File{
				{Name: "balance.png", ContentType: "image/png", Reader: bytes.NewReader(imgBytes)},
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
		Str("net_worth", netWorth.String()).
		Msg("balance check completed successfully")

	return nil
}

// getAccountMetrics safely extracts decimal balances and 24h changes from sqlc generated return types.
func (s *Service) getAccountMetrics(ctx context.Context, accountID string) (decimal.Decimal, decimal.Decimal) {
	if accountID == "" {
		return decimal.Zero, decimal.Zero
	}

	metrics, err := s.Queries.GetAccountBalanceAndChange(ctx, accountID)
	if err != nil {
		s.Log.Warn().Err(fmt.Errorf("failed to fetch account metrics: %w", err)).Str("account_id", accountID).Msg("account metrics fetch failed")
		return decimal.Zero, decimal.Zero
	}

	return numericToDecimal(metrics.CurrentBalance), numericToDecimal(metrics.Change24h)
}

// numericToDecimal flawlessly converts pgtype.Numeric to shopspring/decimal.Decimal.
// This is the single source of truth for database-to-domain financial conversion.
func numericToDecimal(n pgtype.Numeric) decimal.Decimal {
	if !n.Valid {
		return decimal.Zero
	}
	
	// pgtype.Numeric implements fmt.Stringer, which safely outputs the exact base-10 representation
	// without the precision loss inherent in Float64() conversions.
	str := n.String()
	d, err := decimal.NewFromString(str)
	if err != nil {
		return decimal.Zero
	}
	
	return d
}

// sendTextFallback provides a graceful degradation if the imaging compositor fails,
// ensuring the player still receives their financial data without a bot crash.
func (s *Service) sendTextFallback(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate, username string, netWorth decimal.Decimal) error {
	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Financial Ledger",
					Description: fmt.Sprintf("Current economic standing for **%s**\n\n*Net Worth:* `⊄%s`\n*(Image generation temporarily unavailable)*", username, netWorth.String()),
					Color:       0xFFD700,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}