package sima

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jcderner/sima/internal/core"
)

type TMState int

const (
	IDLE TMState = iota
	PAUSED
	RUNNING
	STOPPED
)

var tmstateNames = map[TMState]string{
	IDLE:    "IDLE",
	PAUSED:  "PAUSED",
	RUNNING: "RUNNING",
	STOPPED: "STOPPED",
}

func (s TMState) String() string {
	return tmstateNames[s]
}

// TimeMachine (TM) provides a discrete event simulation (DES).
//
// A client can schedule a parameterless function to be execute after a certain simulation time.
// All scheduled function are executed strictly (simulation) time ordered.
// In most applications the scheduled functions schedule other functions which keeps the simulation alive.
//
// When the TM is started, the simulation time proceeds with a certain speed (s. [TimeMachin.Speed] and [TimeMachin.SetSpeed])
// compared to the real time.
// Alternatively and for fastest processing the client may call [TimeMachine.Step] in a loop, which completely ignores the real time.
type TimeMachine struct {
	eventQueue *core.EventQueue
	t          float64 //the actual simulation time in ms.
	cycleTime  int     //every cycleTime [ms] the time machine will check for new events.
	state      TMState
	speed      float64 //ratio between simulation and real time.
	cmds       chan string
	events     chan *core.Event
	done       chan bool
}

// NewTimeMachine creates a reference to a new TimeMachine.
//
//   - The speed is the ratio between simulation and real time. If it is smaller than 0.01 it will be reset to 0.01.
//   - The eventChanSize is the capacity of the events channel.
//     It should be at least the number of events that are expected to be scheduled initially or in one cycle.
//   - The cycleTime is the real time in ms between two checks for new commands and events.
func NewTimeMachine(speed float64, eventChanSize int, cycleTime int) *TimeMachine {
	if speed < 0.01 {
		log.Printf("Speed = %v must not be smaller than 0.01. It will be reset to 0.01.", speed)
		speed = 0.01
	}
	return &TimeMachine{
		eventQueue: core.NewEventQueue(),
		t:          0,
		cycleTime:  cycleTime,
		state:      IDLE,
		speed:      speed,
		cmds:       make(chan string),
		done:       make(chan bool), //channel to signal that a command has been executed.
		events:     make(chan *core.Event, eventChanSize),
	}
}

// T returns the current simulation time of the time machine.
func (tm *TimeMachine) T() float64 {
	return tm.t
}

// State returns the state of the TimeMachine. Possible states are:
//   - IDLE: The initial state before the TM starts.
//   - RUNNING: The TM is running by executing in (speed * real_time) one scheduled function after another.
//   - PAUSED:  The TM is paused. A paused TM can resume.
//   - STOPPED: The TM is stopped. A stopped TM is stopped forever.
func (tm *TimeMachine) State() TMState {
	return tm.state
}

// Speed returns the speed of the simulation; i.e.: the ratio between simulation time and real time duration.
func (tm *TimeMachine) Speed() float64 {
	return tm.speed
}

// Schedule schedules a function f to be executed dt ms after the current simulation time.
// f needs to be a parameterless function (or method or closure) without return value
// If dt is negative, it will be reset to 0.0.
func (tm *TimeMachine) Schedule(dt float64, f func()) {
	if dt < 0 {
		log.Printf("dt = %v < 0. Will be reset to 0.0", dt)
		dt = 0.0
	}
	tm.events <- (&core.Event{T: tm.t + dt, F: f})
}

// Start starts the simulation in a go routine.
//
// It returns true when the start was successful and false otherwise.
func (tm *TimeMachine) Start() (success bool) {
	if tm.state != IDLE {
		log.Printf("TimeMachine is not in IDLE state. Ignoring Start command.")
		return
	}
	go tm.run()
	done := <-tm.done
	if done {
		return true
	} else {
		return false
	}
}

// Pause pauses the simulation.
// It remembers the current real time to continue smoothly after a resume.
//
// It returns true when the TM has paused and false otherwise.
func (tm *TimeMachine) Pause() (success bool) {
	if tm.state != RUNNING {
		log.Printf("TimeMachine is not in RUNNING state. Ignoring Pause command.")
		return
	}
	tm.cmds <- "PAUSE"
	done := <-tm.done
	if done {
		return true
	} else {
		return false
	}
}

