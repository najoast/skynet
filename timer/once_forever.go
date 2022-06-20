package timer

import "time"

type onceForeverTimer struct {
	timer Timer
}

func (t *onceForeverTimer) Stop() bool {
	return t.timer.Stop()
}

func (t *onceForeverTimer) Reset(d time.Duration) bool {
	return t.timer.Reset(d)
}

func OnceForever(once, loop time.Duration, cb func(), stop func(), pcall func(f func())) Timer {
	ret := &onceForeverTimer{}
	ret.timer = Once(once, func() {
		pcall(cb)
		ret.timer = Forever(loop, cb, stop, pcall)
	})
	return ret
}
