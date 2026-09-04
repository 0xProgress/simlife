package business

import (
	"bytes"
	"context"
	"fmt"

	sqlc "github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/imaging"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/shopspring/decimal"
)

// StatusLayoutData defines the data for the business status image.
type StatusLayoutData struct {
	BusinessName     string
	BusinessType     string
	OwnerName        string
	EmployeeCount    int
	ProjectedRevenue decimal.Decimal
	Status           string
}

// HandleStatus processes the /business status command.
func HandleStatus(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate, queries *sqlc.Queries, composer *imaging.Composer) error {
	log := logger.FromContext(ctx, "commands.business")

	playerIDStr, ok := ctx.Value(commands.PlayerIDKey).(string)
	if !ok || playerIDStr == "" {
		log.Error().Msg("player_id missing from context")
		return fmt.Errorf("player_id missing from context")
	}

	var businessID string
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == "business" {
			businessID = opt.StringValue()
		}
	}

	if businessID == "" {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide a valid business ID.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	biz, err := queries.GetBusinessByID(ctx, businessID)
	if err != nil {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Business not found.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	if biz.OwnerID != playerIDStr {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You do not own this business.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	player, err := queries.GetPlayerByID(ctx, playerIDStr)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch player: %w", err)).Str("player_id", playerIDStr).Msg("status check failed")
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	employees, err := queries.GetEmploymentByBusiness(ctx, businessID)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch employees: %w", err)).Str("business_id", businessID).Msg("status check failed")
		return fmt.Errorf("failed to fetch employees: %w", err)
	}

	employeeCount := len(employees)
	baseRevenuePerEmployee := decimal.NewFromInt(100)
	projectedRevenue := decimal.NewFromInt(int64(employeeCount)).Mul(baseRevenuePerEmployee)

	data := StatusLayoutData{
		BusinessName:     biz.Name,
		BusinessType:     biz.BusinessType,
		OwnerName:        player.Username,
		EmployeeCount:    employeeCount,
		ProjectedRevenue: projectedRevenue,
		Status:           biz.Status,
	}

	imgBytes, err := composer.Compose("business", data)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to compose image: %w", err)).Msg("image composition failed")
		return sendStatusTextFallback(sess, i, biz.Name, biz.BusinessType, employeeCount, projectedRevenue, biz.Status)
	}

	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Business Dashboard",
					Description: fmt.Sprintf("Operational status of **%s**", biz.Name),
					Color:       0x4A90D9, // Blue for business/info
					Image:       &discordgo.MessageEmbedImage{URL: "attachment://business_status.png"},
					Footer:      &discordgo.MessageEmbedFooter{Text: "Aether City Chamber of Commerce"},
				},
			},
			Files: []*discordgo.File{
				{Name: "business_status.png", ContentType: "image/png", Reader: bytes.NewReader(imgBytes)},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func sendStatusTextFallback(sess *discordgo.Session, i *discordgo.InteractionCreate, name, bType string, empCount int, projRev decimal.Decimal, status string) error {
	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Business Dashboard",
					Description: fmt.Sprintf("**%s** (%s)\nStatus: %s\nEmployees: %d\nProjected Daily Revenue: `⊄%s`\n*(Image generation temporarily unavailable)*", name, bType, status, empCount, projRev.String()),
					Color:       0x4A90D9,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}