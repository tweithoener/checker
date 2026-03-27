package lib

import (
	"context"
	"errors"
	"net"
	"testing"

	chkr "github.com/tweithoener/checker"
)

func TestDnsCheck(t *testing.T) {
	orig := dnsLookupHost
	defer func() { dnsLookupHost = orig }()

	tests := []struct {
		name        string
		dns         string
		hostname    string
		address     string
		mockResults []string
		mockErr     error
		expected    chkr.State
	}{
		{"OK", "8.8.8.8", "example.com", "1.2.3.4", []string{"1.2.3.4"}, nil, chkr.OK},
		{"MultipleResultsOK", "8.8.8.8", "example.com", "1.2.3.4", []string{"1.1.1.1", "1.2.3.4"}, nil, chkr.OK},
		{"WrongAddress", "8.8.8.8", "example.com", "1.2.3.4", []string{"1.1.1.1"}, nil, chkr.Fail},
		{"LookupError", "8.8.8.8", "example.com", "1.2.3.4", nil, errors.New("dns error"), chkr.Fail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dnsLookupHost = func(ctx context.Context, r *net.Resolver, hostname string) ([]string, error) {
				// We can't easily check the dial function from here without calling r.Dial,
				// but we can trust that the Dns() function set it up. 
				// In a more complex test, we could invoke r.Dial to see if it targets the right DNS server.
				
				if hostname != tt.hostname {
					t.Errorf("expected hostname %s, got %s", tt.hostname, hostname)
				}
				if tt.mockErr != nil {
					return nil, tt.mockErr
				}
				return tt.mockResults, nil
			}

			chk := Dns(tt.dns, tt.hostname, tt.address)
			s, _ := chk(context.Background(), chkr.CheckState{})
			if s != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, s)
			}
		})
	}
}
