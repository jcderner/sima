package core

import (
	"fmt"
	"testing"
	"time"
)

func TestNewTimeMachine(t *testing.T) {
	tm := NewTimeMachine(1.0, 10, 100)
	if tm == nil {
		t.Fatal("Expected TimeMachine to be created")
	}
	if tm.state != IDLE {
		t.Errorf("Expected initial state to be IDLE, got %v", tm.state)
	}
	if tm.speed != 1.0 {
		t.Errorf("Expected initial speed to be 1.0, got %v", tm.speed)
	}
	if len(tm.events) != 0 {
		t.Errorf("Expected initial events channel to be empty, got %v", len(tm.events))
	}
}

func TestTimeMachineStart(t *testing.T) {
	tm := NewTimeMachine(1.0, 10, 100)
	success := tm.Start()
	if !success {
		t.Error("Expected TimeMachine to start successfully")
	}
	if tm.state != RUNNING {
		t.Errorf("Expected state to be RUNNING, got %v", tm.state)
	}
}

func TestTimeMachinePauseResume(t *testing.T) {
	speed := 2.0
	sd := int(100 * 1000 * 1000 / speed)
	sleepDuration, _ := time.ParseDuration(fmt.Sprintf("%vns", sd))
	tm := NewTimeMachine(speed, 10, 1)
	done := make(map[int]bool)
	schedule := func(t int) {
		done[t] = false
		tm.Schedule(float64(t), func() {
			done[t] = true
		})
	}
	schedule(10)
	schedule(20)
	schedule(110)
	schedule(120)
	schedule(210)
	schedule(220)
	schedule(310)
	schedule(320)
	tm.Start()
	time.Sleep(sleepDuration) //after sleep the first 100 ms should be processed.
	if !done[10] {
		t.Errorf("Expected done[10] to be executed")
	}
	if !done[20] {
		t.Errorf("Expected done[20] to be executed")
	}
	if done[110] {
		t.Errorf("Expected done[110] NOT to be executed")
	}
	//pause and resume
	tm.Pause()
	time.Sleep(sleepDuration) //nothing happens during sleep
	tm.Resume()
	time.Sleep(sleepDuration) //the next 100 ms should be processed.
	if !done[110] {
		t.Errorf("Expected done[110] to be executed")
	}
	if !done[120] {
		t.Errorf("Expected done[120] to be executed")
	}
	if done[210] {
		t.Errorf("Expected done[210] NOT to be executed")
	}
	//pause and resume
	tm.Pause()
	time.Sleep(sleepDuration) //nothing happens during sleep
	tm.Resume()
	time.Sleep(sleepDuration) //the next 100 ms should be processed.
	if !done[210] {
		t.Errorf("Expected done[210] to be executed")
	}
	if !done[220] {
		t.Errorf("Expected done[220] to be executed")
	}
	if done[310] {
		t.Errorf("Expected done[310] NOT to be executed")
	}
	//process remaining
	time.Sleep(sleepDuration)
	if !done[310] {
		t.Errorf("Expected done[310] to be executed")
	}
	if !done[320] {
		t.Errorf("Expected done[320] to be executed")
	}
}

func TestTimeMachineStop(t *testing.T) {
	tm := NewTimeMachine(1.0, 10, 2)
	tm.Start()
	tm.Stop()
	time.Sleep(200 * time.Millisecond) // wait for the stop to take effect
	if tm.state != STOPPED {
		t.Errorf("Expected state to be STOPPED, got %v", tm.state)
	}
}

func TestTimeMachineSetSpeed(t *testing.T) {
	tm := NewTimeMachine(1.0, 10, 2)
	tm.Start()
	tm.SetSpeed(2.0)
	time.Sleep(200 * time.Millisecond) // wait for the speed change to take effect
	if tm.speed != 2.0 {
		t.Errorf("Expected speed to be 2.0, got %v", tm.speed)
	}
}

func TestTimeMachineSchedule(t *testing.T) {
	tm := NewTimeMachine(1.0, 10, 2)
	tm.Start()
	executed100 := false
	tm.Schedule(100, func() {
		executed100 = true
	})
	executed0 := false
	tm.Schedule(-10, func() {
		executed0 = true
	})
	time.Sleep(200 * time.Millisecond) // wait for the event to be processed
	if !executed100 {
		t.Errorf("Expected scheduled event 100 to be executed")
	}
	if !executed0 {
		t.Errorf("Expected scheduled event 0 to be executed")
	}
}
