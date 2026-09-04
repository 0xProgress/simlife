package economy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	sqlc "github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

// WorkLayoutData defines the data structure for the work image compositor.
// STRICT RULE: No float64 is used for financial or time values.
type WorkLayoutData struct {
	Username         string
	HoursWorked      string
	WageAccrued      decimal.Decimal
	EmploymentStatus string
	EmployerName     string
}

// HandleWork processes the /work command.
func (s *Service) HandleWork(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	log := logger.FromContext(ctx, "commands.economy")

	playerIDStr, ok := ctx.Value(commands.PlayerIDKey).(string)
	if !ok || playerIDStr == "" {
		log.Error().Msg("player_id missing from context in work handler")
		return fmt.Errorf("player_id missing from context")
	}

	log.Info().Str("player_id", playerIDStr).Msg("work command initiated")

	// 1. Check Redis Daily Work Cooldown
	workCooldownKey := fmt.Sprintf("work_daily:%s", playerIDStr)
	exists, err := s.Redis.Exists(ctx, workCooldownKey).Result()
	if err != nil {
		log.Error().Err(fmt.Errorf("redis cooldown check failed: %w", err)).Str("player_id", playerIDStr).Msg("work command failed")
		return fmt.Errorf("work cooldown check failed: %w", err)
	}
	if exists > 0 {
		ttl, _ := s.Redis.TTL(ctx, workCooldownKey).Result()
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("You are too tired to work again. Please rest for %s.", ttl.Round(time.Minute)),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// 2. Fetch Player
	player, err := s.Queries.GetPlayerByID(ctx, playerIDStr)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch player: %w", err)).Str("player_id", playerIDStr).Msg("work command failed")
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	// 3. Determine Employment Status and Process Labor
	hoursWorked := decimal.NewFromInt(8) // Standard 8-hour work day
	hoursNumeric := pgtype.Numeric{}
	if err := hoursNumeric.Scan(hoursWorked.String()); err != nil {
		return fmt.Errorf("failed to convert hours to numeric: %w", err)
	}

	var wageAccrued decimal.Decimal
	var employmentStatus string
	var employerName string

	employment, err := s.Queries.GetActiveEmployment(ctx, player.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		log.Error().Err(fmt.Errorf("failed to check employment status: %w", err)).Str("player_id", playerIDStr).Msg("work command failed")
		return fmt.Errorf("failed to check employment status: %w", err)
	}

	if err == nil && employment.ID != "" {
		// Employed: Accrue labor for settlement
		employmentStatus = "Employed"
		employerName = employment.BusinessName

		_, err = s.Queries.UpsertDailyLabor(ctx, sqlc.UpsertDailyLaborParams{
			PlayerID:    player.ID,
			BusinessID:  employment.BusinessID,
			EconomicDay: 1, // TODO: Replace with dynamic economic day from world state
			HoursWorked: hoursNumeric,
		})
		if err != nil {
			log.Error().Err(fmt.Errorf("failed to log daily labor: %w", err)).Str("player_id", playerIDStr).Msg("work command failed")
			return fmt.Errorf("failed to log daily labor: %w", err)
		}

		// Wage accrued is calculated but paid at settlement
		wageRate := decimal.NewFromInt(employment.WageRate)
		wageAccrued = wageRate.Mul(hoursWorked)
	} else {
		// Unemployed: Apply base city wage rate (paid immediately from Treasury)
		employmentStatus = "Unemployed (City Labor)"
		employerName = "Aether City Treasury"
		baseWageRate := decimal.NewFromInt(150) // Base city wage per day

		treasury, err := s.Queries.GetTreasuryAccount(ctx)
		if err != nil {
			log.Error().Err(fmt.Errorf("failed to fetch treasury account: %w", err)).Msg("work command failed")
			return fmt.Errorf("failed to fetch treasury account: %w", err)
		}

		wallet, err := s.Queries.GetAccountByType(ctx, sqlc.GetAccountByTypeParams{
			PlayerID:    player.ID,
			AccountType: sqlc.AccountTypeWALLET,
		})
		if err != nil {
			log.Error().Err(fmt.Errorf("failed to fetch wallet account: %w", err)).Str("player_id", playerIDStr).Msg("work command failed")
			return fmt.Errorf("failed to fetch wallet account: %w", err)
		}

		// Execute immediate transfer from Treasury to Wallet via the Ledger
		// Note: s.Ledger.Transfer must be implemented in internal/economy/ledger.go to handle 
		// the double-entry bookkeeping with SERIALIZABLE isolation.
		err = s.Ledger.Transfer(ctx, treasury.ID, wallet.ID, baseWageRate, "WAGE", player.ID, "Base city wage for day labor")
		if err != nil {
			log.Error().Err(fmt.Errorf("failed to transfer base wage: %w", err)).Str("player_id", playerIDStr).Msg("work command failed")
			return fmt.Errorf("failed to transfer base wage: %w", err)
		}

		wageAccrued = baseWageRate
	}

	// 4. Set Redis Cooldown using Config value
	cooldownDuration := time.Duration(s.Cfg.WorkCooldownMinutes) * time.Minute
	err = s.Redis.Set(ctx, workCooldownKey, true, cooldownDuration).Err()
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to set work cooldown: %w", err)).Str("player_id", playerIDStr).Msg("work cooldown set failed")
		// Non-fatal, we can still proceed, but log it
	}

	// 5. Compose Image and Respond
	data := WorkLayoutData{
		Username:         player.Username,
		HoursWorked:      hoursWorked.String(),
		WageAccrued:      wageAccrued,
		EmploymentStatus: employmentStatus,
		EmployerName:     employerName,
	}

	imgBytes, err := s.Composer.Compose("work", data)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to compose work image: %w", err)).Str("player_id", playerIDStr).Msg("image composition failed")
		return s.sendWorkTextFallback(ctx, sess, i, player.Username, wageAccrued, employmentStatus, employerName)
	}

	err = sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Labor Report",
					Description: fmt.Sprintf("Your work has been logged. Wages will be distributed at the next settlement."),
					Color:       0x4CAF76, // Green for income
					Image:       &discordgo.MessageEmbedImage{URL: "attachment://work.png"},
					Footer:      &discordgo.MessageEmbedFooter{Text: "Day 1 | Next settlement in 4h 20m"}, // TODO: Dynamic day
				},
			},
			Files: []*discordgo.File{
				{Name: "work.png", ContentType: "image/png", Reader: bytes.NewReader(imgBytes)},
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
		Str("wage_accrued", wageAccrued.String()).
		Str("status", employmentStatus).
		Msg("work command completed successfully")

	return nil
}

// sendWorkTextFallback provides a graceful degradation if the imaging compositor fails,
// ensuring the player still receives their labor report without a bot crash.
func (s *Service) sendWorkTextFallback(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate, username string, wageAccrued decimal.Decimal, status, employer string) error {
	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Labor Report",
					Description: fmt.Sprintf("**%s**\n\nStatus: %s\nEmployer: %s\nWage Accrued: `⊄%s`\n\n*(Image generation temporarily unavailable)*", username, status, employer, wageAccrued.String()),
					Color:       0x4CAF76,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}