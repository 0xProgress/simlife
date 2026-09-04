package world

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

// EventType categorizes world events.
type EventType string

const (
	EventRandom    EventType = "RANDOM"
	EventScheduled EventType = "SCHEDULED"
	EventPlayer    EventType = "PLAYER"
)

// EventEffect defines how an event modifies the economy.
type EventEffect struct {
	WageMultiplier     decimal.Decimal `json:"wage_multiplier"`     // 1.0 = no change
	PriceMultiplier    decimal.Decimal `json:"price_multiplier"`    // 1.0 = no change
	TaxMultiplier      decimal.Decimal `json:"tax_multiplier"`      // 1.0 = no change
	TargetSector       string          `json:"target_sector"`       // e.g., "CONSTRUCTION", "ALL"
	Description        string          `json:"description"`
}

// WorldEvent defines a possible world event.
type WorldEvent struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Type        EventType    `json:"type"`
	Weight      int          `json:"weight"` // Relative probability (higher = more likely)
	Effect      EventEffect  `json:"effect"`
	Duration    time.Duration `json:"duration"` // How long the effect lasts
	FlavorText  string       `json:"flavor_text"` // Narrative description for news embed
}

// EventEngine manages the world events system.
type EventEngine struct {
	queries *sqlc.Queries
	nats    *nats.Conn
	redis   *redis.Client
	events  []*WorldEvent
	rng     *rand.Rand
	log     zerolog.Logger
}

// NewEventEngine initializes the event engine with the predefined event catalog.
func NewEventEngine(q *sqlc.Queries, n *nats.Conn, r *redis.Client, log zerolog.Logger) *EventEngine {
	return &EventEngine{
		queries: q,
		nats:    n,
		redis:   r,
		events:  defaultEventCatalog(),
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
		log:     log.With().Str("component", "world.events").Logger(),
	}
}

// defaultEventCatalog returns the built-in catalog of world events.
func defaultEventCatalog() []*WorldEvent {
	return []*WorldEvent{
		{
			ID:   "construction_boom",
			Name: "Construction Boom",
			Type: EventRandom,
			Weight: 15,
			Effect: EventEffect{
				WageMultiplier: decimal.NewFromFloat(1.15), // +15% wages
				TargetSector:   "CONSTRUCTION",
				Description:    "A surge in building projects drives up construction wages.",
			},
			Duration:   24 * time.Hour,
			FlavorText: "🏗️ Construction Boom! All physical workers earn 15% more today.",
		},
		{
			ID:   "market_crash",
			Name: "Market Correction",
			Type: EventRandom,
			Weight: 5,
			Effect: EventEffect{
				PriceMultiplier: decimal.NewFromFloat(0.85), // -15% prices
				TargetSector:    "ALL",
				Description:     "A market correction drives down prices across the board.",
			},
			Duration:   24 * time.Hour,
			FlavorText: "📉 Market Correction! Prices across all sectors drop 15% today.",
		},
		{
			ID:   "tech_innovation",
			Name: "Technological Breakthrough",
			Type: EventRandom,
			Weight: 10,
			Effect: EventEffect{
				WageMultiplier: decimal.NewFromFloat(1.10), // +10% wages
				TargetSector:   "TECHNOLOGY",
				Description:    "A new technology increases productivity in the tech sector.",
			},
			Duration:   24 * time.Hour,
			FlavorText: "💡 Technological Breakthrough! Tech workers earn 10% more today.",
		},
		{
			ID:   "harvest_festival",
			Name: "Harvest Festival",
			Type: EventScheduled,
			Weight: 20,
			Effect: EventEffect{
				PriceMultiplier: decimal.NewFromFloat(0.90), // -10% food prices
				TargetSector:    "FOOD",
				Description:     "A bountiful harvest drives down food prices.",
			},
			Duration:   48 * time.Hour,
			FlavorText: "🌾 Harvest Festival! Food prices drop 10% for the next two days.",
		},
		{
			ID:   "tax_holiday",
			Name: "Tax Holiday",
			Type: EventScheduled,
			Weight: 3,
			Effect: EventEffect{
				TaxMultiplier: decimal.NewFromFloat(0.50), // -50% taxes
				TargetSector:  "ALL",
				Description:   "The city declares a tax holiday to stimulate spending.",
			},
			Duration:   24 * time.Hour,
			FlavorText: "🎉 Tax Holiday! All taxes are halved for today only.",
		},
		{
			ID:   "service_demand_surge",
			Name: "Service Demand Surge",
			Type: EventRandom,
			Weight: 12,
			Effect: EventEffect{
				WageMultiplier: decimal.NewFromFloat(1.12), // +12% wages
				TargetSector:   "SERVICE",
				Description:    "A surge in customer demand boosts service worker wages.",
			},
			Duration:   24 * time.Hour,
			FlavorText: "🍽️ Service Demand Surge! Service workers earn 12% more today.",
		},
		{
			ID:   "industrial_slowdown",
			Name: "Industrial Slowdown",
			Type: EventRandom,
			Weight: 8,
			Effect: EventEffect{
				WageMultiplier: decimal.NewFromFloat(0.90), // -10% wages
				TargetSector:   "MANUFACTURING",
				Description:    "A slowdown in manufacturing reduces industrial wages.",
			},
			Duration:   24 * time.Hour,
			FlavorText: "🏭 Industrial Slowdown. Manufacturing wages drop 10% today.",
		},
		{
			ID:   "neutral_day",
			Name: "Ordinary Day",
			Type: EventRandom,
			Weight: 27, // High weight = most common
			Effect: EventEffect{
				WageMultiplier:  decimal.NewFromFloat(1.0),
				PriceMultiplier: decimal.NewFromFloat(1.0),
				TaxMultiplier:   decimal.NewFromFloat(1.0),
				TargetSector:    "ALL",
				Description:     "A quiet day in Aether City.",
			},
			Duration:   24 * time.Hour,
			FlavorText: "☀️ A quiet day in Aether City. Business as usual.",
		},
	}
}