// Resume resumes the simulation.
// It calculates an offset from the last pause (and previous offsets) to continue smoothly.
//
// It returns true when the TM has resumed running and false otherwise.
func (tm *TimeMachine) Resume() (success bool) {
	if tm.state != PAUSED {
		log.Printf("TimeMachine is not in PAUSED state. Ignoring Resume command.")
		return false
	}
	tm.cmds <- "RESUME"
	done := <-tm.done
	if done {
		return true
	} else {
		return false
	}
}

// Stops the TimeMachine with the next cycle.
//
// It returns true when the TM has stopped and false otherwise.
func (tm *TimeMachine) Stop() (success bool) {
	if tm.state != RUNNING && tm.state != PAUSED {
		log.Printf("TimeMachine is not in RUNNING or PAUSED state. Ignoring Stop command.")
		return
	}
	tm.cmds <- "STOP"
	done := <-tm.done
	if done {
		return true
	} else {
		return false
	}
}

// SetSpeed sets the speed of the simulation.
// If the speed is smaller than 0.01 then it will be reset to 0.01.
func (tm *TimeMachine) SetSpeed(speed float64) (success bool) {
	if speed < 0.01 {
		log.Printf("Speed = %v must not be smaller than 0.01. It will be reset to 0.01.", speed)
		speed = 0.01
	}
	tm.cmds <- "SetSpeed " + strconv.FormatFloat(speed, 'f', -1, 64)
	done := <-tm.done
	if done {
		return true
	} else {
		return false
	}
}

func (tm *TimeMachine) run() {
	if tm.state != IDLE {
		tm.done <- false //signal that the command could not be started.
		return
	}
	tm.state = RUNNING
	tm.done <- true                                 //signal that the command has been started.
	tick := (time.Duration)(1000000 * tm.cycleTime) //cycleTime ms.
	var tReal time.Time
	var tReal_pause time.Time //timestamp of the last pause
	var tReal_offset int64    // cumulated offset in ms due to pause and resume
	ticker := time.NewTicker(tick)
	tReal_start := time.Now() //the start in real time
main:
	for {
		tReal = <-ticker.C
		//check for events and commands
	eventLoop:
		for {
			select {
			case ev := <-tm.events:
				tm.eventQueue.Add(ev)
			default:
				break eventLoop
			}
		}
		select {
		case cmd := <-tm.cmds:
			if cmd == "PAUSE" {
				if tm.state != RUNNING {
					tm.done <- false
				} else {
					tm.state = PAUSED
					tReal_pause = tReal
					tm.done <- true
				}
			} else if cmd == "STOP" {
				ticker.Stop()
				tm.state = STOPPED
				tm.done <- true
				break main
			} else if cmd == "RESUME" {
				if tm.state != PAUSED {
					tm.done <- false
				} else {
					tm.state = RUNNING
					tReal_offset += tReal.Sub(tReal_pause).Milliseconds()
					tm.done <- true
				}
			} else if strings.Fields(cmd)[0] == "SetSpeed" {
				speed, err := strconv.ParseFloat(strings.Fields(cmd)[1], 64)
				if err == nil {
					tm.speed = speed
					tm.done <- true
				} else {
					log.Printf("Could not convert the speed in command: %s", cmd)
					tm.done <- false
				}
			} else {
				log.Printf("Could not recognize command: %s", cmd)
			}
		default:
		}
		//run events from the "past" of the real time
		dtReal := tReal.Sub(tReal_start).Milliseconds() - tReal_offset //time passed in ms since start minus offset
		if tm.state == RUNNING {
			for {
				tNext, ok := tm.eventQueue.NextT()
				if !ok {
					break //no more events
				}
				if tNext < float64(dtReal)*tm.speed {
					tm.Step()
				} else {
					break
				}
			}
		}
	}
}

// Step processes the next event.
func (tm *TimeMachine) Step() {
	ev := tm.eventQueue.Next()
	tm.t = ev.T
	ev.F()
	//check for new events
	for {
		select {
		case ev := <-tm.events:
			tm.eventQueue.Add(ev)
		default:
			return
		}
	}

}
