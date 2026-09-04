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

// MessageHandler is the callback signature for NATS message handlers.
// The handler receives the decoded event data and must return quickly —
// long-running work should be dispatched to a goroutine.
type MessageHandler func(ctx context.Context, envelope EventEnvelope) error

// Subscriber manages NATS JetStream subscriptions with durable consumers.
// It handles reconnection, message acknowledgment, and graceful shutdown.
type Subscriber struct {
	js           nats.JetStreamContext
	conn         *nats.Conn
	subscriptions []*nats.Subscription
	log          zerolog.Logger
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewSubscriber initializes the event subscriber.
func NewSubscriber(conn *nats.Conn, log zerolog.Logger) (*Subscriber, error) {
	js, err := conn.JetStream(nats.MaxWait(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize JetStream context: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Subscriber{
		js:            js,
		conn:          conn,
		subscriptions: make([]*nats.Subscription, 0),
		log:           log.With().Str("component", "events.subscriber").Logger(),
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}

// Subscribe registers a durable consumer for a given subject and wires the handler.
// The durable name ensures messages are not lost if the subscriber restarts.
func (s *Subscriber) Subscribe(subject, durableName string, handler MessageHandler) error {
	log := logger.FromContext(s.ctx, "events.subscriber")

	sub, err := s.js.Subscribe(subject, func(msg *nats.Msg) {
		var envelope EventEnvelope
		if err := json.Unmarshal(msg.Data, &envelope); err != nil {
			log.Error().Err(fmt.Errorf("failed to unmarshal event: %w", err)).
				Str("subject", subject).
				Msg("dropping malformed event")
			_ = msg.Ack() // Ack to prevent redelivery of malformed messages
			return
		}

		// Create a context with timeout for the handler to prevent blocking
		handlerCtx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
		defer cancel()

		if err := handler(handlerCtx, envelope); err != nil {
			log.Error().Err(fmt.Errorf("handler failed: %w", err)).
				Str("subject", subject).
				Str("msg_id", envelope.Timestamp.String()).
				Msg("event handler failed")
			_ = msg.Nak() // Request redelivery
			return
		}

		_ = msg.Ack()
	}, nats.Durable(durableName), nats.ManualAck(), nats.AckWait(60*time.Second))

	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
	}

	s.subscriptions = append(s.subscriptions, sub)
	log.Info().Str("subject", subject).Str("durable", durableName).Msg("subscription registered")
	return nil
}

// Stop gracefully drains all subscriptions and cancels the subscriber context.
func (s *Subscriber) Stop() {
	s.cancel()
	for _, sub := range s.subscriptions {
		_ = sub.Drain()
	}
	s.log.Info().Msg("event subscriber stopped")
}