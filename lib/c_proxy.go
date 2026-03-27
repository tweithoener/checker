package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	chkr "github.com/tweithoener/checker"
)

// ProxyArgs defines the arguments for a proxy connectivity check.
type ProxyArgs struct {
	Method   string
	Request  string
	Proxy    string
	Expected int
}

type proxyMaker struct{}

var proxyMkr = proxyMaker{}

func (proxyMaker) Maker() string {
	return "Proxy"
}

func (proxyMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	args := ProxyArgs{}
	if err := json.Unmarshal(j, &args); err != nil {
		return args, fmt.Errorf("can't unmarshal Proxy arguments: %v", err)
	}
	return args, nil
}

func (proxyMaker) FromConfig(c chkr.CheckConfig) (chkr.Check, error) {
	args, ok := c.Args.(ProxyArgs)
	if !ok {
		return nil, fmt.Errorf("configured arguments are not Proxy arguments")
	}
	return Proxy(args.Method, args.Request, args.Proxy, args.Expected), nil
}

// Proxy returns a check that performs an HTTP request through a specified proxy.
func Proxy(method, request, proxy string, expected int) chkr.Check {
	return func(ctx context.Context, cs chkr.CheckState) (chkr.State, string) {
		req, err := http.NewRequestWithContext(ctx, method, request, nil)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Failed to create request: %v", err)
		}

		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Failed to parse proxy URL: %v", err)
		}
		cl := http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
		resp, err := cl.Do(req)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != expected {
			return chkr.Fail, fmt.Sprintf("Unexpected status code: %d (expected %d)", resp.StatusCode, expected)
		}

		return chkr.OK, ""
	}
}
