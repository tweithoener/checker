package checker

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Check defines the signature for a health check function.
type Check func(ctx context.Context, h History) (s State, message string)

// Notifier defines the signature for a notification handler.
type Notifier func(ctx context.Context, name string, h History)

// Checker manages checks and notifiers, executing them at a configured interval.
type Checker struct {
	checks    []*meta
	notifiers []Notifier
	interval  time.Duration

	wg     sync.WaitGroup
	quit   chan struct{}
	ctx    context.Context
	cancel context.CancelFunc

	serverConfig ServerConfig
	httpServer   *http.Server
}

// State represents the result state of a check.
type State string

const (
	OK      State = "OK"
	Warn    State = "Warning"
	Fail    State = "Failed"
	Skipped State = "Skipped"
)

// History provides historic information for a check.
type History interface {
	Last(s State) time.Time
	State() State
	Since() time.Time
	Message() string
	Streak() int
	Name() string
	String() string
}

// New initializes and returns a new Checker.
func New() *Checker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Checker{
		checks:    []*meta{},
		notifiers: []Notifier{},
		interval:  5 * time.Minute,
		quit:      make(chan struct{}),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// AddCheck registers a new check with the specified name.
func (chkr *Checker) AddCheck(name string, chk Check) {
	chkr.checks = append(chkr.checks, &meta{
		call: chk,

		history: history{
			name:    name,
			last:    make(map[State]time.Time),
			state:   OK,
			since:   time.Now(),
			message: "",
			streak:  0,
		},
	})
}

// AddNotifier registers a new notifier.
func (chkr *Checker) AddNotifier(n Notifier) {
	chkr.notifiers = append(chkr.notifiers, n)
}

// Start begins the periodic execution of registered checks.
func (chkr *Checker) Start() {
	chkr.startHttpServer()

	if len(chkr.checks) == 0 {
		return
	}
	chkr.wg.Add(1)
	go func() {
		defer chkr.wg.Done()

		cnt := len(chkr.checks)
		te := chkr.interval / time.Duration(cnt)
		ticker := time.NewTicker(te)
		defer ticker.Stop()

		log.Printf("Checker starts with %d checks. Running a check every %d millis.", cnt, te/time.Millisecond)

		idx := 0
		for {
			select {
			case <-chkr.quit:
				return
			case <-ticker.C:
				chkr.runCheck(chkr.checks[idx])
				idx = (idx + 1) % cnt
			}
		}
	}()
}

func (chkr *Checker) runCheck(meta *meta) {
	chkr.wg.Add(1)
	go func() {
		defer chkr.wg.Done()

		if !meta.mu.TryLock() {
			log.Printf("Check %s still running - skipping\n", meta.name)
			return
		}
		defer meta.mu.Unlock()

		now := time.Now()
		s, msg := meta.call(chkr.ctx, meta)

		if s == Skipped {
			return
		}
		meta.message = msg
		meta.streak++
		stateChange := s != meta.state
		meta.state = s
		meta.last[s] = now
		if stateChange {
			meta.streak = 1
			meta.since = now
		}
		if stateChange || meta.state != OK {
			snap := meta.snapshot()

			for _, n := range chkr.notifiers {
				chkr.wg.Add(1)
				go func(notifier Notifier, h History) {
					defer chkr.wg.Done()
					notifier(chkr.ctx, meta.name, h)
				}(n, snap)
			}
		}
	}()
}

// Shutdown gracefully stops the Checker, bounded by the provided context.
func (chkr *Checker) Shutdown(ctx context.Context) {
	chkr.stopHttpServer(ctx)
	close(chkr.quit)

	allDone := make(chan struct{})
	go func() {
		defer close(allDone)
		chkr.wg.Wait()
	}()

	select {
	case <-allDone:
		log.Println("Done. Clean shutdown.")
	case <-ctx.Done():
		log.Println("Grace period expired. Forced stop of remaining checks.")
		chkr.cancel()
	}
}

// SetInterval updates the duration between check executions.
func (chkr *Checker) SetInterval(interval time.Duration) {
	chkr.interval = interval
}

// State returns the aggregated state of all registered checks.
func (chkr *Checker) State() State {
	warn := false
	for _, chk := range chkr.checks {
		if chk.state == Fail {
			return Fail
		}
		if chk.state == Warn {
			warn = true
		}
	}
	if warn {
		return Warn
	}
	return OK
}

type history struct {
	name    string
	last    map[State]time.Time
	since   time.Time
	message string
	state   State
	streak  int
}

type meta struct {
	history

	mu   sync.RWMutex
	call Check
}

func (h *history) Name() string {
	return h.name
}

func (h *history) Last(state State) time.Time {
	return h.last[state]
}

func (h *history) Since() time.Time {
	return h.since
}

func (h *history) State() State {
	return h.state
}

func (h *history) Streak() int {
	return h.streak
}

func (h *history) Message() string {
	return h.message
}

func (h *history) String() string {
	return fmt.Sprintf("%s %s: %s (%d times since %s)", h.State(), h.Name(), h.Message(), h.Streak(), h.Since().Local().Format("2006-01-02 15:04:05"))
}

func (m *meta) snapshot() History {
	snapSince := make(map[State]time.Time)
	for k, v := range m.last {
		snapSince[k] = v
	}
	return &history{
		last:    snapSince,
		message: m.message,
		state:   m.state,
		since:   m.since,
		streak:  m.streak,
		name:    m.name,
	}
}
