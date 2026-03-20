package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	chkr "github.com/tweithoener/checker"
)

type HttpArgs struct {
	Method   string
	Url      string
	Expected int
}

type httpMaker struct{}

var httpMkr = httpMaker{}

func (httpMaker) Maker() string {
	return "Http"
}

func (httpMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	args := HttpArgs{}
	if err := json.Unmarshal(j, &args); err != nil {
		return args, fmt.Errorf("can't unmarshal Http arguments: %v", err)
	}
	return args, nil
}

func (httpMaker) FromConfig(c chkr.CheckConfig) (chkr.Check, error) {
	args, ok := c.Args.(HttpArgs)
	if !ok {
		return nil, fmt.Errorf("configured arguments are not Http arguments")
	}
	return Http(args.Method, args.Url, args.Expected), nil
}

func Http(method, url string, expected int) chkr.Check {
	return func(ctx context.Context, h chkr.History) (chkr.State, string) {
		req, err := http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Failed to create request: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
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
