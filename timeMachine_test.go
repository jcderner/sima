package sima

import (
	"fmt"
	"log"
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

func TestTimeMachineCommands(t *testing.T) {
	tm := NewTimeMachine(1.0, 10, 2)
	if !tm.Start() {
		t.Error("Expected Start to return true")
	}
	if !tm.Pause() {
		t.Error("Expected Pause to return true")
	}
	if !tm.Resume() {
		t.Error("Expected Resume to return true")
	}
	if !tm.Pause() {
		t.Error("Expected Pause to return true")
	}
	if !tm.Resume() {
		t.Error("Expected Resume to return true")
	}
	if tm.Resume() {
		t.Error("Expected Pause to return false")
	}
	if tm.Resume() {
		t.Error("Expected Resume to return fale")
	}
	if !tm.Pause() {
		t.Error("Expected Pause to return true")
	}
	if tm.Pause() {
		t.Error("Expected Pause to return false")
	}
	if tm.Pause() {
		t.Error("Expected Pause to return false")
	}
	if !tm.Stop() {
		t.Error("Expected Stop to return true")
	}
	if tm.Resume() {
		t.Error("Expected Resume to return false")
	}
	if tm.Stop() {
		t.Error("Expected Stop to return false")
	}
}
func TestTimeMachinePauseResume(t *testing.T) {
	speed := 10.0
	sleepDuration := (time.Duration)(int64(100 * 1000 * 1000 / speed))
	// sleepDuration, _ := time.ParseDuration(fmt.Sprintf("%vns", sd))
	tm := NewTimeMachine(speed, 10, 1)
	done := make(map[int]bool)
	schedule := func(t int) {
		done[t] = false
		tm.Schedule(float64(t), func() {
			done[t] = true
		})
	}
	schedule(50)
	schedule(60)
	schedule(150)
	schedule(160)
	schedule(250)
	schedule(260)
	schedule(350)
	schedule(360)
	tm.Start()
	time.Sleep(sleepDuration)
	if !done[50] {
		t.Errorf("Expected done[50] to be executed")
	}
	if !done[60] {
		t.Errorf("Expected done[60] to be executed")
	}
	if done[150] {
		t.Errorf("Expected done[150] NOT to be executed")
	}
	//pause and resume
	tm.Pause()
	time.Sleep(sleepDuration)
	tm.Resume()
	time.Sleep(sleepDuration)
	if !done[150] {
		t.Errorf("Expected done[150] to be executed")
	}
	if !done[160] {
		t.Errorf("Expected done[160] to be executed")
	}
	if done[250] {
		t.Errorf("Expected done[250] NOT to be executed")
	}
	//pause and resume
	tm.Pause()
	time.Sleep(sleepDuration)
	tm.Resume()
	time.Sleep(sleepDuration)
	if !done[250] {
		t.Errorf("Expected done[250] to be executed")
	}
	if !done[260] {
		t.Errorf("Expected done[260] to be executed")
	}
	if done[350] {
		t.Errorf("Expected done[350] NOT to be executed")
	}
	//process remaining
	time.Sleep(sleepDuration)
	if !done[350] {
		t.Errorf("Expected done[350] to be executed")
	}
	if !done[360] {
		t.Errorf("Expected done[360] to be executed")
	}
	log.Printf("Last Pause")
	tm.Pause()
	tm.Resume()
	tm.Stop()
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
	tm.Stop()
}

// Example shows how to use the TimeMachine.
// Before Start the ping function is scheduled at t=0 ms.
// The TimeMachine then runs ping and schedules pong at t=20 ms.
// Pong in turn schedules ping at t=100 (=20+80) ms.
// Finally ping schedules pong at t=120 (=100+20) ms.
//
// Since the speed is 2.0, the program needs (roughly) 60 ms.
// The real time granularity is given by the cycle time (here: 1 ms)
// and taken into account some imprecision in real time due to the OS,
// it save to sleep for 70 ms for getting 4 events (ping, pong, ping, pong).
func Example() {
	tm := NewTimeMachine(2.0, 10, 1)
	var ping func()
	var pong func()
	ping = func() {
		fmt.Println(tm.T(), ": Ping!")
		tm.Schedule(20.0, pong)
	}
	pong = func() {
		fmt.Println(tm.T(), ": Pong!")
		tm.Schedule(80.0, ping)
	}
	tm.Schedule(0.0, ping)
	tm.Start()
	time.Sleep(70 * time.Millisecond) // wait for ping, pong, ping, pong
	tm.Stop()
	// Output:
	// 0 : Ping!
	// 20 : Pong!
	// 100 : Ping!
	// 120 : Pong!
}

// See TimeMachine running for 10 s., pausing for 10 s. and then smoothly continuing.
// (Play with the parameter to see the effect more clearly.)
func ExampleTimeMachine_Pause() {
	tm := NewTimeMachine(1.0, 100, 10)
	var ping func()
	var pong func()
	ping = func() {
		fmt.Println(tm.T(), ": Ping!")
		tm.Schedule(20.0, pong)
	}
	pong = func() {
		fmt.Println(tm.T(), ": Pong!")
		tm.Schedule(80.0, ping)
	}
	tm.Schedule(0.0, ping)
	tm.Start()
	time.Sleep(90 * time.Millisecond)
	tm.Pause()
	time.Sleep(100 * time.Millisecond)
	tm.Resume()
	time.Sleep(50 * time.Millisecond)
	tm.Pause() // no use to pause, but it doesn't hurt. You can omit this call.
	tm.Stop()
	// Output:
	// 0 : Ping!
	// 20 : Pong!
	// 100 : Ping!
	// 120 : Pong!
}
