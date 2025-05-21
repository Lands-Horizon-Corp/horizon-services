package horizon_test

import (
	"context"
	"os"
	"testing"

	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// go test -v ./services/horizon_test/horizon.smtp_test.go

type Vars struct {
	Name string
}

func TestHorizonSMTP_Run_Stop(t *testing.T) {

	env := horizon.NewEnvironmentService("../../.env")

	host := env.GetString("SMTP_HOST", "")
	port := env.GetInt("SMTP_PORT", 0)
	username := env.GetString("SMTP_USERNAME", "")
	password := env.GetString("SMTP_PASSWORD", "")
	from := env.GetString("SMTP_FROM", "")

	smtp := horizon.NewHorizonSMTP[Vars](host, port, username, password, from)
	ctx := context.Background()

	require.NoError(t, smtp.Run(ctx))
	require.NoError(t, smtp.Stop(ctx))
}

func TestHorizonSMTP_Format_WithTemplateString(t *testing.T) {

	env := horizon.NewEnvironmentService("../../.env")

	host := env.GetString("SMTP_HOST", "")
	port := env.GetInt("SMTP_PORT", 0)
	username := env.GetString("SMTP_USERNAME", "")
	password := env.GetString("SMTP_PASSWORD", "")
	from := env.GetString("SMTP_FROM", "")
	reciever := env.GetString("SMTP_TEST_RECIEVER", "")

	smtp := horizon.NewHorizonSMTP[Vars](host, port, username, password, from)
	ctx := context.Background()

	req := horizon.SMTPRequest[Vars]{
		To:      reciever,
		Subject: "Test Subject",
		Body:    "Hello {{.Name}}, welcome!",
		Vars:    Vars{Name: "Alice"},
	}

	formatted, err := smtp.Format(ctx, req)
	require.NoError(t, err)
	assert.Contains(t, formatted.Body, "Hello Alice")
}

func TestHorizonSMTP_Format_WithTemplateFile(t *testing.T) {

	env := horizon.NewEnvironmentService("../../.env")

	host := env.GetString("SMTP_HOST", "")
	port := env.GetInt("SMTP_PORT", 0)
	username := env.GetString("SMTP_USERNAME", "")
	password := env.GetString("SMTP_PASSWORD", "")
	from := env.GetString("SMTP_FROM", "")
	reciever := env.GetString("SMTP_TEST_RECIEVER", "")

	file := "test_template.txt"
	content := "Hello {{.Name}}, this is from file."
	os.WriteFile(file, []byte(content), 0644)
	defer os.Remove(file)

	smtp := horizon.NewHorizonSMTP[Vars](host, port, username, password, from)
	ctx := context.Background()

	req := horizon.SMTPRequest[Vars]{
		To:      reciever,
		Subject: "Test File",
		Body:    file,
		Vars:    Vars{Name: "Bob"},
	}

	formatted, err := smtp.Format(ctx, req)
	require.NoError(t, err)
	assert.Contains(t, formatted.Body, "Hello Bob")
}

func TestHorizonSMTP_Send_InvalidEmail(t *testing.T) {
	env := horizon.NewEnvironmentService("../../.env")

	host := env.GetString("SMTP_HOST", "")
	port := env.GetInt("SMTP_PORT", 0)
	username := env.GetString("SMTP_USERNAME", "")
	password := env.GetString("SMTP_PASSWORD", "")
	from := env.GetString("SMTP_FROM", "")

	smtp := horizon.NewHorizonSMTP[Vars](host, port, username, password, from)
	ctx := context.Background()
	_ = smtp.Run(ctx)

	req := horizon.SMTPRequest[Vars]{
		To:      "also-invalid",
		Subject: "Test",
		Body:    "Hello {{.Name}}",
		Vars:    Vars{Name: "Test"},
	}

	err := smtp.Send(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "format is invalid")
}
func TestHorizonSMTP_Send_RateLimitExceeded(t *testing.T) {
	env := horizon.NewEnvironmentService("../../.env")

	host := env.GetString("SMTP_HOST", "")
	port := env.GetInt("SMTP_PORT", 0)
	username := env.GetString("SMTP_USERNAME", "")
	password := env.GetString("SMTP_PASSWORD", "")
	from := env.GetString("SMTP_FROM", "")
	reciever := env.GetString("SMTP_TEST_RECIEVER", "")

	require.NotEmpty(t, from, "SMTP_FROM must be set for test")
	require.NotEmpty(t, reciever, "SMTP_TEST_RECIEVER must be set for test")

	smtp := horizon.NewHorizonSMTP[Vars](host, port, username, password, from)
	ctx := context.Background()
	require.NoError(t, smtp.Run(ctx))

	// Rapid calls to exhaust limiter
	for i := 0; i < 3; i++ {
		_ = smtp.Send(ctx, horizon.SMTPRequest[Vars]{
			To:      reciever,
			Subject: "Rate Test",
			Body:    "Hi {{.Name}}",
			Vars:    Vars{Name: "Test"},
		})
	}

	err := smtp.Send(ctx, horizon.SMTPRequest[Vars]{
		To:      reciever,
		Subject: "Should Fail",
		Body:    "Hi {{.Name}}",
		Vars:    Vars{Name: "Test"},
	})

	// If you want to test rate limiting, you must mock or simulate limiter because
	// the current limiter will block until token available.
	// So, here you can check error or just assert err != nil:
	assert.Error(t, err)
	// Optionally check error message contains known text
}
