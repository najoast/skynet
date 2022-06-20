package skynet

import "fmt"

// newSessionId create a new SessionId, don't worry about overflow,
// as long as it doesn't repeat in a short time.
func (actor *Actor) newSessionId() int {
	actor.session++
	return actor.session
}

// Call asynchronous request + response, non-blocking, the response
// is executed through the callback function.
func (actor *Actor) Call(target *Actor, cb AckCb, fname string, args ...interface{}) error {
	sessionId := actor.newSessionId()
	actor.sess2AckCb[sessionId] = cb

	err := sendTo(target, &Message{
		Fname:     fname,
		Args:      args,
		typ:       messageTypeCallReq,
		ackChan:   actor.ch,
		sessionId: sessionId,
	})

	if err != nil {
		delete(actor.sess2AckCb, sessionId)
	}
	return err
}

func (actor *Actor) CallByName(name string, cb AckCb, fname string, args ...interface{}) error {
	if target, exist := uniqueActors[name]; exist {
		return actor.Call(target, cb, fname, args...)
	} else {
		return fmt.Errorf("Actor %s not found", name)
	}
}

// SetDispatcher set the actor's message dispatch function.
func (actor *Actor) SetDispatcher(dispatcher Dispatcher) {
	actor.dispatcher = dispatcher
}

func (actor *Actor) checkDispatcher(msg *Message) Dispatcher {
	if actor.dispatcher == nil && actor.Logger != nil {
		actor.Logger.Warnf("Does'n have a Dispatcher, but received a %v message: %s(%v)",
			msg.typ, msg.Fname, msg.Args)
	}
	return actor.dispatcher
}

// dispatch is the main goroutine of the Actor.
func (actor *Actor) dispatch() {
	for {
		if actor.exited {
			break
		}

		msg := <-actor.ch
		actor.pcall(func() {
			switch msg.typ {
			case messageTypeSend:
				if dispatcher := actor.checkDispatcher(msg); dispatcher != nil {
					dispatcher(msg)
				}

			case messageTypeCallReq:
				if dispatcher := actor.checkDispatcher(msg); dispatcher != nil {
					ack := dispatcher(msg)
					if msg.ackChan != nil {
						msg.ackChan <- &Message{
							typ:       messageTypeCallAck,
							ack:       ack,
							sessionId: msg.sessionId,
						}
					}
				}

			case messageTypeCallAck:
				if cb, exist := actor.sess2AckCb[msg.sessionId]; exist {
					defer delete(actor.sess2AckCb, msg.sessionId)
					cb(msg.ack)
				}

			case messageTypeOnceTimer:
				if cb, exist := actor.sess2TimerCb[msg.sessionId]; exist {
					defer delete(actor.sess2TimerCb, msg.sessionId)
					cb()
				}

			case messageTypeForeverTimer:
				if cb, exist := actor.sess2TimerCb[msg.sessionId]; exist {
					cb()
				}
			}
		})
	}
}

// Exit an Actor, after exiting, you can no longer send messages to the Actor,
// and the Actor main coroutine will also exit.
func (actor *Actor) Exit() {
	actor.exited = true
}

// Run function f inside the main actor goroutine.
func (actor *Actor) Run(f func()) {
	actor.Timer.Once(0, f)
}

func (actor *Actor) GetName() string {
	return actor.name
}

func (actor *Actor) IsExited() bool {
	return actor.exited
}
