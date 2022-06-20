package skynet

import (
	"fmt"

	"github.com/najoast/skynet/log"
	"github.com/najoast/skynet/util"
)

// NewActor create a new actor.
//
// The created Actor must exit when not in use, otherwise there will be
// a goroutine leak.
func NewActor(name string, pcall func(f func()), main MainFunc) *Actor {
	if pcall == nil {
		pcall = util.Pcall
	}

	actor := &Actor{
		ch:           make(chan *Message),
		session:      0,
		sess2AckCb:   make(map[int]AckCb),
		sess2TimerCb: make(map[int]func()),
		pcall:        pcall,
		name:         name,
		exited:       false,
		Logger: &log.DefaultLogger{
			Level:  log.Debug,
			Prefix: fmt.Sprintf("[actor %s] ", name),
		},
		Timer: &SkynetTimer{},
	}
	actor.Timer.actor = actor

	go actor.dispatch()
	pcall(func() { main(actor) })

	return actor
}

// Call in "request-response" mode to let the source Actor call the function of the target Actor,
// and pass the return value into cb and execute it in the source main coroutine.
func Call(source, target *Actor, cb AckCb, fname string, args ...interface{}) error {
	if source != nil && target != nil {
		source.call(target, cb, fname, args...)
		return nil
	}
	return fmt.Errorf("Actor not found")
}

// Send unidirectionally sends a message to the target Actor.
func Send(target *Actor, fname string, args ...interface{}) error {
	return sendTo(target, &Message{
		fname: fname,
		args:  args,
		typ:   messageTypeSend,
	})
}

// sendTo send msg to target.
func sendTo(target *Actor, msg *Message) error {
	if target.exited {
		return fmt.Errorf("Actor %s has exited", target.name)
	}
	target.ch <- msg
	return nil
}

// All Unique Actors
var uniqueActors map[string]*Actor = make(map[string]*Actor)

// UniqueActor create a unique Actor, if it already exists, return it directly.
func UniqueActor(name string, pcall func(f func()), main MainFunc) *Actor {
	if actor, exist := uniqueActors[name]; exist {
		return actor
	}
	actor := NewActor(name, pcall, main)
	uniqueActors[name] = actor
	return actor
}

// SendByName calls Send directly with the name.
func SendByName(name string, fname string, args ...interface{}) error {
	actor, exist := uniqueActors[name]
	if !exist {
		return fmt.Errorf("Actor %s not found", name)
	}
	Send(actor, fname, args)
	return nil
}

// CallByName calls Call directly with the name.
func CallByName(source, target string, cb AckCb, fname string, args ...interface{}) error {
	return Call(uniqueActors[source], uniqueActors[target], cb, fname, args...)
}
