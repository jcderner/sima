package core

import (
	"log"
	"strconv"
	"strings"
	"time"
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

// TimeMachine (TM) is the core of the simulation.
//Events are fed into the events channel and are processed, once the TM is started.
//Take care to provide the events with sufficient capacity (s. eventChanSize in [NewTimeMachine]).
//
//Furthermore there is a cmds channel to control the TM.
//The following commands are recognized:
//  - "PAUSE": Pauses the simulation.
//  - "RESUME": Resumes the simulation.
//  - "STOP": Stops the simulation.
//  - "SetSpeed <speed>": Sets the speed of the simulation.
//
//The commands are put into the cmds channel by the respective methods.

type TimeMachine struct {
	eventQueue *EventQueue
	t          float64 //the actual simulation time in ms.
	cycleTime  int     //every cycleTime [ms] the time machine will check for new events.
	state      TMState
	speed      float64 //ratio between simulation and real time.
	cmds       chan string
	events     chan *Event
}

// NewTimeMachine creates a reference to a new TimeMachine.
// The speed is the ratio between simulation and real time.
// The eventChanSize is the capacity of the events channel. It should be at least the number of events that are expected to be scheduled initially or in one cycle.
// The cycleTime is the time in ms between two checks for new commands and events.
func NewTimeMachine(speed float64, eventChanSize int, cycleTime int) *TimeMachine {
	return &TimeMachine{
		eventQueue: NewEventQueue(),
		t:          0,
		cycleTime:  cycleTime,
		state:      IDLE,
		speed:      speed,
		cmds:       make(chan string),
		events:     make(chan *Event, eventChanSize),
	}
}

// T returns the current simulation time of the time machine.
func (tm *TimeMachine) T() float64 {
	return tm.t
}

func (tm *TimeMachine) State() TMState {
	return tm.state
}

func (tm *TimeMachine) Speed() float64 {
	return tm.speed
}

// Schedule schedules a function f to be executed at dt after the current simulation time.
// If dt is negative, it will be reset to 0.0.
func (tm *TimeMachine) Schedule(dt float64, f func()) {
	if dt < 0 {
		log.Printf("dt = %v < 0. Will be reset to 0.0", dt)
		dt = 0.0
	}
	tm.events <- (&Event{tm.t + dt, f})
}

func (tm *TimeMachine) Pause() {
	tm.cmds <- "PAUSE"
}

func (tm *TimeMachine) Resume() {
	tm.cmds <- "RESUME"
}

func (tm *TimeMachine) Stop() {
	tm.cmds <- "STOP"
}

// SetSpeed sets the speed of the simulation.
// If the speed is smaller than 0.01 then it will be reset to 0.01.
func (tm *TimeMachine) SetSpeed(speed float64) {
	if speed < 0.01 {
		log.Printf("Speed = %v must not be smaller than 0.01. It will be reset to 0.01.", speed)
		speed = 0.01
	}
	tm.cmds <- "SetSpeed " + strconv.FormatFloat(speed, 'f', -1, 64)
}

// Start starts the simulation.
// It returns true if the simulation was started successfully.
func (tm *TimeMachine) Start() (success bool) {
	if tm.state != IDLE {

		return false
	}
	tm.state = RUNNING
	go tm.run()
	return true
}

func (tm *TimeMachine) run() {
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
				tm.eventQueue.add(ev)
			default:
				break eventLoop
			}
		}
	cmdLoop:
		for {
			select {
			case cmd := <-tm.cmds:
				if cmd == "PAUSE" {
					if tm.state != RUNNING {
						log.Printf("TimeMachine is not running. Ignoring PAUSE command.")
					} else {
						tm.state = PAUSED
						tReal_pause = tReal
					}
				} else if cmd == "STOP" {
					ticker.Stop()
					tm.state = STOPPED
					break main
				} else if cmd == "RESUME" {
					if tm.state != PAUSED {
						log.Printf("TimeMachine is not paused. Ignoring RESUME command.")
					} else {
						tm.state = RUNNING
						tReal_offset += tReal.Sub(tReal_pause).Milliseconds()
					}
				} else if strings.Fields(cmd)[0] == "SetSpeed" {
					speed, err := strconv.ParseFloat(strings.Fields(cmd)[1], 64)
					if err == nil {
						tm.speed = speed
					} else {
						log.Printf("Could not convert the speed in command: %s", cmd)
					}
				} else {
					log.Printf("Could not recognize command: %s", cmd)
				}
			default:
				break cmdLoop
			}
		}
		//run events from the "past" of the real time
		dtReal := tReal.Sub(tReal_start).Milliseconds() - tReal_offset //time passed in ms since start minus offset
		if tm.state == RUNNING {
			for {
				tNext, ok := tm.eventQueue.nextT()
				if !ok {
					return
				}
				if tNext < float64(dtReal)*tm.speed {
					//log.Printf("t: %v, tNext: %v,  dtReal: %v, tReal: %v", tm.t, tNext, dtReal, tReal)
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
	ev := tm.eventQueue.next()
	tm.t = ev.t
	ev.f()
	//check for new events
	for {
		select {
		case ev := <-tm.events:
			tm.eventQueue.add(ev)
		default:
			return
		}
	}

}
