package horizon

import (
	"context"

	"gorm.io/gorm"
)

// SQLDatabase defines the interface for PostgreSQL operations
type SQLDatabase interface {
	// Start initializes the connection pool with the database
	Start(ctx context.Context) error

	// Stop closes all database connections
	Stop(ctx context.Context) error

	// Client returns the active GORM database client
	Client() *gorm.DB

	// Ping checks if the database is reachable
	Ping(ctx context.Context) error
}
