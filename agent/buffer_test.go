package agent

import (
	"slices"
	"testing"

	chkr "github.com/tweithoener/checker"
)

func TestNewEventBuffer(t *testing.T) {
	b1 := NewEventBuffer(0)
	if b1.limit != 100 {
		t.Errorf("expected default limit 100, got %d", b1.limit)
	}

	b2 := NewEventBuffer(5)
	if b2.limit != 5 {
		t.Errorf("expected limit 5, got %d", b2.limit)
	}
}

func TestEventBuffer_RingBehavior(t *testing.T) {
	b := NewEventBuffer(3)

	b.Add(chkr.CheckState{Name: "check1", State: chkr.OK})
	b.Add(chkr.CheckState{Name: "check2", State: chkr.Fail})
	b.Add(chkr.CheckState{Name: "check3", State: chkr.Warn})

	if b.Len() != 3 {
		t.Fatalf("expected len 3, got %d", b.Len())
	}

	evs := slices.Collect(b.Events())
	if evs[0].Name != "check1" || evs[2].Name != "check3" {
		t.Fatalf("unexpected event order before wrap: %v", evs)
	}

	// Add 4th element, should overwrite "check1"
	b.Add(chkr.CheckState{Name: "check4", State: chkr.OK})

	if b.Len() != 3 {
		t.Fatalf("expected len 3 after wrap, got %d", b.Len())
	}

	evsAfterWrap := slices.Collect(b.Events())
	if evsAfterWrap[0].Name != "check2" || evsAfterWrap[2].Name != "check4" {
		t.Fatalf("unexpected event order after wrap: %v", evsAfterWrap)
	}
}

func TestEventBuffer_Cache(t *testing.T) {
	b := NewEventBuffer(5)
	b.Add(chkr.CheckState{Name: "check1", State: chkr.OK})

	evs1 := slices.Collect(b.Events())

	if b.isDirty {
		t.Fatal("buffer should not be dirty after Events() call")
	}

	evs2 := slices.Collect(b.Events())

	// Simply verifying that len is stable and cache isn't broken
	if len(evs1) != 1 || len(evs2) != 1 {
		t.Fatalf("cache broke array length")
	}

	// Adding a new event should invalidate the cache
	b.Add(chkr.CheckState{Name: "check2", State: chkr.Warn})
	if !b.isDirty {
		t.Fatal("buffer should be dirty after Add() call")
	}

	evs3 := slices.Collect(b.Events())
	if len(evs3) != 2 {
		t.Fatalf("expected 2 events, got %d", len(evs3))
	}
	if b.isDirty {
		t.Fatal("buffer should be clean after second Events() call")
	}
}
