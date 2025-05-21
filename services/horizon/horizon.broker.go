package horizon

import (
	"context"
)

// MessageBroker defines the interface for pub/sub messaging systems
type MessageBroker interface {
	// Run connects to a broker cluster
	Run(ctx context.Context, brokers []string) error

	// Stop closes all producer/consumer connections
	Stop(ctx context.Context) error

	// Publish sends a message to a single topic
	Publish(ctx context.Context, topic string, payload []byte) error

	// DispatchBatch sends a message to multiple topics
	DispatchBatch(ctx context.Context, topics []string, payload []byte)

	// Subscribe registers a message handler for a topic
	Subscribe(ctx context.Context, topic string, handler func([]byte) error) error
}
