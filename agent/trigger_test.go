package agent

import (
	"testing"
	"time"

	chkr "github.com/tweithoener/checker"
)

// Helper to create a deterministic buffer with specific event timestamps for testing.
func createBufferWithEvents(events ...Event) *EventBuffer {
	b := NewEventBuffer(100)
	for _, ev := range events {
		b.events[b.tail] = ev
		b.tail = (b.tail + 1) % b.limit
		b.count++
	}
	// Important: Make dirty to force a cache rebuild of Events()
	b.isDirty = true
	return b
}

func TestVolumeTrigger(t *testing.T) {
	now := time.Now()
	b := createBufferWithEvents(
		Event{Name: "c1", ReceivedAt: now.Add(-10 * time.Minute)},
		Event{Name: "c2", ReceivedAt: now.Add(-5 * time.Minute)},
		Event{Name: "c3", ReceivedAt: now.Add(-1 * time.Minute)},
	)

	// c1 is old, c2 and c3 are new -> 2 new events
	lastAnalysis := now.Add(-6 * time.Minute)

	trigger2 := NewVolumeTrigger(2)
	if !trigger2.ShouldTrigger(b, lastAnalysis) {
		t.Errorf("Expected VolumeTrigger to fire (2 new events match threshold 2)")
	}

	trigger3 := NewVolumeTrigger(3)
	if trigger3.ShouldTrigger(b, lastAnalysis) {
		t.Errorf("VolumeTrigger should not fire (only 2 new events, threshold is 3)")
	}

	// First run without lastAnalysis
	if !trigger3.ShouldTrigger(b, time.Time{}) {
		t.Errorf("VolumeTrigger should fire if lastAnalysis is zero and total events >= threshold")
	}
}

func TestTimeTrigger(t *testing.T) {
	now := time.Now()
	b := createBufferWithEvents(
		Event{Name: "c1", ReceivedAt: now.Add(-1 * time.Minute)},
	)

	lastAnalysis := now.Add(-10 * time.Minute)

	trigger := NewTimeTrigger(5 * time.Minute)
	if !trigger.ShouldTrigger(b, lastAnalysis) {
		t.Errorf("Expected TimeTrigger to fire (10m passed, 1 new event)")
	}

	// Test case: Time passed, but NO new events
	bEmpty := createBufferWithEvents()
	if trigger.ShouldTrigger(bEmpty, lastAnalysis) {
		t.Errorf("TimeTrigger should NOT fire if there are no new events, even if interval passed")
	}

	// Test case: Not enough time passed
	lastAnalysisRecent := now.Add(-2 * time.Minute)
	if trigger.ShouldTrigger(b, lastAnalysisRecent) {
		t.Errorf("TimeTrigger should NOT fire if interval has not yet passed")
	}
}

func TestStateTrigger(t *testing.T) {
	now := time.Now()
	b := createBufferWithEvents(
		Event{Name: "c1", CheckState: chkr.CheckState{State: chkr.Fail}, ReceivedAt: now.Add(-10 * time.Minute)}, // Old fail
		Event{Name: "c2", CheckState: chkr.CheckState{State: chkr.Fail}, ReceivedAt: now.Add(-1 * time.Minute)},  // New fail
		Event{Name: "c3", CheckState: chkr.CheckState{State: chkr.OK}, ReceivedAt: now.Add(-1 * time.Minute)},    // New OK
	)

	lastAnalysis := now.Add(-5 * time.Minute)

	trigger1 := NewStateTrigger(chkr.Fail, 1)
	if !trigger1.ShouldTrigger(b, lastAnalysis) {
		t.Errorf("Expected StateTrigger to fire for 1 new Fail event")
	}

	trigger2 := NewStateTrigger(chkr.Fail, 2)
	if trigger2.ShouldTrigger(b, lastAnalysis) {
		t.Errorf("StateTrigger should not fire (only 1 NEW fail event exists, c1 is too old)")
	}
}

func TestCompositeTriggers(t *testing.T) {
	now := time.Now()
	b := createBufferWithEvents(
		Event{Name: "c1", CheckState: chkr.CheckState{State: chkr.Fail}, ReceivedAt: now.Add(-1 * time.Minute)},
		Event{Name: "c2", CheckState: chkr.CheckState{State: chkr.Fail}, ReceivedAt: now.Add(-1 * time.Minute)},
	)
	lastAnalysis := now.Add(-10 * time.Minute)

	volTrigger := NewVolumeTrigger(2)             // Evaluates to true
	stateTrigger := NewStateTrigger(chkr.Fail, 3) // Evaluates to false

	andT := NewAndTrigger(volTrigger, stateTrigger)
	if andT.ShouldTrigger(b, lastAnalysis) {
		t.Errorf("AndTrigger should be false because one sub-trigger is false")
	}

	orT := NewOrTrigger(volTrigger, stateTrigger)
	if !orT.ShouldTrigger(b, lastAnalysis) {
		t.Errorf("OrTrigger should be true because at least one sub-trigger is true")
	}
}
