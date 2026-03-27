package lib

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	chkr "github.com/tweithoener/checker"
)

func TestHttp(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	chk := Http("GET", ts.URL, http.StatusAccepted)
	s, _ := chk(context.Background(), chkr.CheckState{})
	if s != chkr.OK {
		t.Errorf("Http check should be OK for status 202, got %s", s)
	}

	chk = Http("GET", ts.URL, http.StatusOK)
	s, _ = chk(context.Background(), chkr.CheckState{})
	if s != chkr.Fail {
		t.Errorf("Http check should be Fail for status 200 (expected 202), got %s", s)
	}

	// Test invalid URL
	chk = Http("GET", "http://invalid-url-that-does-not-exist.example.com", 200)
	s, _ = chk(context.Background(), chkr.CheckState{})
	if s != chkr.Fail {
		t.Error("Http check should fail for invalid URL")
	}
}
