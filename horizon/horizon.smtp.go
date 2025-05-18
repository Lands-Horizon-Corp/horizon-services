package horizon

import "context"

// SMTPRequest represents a templated SMTP request with dynamic variables
type SMTPRequest[T any] struct {
	To      string // Recipient SMTP address
	Subject string // SMTP subject line
	Body    string // Template body with placeholders
	Vars    T      // Dynamic variables for template interpolation
}

// SMTPService defines the interface for SMTP processing and delivery
type SMTPService interface {
	// Start initializes the SMTP client with rate limiting
	Start(ctx context.Context, rateLimit int) error

	// Stop disables SMTP sending and cleans up resources
	Stop(ctx context.Context) error

	// FormatSMTP processes template and injects variables
	FormatSMTP(ctx context.Context, req SMTPRequest[any]) (SMTPRequest[any], error)

	// SendSMTP dispatches the formatted SMTP to the recipient
	SendSMTP(ctx context.Context, req SMTPRequest[any]) error
}
