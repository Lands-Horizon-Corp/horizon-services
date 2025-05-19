package horizon

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"sync"

	"github.com/microcosm-cc/bluemonday"
	"github.com/rotisserie/eris"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	"golang.org/x/time/rate"
)

/*

// Replace with your actual Twilio credentials
accountSID := "ACXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
authToken := "your_twilio_auth_token"
sender := "+1234567890" // Must be a Twilio-verified number

// Initialize the SMS service with a 160-character limit
smsService := horizon.NewHorizonSMS[horizon.AlertVars](accountSID, authToken, sender, 160)

ctx := context.Background()

// Start the service
if err := smsService.Run(ctx); err != nil {
	log.Fatalf("failed to start SMS service: %v", err)
}
defer smsService.Stop(ctx)

// Define the message with a template
message := horizon.SMSRequest[horizon.AlertVars]{
	To: "+19876543210", // Recipient number
	Body: "Hello {{.Name}},\nAlert Code: {{.Code}}\nDetails: {{.Message}}", // Templated body
	Vars: horizon.AlertVars{
		Name:    "John Doe",
		Code:    "ALRT-001",
		Message: "Suspicious activity detected on your account.",
	},
}

// Send the SMS
err := smsService.Send(ctx, message)
if err != nil {
	log.Fatalf("failed to send SMS: %v", err)
}

fmt.Println("SMS successfully sent.")
*/

// SMSRequest represents a templated SMS message with dynamic variables
type SMSRequest[T any] struct {
	To   string // Recipient phone number
	Body string // Message template body
	Vars T      // Dynamic variables for template interpolation
}

// SMSService defines the interface for SMS processing and delivery
type SMSService[T any] interface {
	// Start initializes the SMS gateway with rate limiting
	Run(ctx context.Context) error

	// Stop shuts down the SMS gateway connection
	Stop(ctx context.Context) error

	// FormatSMS processes template and injects variables
	Format(ctx context.Context, req SMSRequest[T]) (*SMSRequest[T], error)

	// SendSMS dispatches the formatted message to the recipient
	Send(ctx context.Context, req SMSRequest[T]) error
}

type HorizonSMS[T any] struct {
	limiterOnce sync.Once
	limiter     *rate.Limiter
	twilio      *twilio.RestClient

	accountSID    string // Twilio Account SID
	authToken     string // Twilio Auth Token
	sender        string // Sender phone number registered with Twilio
	maxCharacters int32  // Maximum allowed length for SMS body
}

func NewHorizonSMS[T any](accountSID, authToken, sender string, maxCharacters int32) SMSService[T] {
	return &HorizonSMS[T]{
		accountSID:    accountSID,
		authToken:     authToken,
		sender:        sender,
		maxCharacters: maxCharacters,
	}
}

// Run implements SMSService.
func (h *HorizonSMS[T]) Run(ctx context.Context) error {
	h.limiterOnce.Do(func() {
		h.limiter = rate.NewLimiter(1, 3)
	})
	h.twilio = twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: h.accountSID,
		Password: h.authToken,
	})
	return nil
}

// Stop implements SMSService.
func (h *HorizonSMS[T]) Stop(ctx context.Context) error {
	h.twilio = nil
	h.limiter = nil
	return nil
}

// FormatSMS implements SMSService.
func (h *HorizonSMS[T]) Format(ctx context.Context, req SMSRequest[T]) (*SMSRequest[T], error) {
	tmpl, err := template.New("email").Parse(req.Body)
	if err != nil {
		return nil, eris.Wrap(err, "parse template failed")
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, req.Vars); err != nil {
		return nil, eris.Wrap(err, "execute template failed")
	}
	req.Body = buf.String()
	return &req, nil
}

// Send formats and sends the SMS using Twilio
func (h *HorizonSMS[T]) Send(ctx context.Context, req SMSRequest[T]) error {
	// Validate recipient phone number
	if !IsValidPhoneNumber(req.To) {
		return fmt.Errorf("invalid recipient phone number format: %s", req.To)
	}

	// Validate sender phone number
	if !IsValidPhoneNumber(h.sender) {
		return fmt.Errorf("invalid admin phone number format: %s", h.sender)
	}

	// Ensure SMS body length is within limit
	if len(req.Body) > int(h.maxCharacters) {
		return fmt.Errorf("SMS body exceeds %d characters (actual: %d)", h.maxCharacters, len(req.Body))
	}

	// Apply rate limit to prevent sending too fast
	if err := h.limiter.Wait(ctx); err != nil {
		return eris.Wrap(err, "rate limit wait failed")
	}

	// Ensure the request is still allowed after waiting
	if !h.limiter.Allow() {
		return fmt.Errorf("rate limit exceeded for sending SMS")
	}

	// Sanitize message body to remove harmful content
	req.Body = bluemonday.UGCPolicy().Sanitize(req.Body)

	// Re-inject template variables after sanitizing
	finalBody, err := h.Format(ctx, req)
	if err != nil {
		return eris.Wrap(err, "failed to inject variables into body")
	}
	req.Body = finalBody.Body

	// Build Twilio message parameters
	params := &openapi.CreateMessageParams{}
	params.SetTo(req.To)
	params.SetFrom(h.sender)
	params.SetBody(req.Body)

	// Send the message using Twilio's REST API
	_, err = h.twilio.Api.CreateMessage(params)
	if err != nil {
		return eris.Wrap(err, "failed to send SMS via Twilio")
	}
	return nil
}
