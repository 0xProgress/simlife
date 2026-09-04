package economy

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
)

// WorkLayoutData defines the data structure for the work image compositor.
type WorkLayoutData struct {
	Username         string
	HoursWorked      float64
	WageAccrued      float64
	EmploymentStatus string
}

// HandleWork processes the /work command.
func (s *Service) HandleWork(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate) error {
	playerIDStr := ctx.Value(commands.PlayerIDKey).(string)

	// 1. Check Redis Cooldown
	cooldownKey := fmt.Sprintf("work_cooldown:%s", playerIDStr)
	exists, err := s.Redis.Exists(ctx, cooldownKey).Result()
	if err == nil && exists > 0 {
		ttl, _ := s.Redis.TTL(ctx, cooldownKey).Result()
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("You are too tired to work again. Please rest for %s.", ttl.Round(time.Second)),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	player, err := s.Queries.GetPlayerByDiscordID(ctx, playerIDStr)
	if err != nil {
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	// 2. Determine Employment and Wage
	isEmployed := false
	wageRate := 15.0 // Base city wage rate
	hoursWorked := 1.0
	employmentStatus := "Unemployed (Base City Wage)"

	if isEmployed {
		employmentStatus = "Employed"
	}

	wageAccrued := wageRate * hoursWorked

	// 3. Credit Labor to Daily Ledger
	currentDay := int32(142)
	dummyCityBusinessID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	// Convert float64 to pgtype.Numeric for sqlc
	var pgHours pgtype.Numeric
	if err := pgHours.Scan(hoursWorked); err != nil {
		return fmt.Errorf("failed to convert hours to numeric: %w", err)
	}

	_, err = s.Queries.UpsertDailyLabor(ctx, db.UpsertDailyLaborParams{
		PlayerID:    player.ID,
		BusinessID:  dummyCityBusinessID,
		EconomicDay: currentDay,
		HoursWorked: pgHours,
	})
	if err != nil {
		return fmt.Errorf("failed to log daily labor: %w", err)
	}

	// 4. Set Redis Cooldown using Config value
	cooldownDuration := time.Duration(s.Cfg.WorkCooldownMinutes) * time.Minute
	s.Redis.Set(ctx, cooldownKey, true, cooldownDuration)

	// 5. Compose Image and Respond
	data := WorkLayoutData{
		Username:         player.Username,
		HoursWorked:      hoursWorked,
		WageAccrued:      wageAccrued,
		EmploymentStatus: employmentStatus,
	}

	imgBytes, err := s.Composer.Compose("work", data)
	if err != nil {
		return fmt.Errorf("failed to compose work image: %w", err)
	}

	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Labor Report",
					Description: "Your work has been logged for the next settlement.",
					Color:       0xFFD700,
					Image:       &discordgo.MessageEmbedImage{URL: "attachment://work.png"},
					Footer:      &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Day %d | Next settlement in 4h 20m", currentDay)},
				},
			},
			Files: []*discordgo.File{
				{Name: "work.png", ContentType: "image/png", Reader: bytes.NewReader(imgBytes)},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}
