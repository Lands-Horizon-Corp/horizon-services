package horizon

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rotisserie/eris"
)

/*

   storageSvc := NewHorizonStorageService(
        storageHost,
        accessKey,
        secretKey,
        driver,
        region,
        bucket,
        port,
        maxSize,
    )

    // Initialize connection and bucket
    if err := storageSvc.Run(ctx); err != nil {
        panic(err)
    }

    // Upload from path
    storage1, err := storageSvc.Upload(ctx, "/path/to/file.jpg", nil)
    if err != nil {
        fmt.Println("Path upload error:", err)
    } else {
        fmt.Printf("Uploaded via path: %+v", storage1)
	}

*/

// Storage represents metadata for uploaded files
type Storage struct {
	FileName   string // Original file name
	FileSize   int64  // File size in bytes
	FileType   string // MIME type
	StorageKey string // Unique Storage identifier
	URL        string // Public access URL
	BucketName string // Storage bucket name
	Status     string // Upload status: pending, cancelled, corrupt, completed, progress
	Progress   int64  // Upload progress percentage
}

// ProgressCallback defines a function type for upload progress updates
type ProgressCallback func(progress int64, total int64, Storage *Storage)

// StorageService defines the interface for cloud Storage operations
type StorageService interface {
	// Run initializes connections to Storage providers
	Run(ctx context.Context) error

	// Upload implements StorageService and delegates based on input type:
	// - string: local file path
	// - []byte: binary data
	// - *multipart.FileHeader: HTTP uploaded file
	Upload(ctx context.Context, file any, opts ProgressCallback) (*Storage, error)

	// UploadFromBinary uploads raw byte data
	UploadFromBinary(ctx context.Context, data []byte, opts ProgressCallback) (*Storage, error)

	// UploadFromHeader handles multipart form file uploads
	UploadFromHeader(ctx context.Context, hdr *multipart.FileHeader, opts ProgressCallback) (*Storage, error)

	// UploadFromPath uploads a local file by path
	UploadFromPath(ctx context.Context, path string, opts ProgressCallback) (*Storage, error)

	// GeneratePresignedURL creates a time-limited access URL for a stored file
	GeneratePresignedURL(ctx context.Context, Storage *Storage, expiry time.Duration) (string, error)

	// DeleteFile permanently removes a file from Storage
	DeleteFile(ctx context.Context, Storage *Storage) error

	// GenerateUniqueName creates a collision-resistant filename
	GenerateUniqueName(ctx context.Context, originalName string) (string, error)
}

type progressReader struct {
	reader    io.Reader
	callback  ProgressCallback
	total     int64
	readSoFar int64
	Storage   *Storage
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.readSoFar += int64(n)
		percent := pr.readSoFar * 100 / pr.total
		if percent > 100 {
			percent = 100
		}
		pr.Storage.Progress = percent
		if pr.callback != nil {
			pr.callback(percent, 100, pr.Storage)
		}
	}
	return n, err
}

type HorizonStorage struct {
	storageHost       string
	storageAccessKey  string
	storageSecreetKey string
	storageDriver     string
	storageRegion     string
	storageBucket     string
	storageAPIPort    int16
	maxFileSize       int64
	Storage           *minio.Client
}

func NewHorizonStorageService(
	storageHost,
	storageAccessKey,
	storageSecreetKey,
	storageDriver,
	storageRegion,
	storageBucket string,
	storageAPIPort int16,
	maxFileSize int64,
) StorageService {
	return &HorizonStorage{
		storageHost: storageHost,

		storageAccessKey:  storageAccessKey,
		storageSecreetKey: storageSecreetKey,
		storageDriver:     storageDriver,
		storageRegion:     storageRegion,
		storageBucket:     storageBucket,
		storageAPIPort:    storageAPIPort,
		maxFileSize:       maxFileSize,
		Storage:           nil,
	}
}

