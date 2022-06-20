package skynet

// newSessionId create a new SessionId, don't worry about overflow,
// as long as it doesn't repeat in a short time.
func (actor *Actor) newSessionId() int {
	actor.session++
	return actor.session
}

// call asynchronous request + response, non-blocking, the response
// is executed through the callback function.
func (actor *Actor) call(target *Actor, cb AckCb, fname string, args ...interface{}) error {
	sessionId := actor.newSessionId()
	actor.sess2AckCb[sessionId] = cb

	err := sendTo(target, &Message{
		fname:     fname,
		args:      args,
		typ:       messageTypeCallReq,
		ackChan:   actor.ch,
		sessionId: sessionId,
	})

	if err != nil {
		delete(actor.sess2AckCb, sessionId)
	}
	return err
}

// SetDispatcher set the actor's message dispatch function.
func (actor *Actor) SetDispatcher(dispatcher Dispatcher) {
	actor.dispatcher = dispatcher
}

func (actor *Actor) checkDispatcher(msg *Message) Dispatcher {
	if actor.dispatcher == nil && actor.Logger != nil {
		actor.Logger.Warnf("Does'n have a Dispatcher, but received a %v message: %s(%v)",
			msg.typ, msg.fname, msg.args)
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
					cb(msg.ack)
					delete(actor.sess2AckCb, msg.sessionId)
				}

			case messageTypeOnceTimer:
				if cb, exist := actor.sess2TimerCb[msg.sessionId]; exist {
					cb()
					delete(actor.sess2TimerCb, msg.sessionId)
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
