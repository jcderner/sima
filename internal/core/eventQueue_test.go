package core

import (
	rand "math/rand/v2"
	"testing"
)

func TestEventQueue(t *testing.T) {
	eq := NewEventQueue()
	qty := 999
	for i := 0; i < qty; i++ {
		eq.Add(&Event{T: float64(rand.IntN(10 * qty)), F: func() {}})
	}
	if eq.Len() != qty {
		t.Errorf("expected length %d, got %d", qty, eq.Len())
	}
	lastEvent := &Event{T: 0.0, F: func() {}}
	for {
		nextT, ok := eq.NextT()
		if !ok {
			break
		}
		nextEvent := eq.Next()
		if lastEvent.T > nextEvent.T {
			t.Errorf("expected next event time %v to be greater than last event time %v", nextEvent.T, lastEvent.T)
		}
		if nextT != nextEvent.T {
			t.Errorf("expected next event time %v to be equal to nextT %v", nextEvent.T, nextT)
		}
		lastEvent = nextEvent
	}
}
