/*
Sima provides a simple discrete event simulation (DES).
Events are scheduled as parameterless functions without return value at certain simulation time in a TimeMachine (TM).

We distinguish two different notions of time:
  - The real time: This is the normal time that can be read from the wall clock.
  - The simulation time; This is the time attached to the event. It is a float64 and is interpreted
    as the amount of milliseconds since the start of the simulation.
    The simulation time flows (roughly) by the factor [TimeMachine.Speed] faster than the real time.

The TM - once started - executes the scheduled function in a strictly time ordered manner.

Alternatively and for fastest processing the client may call [TimeMachine.Step] in a loop, which completely ignores the real time.
*/
package sima
