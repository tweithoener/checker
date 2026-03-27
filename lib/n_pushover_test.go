package lib

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"testing"

	po "github.com/gregdel/pushover"
	chkr "github.com/tweithoener/checker"
)

func TestPushover(t *testing.T) {
	orig := sendPushoverMessage
	defer func() { sendPushoverMessage = orig }()

	tests := []struct {
		name             string
		state            chkr.State
		expectedPriority int
		expectedSound    string
		mockErr          error
		expectLogMsg     string
	}{
		{"OK", chkr.OK, po.PriorityNormal, po.SoundIncoming, nil, ""},
		{"Warn", chkr.Warn, po.PriorityHigh, po.SoundBike, nil, ""},
		{"Fail", chkr.Fail, po.PriorityHigh, po.SoundCosmic, nil, ""},
		{"Error", chkr.OK, po.PriorityNormal, po.SoundIncoming, errors.New("api unavailable"), "api unavailable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock log output to capture errors
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(os.Stderr)

			called := false
			sendPushoverMessage = func(app *po.Pushover, message *po.Message, recipient *po.Recipient) (*po.Response, error) {
				called = true
				
				// Validate properties
				if message.Priority != tt.expectedPriority {
					t.Errorf("Expected priority %d, got %d", tt.expectedPriority, message.Priority)
				}
				if message.Sound != tt.expectedSound {
					t.Errorf("Expected sound %s, got %s", tt.expectedSound, message.Sound)
				}
				
				// Validate title format
				expectedTitle := string(tt.state) + " my-check"
				if message.Title != expectedTitle {
					t.Errorf("Expected title '%s', got '%s'", expectedTitle, message.Title)
				}
				
				// Validate message format
				if !strings.HasPrefix(message.Message, "PREFIX-") {
					t.Errorf("Expected message to start with PREFIX-, got %s", message.Message)
				}

				if tt.mockErr != nil {
					return nil, tt.mockErr
				}
				return &po.Response{}, nil
			}

			not := Pushover("PREFIX-", "app-token", "user-token")
			
			cs := chkr.CheckState{
				Name:  "my-check",
				State: tt.state,
			}
			not(context.Background(), "my-check", cs)

			if !called {
				t.Error("sendPushoverMessage was not called")
			}

			// Check log output for errors
			output := buf.String()
			if tt.expectLogMsg != "" {
				if !strings.Contains(output, tt.expectLogMsg) {
					t.Errorf("Expected log output to contain '%s', got '%s'", tt.expectLogMsg, output)
				}
			} else {
				if output != "" {
					t.Errorf("Expected no log output, got '%s'", output)
				}
			}
		})
	}
}