// SelectDailyEvent picks a random event for the current economic day based on weights.
func (e *EventEngine) SelectDailyEvent(ctx context.Context) (*WorldEvent, error) {
	log := logger.FromContext(ctx, "world.events")

	// Calculate total weight
	totalWeight := 0
	for _, event := range e.events {
		totalWeight += event.Weight
	}

	if totalWeight == 0 {
		return nil, fmt.Errorf("event catalog is empty")
	}

	// Weighted random selection
	roll := e.rng.Intn(totalWeight)
	cumulative := 0
	for _, event := range e.events {
		cumulative += event.Weight
		if roll < cumulative {
			log.Info().
				Str("event_id", event.ID).
				Str("event_name", event.Name).
				Int("roll", roll).
				Int("total_weight", totalWeight).
				Msg("daily event selected")

			return event, nil
		}
	}

	// Fallback (should never reach here)
	return e.events[len(e.events)-1], nil
}

// ApplyEvent records the selected event and publishes it to NATS for the news system.
func (e *EventEngine) ApplyEvent(ctx context.Context, event *WorldEvent, economicDay int32) error {
	log := logger.FromContext(ctx, "world.events")

	// 1. Record event in database
	effectJSON, err := json.Marshal(event.Effect)
	if err != nil {
		return fmt.Errorf("failed to marshal event effect: %w", err)
	}

	err = e.queries.RecordWorldEvent(ctx, sqlc.RecordWorldEventParams{
		EventID:       event.ID,
		EventName:     event.Name,
		EventType:     string(event.Type),
		EconomicDay:   economicDay,
		Effect:        effectJSON,
		FlavorText:    event.FlavorText,
		ExpiresAt:     time.Now().Add(event.Duration),
	})
	if err != nil {
		return fmt.Errorf("failed to record world event: %w", err)
	}

	// 2. Cache active event in Redis for quick lookup by /work and pricing engine
	eventData := map[string]interface{}{
		"id":          event.ID,
		"name":        event.Name,
		"effect":      event.Effect,
		"expires_at":  time.Now().Add(event.Duration).Unix(),
		"flavor_text": event.FlavorText,
	}
	eventJSON, _ := json.Marshal(eventData)
	e.redis.Set(ctx, "world:active_event", eventJSON, event.Duration+time.Hour)

	// 3. Publish to NATS for the economic news system
	natsMsg := map[string]interface{}{
		"event":        event,
		"economic_day": economicDay,
		"timestamp":    time.Now().UTC(),
	}
	msgJSON, _ := json.Marshal(natsMsg)
	if err := e.nats.Publish("economic.news", msgJSON); err != nil {
		log.Warn().Err(err).Msg("failed to publish event to NATS, news bulletin may be delayed")
	}

	log.Info().
		Str("event_id", event.ID).
		Str("event_name", event.Name).
		Int32("economic_day", economicDay).
		Msg("world event applied and published")

	return nil
}

// GetActiveEvent retrieves the currently active world event from Redis cache.
// Returns nil if no event is active.
func (e *EventEngine) GetActiveEvent(ctx context.Context) (*ActiveEvent, error) {
	cached, err := e.redis.Get(ctx, "world:active_event").Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read active event from cache: %w", err)
	}

	var event ActiveEvent
	if err := json.Unmarshal([]byte(cached), &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal active event: %w", err)
	}

	// Check if expired
	if time.Now().Unix() > event.ExpiresAt {
		return nil, nil
	}

	return &event, nil
}

// ActiveEvent represents the currently active world event.
type ActiveEvent struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Effect      EventEffect `json:"effect"`
	ExpiresAt   int64       `json:"expires_at"`
	FlavorText  string      `json:"flavor_text"`
}

// GetEventModifier returns the economic modifier for a specific sector.
// This is used by the /work command and pricing engine to adjust wages and prices.
func (e *EventEngine) GetEventModifier(ctx context.Context, sector string) EventEffect {
	activeEvent, err := e.GetActiveEvent(ctx)
	if err != nil || activeEvent == nil {
		// No active event — return neutral modifiers
		return EventEffect{
			WageMultiplier:  decimal.NewFromFloat(1.0),
			PriceMultiplier: decimal.NewFromFloat(1.0),
			TaxMultiplier:   decimal.NewFromFloat(1.0),
			TargetSector:    "NONE",
		}
	}

	// If the event targets "ALL" or matches the requested sector, apply it
	if activeEvent.Effect.TargetSector == "ALL" || activeEvent.Effect.TargetSector == sector {
		return activeEvent.Effect
	}

	// Event doesn't affect this sector — return neutral
	return EventEffect{
		WageMultiplier:  decimal.NewFromFloat(1.0),
		PriceMultiplier: decimal.NewFromFloat(1.0),
		TaxMultiplier:   decimal.NewFromFloat(1.0),
		TargetSector:    "NONE",
	}
}