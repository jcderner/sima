package core

import (
	rand "math/rand/v2"
	"testing"
)

func TestEventQueue(t *testing.T) {
	eq := NewEventQueue()
	qty := 999
	for i := 0; i < qty; i++ {
		eq.add(&Event{t: float64(rand.IntN(10 * qty)), f: func() {}})
	}
	if eq.Len() != qty {
		t.Errorf("expected length %d, got %d", qty, eq.Len())
	}
	lastEvent := &Event{t: 0.0, f: func() {}}
	for {
		nextT, ok := eq.nextT()
		if !ok {
			break
		}
		nextEvent := eq.next()
		if lastEvent.t > nextEvent.t {
			t.Errorf("expected next event time %v to be greater than last event time %v", nextEvent.t, lastEvent.t)
		}
		if nextT != nextEvent.t {
			t.Errorf("expected next event time %v to be equal to nextT %v", nextEvent.t, nextT)
		}
		lastEvent = nextEvent
	}
}
