package timer

import (
	"time"

	"github.com/najoast/skynet/util"
)

type foreverTimer struct {
	ticker *time.Ticker
	stop   chan struct{}
}

// Stop prevents the Timer from firing.
func (t *foreverTimer) Stop() bool {
	t.stop <- struct{}{}
	return true
}

// Reset changes the timer to expire after duration d.
func (t *foreverTimer) Reset(d time.Duration) bool {
	t.ticker.Reset(d)
	return true
}

// Forever create a timer that runs forever.
func Forever(d time.Duration, cb func(), stop func(), pcall func(f func())) Timer {
	t := &foreverTimer{
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
