package timer

import "time"

// Once create a timer that runs once.
func Once(d time.Duration, cb func()) Timer {
	return time.AfterFunc(d, cb)
}
