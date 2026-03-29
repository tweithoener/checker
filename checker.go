package checker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// Check defines the signature for a health check function.
type Check func(ctx context.Context, cs CheckState) (s State, message string)

// Notifier defines the signature for a notification handler.
type Notifier func(ctx context.Context, cs CheckState)

// Checker manages checks and notifiers, executing them at a configured interval.
type Checker struct {
	mu sync.RWMutex

	running    bool
	name       string
	checks     []*meta
	notifiers  []Notifier
	peerStates map[string]PeerState
	interval   time.Duration

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

// New initializes and returns a new Checker.
func New() *Checker {
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("can't get hostname: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Checker{
		name:       hostname,
		checks:     []*meta{},
		notifiers:  []Notifier{},
		peerStates: map[string]PeerState{},
		interval:   5 * time.Minute,
		quit:       make(chan struct{}),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// AddCheck registers a new check with the specified name.
func (chkr *Checker) AddCheck(name string, chk Check) error {
	chkr.mu.Lock()
	defer chkr.mu.Unlock()
	if chkr.running {
		return errors.New("can't change configuration while checker is running")
	}
	for _, c := range chkr.checks {
		if c.Name == name {
			return fmt.Errorf("check '%s' already exists", name)
		}
	}
	chkr.checks = append(chkr.checks, &meta{
		call:   chk,
		isPeer: false,

		CheckState: CheckState{
			Name:    name,
			Last:    make(map[State]time.Time),
			State:   OK,
			Since:   time.Now(),
			Message: "",
			Streak:  0,
		},
	})
	return nil
}

// AddPeer adds a peer for p2p monitoring. Address must consist of hostname or ip address plus port
// (e.g. "example.com:8080").
// Note: Only peers that are added using this method (or the respective JSON config)
// will be probed actively. However, you might see more peers in the checker web page.
// This happens when direct peers have more peers. Peer states are passed on throughout the
// monitoring network. But only state changes in direct peers will result in notifications.
func (chkr *Checker) AddPeer(address string) error {
	chkr.mu.Lock()
	defer chkr.mu.Unlock()
	if chkr.running {
		return errors.New("can't change configuration while checker is running")
	}
	for _, c := range chkr.checks {
		if c.Name == "Peer "+address {
			return fmt.Errorf("peer '%s' already exists", address)
		}
	}
	chk := chkr.peerCheck(address)
	chkr.checks = append(chkr.checks, &meta{
		call:   chk,
		isPeer: true,

		CheckState: CheckState{
			Name:    "Peer " + address,
			Last:    make(map[State]time.Time),
			State:   OK,
			Since:   time.Now(),
			Message: "",
			Streak:  0,
		},
	})
	return nil
}

// AddNotifier registers a new notifier.
func (chkr *Checker) AddNotifier(n Notifier) error {
	chkr.mu.Lock()
	defer chkr.mu.Unlock()
	if chkr.running {
		return errors.New("can't change configuration while checker is running")
	}
	chkr.notifiers = append(chkr.notifiers, n)
	return nil
}

// Start begins the periodic execution of registered checks.
func (chkr *Checker) Start() error {
	chkr.mu.Lock()
	defer chkr.mu.Unlock()
	if chkr.running {
		return errors.New("checker is already running")
	}
	if len(chkr.checks) == 0 {
		return errors.New("no checks, no peers")
	}
	chkr.running = true
	chkr.startHttpServer()

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
	return nil
}

func (chkr *Checker) runCheck(meta *meta) {
	chkr.wg.Add(1)
	go func() {
		defer chkr.wg.Done()

		if !meta.mu.TryLock() {
			log.Printf("Check %s still running - skipping\n", meta.Name)
			return
		}
		defer meta.mu.Unlock()

		now := time.Now()
		s, msg := meta.call(chkr.ctx, meta.CheckState)

		if s == Skipped {
			return
		}
		meta.Message = msg
		meta.Streak++
		stateChange := s != meta.State
		meta.State = s
		meta.Last[s] = now
		if stateChange {
			meta.Streak = 1
			meta.Since = now
		}
		if stateChange || meta.State != OK {
			snap := meta.snapshotNoLock()

			for _, n := range chkr.notifiers {
				chkr.wg.Add(1)
				go func(notifier Notifier, cs CheckState) {
					defer chkr.wg.Done()
					notifier(chkr.ctx, cs)
				}(n, snap)
			}
		}
	}()
}

// Shutdown gracefully stops the Checker, bounded by the provided context.
func (chkr *Checker) Shutdown(ctx context.Context) error {
	if !chkr.running {
		return errors.New("checker is not running")
	}
	defer func() {
		chkr.running = false
	}()

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
	return nil
}

// SetInterval updates the duration between check executions.
// It determines how often the entire suite of checks is iterated over.
func (chkr *Checker) SetInterval(interval time.Duration) {
	chkr.interval = interval
}

// SetName sets the name of this checker instance.
// If this method is not used or called with an empty string,
// the system's hostname will be used automatically.
func (chkr *Checker) SetName(name string) {
	if name == "" {
		var err error
		name, err = os.Hostname()
		if err != nil {
			log.Printf("can't get system's hostname: %v", err)
			name = ""
		}
	}
	chkr.name = name
}

// Name returns the configured name of the checker instance.
func (chkr *Checker) Name() string {
	return chkr.name
}

func (chkr *Checker) snapshot() CheckerState {
	chkr.mu.RLock()
	defer chkr.mu.RUnlock()
	st := CheckerState{
		PeerState: PeerState{
			Name:   chkr.name,
			Checks: map[string]CheckState{},
		},
		PeerStates: make(map[string]PeerState, len(chkr.peerStates)),
	}
	for _, m := range chkr.checks {
		if m.isPeer {
			continue
		}
		st.Checks[m.Name] = m.snapshot()
	}
	for addr, ps := range chkr.peerStates {
		st.PeerStates[addr] = ps.deepCopy()
	}
	state, message := summary(st.Checks)
	st.State = state
	st.Summary = message
	return st
}

type meta struct {
	CheckState

	mu     sync.RWMutex
	call   Check
	isPeer bool
}

func (m *meta) snapshotNoLock() CheckState {
	return m.CheckState.deepCopy()
}
func (m *meta) snapshot() CheckState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.snapshotNoLock()
}
