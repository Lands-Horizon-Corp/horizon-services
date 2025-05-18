package horizon

import "context"

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
