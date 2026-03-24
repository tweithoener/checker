package checker

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// StateTransfer is used to transfer json encoded state information between
// checker instances using the checker server and Peer check from the standard lib.
type StateTransfer struct {
	State  State                `json:"state"`
	Checks []CheckStateTransfer `json:"checks"`
}

// CheckStateTransfer represents the state of a single check for JSON transfer.
type CheckStateTransfer struct {
	Name    string    `json:"name"`
	State   State     `json:"state"`
	Message string    `json:"message"`
	Streak  int       `json:"streak"`
	Since   time.Time `json:"since"`
}

// EnableServer enables the peer-to-peer monitoring server on the provided listen address (e.g. ":8080")
func (chkr *Checker) EnableServer(listen string) {
	chkr.serverConfig.Enabled = true
	chkr.serverConfig.Listen = listen
}

// ServeHTTP provides the current state of the checker as a JSON response.
func (chkr *Checker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	st := StateTransfer{
		State: chkr.State(),
	}
	for _, m := range chkr.checks {
		m.mu.RLock()
		st.Checks = append(st.Checks, CheckStateTransfer{
			Name:    m.name,
			State:   m.state,
			Message: m.message,
			Streak:  m.streak,
			Since:   m.since,
		})
		m.mu.RUnlock()
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(st); err != nil {
		log.Printf("Error encoding state: %v", err)
	}
}

func (chkr *Checker) startHttpServer() {
	if !chkr.serverConfig.Enabled {
		return
	}
	mux := http.NewServeMux()
	mux.Handle("/", chkr)
	chkr.httpServer = &http.Server{
		Addr:    chkr.serverConfig.Listen,
		Handler: mux,
	}

	chkr.wg.Add(1)
	go func() {
		defer chkr.wg.Done()
		log.Printf("Starting checker HTTP server on %s", chkr.serverConfig.Listen)
		if err := chkr.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()
}

func (chkr *Checker) stopHttpServer(ctx context.Context) {
	if chkr.httpServer != nil {
		chkr.httpServer.Shutdown(ctx)
	}
}
