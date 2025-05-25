package jetstream

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	sandwich "github.com/WelcomerTeam/Sandwich-Daemon"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type JetstreamProducerProvider struct {
	client  jetstream.JetStream
	channel string
}

type JetstreamProducer struct {
	*JetstreamProducerProvider
}

func NewJetstreamProducerProvider(ctx context.Context, address, channel string, natsOptions []nats.Option, jetstreamOptions []jetstream.JetStreamOpt) (sandwich.ProducerProvider, error) {
	provider := &JetstreamProducerProvider{
		client:  nil,
		channel: channel,
	}

	nc, err := nats.Connect(address, natsOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	js, err := jetstream.New(nc, jetstreamOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream client: %w", err)
	}

	provider.client = js

	return provider, nil
}

func (p *JetstreamProducerProvider) GetProducer(ctx context.Context, applicationIdentifier, clientName string) (sandwich.Producer, error) {
	_, err := p.client.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:              p.channel,
		Subjects:          []string{p.channel},
		Retention:         jetstream.WorkQueuePolicy,
		Discard:           jetstream.DiscardOld,
		MaxAge:            5 * time.Minute,
		Storage:           jetstream.MemoryStorage,
		MaxMsgsPerSubject: 1_000_000,
		MaxMsgSize:        math.MaxInt32,
		NoAck:             false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create or update stream: %w", err)
	}

	return &JetstreamProducer{p}, nil
}

func (p *JetstreamProducer) Publish(ctx context.Context, shard *sandwich.Shard, payload *sandwich.ProducedPayload) error {
	println("PRODUCE", shard.Application.Identifier, shard.ShardID, payload.Type)
	payloadData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal produced payload: %w", err)
	}

	_, err = p.client.Publish(ctx, p.channel, payloadData)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (p *JetstreamProducer) Close() error {
	return nil
}
