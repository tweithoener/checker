package lib

import (
	"testing"

	chkr "github.com/tweithoener/checker"
)

func TestCheckMakers(t *testing.T) {
	tests := []struct {
		maker chkr.CheckMaker
		json  string
	}{
		{cmdMkr, `{"Command": "ls", "Args": ["-l"]}`},
		{cpuMkr, `{"WarnPercent": 80, "FailPercent": 90}`},
		{diskMkr, `{"Path": "/", "WarnPercent": 80, "FailPercent": 90}`},
		{dnsMkr, `{"Dns": "8.8.8.8", "Hostname": "example.com", "Address": "1.2.3.4"}`},
		{failMkr, `{"Check": {"Maker": "Cpu", "Name": "mock-cpu", "Args": {"WarnPercent": 80}}}`},
		{httpMkr, `{"Method": "GET", "Url": "http://example.com", "Expected": 200}`},
		{loadMkr, `{"WarnLoad5": 2.0, "FailLoad5": 4.0}`},
		{memMkr, `{"WarnPercent": 80, "FailPercent": 90}`},
		{pingMkr, `{"Address": "1.2.3.4", "WarnMillis": 100, "FailMillis": 200}`},
		{procExistsMkr, `{"Name": "nginx"}`},
		{proxyMkr, `{"Method": "GET", "Request": "http://example.com", "Proxy": "http://proxy:8080", "Expected": 200}`},
		{sshMkr, `{"Host": "example.com", "User": "admin", "Command": "uptime"}`},
		{swapMkr, `{"WarnPercent": 80, "FailPercent": 90}`},
		{sysProcsMkr, `{"WarnCount": 100, "FailCount": 200}`},
		{uptimeMkr, `{"MinMinutes": 10}`},
	}

	for _, tt := range tests {
		t.Run(tt.maker.Maker(), func(t *testing.T) {
			// 1. Test UnmarshalArgs
			args, err := tt.maker.UnmarshalArgs([]byte(tt.json))
			if err != nil {
				t.Fatalf("UnmarshalArgs failed: %v", err)
			}

			// 2. Test FromConfig with valid args
			chk, err := tt.maker.FromConfig(chkr.CheckConfig{Args: args})
			if err != nil {
				t.Fatalf("FromConfig failed: %v", err)
			}
			if chk == nil {
				t.Fatal("FromConfig returned nil Check")
			}

			// 3. Test FromConfig with invalid args
			_, err = tt.maker.FromConfig(chkr.CheckConfig{Args: struct{}{}})
			if err == nil {
				t.Error("FromConfig should fail when given wrong argument types")
			}
		})
	}
}

func TestNotifierMakers(t *testing.T) {
	tests := []struct {
		maker chkr.NotifierMaker
		json  string
	}{
		{loggingMkr, `{"Prefix": "LOG: "}`},
		{debugMkr, `{"Prefix": "DEBUG: "}`},
		{pushoverMkr, `{"Prefix": "ERR: ", "App": "app-token", "Recipient": "user-token"}`},
		{emailMkr, `{"SmtpServer": "smtp.example.com:587", "User": "user", "Password": "pwd", "To": ["admin@example.com"], "Template": "Body"}`},
		{lessMkr, `{"Notifier": {"Maker": "Logging", "Args": {"Prefix": "LOG: "}}}`},
	}

	for _, tt := range tests {
		t.Run(tt.maker.Maker(), func(t *testing.T) {
			// 1. Test UnmarshalArgs
			args, err := tt.maker.UnmarshalArgs([]byte(tt.json))
			if err != nil {
				t.Fatalf("UnmarshalArgs failed: %v", err)
			}

			// 2. Test FromConfig with valid args
			not, err := tt.maker.FromConfig(chkr.NotifierConfig{Args: args})
			if err != nil {
				t.Fatalf("FromConfig failed: %v", err)
			}
			if not == nil {
				t.Fatal("FromConfig returned nil Notifier")
			}

			// 3. Test FromConfig with invalid args
			_, err = tt.maker.FromConfig(chkr.NotifierConfig{Args: struct{}{}})
			if err == nil {
				t.Error("FromConfig should fail when given wrong argument types")
			}
		})
	}
}
