package horizon

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"sync"

	"github.com/microcosm-cc/bluemonday"
	"github.com/rotisserie/eris"
	"golang.org/x/time/rate"
)

/*
func main() {
    // Initialize HorizonSMTP
    smtpService := NewHorizonSMTP("smtp.example.com", 587, "your_username", "your_password", "your_email@example.com")

    // Run the SMTP service (initialize resources like rate limiter)
    if err := smtpService.Run(context.Background()); err != nil {
        fmt.Println("Failed to run SMTP service:", err)
        return
    }

    // Example SMTP request
    emailRequest := SMTPRequest{
        To:      "recipient@example.com",
        Subject: "Test Email",
        Body:    "Hello {{.Name}},\n\nThis is a test email.",
        Vars: struct {
            Name string
        }{
            Name: "John Doe",
        },
    }

    // Format the email body with dynamic variables
    formattedRequest, err := smtpService.Format(context.Background(), emailRequest)
    if err != nil {
        fmt.Println("Failed to format email:", err)
        return
    }

    // Send the formatted email
    if err := smtpService.Send(context.Background(), *formattedRequest); err != nil {
        fmt.Println("Failed to send email:", err)
        return
    }

    fmt.Println("Email sent successfully!")
}

*/

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
		h.limiter = rate.NewLimiter(rate.Limit(10), 5) // 10 rps, burst 5
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
	var tmplBody string

	if err := isValidFilePath(req.Body); err == nil {
		content, err := os.ReadFile(req.Body)
		if err != nil {
			return nil, eris.Wrap(err, "failed to read template file")
		}
		tmplBody = string(content)
	} else {
		tmplBody = req.Body
	}

	tmpl, err := template.New("email").Parse(tmplBody)
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
	if !IsValidEmail(req.To) {
		return eris.New("Recipient email format is invalid")
	}
	if !IsValidEmail(h.from) {
		return eris.New("Admin email format is invalid")
	}

	// Wait for rate limiter token (blocking)
	if err := h.limiter.Wait(ctx); err != nil {
		return eris.Wrap(err, "rate limit wait failed")
	}

	// Sanitize and format body
	req.Body = bluemonday.UGCPolicy().Sanitize(req.Body)
	finalBody, err := h.Format(ctx, req)
	if err != nil {
		return eris.Wrap(err, "failed to inject variables into body")
	}
	req.Body = finalBody.Body

	auth := smtp.PlainAuth("", h.username, h.password, h.host)
	addr := fmt.Sprintf("%s:%d", h.host, h.port)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", h.from, req.To, req.Subject, req.Body)

	if err := smtp.SendMail(addr, auth, h.from, []string{req.To}, []byte(msg)); err != nil {
		return eris.Wrap(err, "smtp send failed")
	}
	return nil
}
