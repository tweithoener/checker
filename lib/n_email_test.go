package lib

import (
	"bytes"
	"context"
	"log"
	"os"
	"strings"
	"testing"

	chkr "github.com/tweithoener/checker"
	"github.com/wneessen/go-mail"
)

func TestEmail(t *testing.T) {
	orig := sendEmailMsg
	defer func() { sendEmailMsg = orig }()

	hostname, _ := os.Hostname()
	shortHostname := strings.Split(hostname, ".")[0]

	tests := []struct {
		name         string
		smtpServer   string
		user         string
		password     string
		to           []string
		opts         []EmailOption
		checkName    string
		checkState   chkr.CheckState
		expectedTo   []string
		expectedFrom string
		expectedSubj string
		expectedBody string
		mockErr      error
		expectLogMsg string
	}{
		{
			name:       "Default Template and From",
			smtpServer: "smtp.example.com:587",
			user:       "user",
			password:   "pass",
			to:         []string{"admin@example.com"},
			checkName:  "test-check",
			checkState: chkr.CheckState{Name: "test-check", State: chkr.Fail, Message: "failed badly"},
			expectedTo: []string{"admin@example.com"},
			expectedFrom: "user@" + shortHostname,
			expectedSubj: "[Checker] Failed test-check",
			expectedBody: "Message: failed badly",
		},
		{
			name:       "Custom Template, From and To",
			smtpServer: "smtp.example.com:587",
			user:       "user@example.com",
			password:   "pass",
			to:         []string{"admin@example.com", "dev@example.com"},
			opts: []EmailOption{
				WithFrom("alerts@example.com"),
				WithTemplate("CRITICAL: {{.Name}} is {{.State}}"),
			},
			checkName:  "db-check",
			checkState: chkr.CheckState{Name: "db-check", State: chkr.Fail},
			expectedTo: []string{"admin@example.com", "dev@example.com"},
			expectedFrom: "alerts@example.com",
			expectedSubj: "[Checker] Failed db-check",
			expectedBody: "CRITICAL: db-check is Failed",
		},
		{
			name:       "SMTP Error",
			smtpServer: "smtp.example.com:587",
			user:       "user@example.com",
			password:   "pass",
			to:         []string{"user@example.com"},
			checkName:  "test-check",
			checkState: chkr.CheckState{Name: "test-check", State: chkr.Fail},
			expectedTo: []string{"user@example.com"},
			expectedFrom: "user@example.com",
			expectedSubj: "[Checker] Failed test-check",
			expectedBody: "Check test-check is in state Failed",
			mockErr:      context.DeadlineExceeded,
			expectLogMsg: "Can't send email notification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(os.Stderr)

			called := false
			sendEmailMsg = func(smtpServer string, user string, password string, msg *mail.Msg) error {
				called = true
				if smtpServer != tt.smtpServer {
					t.Errorf("Expected addr %s, got %s", tt.smtpServer, smtpServer)
				}
				
				var msgBuf bytes.Buffer
				_, err := msg.WriteTo(&msgBuf)
				if err != nil {
					t.Fatalf("Failed to write message to buffer: %v", err)
				}
				msgStr := msgBuf.String()

				if !strings.Contains(msgStr, "From: <"+tt.expectedFrom+">") {
					t.Errorf("Expected from %s, got message:\n%s", tt.expectedFrom, msgStr)
				}
				
				for _, expectedTo := range tt.expectedTo {
					if !strings.Contains(msgStr, "<"+expectedTo+">") && !strings.Contains(msgStr, expectedTo) {
						t.Errorf("Expected To %s, got message:\n%s", expectedTo, msgStr)
					}
				}

				if !strings.Contains(msgStr, "Subject: "+tt.expectedSubj) {
					t.Errorf("Expected message to contain subject '%s', got '%s'", tt.expectedSubj, msgStr)
				}
				if !strings.Contains(msgStr, tt.expectedBody) {
					t.Errorf("Expected message to contain body '%s', got '%s'", tt.expectedBody, msgStr)
				}

				return tt.mockErr
			}

			not := Email(tt.smtpServer, tt.user, tt.password, tt.to, tt.opts...)
			not(context.Background(), tt.checkState)

			if !called {
				t.Error("sendEmailMsg was not called")
			}

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