// Run implements StorageService.
func (h HorizonStorage) Run(ctx context.Context) error {
	client, err := minio.New(fmt.Sprintf("%s:%d", h.storageHost, h.storageAPIPort), &minio.Options{
		Creds:  credentials.NewStaticV4(h.storageAccessKey, h.storageSecreetKey, ""),
		Secure: false,
		Region: h.storageRegion,
		BucketLookup: func() minio.BucketLookupType {
			if h.storageHost == "s3" {
				return minio.BucketLookupDNS
			}
			return minio.BucketLookupPath
		}(),
	})
	if err != nil {
		return eris.Wrap(err, "failed to initialize MinIO client")
	}

	h.Storage = client
	exists, err := client.BucketExists(ctx, h.storageBucket)
	if err != nil {
		return eris.Wrap(err, "failed to check bucket exists")
	}
	if !exists {
		err = client.MakeBucket(ctx, h.storageBucket, minio.MakeBucketOptions{Region: h.storageRegion})
		if err != nil {
			return eris.Wrapf(err, "failed to create bucket %s", h.storageBucket)
		}
	}
	return nil
}

// DeleteFile implements StorageService.
func (h HorizonStorage) DeleteFile(ctx context.Context, Storage *Storage) error {
	if h.Storage == nil {
		return eris.New("not initialized")
	}
	if strings.TrimSpace(Storage.StorageKey) == "" {
		return eris.New("empty key")
	}
	err := h.Storage.RemoveObject(ctx, Storage.BucketName, Storage.StorageKey, minio.RemoveObjectOptions{})
	if err != nil {
		return eris.Wrapf(err, "delete %s failed", Storage.FileName)
	}
	return nil
}

// GeneratePresignedURL implements StorageService.
func (h HorizonStorage) GeneratePresignedURL(ctx context.Context, Storage *Storage, expiry time.Duration) (string, error) {
	if h.Storage == nil {
		return "", eris.New("not initialized")
	}
	if Storage.StorageKey == "" {
		return "", eris.New("Storage key do not exists")
	}
	u, err := h.Storage.PresignedGetObject(ctx, Storage.BucketName, Storage.StorageKey, 24*time.Hour, nil)
	if err != nil {
		return "", eris.Wrap(err, "presign failed")
	}
	return u.String(), nil
}

// GenerateUniqueName implements StorageService.
func (h HorizonStorage) GenerateUniqueName(ctx context.Context, name string) (string, error) {
	if h.Storage == nil {
		return "", eris.New("not initialized")
	}
	if name == "" {
		return "", eris.New("Please include name")
	}
	return fmt.Sprintf("%s-%d-%s", time.Now().Format("20060102150405"), os.Getpid(), name), nil
}

func (h *HorizonStorage) Upload(
	ctx context.Context,
	input any,
	callback ProgressCallback,
) (*Storage, error) {
	switch v := input.(type) {
	case string:
		return h.UploadFromPath(ctx, v, callback)
	case []byte:
		return h.UploadFromBinary(ctx, v, callback)
	case *multipart.FileHeader:
		return h.UploadFromHeader(ctx, v, callback)
	default:
		return nil, eris.Errorf("unsupported upload type %T", input)
	}
}

// UploadFromBinary implements StorageService.
func (h HorizonStorage) UploadFromBinary(ctx context.Context, data []byte, callback ProgressCallback) (*Storage, error) {
	if h.Storage == nil {
		return nil, eris.New("Storage not initialized")
	}

	// Prepare reader and size
	reader := bytes.NewReader(data)
	size := int64(len(data))

	// Predict content type
	buf := make([]byte, 512)
	n, _ := reader.Read(buf)
	contentType := http.DetectContentType(buf[:n])
	reader.Seek(0, io.SeekStart)

	fileName, err := h.GenerateUniqueName(ctx, "file")
	if err != nil {
		return nil, err
	}
	Storage := &Storage{
		FileName: fileName,
		FileSize: size,
		FileType: contentType,
		Status:   "progress",
	}

	pr := &progressReader{
		reader:   reader,
		callback: callback,
		total:    size,
		Storage:  Storage,
	}

	uploadInfo, err := h.Storage.PutObject(
		ctx,
		h.storageBucket,
		fileName,
		pr,
		size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return nil, eris.Wrap(err, "failed to upload binary data")
	}

	// Update record
	Storage.StorageKey = uploadInfo.Key
	Storage.BucketName = h.storageBucket

	// Generate URL
	url, err := h.GeneratePresignedURL(ctx, Storage, 5*time.Minute)
	if err != nil {
		return nil, eris.Wrap(err, "failed to generate presigned URL")
	}

	Storage.URL = url
	Storage.Status = "completed"
	Storage.Progress = uploadInfo.Size

	return Storage, nil

}

