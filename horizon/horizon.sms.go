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

	accountSID    string
	authToken     string
	sender        string
	maxCharacters int32
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

// SendSMS implements SMSService.
func (h *HorizonSMS[T]) Send(ctx context.Context, req SMSRequest[T]) error {
	if !IsValidPhoneNumber(req.To) {
		err := fmt.Errorf("invalid recipient phone number format: %s", req.To)
		return err
	}
	if !IsValidPhoneNumber(h.sender) {
		err := fmt.Errorf("invalid admin phone number format: %s", req.To)
		return err
	}
	if len(req.Body) > int(h.maxCharacters) {
		err := fmt.Errorf("SMS body exceeds %d characters (actual: %d)", h.maxCharacters, len(req.Body))
		return err
	}

	// Apply rate limiting to control send rate
	if err := h.limiter.Wait(ctx); err != nil {
		return eris.Wrap(err, "rate limit wait failed")
	}

	// Check if rate limiter exceed
	if !h.limiter.Allow() {
		err := fmt.Errorf("rate limit exceeded for sending SMS")
		return err
	}

	// Sanitize the raw body template for safety
	req.Body = bluemonday.UGCPolicy().Sanitize(req.Body)

	// Inject dynamic variables into the sanitized body
	finalBody, err := h.Format(ctx, req)
	if err != nil {
		return eris.Wrap(err, "failed to inject variables into body")
	}
	req.Body = finalBody.Body

	params := &openapi.CreateMessageParams{}
	params.SetTo(req.To)
	params.SetFrom(h.sender)

	_, err = h.twilio.Api.CreateMessage(params)
	if err != nil {
		return eris.Wrap(err, "failed to send SMS via Twilio")
	}
	return nil
}
