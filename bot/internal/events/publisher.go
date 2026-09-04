package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
)

// Subject constants for NATS JetStream. These are the authoritative channel names
// used across all services. Never hardcode subject strings elsewhere.
const (
	SubjectSettlementStart    = "settlement.start"
	SubjectSettlementComplete = "settlement.complete"
	SubjectMarketTrade        = "market.trade"
	SubjectEconomicNews       = "economic.news"
	SubjectPlayerNotification = "player.notification"
	SubjectAnomalyAlert       = "anomaly.alert"
)

// Publisher wraps a NATS JetStream connection for typed, structured event publishing.
// All events are JSON-encoded and include a timestamp for downstream ordering.
type Publisher struct {
	js   nats.JetStreamContext
	conn *nats.Conn
	log  zerolog.Logger
}

// NewPublisher initializes the event publisher from an existing NATS connection.
// It verifies JetStream is available before returning.
func NewPublisher(conn *nats.Conn, log zerolog.Logger) (*Publisher, error) {
	js, err := conn.JetStream(nats.MaxWait(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize JetStream context: %w", err)
	}

	return &Publisher{
		js:   js,
		conn: conn,
		log:  log.With().Str("component", "events.publisher").Logger(),
	}, nil
}

// EventEnvelope wraps all published events with metadata for ordering and tracing.
type EventEnvelope struct {
	Subject   string      `json:"subject"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// Publish sends a typed event to the NATS JetStream. The payload is JSON-encoded
// and wrapped in an EventEnvelope. Failures are logged but do not propagate —
// event publishing is best-effort to prevent blocking the caller's primary flow.
func (p *Publisher) Publish(ctx context.Context, subject string, data interface{}) error {
	log := logger.FromContext(ctx, "events.publisher")

	envelope := EventEnvelope{
		Subject:   subject,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}

	payload, err := json.Marshal(envelope)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to marshal event: %w", err)).
			Str("subject", subject).
			Msg("event publish aborted")
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// PublishAsync with a short timeout prevents blocking the caller on NATS backpressure
	msgID := fmt.Sprintf("%s-%d", subject, time.Now().UnixNano())
	ack, err := p.js.PublishMsgAsync(&nats.Msg{
		Subject: subject,
		Data:    payload,
		Header:  nats.Header{"Nats-Msg-Id": {msgID}},
	}, nats.MsgId(msgID))

	// Wait for ack synchronously with a short timeout
	select {
	case <-ack:
		log.Debug().Str("subject", subject).Str("msg_id", msgID).Msg("event published")
		return nil
	case err := <-err:
		log.Error().Err(fmt.Errorf("nats publish failed: %w", err)).
			Str("subject", subject).
			Msg("event publish failed")
		return fmt.Errorf("nats publish failed: %w", err)
	case <-time.After(2 * time.Second):
		log.Warn().Str("subject", subject).Msg("nats publish timed out waiting for ack")
		return fmt.Errorf("nats publish timed out")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// PublishSettlementComplete publishes the settlement completion event that triggers
// the Python analytics pipeline and the economic news bulletin.
func (p *Publisher) PublishSettlementComplete(ctx context.Context, economicDay int32, duration time.Duration) error {
	return p.Publish(ctx, SubjectSettlementComplete, map[string]interface{}{
		"economic_day": economicDay,
		"duration_ms":  duration.Milliseconds(),
		"status":       "success",
	})
}

// PublishMarketTrade publishes a real-time trade event for the Activity frontend's
// WebSocket hub to forward to connected clients viewing the Market panel.
func (p *Publisher) PublishMarketTrade(ctx context.Context, tradeID, buyerID, sellerID, itemType string, quantity int32, pricePerUnit string) error {
	return p.Publish(ctx, SubjectMarketTrade, map[string]interface{}{
		"trade_id":       tradeID,
		"buyer_id":       buyerID,
		"seller_id":      sellerID,
		"item_type":      itemType,
		"quantity":       quantity,
		"price_per_unit": pricePerUnit,
	})
}

// PublishPlayerNotification publishes a notification event for delivery via Discord DM
// or channel message. The notification service subscriber handles actual delivery.
func (p *Publisher) PublishPlayerNotification(ctx context.Context, playerID, notificationType, message string) error {
	return p.Publish(ctx, SubjectPlayerNotification, map[string]interface{}{
		"player_id":         playerID,
		"notification_type": notificationType,
		"message":           message,
	})
}

// PublishAnomalyAlert publishes an anomaly detection alert for operator review.
func (p *Publisher) PublishAnomalyAlert(ctx context.Context, flagType string, implicatedPlayers []string, evidence string) error {
	return p.Publish(ctx, SubjectAnomalyAlert, map[string]interface{}{
		"flag_type":          flagType,
		"implicated_players": implicatedPlayers,
		"evidence":           evidence,
	})
}