package timer

import (
	"time"

	"github.com/najoast/skynet/util"
)

type ForeverTimer struct {
	ticker *time.Ticker
	stop   chan struct{}
}

// Stop prevents the Timer from firing.
func (t *ForeverTimer) Stop() {
	t.stop <- struct{}{}
}

// Reset changes the timer to expire after duration d.
func (t *ForeverTimer) Reset(d time.Duration) {
	t.ticker.Reset(d)
}

// Forever create a timer that runs forever.
func Forever(d time.Duration, cb func(), stop func(), pcall func(f func())) *ForeverTimer {
	t := &ForeverTimer{
		ticker: time.NewTicker(d),
		stop:   make(chan struct{}),
	}

	if pcall == nil {
		pcall = util.Pcall
	}

	go func() {
		for {
			select {
			case <-t.ticker.C:
				pcall(cb)
			case <-t.stop:
				pcall(stop)
				t.ticker.Stop()
				return
			}
		}
	}()

	return t
}

type OnceTimer = time.Timer

// Once create a timer that runs once.
func Once(d time.Duration, cb func()) *OnceTimer {
	return time.AfterFunc(d, cb)
}
