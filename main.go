package main

import (
	"fmt"

	"github.com/jcderner/sima/core"
)

func main() {
	fmt.Println("Starting the time machine...")
	startTimeMachine()
}

func startTimeMachine() {
	tm := core.NewTimeMachine(1.0, 100, 10)
	p := NewPinger(tm)
	tm.Schedule(2000.0, p.pong)
	tm.Schedule(1000.0, p.ping)
	tm.Start()
}

type Pinger struct {
	tm *core.TimeMachine
}

func NewPinger(tm *core.TimeMachine) *Pinger {
	return &Pinger{tm: tm}
}
func (p Pinger) ping() {
	fmt.Println(p.tm.T(), ": Ping!")
	p.tm.Schedule(2000.0, p.pong)
}
func (p Pinger) pong() {
	fmt.Println(p.tm.T(), ": Pong!")
	p.tm.Schedule(2000.0, p.ping)
}
