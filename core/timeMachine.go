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
)

var tmstateNames = map[TMState]string{
	IDLE:    "IDLE",
	PAUSED:  "PAUSED",
	RUNNING: "RUNNING",
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

func (tm *TimeMachine) T() float64 {
	return tm.t
}

func (tm *TimeMachine) State() TMState {
	return tm.state
}

func (tm *TimeMachine) Speed() float64 {
	return tm.speed
}

func (tm *TimeMachine) Schedule(dt float64, f func()) {
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

func (tm *TimeMachine) SetSpeed(speed float64) {
	tm.cmds <- "SetSpeed " + strconv.FormatFloat(speed, 'f', -1, 64)
}

func (tm *TimeMachine) Start() (success bool) {
	if tm.state != IDLE {

		return false
	}
	tm.state = RUNNING
	tm.run()
	return true
}

func (tm *TimeMachine) run() {
	tick := (time.Duration)(1000000 * tm.cycleTime) //cycleTime ms.
	t_start := tm.t
	tReal_start := time.Now() //the start in real time
	var tReal time.Time
	ticker := time.NewTicker(tick)
main:
	for {
		tReal = <-ticker.C
		dtReal := tReal.Sub(tReal_start).Milliseconds() //ms
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
	cmdLoop:
		for {
			select {
			case cmd := <-tm.cmds:
				if cmd == "PAUSE" {
					tm.state = PAUSED
					ticker.Stop()
				} else if cmd == "STOP" {
					ticker.Stop()
					break main
				} else if cmd == "RESUME" {
					tm.state = RUNNING
					t_start = tm.t
					tReal_start = time.Now()
					ticker.Reset(tick)
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
		if tm.state == RUNNING {
			for {
				tNext, ok := tm.eventQueue.NextT()
				if !ok {
					break
				}
				if (tNext - t_start) < float64(dtReal)*tm.speed {
					tm.Step()
				} else {
					break
				}
			}
		}
	}
}

func (tm *TimeMachine) Step() {
	ev := tm.eventQueue.Next()
	tm.t = ev.t
	ev.f()
}
