package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	chkr "github.com/tweithoener/checker"
)

// generateRealisticTimeline generates 2-3 days of monitoring data for a small company.
func generateRealisticTimeline() []Event {
	var events []Event

	// Start: Monday 00:00:00
	start := time.Date(2026, time.March, 2, 0, 0, 0, 0, time.UTC)

	// Simulate Day 1 and Day 2 (Normal Operations)
	for day := 0; day < 2; day++ {
		baseDate := start.AddDate(0, 0, day)

		// 03:00 AM - ISP forced disconnection (PBX SIP Trunk drops)
		events = append(events, Event{
			Name:       "SIP Trunk Ping",
			CheckState: chkr.CheckState{State: chkr.Fail, Message: "timeout"},
			ReceivedAt: baseDate.Add(3 * time.Hour),
		})
		// 03:05 AM - ISP reconnects
		events = append(events, Event{
			Name:       "SIP Trunk Ping",
			CheckState: chkr.CheckState{State: chkr.OK, Message: "reachable (20ms)"},
			ReceivedAt: baseDate.Add(3*time.Hour + 5*time.Minute),
		})

		// 08:00 AM to 09:30 AM - Employees arrive, PCs turn on
		for pc := 1; pc <= 50; pc++ {
			arrivalOffset := time.Duration(pc*2) * time.Minute
			events = append(events, Event{
				Name:       fmt.Sprintf("PC-%02d Ping", pc),
				CheckState: chkr.CheckState{State: chkr.OK, Message: "reachable"},
				ReceivedAt: baseDate.Add(8*time.Hour + arrivalOffset),
			})
		}

		// 17:00 PM to 18:30 PM - Employees leave, PCs turn off
		for pc := 1; pc <= 50; pc++ {
			leaveOffset := time.Duration(pc*2) * time.Minute
			events = append(events, Event{
				Name:       fmt.Sprintf("PC-%02d Ping", pc),
				CheckState: chkr.CheckState{State: chkr.Fail, Message: "host down"},
				ReceivedAt: baseDate.Add(17*time.Hour + leaveOffset),
			})
		}
	}

	// Day 3: Disaster Strikes!
	day3 := start.AddDate(0, 0, 2)

	// 08:00 AM - PCs turn on normally
	for pc := 1; pc <= 50; pc++ {
		arrivalOffset := time.Duration(pc*2) * time.Minute
		events = append(events, Event{
			Name:       fmt.Sprintf("PC-%02d Ping", pc),
			CheckState: chkr.CheckState{State: chkr.OK, Message: "reachable"},
			ReceivedAt: day3.Add(8*time.Hour + arrivalOffset),
		})
	}

	// 14:00 PM - The Cascade begins
	disasterStart := day3.Add(14 * time.Hour)

	events = append(events,
		Event{
			Name:       "AppServer CPU",
			CheckState: chkr.CheckState{State: chkr.Warn, Message: "85% load"},
			ReceivedAt: disasterStart,
		},
		Event{
			Name:       "AppServer CPU",
			CheckState: chkr.CheckState{State: chkr.Fail, Message: "99% load"},
			ReceivedAt: disasterStart.Add(2 * time.Minute),
		},
		Event{
			Name:       "AppServer Disk /var/log",
			CheckState: chkr.CheckState{State: chkr.Fail, Message: "100% full (No space left on device)"},
			ReceivedAt: disasterStart.Add(15 * time.Minute),
		},
		Event{
			Name:       "Database Connection",
			CheckState: chkr.CheckState{State: chkr.Fail, Message: "connection refused"},
			ReceivedAt: disasterStart.Add(20 * time.Minute),
		},
		Event{
			Name:       "WebFrontend HTTP",
			CheckState: chkr.CheckState{State: chkr.Fail, Message: "HTTP 503 Service Unavailable"},
			ReceivedAt: disasterStart.Add(25 * time.Minute),
		},
	)

	return events
}

// mockNotifier captures the final notification sent by the AIAgent.
type mockNotifier struct {
	mu       sync.Mutex
	captured []chkr.CheckState
}

func (m *mockNotifier) Notify(_ context.Context, cs chkr.CheckState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.captured = append(m.captured, cs)
}

