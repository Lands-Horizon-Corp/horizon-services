package horizon_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lands-horizon/horizon-server/horizon"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RUN TEST: go test ./horizon_test/horizon.storage_test.go
func setupStorageService(t *testing.T) *horizon.HorizonStorage {
	env := horizon.NewEnvironmentService("../.env")

	svc := horizon.NewHorizonStorageService(
		env.GetString("STORAGE_HOST", "localhost"),
		env.GetString("STORAGE_ACCESS_KEY", "minioadmin"),
		env.GetString("STORAGE_SECRET_KEY", "minioadmin"),
		env.GetString("STORAGE_DRIVER", "minio"),
		env.GetString("STORAGE_REGION", "us-east-1"),
		env.GetString("STORAGE_BUCKET", "my-bucket"),
		env.GetInt16("STORAGE_API_PORT", 9000),
		env.GetInt64("STORAGE_MAX_SIZE", 10*1024*1024), // 10MB fallback
	).(*horizon.HorizonStorage)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Initialize service and bucket
	err := svc.Run(ctx)
	require.NoError(t, err, "Failed to initialize storage service")
	// Debug: Check if Storage is initialized
	if svc.Storage == nil {
		t.Fatal("Storage client is nil after Run()")
	}
	cleanupBucket(t, svc)
	return svc
}

func cleanupBucket(t *testing.T, svc *horizon.HorizonStorage) {
	env := horizon.NewEnvironmentService("../.env")

	ctx := context.Background()
	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)
		for obj := range svc.Storage.ListObjects(ctx, env.GetString("STORAGE_BUCKET", "my-bucket"), minio.ListObjectsOptions{Recursive: true}) {
			if obj.Err != nil {
				t.Logf("Error listing objects: %v", obj.Err)
				return
			}
			objectsCh <- obj
		}
	}()

	for obj := range objectsCh {
		err := svc.Storage.RemoveObject(ctx, env.GetString("STORAGE_BUCKET", "my-bucket"), obj.Key, minio.RemoveObjectOptions{})
		require.NoError(t, err, "Failed to delete test object")
	}
}

// Test 1: Successful Initialization
func TestRun(t *testing.T) {
	env := horizon.NewEnvironmentService("../.env")
	svc := setupStorageService(t)
	fmt.Println(env.GetString("STORAGE_HOST", "localhost"))
	// fmt.Println(svc)
	fmt.Println("--")
	assert.NotNil(t, svc.Storage, "MinIO client not initialized")
}

// Test 2: Upload from Local Path
func TestUploadFromPath(t *testing.T) {
	svc := setupStorageService(t)
	ctx := context.Background()

	// Create temp file
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := []byte("test content")
	_, err = tmpFile.Write(content)
	require.NoError(t, err)
	tmpFile.Close()

	// Test upload
	storage, err := svc.UploadFromPath(ctx, tmpFile.Name(), nil)
	require.NoError(t, err)

	// Validate metadata
	assert.Equal(t, filepath.Base(tmpFile.Name()), storage.FileName)
	assert.Equal(t, int64(len(content)), storage.FileSize)
	assert.Equal(t, "text/plain; charset=utf-8", storage.FileType)
	assert.NotEmpty(t, storage.URL)
}

// Test 3: Upload Binary Data
func TestUploadFromBinary(t *testing.T) {
	env := horizon.NewEnvironmentService("../.env")
	svc := setupStorageService(t)
	ctx := context.Background()

	content := []byte("binary data")
	storage, err := svc.UploadFromBinary(ctx, content, nil)
	require.NoError(t, err)

	// Verify stored content
	reader, err := svc.Storage.GetObject(ctx, env.GetString("STORAGE_BUCKET", "my-bucket"), storage.StorageKey, minio.GetObjectOptions{})
	require.NoError(t, err)
	defer reader.Close()

	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, content, data)
}

// Test 4: Upload Multipart File
func TestUploadFromHeader(t *testing.T) {
	svc := setupStorageService(t)
	ctx := context.Background()

	// Simulate multipart file
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.txt")
	require.NoError(t, err)
	_, err = part.Write([]byte("multipart content"))
	require.NoError(t, err)
	writer.Close()

	// Parse multipart request
	req, _ := http.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	form, err := multipart.NewReader(body, writer.Boundary()).ReadForm(32 << 20)
	require.NoError(t, err)
	fileHeader := form.File["file"][0]

	// Test upload
	storage, err := svc.UploadFromHeader(ctx, fileHeader, nil)
	require.NoError(t, err)
	assert.Equal(t, "test.txt", storage.FileName)
}

// Test 5: Generate Presigned URL
func TestGeneratePresignedURL(t *testing.T) {
	svc := setupStorageService(t)
	ctx := context.Background()

	// Upload test file
	storage, err := svc.UploadFromBinary(ctx, []byte("presigned content"), nil)
	require.NoError(t, err)

	// Generate URL
	url, err := svc.GeneratePresignedURL(ctx, storage, time.Hour)
	require.NoError(t, err)

	// Verify URL accessibility
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Test 6: Delete File
func TestDeleteFile(t *testing.T) {
	env := horizon.NewEnvironmentService("../.env")
	svc := setupStorageService(t)
	ctx := context.Background()

	storage, err := svc.UploadFromBinary(ctx, []byte("delete me"), nil)
	require.NoError(t, err)

	// Delete file
	err = svc.DeleteFile(ctx, storage)
	require.NoError(t, err)

	// Verify deletion
	_, err = svc.Storage.StatObject(ctx, env.GetString("STORAGE_BUCKET", "my-bucket"), storage.StorageKey, minio.StatObjectOptions{})
	assert.Error(t, err)
}

// Test 7: Progress Callback
func TestProgressCallback(t *testing.T) {
	svc := setupStorageService(t)
	ctx := context.Background()

	var progressUpdates []int64
	callback := func(progress int64, _ int64, _ *horizon.Storage) {
		progressUpdates = append(progressUpdates, progress)
	}

	// Upload large enough data to trigger multiple progress updates
	data := bytes.Repeat([]byte{0}, 5*1024*1024) // 5MB
	_, err := svc.UploadFromBinary(ctx, data, callback)
	require.NoError(t, err)

	assert.NotEmpty(t, progressUpdates, "No progress updates received")
	assert.Equal(t, int64(100), progressUpdates[len(progressUpdates)-1], "Final progress not 100%")
}

// Test 8: Invalid File Path
func TestUploadInvalidPath(t *testing.T) {
	svc := setupStorageService(t)
	ctx := context.Background()

	_, err := svc.UploadFromPath(ctx, "/invalid/path/to/file.txt", nil)
	assert.Error(t, err)
}

// Test 9: Unique Filename Generation
func TestGenerateUniqueName(t *testing.T) {
	svc := setupStorageService(t)
	ctx := context.Background()

	name1, err := svc.GenerateUniqueName(ctx, "file.txt")
	require.NoError(t, err)
	name2, err := svc.GenerateUniqueName(ctx, "file.txt")
	require.NoError(t, err)

	assert.NotEqual(t, name1, name2, "Generated names must be unique")
	assert.Contains(t, name1, "file.txt", "Original name not preserved")
}
