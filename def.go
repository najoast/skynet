package skynet

import "github.com/najoast/skynet/log"

// Each message sent to an actor has a message type, which is used
// to do different logic processing after receiving the message.
type messageType int

const (
	// A message sent directly to the Actor, with no return value.
	messageTypeSend messageType = iota
	// "request" in request-response mode.
	messageTypeCallReq
	// "response" in request-response mode.
	messageTypeCallAck
	// OnceTimer's message.
	messageTypeOnceTimer
	// ForeverTimer's message.
	messageTypeForeverTimer
)

// Message is a message sent to the Actor.
type Message struct {
	fname     string
	args      []interface{}
	typ       messageType
	ackChan   chan *Message
	ack       Ack
	sessionId int
}

// Ack is the "response content" in request-response mode.
type Ack []interface{}

// AckCb is a callback function in request-response mode.
type AckCb func(Ack)

// Dispatcher is the message dispatch function.
type Dispatcher func(msg *Message) Ack

// Actor is the actor object.
type Actor struct {
	ch           chan *Message
	session      int
	sess2AckCb   map[int]AckCb
	sess2TimerCb map[int]func()
	pcall        func(f func())
	dispatcher   Dispatcher
	name         string
	exited       bool
	Logger       log.Logger
	Timer        *SkynetTimer
}

// MainFunc is the entry function executed when the Actor is initialized.
type MainFunc func(actor *Actor)
