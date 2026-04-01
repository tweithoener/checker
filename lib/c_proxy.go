package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

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

var proxyDoRequest = func(cl *http.Client, req *http.Request) (*http.Response, error) {
	return cl.Do(req)
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

		// intentionally creating a new client for every check as we want
		// to check reachability, tls handshake and proxy response
		cl := &http.Client{
			Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
			Timeout:   20 * time.Second,
		}
		resp, err := proxyDoRequest(cl, req)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Request failed: %v", err)
		}
		defer func() {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}()

		if resp.StatusCode != expected {
			return chkr.Fail, fmt.Sprintf("Unexpected status code: %d (expected %d)", resp.StatusCode, expected)
		}

		return chkr.OK, ""
	}
}
