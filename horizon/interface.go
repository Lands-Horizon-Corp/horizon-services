package horizon

import (
	"context"
	"io"
	"mime/multipart"
	"time"

	"github.com/gocql/gocql"
	"github.com/labstack/echo/v4"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// Scheduler defines the interface for scheduling and managing cron jobs
type Scheduler interface {
	// Start initializes the scheduler and in-memory job store
	Start(ctx context.Context) error

	// Stop gracefully shuts down the scheduler and clears all timers
	Stop(ctx context.Context) error

	// CreateJob registers a new cron-style job with the specified schedule
	CreateJob(ctx context.Context, jobID string, schedule cron.Schedule, task func() error) error

	// RunJobNow immediately executes a scheduled job by ID
	RunJobNow(ctx context.Context, jobID string) error

	// RemoveJob deletes a job by its ID from the scheduler
	RemoveJob(ctx context.Context, jobID string) error

	// ListJobs returns all registered job IDs
	ListJobs(ctx context.Context) ([]string, error)
}

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

// Cache defines the interface for Redis operations
type Cache interface {
	// Start initializes the Redis connection pool
	Start(ctx context.Context) error

	// Stop gracefully shuts down all Redis connections
	Stop(ctx context.Context) error

	// Ping checks Redis server health
	Ping(ctx context.Context) error

	// Get retrieves a value by key from Redis
	Get(ctx context.Context, key string) (string, error)

	// Set stores a value with TTL expiration
	Set(ctx context.Context, key string, value any, ttl time.Duration) error

	// Exists checks if a key exists in the cache
	Exists(ctx context.Context, key string) (bool, error)

	// Delete removes a key from the cache
	Delete(ctx context.Context, key string) error
}

// EmailRequest represents a templated email request with dynamic variables
type EmailRequest[T any] struct {
	To      string // Recipient email address
	Subject string // Email subject line
	Body    string // Template body with placeholders
	Vars    T      // Dynamic variables for template interpolation
}

// EmailService defines the interface for email processing and delivery
type EmailService interface {
	// Start initializes the SMTP client with rate limiting
	Start(ctx context.Context, rateLimit int) error

	// Stop disables email sending and cleans up resources
	Stop(ctx context.Context) error

	// FormatEmail processes template and injects variables
	FormatEmail(ctx context.Context, req EmailRequest[any]) (EmailRequest[any], error)

	// SendEmail dispatches the formatted email to the recipient
	SendEmail(ctx context.Context, req EmailRequest[any]) error
}

// SMSRequest represents a templated SMS message with dynamic variables
type SMSRequest[T any] struct {
	To   string // Recipient phone number
	Body string // Message template body
	Vars T      // Dynamic variables for template interpolation
}

// SMSService defines the interface for SMS processing and delivery
type SMSService interface {
	// Start initializes the SMS gateway with rate limiting
	Start(ctx context.Context, rateLimit int) error

	// Stop shuts down the SMS gateway connection
	Stop(ctx context.Context) error

	// FormatSMS processes template and injects variables
	FormatSMS(ctx context.Context, req SMSRequest[any]) (SMSRequest[any], error)

	// SendSMS dispatches the formatted message to the recipient
	SendSMS(ctx context.Context, req SMSRequest[any]) error
}

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

// MediaType defines supported WebRTC media types
type MediaType string

const (
	Video MediaType = "video"
	Audio MediaType = "audio"
)

// MediaRequest contains WebRTC session negotiation details
type MediaRequest struct {
	ID        string    // Unique session identifier
	From      string    // Participant ID
	To        string    // Target participant ID (empty for broadcast)
	Type      MediaType // Media stream type
	SDP       string    // Session Description Protocol offer/answer
	Candidate string    // ICE candidate information
	Timestamp time.Time // Request timestamp
}

// MediaSession tracks active WebRTC sessions
type MediaSession struct {
	ID           string     // Session identifier
	Type         MediaType  // Session media type
	StartedAt    time.Time  // Session start time
	EndedAt      *time.Time // Session end time (nil if active)
	Participants []string   // List of participant IDs
}

// MediaService defines the interface for WebRTC communication
type MediaService interface {
	// Start initializes signaling servers and ICE agents
	Start(ctx context.Context) error

	// Stop terminates all active sessions and releases resources
	Stop(ctx context.Context) error

	// CreateSession initiates a new media session
	CreateSession(ctx context.Context, req MediaRequest) (*MediaSession, error)

	// JoinSession adds a participant to an existing session
	JoinSession(ctx context.Context, sessionID string, req MediaRequest) error

	// EndSession terminates a media session
	EndSession(ctx context.Context, sessionID string) error

	// SendSDP exchanges session description protocol information
	SendSDP(ctx context.Context, sessionID string, sdp string) error

	// SendCandidate exchanges ICE candidate information
	SendCandidate(ctx context.Context, sessionID string, candidate string) error

	// OnCandidate registers ICE candidate handler
	OnCandidate(ctx context.Context, handler func(string))

	// OnSessionEnded registers session termination handler
	OnSessionEnded(ctx context.Context, handler func(MediaSession))

	// ListSessions returns all active media sessions
	ListSessions(ctx context.Context) ([]MediaSession, error)
}

// SecurityUtils provides cryptographic and security-related functions
type SecurityUtils interface {
	// GenerateUUID creates a new UUIDv4
	GenerateUUID(ctx context.Context) (string, error)

	// HashPassword creates an Argon2 hashed password
	HashPassword(ctx context.Context, password string) (string, error)

	// VerifyPassword compares a password with its hash
	VerifyPassword(ctx context.Context, hash, password string) (bool, error)

	// SanitizeHTML removes potentially dangerous HTML content
	SanitizeHTML(ctx context.Context, input string) (string, error)

	// Encrypt performs AES encryption on plaintext
	Encrypt(ctx context.Context, plaintext, key string) (string, error)

	// Decrypt performs AES decryption on ciphertext
	Decrypt(ctx context.Context, ciphertext, key string) (string, error)
}

// OTPService manages one-time password generation and validation
type OTPService interface {
	// Generate creates a new OTP code for a key
	Generate(ctx context.Context, key string) (string, error)

	// Verify checks a code against the stored OTP
	Verify(ctx context.Context, key, code string) (bool, error)

	// Revoke invalidates an existing OTP code
	Revoke(ctx context.Context, key string) error
}

// TokenService manages JWT token lifecycle
type TokenService[T any] interface {
	// GetToken extracts and validates token from request context
	GetToken(ctx context.Context, c echo.Context) (T, error)

	// CleanToken removes token from response context
	CleanToken(ctx context.Context, c echo.Context)

	// VerifyToken validates a token string and returns claims
	VerifyToken(ctx context.Context, value string) T

	// SetToken creates and sets a new token in response context
	SetToken(ctx context.Context, c echo.Context, claim T) error

	// GenerateToken creates a new signed token with claims
	GenerateToken(ctx context.Context, claims T) (string, error)
}

// CSRFService manages Cross-Site Request Forgery protection
type CSRFService interface {
	// GenerateToken creates a new CSRF token for a session
	GenerateToken(ctx context.Context, sessionID string) (string, error)

	// VerifyToken validates a CSRF token against session ID
	VerifyToken(ctx context.Context, sessionID string, token string) (bool, error)

	// RevokeToken invalidates all CSRF tokens for a session
	RevokeToken(ctx context.Context, sessionID string) error

	// Middleware provides Echo middleware for CSRF protection
	Middleware() echo.MiddlewareFunc
}
