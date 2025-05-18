package horizon

import (
	"context"

	"github.com/gocql/gocql"
)

// NoSQLDatabase defines the interface for ScyllaDB operations
type NoSQLDatabase interface {
	// Start establishes the connection pool with the database cluster
	Start(ctx context.Context) error

	// Stop closes all active sessions and connections
	Stop(ctx context.Context) error

	// Client returns the configured Cassandra/ScyllaDB cluster configuration
	Client() *gocql.ClusterConfig

	// Ping verifies database connectivity
	Ping(ctx context.Context) error
}
