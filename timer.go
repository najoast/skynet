package skynet

import (
	"fmt"
	"time"

	"github.com/najoast/skynet/timer"
)

// SkynetTimer means SkyNet TIMER
type SkynetTimer struct {
	actor *Actor
}

// Once create a timer that executes only once.
func (snt *SkynetTimer) Once(d time.Duration, f func()) (timer.Timer, error) {
	if d < 0 {
		return nil, fmt.Errorf("invalid duration")
	}

	sessionId := snt.actor.newSessionId()
	snt.actor.sess2TimerCb[sessionId] = f

	if d == 0 {
		sendTo(snt.actor, &Message{
			typ:       messageTypeOnceTimer,
			sessionId: sessionId,
		})
		return nil, nil
	} else {
		return timer.Once(d, func() {
			sendTo(snt.actor, &Message{
				typ:       messageTypeOnceTimer,
				sessionId: sessionId,
			})
		}), nil
	}
}

// Forever create a timer that runs forever.
func (snt *SkynetTimer) Forever(d time.Duration, f func()) (timer.Timer, error) {
	return snt.newForeverTimer(0, d, f)
}

// OnceForever creates a timer that first runs once by the "Once" duration, then runs
// forever by the "Forever" duration.
func (snt *SkynetTimer) OnceForever(once, loop time.Duration, f func()) (timer.Timer, error) {
	return snt.newForeverTimer(once, loop, f)
}

func (snt *SkynetTimer) newForeverTimer(once, loop time.Duration, f func()) (timer.Timer, error) {
	if loop <= 0 {
		return nil, fmt.Errorf("invalid duration")
	}

	sessionId := snt.actor.newSessionId()
	snt.actor.sess2TimerCb[sessionId] = f

	var ret timer.Timer

	cb := func() {
		err := sendTo(snt.actor, &Message{
			typ:       messageTypeForeverTimer,
			sessionId: sessionId,
		})
		if err != nil && snt != nil {
			ret.Stop()
		}
	}

	stop := func() {
		delete(snt.actor.sess2TimerCb, sessionId)
	}

	if once > 0 {
		ret = timer.OnceForever(once, loop, cb, stop, snt.actor.pcall)
	} else {
		ret = timer.Forever(loop, cb, stop, snt.actor.pcall)
	}
	return ret, nil
}