func TestAIAgent_Integration(t *testing.T) {
	// 1. Mock the LLM HTTP Client
	originalDo := doHttpRequest
	defer func() { doHttpRequest = originalDo }()

	var capturedPrompt string

	doHttpRequest = func(req *http.Request) (*http.Response, error) {
		// Read the body the agent generated
		bodyBytes, _ := io.ReadAll(req.Body)
		var parsedReq chatRequest
		json.Unmarshal(bodyBytes, &parsedReq)

		// Capture the user prompt (which contains our events)
		for _, msg := range parsedReq.Messages {
			if msg.Role == "user" {
				capturedPrompt = msg.Content
			}
		}

		// Mock a successful LLM response
		mockResp := chatResponse{}
		mockResp.Choices = append(mockResp.Choices, struct {
			Message chatMessage `json:"message"`
		}{
			Message: chatMessage{
				Role:    "assistant",
				Content: "Analysis complete. Detected predictable transient errors (SIP Trunk nightly drop, PCs powering down) and a critical cascading failure starting with AppServer CPU load, leading to disk full, database crash, and web service outage.",
			},
		})

		respBytes, _ := json.Marshal(mockResp)
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader(respBytes)),
		}, nil
	}

	// 2. Setup the AIAgent
	outNotifier := &mockNotifier{}
	llmClient := NewRESTClient("https://api.mock.local/v1/chat/completions", "mock-token", "gpt-4-mock")

	// We use an AndTrigger: At least 5 fails AND at least 5 minutes passed since last analysis
	failTrigger := NewStateTrigger(chkr.Fail, 5)
	timeTrigger := NewTimeTrigger(5 * time.Minute)
	compositeTrigger := NewAndTrigger(failTrigger, timeTrigger)

	agent := New(llmClient, outNotifier.Notify,
		WithTrigger(compositeTrigger),
		WithBufferLimit(500), // Large enough to hold our 3 days of events
	)

	// 3. Inject the generated timeline
	timeline := generateRealisticTimeline()

	if len(timeline) < 100 {
		t.Fatalf("Timeline too short, generated %d events", len(timeline))
	}

	// We push events into the agent sequentially.
	for _, ev := range timeline {
		// We pass our simulated timestamps indirectly. In reality, the checker passes CheckState.
		// However, the agent stamps 'time.Now()'. To test historical data without waiting 3 days,
		// we inject directly into the buffer for this test, mimicking the agent's internal behavior,
		// OR we adjust the test to force the timestamp.

		// For the sake of the test, we'll acquire the lock, add the event manually to keep the simulated time,
		// and then evaluate the trigger manually, exactly as Notifier() does.
		agent.mu.Lock()

		agent.buffer.events[agent.buffer.tail] = ev // Inject with simulated time
		agent.buffer.tail = (agent.buffer.tail + 1) % agent.buffer.limit
		if agent.buffer.count < agent.buffer.limit {
			agent.buffer.count++
		} else {
			agent.buffer.head = (agent.buffer.head + 1) % agent.buffer.limit
		}
		agent.buffer.isDirty = true

		shouldTrigger := agent.trigger.ShouldTrigger(agent.buffer, agent.lastAnalysis)
		if shouldTrigger {
			eventsToAnalyze := agent.buffer.Events()
			prevAnalysis := agent.lastAnalysis
			agent.lastAnalysis = ev.ReceivedAt // Use simulated time for lastAnalysis

			// Call the LLM synchronously for the test
			agent.mu.Unlock()
			analysis, _ := agent.client.Analyze(context.Background(), agent.systemPrompt, eventsToAnalyze, prevAnalysis)
			outNotifier.Notify(context.Background(), chkr.CheckState{State: chkr.OK, Message: analysis})
			agent.mu.Lock()
		}

		agent.mu.Unlock()
	}

	// 4. Verification
	// The trigger should have fired exactly once (when the disaster hit 5 Fail states)
	outNotifier.mu.Lock()
	defer outNotifier.mu.Unlock()

	if len(outNotifier.captured) == 0 {
		t.Fatalf("Expected output notifier to be called, but it wasn't")
	}

	finalMsg := outNotifier.captured[0].Message
	if !strings.Contains(finalMsg, "cascading failure") {
		t.Errorf("Expected analysis to contain 'cascading failure', got: %s", finalMsg)
	}

	if !strings.Contains(capturedPrompt, "PC-01 Ping") || !strings.Contains(capturedPrompt, "AppServer CPU") {
		t.Errorf("Captured prompt did not contain expected events. Prompt snippet: %s", capturedPrompt[:100])
	}
}
