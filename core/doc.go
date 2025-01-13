/*
Package core provides the core functionality for the discrete event simulation (DES).

The [EventQueue] stores events that shall happen at a certain time and retrieves them
(via [EventQueue.next]) in a strictly time ordered manner. (Order of events at equal time are arbitrary)

The [TimeMachine] runs a scenario by calling [EventQueue.next] in a loop and processing the retrieved events
which in turn may produce new events.

We distinguish two different notions of time:
  - The simulation time; This is the time attached to the event. It is a float64 and is interpreted
    as the amount of milliseconds since the start of the simulation.
  - The real time: This is the normal time that you can read from your watch.
*/
package core
