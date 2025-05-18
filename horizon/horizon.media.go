package horizon

import (
	"context"
	"time"
)

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