// UploadFromHeader implements StorageService.
func (h HorizonStorage) UploadFromHeader(ctx context.Context, file *multipart.FileHeader, callback ProgressCallback) (*Storage, error) {
	// Ensure Storage client is initialized
	if h.Storage == nil {
		return nil, eris.New("Storage not initialized")
	}
	src, err := file.Open()
	if err != nil {
		return nil, eris.Wrapf(err, "open %s failed", file.Filename)
	}
	defer src.Close()
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Generate unique file name
	fileName, err := h.GenerateUniqueName(ctx, file.Filename)
	if err != nil {
		return nil, eris.Wrap(err, "failed to generate unique name")
	}
	Storage := &Storage{
		FileName: fileName,
		FileSize: file.Size,
		FileType: contentType,
		Status:   "progress",
	}

	pr := &progressReader{reader: src, callback: callback, total: file.Size, Storage: Storage}
	info, err := h.Storage.PutObject(ctx, h.storageBucket, fileName, pr, file.Size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return nil, eris.Wrapf(err, "upload %s failed", file.Filename)
	}

	// Update Storage record
	Storage.StorageKey = info.Key
	Storage.BucketName = h.storageBucket

	// Generate presigned URL
	url, err := h.GeneratePresignedURL(ctx, Storage, 5*time.Minute)
	if err != nil {
		return nil, eris.Wrap(err, "failed to generate presigned URL")
	}

	Storage.URL = url
	Storage.Status = "completed"
	Storage.Progress = info.Size
	return Storage, nil
}

// UploadFromPath implements StorageService.
func (h *HorizonStorage) UploadFromPath(ctx context.Context, path string, callback ProgressCallback,
) (*Storage, error) {
	// Ensure Storage client is initialized
	if h.Storage == nil {
		return nil, eris.New("Storage not initialized")
	}

	// Validate and clean path
	cleanPath := strings.TrimSpace(path)
	if err := isValidFilePath(cleanPath); err != nil {
		return nil, err
	}

	// Open file
	file, err := os.Open(cleanPath)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to open file: %s", cleanPath)
	}
	defer file.Close()

	// Get file info
	info, err := file.Stat()
	if err != nil {
		return nil, eris.Wrapf(err, "failed to stat file: %s", cleanPath)
	}

	// Detect content type
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	contentType := http.DetectContentType(buf[:n])
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, eris.Wrap(err, "failed to reset file reader")
	}

	// Generate unique file name
	fileName, err := h.GenerateUniqueName(ctx, filepath.Base(cleanPath))
	if err != nil {
		return nil, eris.Wrap(err, "failed to generate unique name")
	}

	// Prepare Storage record
	Storage := &Storage{
		FileName: fileName,
		FileSize: info.Size(),
		FileType: contentType,
		Status:   "progress",
	}

	// Wrap reader for progress tracking
	reader := &progressReader{
		reader:   file,
		callback: callback,
		total:    info.Size(),
		Storage:  Storage,
	}

	// Upload to MinIO
	uploadInfo, err := h.Storage.PutObject(
		ctx,
		h.storageBucket,
		fileName,
		reader,
		info.Size(),
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return nil, eris.Wrap(err, "failed to upload file")
	}

	// Update Storage record
	Storage.StorageKey = uploadInfo.Key
	Storage.BucketName = h.storageBucket

	// Generate presigned URL
	url, err := h.GeneratePresignedURL(ctx, Storage, 5*time.Minute)
	if err != nil {
		return nil, eris.Wrap(err, "failed to generate presigned URL")
	}

	Storage.URL = url
	Storage.Status = "completed"
	Storage.Progress = uploadInfo.Size

	return Storage, nil
}
