package business

import (
	"context"
	"fmt"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/commands"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/bwmarrin/discordgo"
	"github.com/shopspring/decimal"
)

// HandleProduce processes the /business produce command.
// In Layer 4, this manually flags the business as "Open" for the day, triggering 
// a projected revenue calculation based on employee count. Actual settlement 
// and ledger transfers are processed by the daily settlement engine.
func HandleProduce(ctx context.Context, sess *discordgo.Session, i *discordgo.InteractionCreate, queries *sqlc.Queries) error {
	log := logger.FromContext(ctx, "commands.business")

	ownerIDStr, ok := ctx.Value(commands.PlayerIDKey).(string)
	if !ok || ownerIDStr == "" {
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

	if biz.OwnerID != ownerIDStr {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You do not own this business.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	employees, err := queries.GetEmploymentByBusiness(ctx, businessID)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch employees: %w", err)).Str("business_id", businessID).Msg("produce check failed")
		return fmt.Errorf("failed to fetch employees: %w", err)
	}

	employeeCount := len(employees)
	if employeeCount == 0 {
		return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You have no employees. Hire workers first to generate revenue.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	baseRevenuePerEmployee := decimal.NewFromInt(100)
	projectedRevenue := decimal.NewFromInt(int64(employeeCount)).Mul(baseRevenuePerEmployee)

	err = queries.MarkBusinessOperated(ctx, biz.ID)
	if err != nil {
		log.Warn().Err(fmt.Errorf("failed to mark business operated: %w", err)).Str("business_id", businessID).Msg("produce warning")
	}

	log.Info().
		Str("business_id", businessID).
		Int("employees", employeeCount).
		Str("projected_revenue", projectedRevenue.String()).
		Msg("business production initiated for the day")

	return sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("**%s** is now open for production today.\nProjected revenue based on %d employees: **⊄%s**.\n*(Settlement will process actual revenue and payroll at the end of the day.)*", biz.Name, employeeCount, projectedRevenue.String()),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}