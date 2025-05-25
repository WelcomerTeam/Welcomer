package jetstream

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func SetupJetstreamConsumer(ctx context.Context, address, channel, clientName string, natsOptions []nats.Option, jetstreamOptions []jetstream.JetStreamOpt) (jetstream.Consumer, error) {
	nc, err := nats.Connect(address, natsOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	js, err := jetstream.New(nc, jetstreamOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream client: %w", err)
	}

	jsc, err := js.CreateOrUpdateConsumer(ctx, channel, jetstream.ConsumerConfig{
		Name:          clientName + "_" + strconv.Itoa(os.Getpid()),
		DeliverPolicy: jetstream.DeliverAllPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: channel,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream consumer: %w", err)
	}

	return jsc, nil
}
