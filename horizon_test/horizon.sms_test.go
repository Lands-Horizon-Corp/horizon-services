package horizon_test

import (
	"context"
	"testing"

	"github.com/lands-horizon/horizon-server/horizon"
	"github.com/stretchr/testify/assert"
)

type AlertVars struct {
	Name    string
	Code    string
	Message string
}

// injectMockTwilio allows test control over the Twilio client
func injectMockTwilio(h *horizon.HorizonSMS[AlertVars]) {
	h.Run(context.Background())
}

func TestSendSMS(t *testing.T) {
	env := horizon.NewEnvironmentService("../.env")

	accountSID := env.GetString("TWILIO_ACCOUNT_SID", "")
	authToken := env.GetString("TWILIO_AUTH_TOKEN", "")
	sender := env.GetString("TWILIO_SENDER", "")
	reciever := env.GetString("TWILIO_TEST_RECIEVER", "")

	h := horizon.NewHorizonSMS[AlertVars](accountSID, authToken, sender, 160).(*horizon.HorizonSMS[AlertVars])
	injectMockTwilio(h)

	tests := []struct {
		name    string
		req     horizon.SMSRequest[AlertVars]
		wantErr bool
	}{
		{
			name: "valid request",
			req: horizon.SMSRequest[AlertVars]{
				To:   reciever,
				Body: "Hi {{.Name}}, alert {{.Code}}: {{.Message}}",
				Vars: AlertVars{Name: "Alice", Code: "A001", Message: "Test alert"},
			},
			wantErr: false,
		},
		{
			name: "invalid recipient number",
			req: horizon.SMSRequest[AlertVars]{
				To:   "invalid-number",
				Body: "Test",
				Vars: AlertVars{},
			},
			wantErr: true,
		},
		{
			name: "body too long",
			req: horizon.SMSRequest[AlertVars]{
				To:   reciever,
				Body: string(make([]byte, 200)), // Exceeds 160 chars
				Vars: AlertVars{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.Send(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err, "expected an error")
			} else {
				assert.NoError(t, err, "did not expect an error")
			}
		})
	}
}
