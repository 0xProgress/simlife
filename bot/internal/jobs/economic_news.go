package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/imaging"
	"github.com/0xProgress/simlife/bot/internal/imaging/layouts"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

// PublishEconomicNews generates and posts the daily economic bulletin to the
// configured Discord news channel. It fetches the latest snapshot from the
// database, composes a branded image, and attaches it to a styled embed.
// If image composition fails, it falls back to a text-only embed so the
// bulletin is never silently dropped.
func PublishEconomicNews(ctx context.Context, sess *discordgo.Session, composer *imaging.Composer, queries *sqlc.Queries, log zerolog.Logger) {
	log = log.With().Str("job", "economic_news").Logger()

	if sess == nil || composer == nil || queries == nil {
		log.Error().Msg("economic news publisher received nil dependency")
		return
	}

	// Determine the target news channel
	channelID := resolveNewsChannelID(queries, log)
	if channelID == "" {
		log.Warn().Msg("no news channel configured; skipping economic news publication")
		return
	}

	// 1. Fetch the latest economic snapshot
	snapshot, err := queries.GetLatestEconomicSnapshot(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Warn().Msg("no economic snapshots found; skipping news publication")
			return
		}
		log.Error().Err(fmt.Errorf("failed to fetch latest snapshot: %w", err)).
			Msg("economic news publication aborted")
		return
	}

	// 2. Parse the snapshot metrics (stored as JSONB)
	metrics, err := parseSnapshotMetrics(snapshot.Metrics)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to parse snapshot metrics: %w", err)).
			Int32("economic_day", snapshot.EconomicDay).
			Msg("economic news publication aborted")
		return
	}

	// 3. Fetch the previous day's snapshot for the price index change calculation
	prevPriceIndex := decimal.NewFromFloat(100.0) // Default launch baseline
	prevSnapshot, err := queries.GetPreviousEconomicSnapshot(ctx, snapshot.EconomicDay)
	if err == nil {
		if prevMetrics, parseErr := parseSnapshotMetrics(prevSnapshot.Metrics); parseErr == nil {
			prevPriceIndex = prevMetrics.PriceIndex
		}
	}

	// Compute index change percentage
	indexChange := decimal.Zero
	if !prevPriceIndex.IsZero() {
		indexChange = metrics.PriceIndex.Sub(prevPriceIndex).Div(prevPriceIndex).Mul(decimal.NewFromFloat(100.0))
	}

	// 4. Fetch the active world event for flavor text (optional — non-fatal if missing)
	flavorText := "A quiet day in Aether City. Business as usual."
	activeEvent, err := queries.GetActiveWorldEvent(ctx)
	if err == nil && activeEvent.FlavorText != "" {
		flavorText = activeEvent.FlavorText
	}

	// 5. Build the layout data
	data := layouts.EconomicNewsData{
		EconDay:     int(snapshot.EconomicDay),
		PriceIndex:  metrics.PriceIndex.InexactFloat64(),
		IndexChange: indexChange.InexactFloat64(),
		Velocity:    metrics.Velocity.InexactFloat64(),
		GiniCoeff:   metrics.InequalityRatio.InexactFloat64(),
		TopEarners:  metrics.TopEarners,
		MoneySupply: metrics.MoneySupply,
		FlavorText:  flavorText,
	}

	// 6. Compose the image
	imgBytes, err := composer.Compose("economic_news", data)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to compose economic news image: %w", err)).
			Msg("falling back to text-only bulletin")
		publishTextOnlyNews(ctx, sess, channelID, data, log)
		return
	}

	// 7. Publish the full bulletin with image
	embed := buildNewsEmbed(data, imgBytes)
	_, err = sess.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: fmt.Sprintf("📰 **%s**", flavorText),
		Embeds:  []*discordgo.MessageEmbed{embed},
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to publish economic news: %w", err)).
			Str("channel_id", channelID).
			Msg("economic news publication failed")
		return
	}

	log.Info().
		Int32("economic_day", snapshot.EconomicDay).
		Str("channel_id", channelID).
		Str("money_supply", metrics.MoneySupply.String()).
		Msg("daily economic news published successfully")
}

// resolveNewsChannelID returns the configured news channel ID. In Layer 1,
// this is a single global channel from config. Future layers may support
// per-district channels.
func resolveNewsChannelID(queries *sqlc.Queries, log zerolog.Logger) string {
	// For Layer 1, we use a hardcoded well-known channel or a config value.
	// The config should expose this; for now, we fall back to a placeholder.
	// In production, this would be s.cfg.NewsChannelID.
	return "" // Caller should set this via config; placeholder to prevent accidental posting
}

