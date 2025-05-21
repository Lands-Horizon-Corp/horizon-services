package src

import (
	"context"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Repository[TData any, TResponse any, TRequest any] interface {

	// Validate
	Validate(ctx echo.Context, v *validator.Validate) (*TRequest, error)

	// Models
	ToModel(data *TData) *TResponse

	// Convert data to anything
	ToModels(data []*TData) []*TResponse

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

// RepositoryParams groups the constructor parameters for NewRepository
type RepositoryParams[TData any, TResponse any, TRequest any] struct {
	Provider *Provider
	Created  func(*TData) ([]string, any)
	Updated  func(*TData) ([]string, any)
	Deleted  func(*TData) ([]string, any)
	Resource func(*TData) *TResponse
	Preloads []string
}

// CollectionManager is a generic implementation of Repository
type CollectionManager[TData any, TResponse any, TRequest any] struct {
	provider *Provider
	created  func(*TData) ([]string, any)
	updated  func(*TData) ([]string, any)
	deleted  func(*TData) ([]string, any)
	resource func(*TData) *TResponse
	preloads []string
}

// NewRepository creates a new CollectionManager instance with the given parameters
func NewRepository[TData any, TResponse any, TRequest any](params RepositoryParams[TData, TResponse, TRequest]) Repository[TData, TResponse, TRequest] {
	return &CollectionManager[TData, TResponse, TRequest]{
		provider: params.Provider,
		created:  params.Created,
		updated:  params.Updated,
		deleted:  params.Deleted,
		resource: params.Resource,
		preloads: params.Preloads,
	}
}

// ToModel implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) ToModel(data *TData) *TResponse {
	if data == nil {
		return nil
	}
	return c.resource(data)
}

// ToModels implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) ToModels(data []*TData) []*TResponse {
	if data == nil {
		return []*TResponse{}
	}
	out := make([]*TResponse, 0, len(data))
	for _, item := range data {
		if m := c.ToModel(item); m != nil {
			out = append(out, m)
		}
	}
	return out
}

// Validate implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) Validate(ctx echo.Context, v *validator.Validate) (*TRequest, error) {
	var req TRequest
	if err := ctx.Bind(&req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := v.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return &req, nil
}

// Count implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) Count(ctx context.Context, fields *TData) (int64, error) {
	panic("unimplemented")
}

// CountWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) CountWithTx(ctx context.Context, tx *gorm.DB, fields *TData) (int64, error) {
	panic("unimplemented")
}

// Create implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) Create(ctx context.Context, entity *TData, preloads ...string) error {
	panic("unimplemented")
}

// CreateMany implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) CreateMany(ctx context.Context, entities []*TData, preloads ...string) error {
	panic("unimplemented")
}

// CreateManyWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) CreateManyWithTx(ctx context.Context, tx *gorm.DB, entities []*TData, preloads ...string) error {
	panic("unimplemented")
}

// CreateWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) CreateWithTx(ctx context.Context, tx *gorm.DB, entity *TData, preloads ...string) error {
	panic("unimplemented")
}

// Delete implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) Delete(ctx context.Context, entity *TData) error {
	panic("unimplemented")
}

// DeleteByID implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) DeleteByID(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

// DeleteByIDWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) DeleteByIDWithTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) error {
	panic("unimplemented")
}

// DeleteMany implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) DeleteMany(ctx context.Context, entities []*TData) error {
	panic("unimplemented")
}

// DeleteManyWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) DeleteManyWithTx(ctx context.Context, tx *gorm.DB, entities []*TData) error {
	panic("unimplemented")
}

// DeleteWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) DeleteWithTx(ctx context.Context, tx *gorm.DB, entity *TData) error {
	panic("unimplemented")
}

// Find implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) Find(ctx context.Context, fields *TData, preloads ...string) ([]*TData, error) {
	panic("unimplemented")
}

// FindOne implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) FindOne(ctx context.Context, fields *TData, preloads ...string) (*TData, error) {
	panic("unimplemented")
}

// FindOneRaw implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) FindOneRaw(ctx context.Context, fields *TData, preloads ...string) (*TResponse, error) {
	panic("unimplemented")
}

// FindRaw implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) FindRaw(ctx context.Context, fields *TData, preloads ...string) ([]*TResponse, error) {
	panic("unimplemented")
}

// GetByID implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) GetByID(ctx context.Context, id uuid.UUID, preloads ...string) (*TData, error) {
	panic("unimplemented")
}

// GetByIDRaw implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) GetByIDRaw(ctx context.Context, id uuid.UUID, preloads ...string) (*TResponse, error) {
	panic("unimplemented")
}

// List implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) List(ctx context.Context, preloads ...string) ([]*TData, error) {
	panic("unimplemented")
}

// ListRaw implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) ListRaw(ctx context.Context, preloads ...string) ([]*TResponse, error) {
	panic("unimplemented")
}

// Update implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) Update(ctx context.Context, entity *TData, preloads ...string) error {
	panic("unimplemented")
}

// UpdateByID implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) UpdateByID(ctx context.Context, id uuid.UUID, entity *TData, preloads ...string) error {
	panic("unimplemented")
}

// UpdateByIDWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) UpdateByIDWithTx(ctx context.Context, tx *gorm.DB, id uuid.UUID, entity *TData, preloads ...string) error {
	panic("unimplemented")
}

// UpdateFields implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) UpdateFields(ctx context.Context, id uuid.UUID, fields *TData, preloads ...string) error {
	panic("unimplemented")
}

// UpdateFieldsWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) UpdateFieldsWithTx(ctx context.Context, tx *gorm.DB, id uuid.UUID, fields *TData, preloads ...string) error {
	panic("unimplemented")
}

// UpdateMany implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) UpdateMany(ctx context.Context, entities []*TData, preloads ...string) error {
	panic("unimplemented")
}

// UpdateManyWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) UpdateManyWithTx(ctx context.Context, tx *gorm.DB, entities []*TData, preloads ...string) error {
	panic("unimplemented")
}

// UpdateWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) UpdateWithTx(ctx context.Context, tx *gorm.DB, entity *TData, preloads ...string) error {
	panic("unimplemented")
}

// Upsert implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) Upsert(ctx context.Context, entity *TData, preloads ...string) error {
	panic("unimplemented")
}

// UpsertMany implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) UpsertMany(ctx context.Context, entities []*TData, preloads ...string) error {
	panic("unimplemented")
}

// UpsertManyWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) UpsertManyWithTx(ctx context.Context, tx *gorm.DB, entities []*TData, preloads ...string) error {
	panic("unimplemented")
}

// UpsertWithTx implements Repository.
func (c *CollectionManager[TData, TResponse, TRequest]) UpsertWithTx(ctx context.Context, tx *gorm.DB, entity *TData, preloads ...string) error {
	panic("unimplemented")
}
