package checker

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
)

// EnableServer enables the peer-to-peer monitoring server on the provided listen address (e.g. ":8080")
// When connecting to this address using a web browser the server responds with an html page displaying
// the current status of this checker instance and all known peers.
func (chkr *Checker) EnableServer(listen string) {
	chkr.serverConfig.Enabled = true
	chkr.serverConfig.Listen = listen
}

// ServeHTTP provides the current state of the checker as a JSON or HTML response,
// depending on the client's Accept header. It also accepts POST requests containing
// peer states to update the local instance's knowledge of the network.
func (chkr *Checker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	acceptsJson := false
	wantsHTML := false
	for _, cts := range r.Header.Values("accept") {
		for ctq := range strings.SplitSeq(cts, ",") {
			ct := strings.Split(ctq, ";q=")[0]
			switch ct {
			case "text/html":
				wantsHTML = true
			case "application/json":
				acceptsJson = true
			}
		}
	}
	if wantsHTML {
		w.Header().Set("Content-Type", "text/html")
		tpl := template.Must(template.New("checker status page").Parse(html))
		if err := tpl.Execute(w, chkr.snapshot()); err != nil {
			slog.Error("error executing template", "error", err)
		}
		return
	}
	if acceptsJson {
		if r.Method == "POST" {
			pss := map[string]PeerState{}
			if err := json.NewDecoder(r.Body).Decode(&pss); err == nil {
				for _, ps := range pss {
					chkr.integratePeerState(ps)
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		snap := chkr.snapshot()
		if err := json.NewEncoder(w).Encode(snap); err != nil {
			slog.Error("error encoding state", "error", err)
		}
		return
	}

	w.WriteHeader(http.StatusBadRequest)
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

	chkr.wg.Go(func() {
		defer chkr.wg.Done()
		slog.Info("starting checker HTTP server", "address", chkr.serverConfig.Listen)
		if err := chkr.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
		}
	})
}

func (chkr *Checker) stopHttpServer(ctx context.Context) {
	if chkr.httpServer != nil {
		chkr.httpServer.Shutdown(ctx)
	}
}

func (chkr *Checker) integratePeerState(ips PeerState) {
	chkr.mu.Lock()
	defer chkr.mu.Unlock()

	ps, ok := chkr.peerStates[ips.Address]
	if !ok {
		// We do not know this peer. add it
		chkr.peerStates[ips.Address] = ips
		return
	}
	anyNewer := false
	allNewer := true
	// peer known. update all check information. we keep the newer one
	for _, ic := range ips.Checks {
		// find check wir same name
		c, ok := ps.Checks[ic.Name]
		if !ok {
			ps.Checks[ic.Name] = ic
			anyNewer = true
			continue
		}
		if c.Since.Before(ic.Since) {
			//ic is newer. keep ic
			ps.Checks[ic.Name] = ic
			anyNewer = true
		} else {
			allNewer = false
		}
	}
	if !anyNewer {
		// unchanged. State and summary of local data ist ok
		return
	}
	if allNewer {
		ps.State = ips.State
		ps.Summary = ips.Summary
	}
	if anyNewer && !allNewer {
		state, summary := summary(ps.Checks)
		ps.State = state
		ps.Summary = summary
	}
	// update the map
	chkr.peerStates[ps.Address] = ps
}

func summary(chks map[string]CheckState) (s State, message string) {
	if len(chks) == 0 {
		return OK, "no checks"
	}
	okCnt, warnCnt, failCnt := 0, 0, 0
	var newestOK, newestWarn, newestFail *CheckState = nil, nil, nil
	s = OK
	for _, chk := range chks {
		switch chk.State {
		case OK:
			okCnt++
			if newestOK == nil || newestOK.Since.Before(chk.Since) {
				newestOK = &chk
			}
		case Warn:
			warnCnt++
			if newestWarn == nil || newestWarn.Since.Before(chk.Since) {
				newestWarn = &chk
			}
		case Fail:
			failCnt++
			if newestFail == nil || newestFail.Since.Before(chk.Since) {
				newestFail = &chk
			}
		}
	}

	newest := newestOK
	if warnCnt > 0 {
		s = Warn
		newest = newestWarn
	}
	if failCnt > 0 {
		s = Fail
		newest = newestFail
	}
	message = fmt.Sprintf("%s (%d %s, %d %s, %d %s). Newest %s", s, okCnt, OK, warnCnt, Warn, failCnt, Fail, newest)
	return
}
