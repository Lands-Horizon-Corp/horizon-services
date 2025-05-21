package horizon

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Repository[TData any, TResponse any, TRequest any] interface {

	// Validate
	Validate(ctx echo.Context, v *validator.Validate) (*TRequest, error)

	// Models
	ToModel(data *TData, mapFunc func(*TData) *TResponse) *TResponse

	// Convert data to anything
	ToModels(data *TData, mapFunc func(*TData) *TResponse) *[]TResponse

	// --- Retrieval ---
	// List retrieves all entities of type T, optionally with related entities specified in preloads.
	List(ctx context.Context, preloads ...string) ([]*TData, error)
	ListRaw(ctx context.Context, preloads ...string) ([]*TResponse, error)

	// GetByID retrieves a single entity by its UUID. Optionally preloads related entities.
	GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*TData, error)
	GetByIDRaw(ctx context.Context, id uuid.UUID, preloads ...string) (*TResponse, error)

	// Find retrieves all entities that match the non-zero fields of the provided struct.
	// Optionally preloads related entities.
	Find(ctx context.Context, fields *TData, preloads ...string) ([]*TData, error)
	FindRaw(ctx context.Context, fields *TData, preloads ...string) ([]*TResponse, error)

	// FindOne retrieves a single entity that matches the non-zero fields of the provided struct.
	// Optionally preloads related entities.
	FindOne(ctx context.Context, fields *TData, preloads ...string) (*TData, error)
	FindOneRaw(ctx context.Context, fields *TData, preloads ...string) (*TResponse, error)

	// --- Aggregation ---

	// Count returns the number of records matching the given fields.
	Count(ctx context.Context, fields *TData) (int64, error)

	// CountWithTx performs Count using the provided GORM transaction.
	CountWithTx(ctx context.Context, tx *gorm.DB, fields *TData) (int64, error)

	// --- Creation ---

	// Create inserts a new record into the database.
	// Optionally preloads related entities after creation.
	Create(ctx context.Context, entity *TData, preloads ...string) error

	// CreateWithTx performs Create within the provided transaction.
	CreateWithTx(ctx context.Context, tx *gorm.DB, entity *TData, preloads ...string) error

	// CreateMany inserts multiple records in a batch.
	CreateMany(ctx context.Context, entities []*TData, preloads ...string) error

	// CreateManyWithTx performs CreateMany within the provided transaction.
	CreateManyWithTx(ctx context.Context, tx *gorm.DB, entities []*TData, preloads ...string) error

	// --- Update ---

	// Update modifies an existing record in the database.
	Update(ctx context.Context, entity *TData, preloads ...string) error

	// UpdateWithTx performs Update within the provided transaction.
	UpdateWithTx(ctx context.Context, tx *gorm.DB, entity *TData, preloads ...string) error

	// UpdateByID updates a record by its UUID.
	UpdateByID(ctx context.Context, id uuid.UUID, entity *TData, preloads ...string) error

	// UpdateByIDWithTx performs UpdateByID within the provided transaction.
	UpdateByIDWithTx(ctx context.Context, tx *gorm.DB, id uuid.UUID, entity *TData, preloads ...string) error

	// UpdateFields updates only the non-zero fields of the entity for the given UUID.
	UpdateFields(ctx context.Context, id uuid.UUID, fields *TData, preloads ...string) error

	// UpdateFieldsWithTx performs UpdateFields within the provided transaction.
	UpdateFieldsWithTx(ctx context.Context, tx *gorm.DB, id uuid.UUID, fields *TData, preloads ...string) error

	// UpdateMany performs a batch update on multiple entities.
	UpdateMany(ctx context.Context, entities []*TData, preloads ...string) error

	// UpdateManyWithTx performs UpdateMany within the provided transaction.
	UpdateManyWithTx(ctx context.Context, tx *gorm.DB, entities []*TData, preloads ...string) error

	// --- Upsert ---

	// Upsert inserts a new record or updates it if it already exists.
	Upsert(ctx context.Context, entity *TData, preloads ...string) error

	// UpsertWithTx performs Upsert within the provided transaction.
	UpsertWithTx(ctx context.Context, tx *gorm.DB, entity *TData, preloads ...string) error

	// UpsertMany performs batch upsert for multiple entities.
	UpsertMany(ctx context.Context, entities []*TData, preloads ...string) error

	// UpsertManyWithTx performs UpsertMany within the provided transaction.
	UpsertManyWithTx(ctx context.Context, tx *gorm.DB, entities []*TData, preloads ...string) error

	// --- Deletion ---

	// Delete removes the specified entity from the database.
	Delete(ctx context.Context, entity *TData) error

	// DeleteWithTx performs Delete within the provided transaction.
	DeleteWithTx(ctx context.Context, tx *gorm.DB, entity *TData) error

	// DeleteByID deletes a record by its UUID.
	DeleteByID(ctx context.Context, id uuid.UUID) error

	// DeleteByIDWithTx performs DeleteByID within the provided transaction.
	DeleteByIDWithTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) error

	// DeleteMany deletes multiple entities in a batch.
	DeleteMany(ctx context.Context, entities []*TData) error

	// DeleteManyWithTx performs DeleteMany within the provided transaction.
	DeleteManyWithTx(ctx context.Context, tx *gorm.DB, entities []*TData) error
}
