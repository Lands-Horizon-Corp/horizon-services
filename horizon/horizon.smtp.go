package horizon

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"
	"sync"

	"github.com/microcosm-cc/bluemonday"
	"github.com/rotisserie/eris"
	"golang.org/x/time/rate"
)

// SMTPRequest represents a templated SMTP request with dynamic variables
type SMTPRequest[T any] struct {
	To      string // Recipient SMTP address
	Subject string // SMTP subject line
	Body    string // Template body with placeholders
	Vars    T      // Dynamic variables for template interpolation
}

type SMTPService[T any] interface {
	// Run initializes internal resources like rate limiter
	Run(ctx context.Context) error

	// Stop cleans up resources
	Stop(ctx context.Context) error

	// Format processes template and injects variables
	Format(ctx context.Context, req SMTPRequest[T]) (*SMTPRequest[T], error)

	// Send dispatches the formatted SMTP to the recipient
	Send(ctx context.Context, req SMTPRequest[T]) error
}

type HorizonSMTP[T any] struct {
	host     string
	port     int
	username string
	password string
	from     string

	limiterOnce sync.Once
	limiter     *rate.Limiter
}

// NewHorizonSMTP constructs a new HorizonSMTP client
func NewHorizonSMTP[T any](host string, port int, username, password string, from string) SMTPService[T] {
	return &HorizonSMTP[T]{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

// Run implements SMTPService.
func (h *HorizonSMTP[T]) Run(ctx context.Context) error {
	h.limiterOnce.Do(func() {
		h.limiter = rate.NewLimiter(rate.Limit(1), 3)
	})
	return nil
}

// Stop implements SMTPService.
func (h *HorizonSMTP[T]) Stop(ctx context.Context) error {
	h.limiter = nil
	return nil
}

// Format implements SMTPService.
func (h *HorizonSMTP[T]) Format(ctx context.Context, req SMTPRequest[T]) (*SMTPRequest[T], error) {
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

// Send implements SMTPService.
func (h *HorizonSMTP[T]) Send(ctx context.Context, req SMTPRequest[T]) error {
	// Validate the recipient email address format
	if !IsValidEmail(req.To) {
		return eris.New("Recipient email format is invalid")
	}
	// Validate the configured sender email address format

	if !IsValidEmail(h.from) {
		return eris.New("Admn email format is invalid")
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

	// Build SMTP authentication and message
	auth := smtp.PlainAuth("", h.username, h.password, h.host)
	addr := fmt.Sprintf("%s:%d", h.host, h.port)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", h.from, req.To, req.Subject, req.Body)

	// Send the email via SMTP
	if err := smtp.SendMail(addr, auth, h.from, []string{req.To}, []byte(msg)); err != nil {
		return eris.Wrap(err, "smtp send failed")
	}
	return nil
}
