package core

import (
	"container/heap"
)

// Event represents a single Event in the simulation
type Event struct {
	T float64
	F func()
}

// EventQueue implements a priority queue for Events.
type EventQueue []*Event

// Len returns the length of the event queue.
func (eq EventQueue) Len() int { return len(eq) }

// Less returns true if the event at index i is less/earlier than the event at index j.
func (eq EventQueue) Less(i, j int) bool {
	return eq[i].T < eq[j].T
}

// Swap swaps the events at index i and j.
// Only to be used by the heap package.
// Never call it directly.
func (eq EventQueue) Swap(i, j int) {
	eq[i], eq[j] = eq[j], eq[i]
}

// Push adds an event to the event queue.
// Only to be used by the heap package.
// Never call it directly.
func (eq *EventQueue) Push(x interface{}) {
	event := x.(*Event)
	*eq = append(*eq, event)
}

// Pop removes and returns the last event from the event queue.
// Only to be used by the heap package.
// Never call it directly.
func (eq *EventQueue) Pop() interface{} {
	old := *eq
	n := len(old)
	event := old[n-1]
	old[n-1] = nil // avoid memory leak
	*eq = old[0 : n-1]
	return event
}

// tNext returns the time of the next event in ms.
// If there are no more events then ok is false.
func (eq EventQueue) NextT() (t float64, ok bool) {
	if len(eq) > 0 {
		ok = true
		t = eq[0].T
	}
	return t, ok
}

// NewEventQueue creates a new EventQueue.
func NewEventQueue() *EventQueue {
	eq := &EventQueue{}
	heap.Init(eq)
	return eq
}

// Add adds a new event to the queue.
func (eq *EventQueue) Add(event *Event) {
	heap.Push(eq, event)
}

// Next retrieves and removes the Next event from the queue.
func (eq *EventQueue) Next() *Event {
	if eq.Len() == 0 {
		return nil
	}
	return heap.Pop(eq).(*Event)
}