// parseSnapshotMetrics deserializes the JSONB metrics blob into a typed struct.
func parseSnapshotMetrics(data []byte) (*SnapshotMetricsParsed, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty metrics payload")
	}

	// Use a permissive parser that tolerates both string and numeric JSON values
	// (the analytics service may post either format depending on its serializer).
	var raw map[string]interface{}
	if err := jsonUnmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics JSON: %w", err)
	}

	m := &SnapshotMetricsParsed{
		TopEarners: []string{},
	}

	if v, ok := raw["money_supply"]; ok {
		m.MoneySupply = parseDecimal(v)
	}
	if v, ok := raw["velocity"]; ok {
		m.Velocity = parseDecimal(v)
	}
	if v, ok := raw["inequality_ratio"]; ok {
		m.InequalityRatio = parseDecimal(v)
	}
	if v, ok := raw["base_wage_rate"]; ok {
		m.BaseWageRate = parseDecimal(v)
	}
	if v, ok := raw["top_earners"]; ok {
		if arr, isArr := v.([]interface{}); isArr {
			for _, item := range arr {
				if s, isStr := item.(string); isStr {
					m.TopEarners = append(m.TopEarners, s)
				}
			}
		}
	}

	return m, nil
}

// parseDecimal converts a JSON value (string, float64, or int) to decimal.Decimal.
func parseDecimal(v interface{}) decimal.Decimal {
	switch val := v.(type) {
	case string:
		d, err := decimal.NewFromString(val)
		if err == nil {
			return d
		}
	case float64:
		return decimal.NewFromFloat(val)
	case int:
		return decimal.NewFromInt(int64(val))
	case int64:
		return decimal.NewFromInt(val)
	case json.Number:
		d, err := decimal.NewFromString(string(val))
		if err == nil {
			return d
		}
	}
	return decimal.Zero
}

// SnapshotMetricsParsed is the typed representation of the JSONB metrics blob.
type SnapshotMetricsParsed struct {
	MoneySupply     decimal.Decimal `json:"money_supply"`
	Velocity        decimal.Decimal `json:"velocity"`
	TopEarners      []string        `json:"top_earners"`
	InequalityRatio decimal.Decimal `json:"inequality_ratio"`
	BaseWageRate    decimal.Decimal `json:"base_wage_rate"`
}

// buildNewsEmbed constructs the Discord embed for the economic bulletin.
func buildNewsEmbed(data layouts.EconomicNewsData, imgBytes []byte) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📊 Day %d Economic Report", data.EconDay),
		Description: fmt.Sprintf("**Price Index:** %.2f | **Velocity:** %.2f | **Gini:** %.2f", data.PriceIndex, data.Velocity, data.GiniCoeff),
		Color:       0xD4A847, // --accent-gold
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Money Supply: ⊄%s", data.MoneySupply.String()),
		},
	}

	if imgBytes != nil {
		embed.Image = &discordgo.MessageEmbedImage{URL: "attachment://economic_news.png"}
	}

	return embed
}

// publishTextOnlyNews is the graceful fallback when image composition fails.
func publishTextOnlyNews(ctx context.Context, sess *discordgo.Session, channelID string, data layouts.EconomicNewsData, log zerolog.Logger) {
	description := fmt.Sprintf(
		"**Price Index:** %.2f\n**Velocity:** %.2f\n**Gini Coefficient:** %.2f\n**Money Supply:** ⊄%s\n\n*Image generation temporarily unavailable.*",
		data.PriceIndex, data.Velocity, data.GiniCoeff, data.MoneySupply.String(),
	)

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📊 Day %d Economic Report", data.EconDay),
		Description: description,
		Color:       0xD4A847,
	}

	_, err := sess.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to publish text-only news: %w", err)).
			Str("channel_id", channelID).
			Msg("text-only news fallback failed")
	}
}

// jsonUnmarshal is a thin wrapper to avoid importing encoding/json at the top
// of every file. It uses the standard library's JSON decoder.
func jsonUnmarshal(data []byte, v interface{}) error {
	return jsonDecoder(data).Decode(v)
}

// jsonDecoder returns a JSON decoder configured for strict parsing.
func jsonDecoder(data []byte) *json.Decoder {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber() // Preserve numeric precision as json.Number
	return dec
}