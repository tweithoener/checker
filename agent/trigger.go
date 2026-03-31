package agent

import (
	"time"

	chkr "github.com/tweithoener/checker"
)

// VolumeTrigger fires if the number of NEW events since the last analysis reaches a threshold.
type VolumeTrigger struct {
	Threshold int
}

func NewVolumeTrigger(threshold int) *VolumeTrigger {
	return &VolumeTrigger{Threshold: threshold}
}

func (t *VolumeTrigger) ShouldTrigger(buffer *EventBuffer, lastAnalysis time.Time) bool {
	if lastAnalysis.IsZero() {
		return buffer.Len() >= t.Threshold
	}

	newCount := 0
	for ev := range buffer.Events() {
		if ev.ReceivedAt.After(lastAnalysis) {
			newCount++
		}
	}
	return newCount >= t.Threshold
}

// TimeTrigger fires if a certain duration has passed since the last analysis.
type TimeTrigger struct {
	Interval time.Duration
}

func NewTimeTrigger(interval time.Duration) *TimeTrigger {
	return &TimeTrigger{Interval: interval}
}

func (t *TimeTrigger) ShouldTrigger(buffer *EventBuffer, lastAnalysis time.Time) bool {
	if lastAnalysis.IsZero() {
		// If never analyzed, trigger if we have at least one event
		return buffer.Len() > 0
	}

	// We only trigger if time has passed AND there are actually new events to analyze
	hasNewEvents := false
	for ev := range buffer.Events() {
		if ev.ReceivedAt.After(lastAnalysis) {
			hasNewEvents = true
			break
		}
	}

	return time.Since(lastAnalysis) >= t.Interval && hasNewEvents
}

// StateTrigger fires if at least N NEW events since the last analysis have a specific state.
type StateTrigger struct {
	State     chkr.State
	Threshold int
}

func NewStateTrigger(state chkr.State, threshold int) *StateTrigger {
	return &StateTrigger{State: state, Threshold: threshold}
}

func (t *StateTrigger) ShouldTrigger(buffer *EventBuffer, lastAnalysis time.Time) bool {
	count := 0
	for ev := range buffer.Events() {
		if !lastAnalysis.IsZero() && !ev.ReceivedAt.After(lastAnalysis) {
			continue // Skip events already analyzed
		}

		if ev.CheckState.State == t.State {
			count++
		}
		if count >= t.Threshold {
			return true
		}
	}
	return false
}

// AndTrigger fires only if all of its sub-triggers fire.
type AndTrigger struct {
	Triggers []Trigger
}

func NewAndTrigger(triggers ...Trigger) *AndTrigger {
	return &AndTrigger{Triggers: triggers}
}

func (t *AndTrigger) ShouldTrigger(buffer *EventBuffer, lastAnalysis time.Time) bool {
	if len(t.Triggers) == 0 {
		return false
	}
	for _, tri := range t.Triggers {
		if !tri.ShouldTrigger(buffer, lastAnalysis) {
			return false
		}
	}
	return true
}

// OrTrigger fires if at least one of its sub-triggers fires.
type OrTrigger struct {
	Triggers []Trigger
}

func NewOrTrigger(triggers ...Trigger) *OrTrigger {
	return &OrTrigger{Triggers: triggers}
}

func (t *OrTrigger) ShouldTrigger(buffer *EventBuffer, lastAnalysis time.Time) bool {
	for _, tri := range t.Triggers {
		if tri.ShouldTrigger(buffer, lastAnalysis) {
			return true
		}
	}
	return false
}
