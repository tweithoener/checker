package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	chkr "github.com/tweithoener/checker"
)

// DnsArgs defines the arguments for a DNS check.
type DnsArgs struct {
	Dns      string
	Hostname string
	Address  string
}

type dnsMaker struct{}

var dnsMkr = dnsMaker{}

func (dnsMaker) Maker() string {
	return "Dns"
}

func (dnsMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	args := DnsArgs{}
	if err := json.Unmarshal(j, &args); err != nil {
		return args, fmt.Errorf("can't unmarshal Dns arguments: %v", err)
	}
	return args, nil
}

func (dnsMaker) FromConfig(c chkr.CheckConfig) (chkr.Check, error) {
	args, ok := c.Args.(DnsArgs)
	if !ok {
		return nil, fmt.Errorf("configured arguments are not Dns arguments")
	}
	return Dns(args.Dns, args.Hostname, args.Address), nil
}

var dnsLookupHost = func(ctx context.Context, r *net.Resolver, hostname string) ([]string, error) {
	return r.LookupHost(ctx, hostname)
}

// Dns returns a check that verifies the resolution of a hostname to a specific address using a given DNS server.
func Dns(dns, hostname, address string) chkr.Check {
	return func(ctx context.Context, cs chkr.CheckState) (chkr.State, string) {
		r := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: 10 * time.Second,
				}
				return d.DialContext(ctx, network, dns+":53")
			},
		}
		ads, err := dnsLookupHost(ctx, r, hostname)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Host lookup failed: %v", err)
		}
		for _, ad := range ads {
			if ad == address {
				return chkr.OK, ""
			}
		}
		return chkr.Fail, fmt.Sprintf("Hostname %s does not resolve to address %s (resolves to %s)", hostname, address, strings.Join(ads, ", "))
	}
}
