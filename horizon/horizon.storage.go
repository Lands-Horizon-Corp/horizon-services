package horizon

import (
	"context"
	"io"
	"mime/multipart"
	"time"
)

// Storage represents metadata for uploaded files
type Storage struct {
	FileName   string // Original file name
	FileSize   int64  // File size in bytes
	FileType   string // MIME type
	StorageKey string // Unique storage identifier
	URL        string // Public access URL
	BucketName string // Storage bucket name
	Status     string // Upload status: pending, cancelled, corrupt, completed, progress
	Progress   int64  // Upload progress percentage
}

// ProgressCallback defines a function type for upload progress updates
type ProgressCallback func(progress int64, total int64, storage *Storage)

// StorageService defines the interface for cloud storage operations
type StorageService interface {
	// Run initializes connections to storage providers
	Run(ctx context.Context) error

	// Upload generic file upload from any io.Reader source
	Upload(ctx context.Context, file io.Reader, opts ProgressCallback) (Storage, error)

	// UploadFromBinary uploads raw byte data
	UploadFromBinary(ctx context.Context, data []byte, opts ProgressCallback) (Storage, error)

	// UploadFromHeader handles multipart form file uploads
	UploadFromHeader(ctx context.Context, hdr *multipart.FileHeader, opts ProgressCallback) (Storage, error)

	// UploadFromPath uploads a local file by path
	UploadFromPath(ctx context.Context, path string, opts ProgressCallback) (Storage, error)

	// GeneratePresignedURL creates a time-limited access URL for a stored file
	GeneratePresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)

	// DeleteFile permanently removes a file from storage
	DeleteFile(ctx context.Context, key string) error

	// GenerateUniqueName creates a collision-resistant filename
	GenerateUniqueName(ctx context.Context, originalName string) string
}
