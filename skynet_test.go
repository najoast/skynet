package skynet

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestHelloWorld(t *testing.T) {
	actor := UniqueActor("hello", nil, func(actor *Actor) {
		actor.Logger.Info("main started!")

		actor.SetDispatcher(func(msg *Message) Ack {
			switch msg.fname {
			case "hello":
				actor.Logger.Infof("hello: %v", msg.args...)
			default:
				actor.Logger.Infof("Unhandled fname %s(%v)!", msg.fname, msg.args)
			}
			return nil
		})

		actor.Logger.Debug("main finished")
	})

	Send(actor, "hello", "world")
	Send(actor, "xxxxx", "a", "b", 1, 2, 3, []int{7, 8, 9})
	SendByName("hello", "hello", "skynet")

	for i := 0; i < 10; i++ {
		if err := Send(actor, "hello", i); err != nil {
			actor.Logger.Error(err)
		}
	}

	actor.Exit()

	for i := 0; i < 10; i++ {
		if err := Send(actor, "hello", i); err != nil {
			actor.Logger.Error(err)
		}
	}

	time.Sleep(time.Second)
}

func TestCall(t *testing.T) {
	server := UniqueActor("simpledb:server", nil, func(actor *Actor) {
		db := make(map[interface{}]interface{})
		actor.SetDispatcher(func(msg *Message) Ack {
			switch msg.fname {
			case "set":
				actor.Logger.Debugf("set %v %v", msg.args...)
				db[msg.args[0]] = msg.args[1]
			case "get":
				return Ack{db[msg.args[0]]}
			default:
				fmt.Printf("Unhandled fname %s(%v)!\n", msg.fname, msg.args)
			}
			return nil
		})
	})
	defer server.Exit()

	client := UniqueActor("simpledb:client", nil, func(actor *Actor) {
		actor.Logger.Info("started")
	})
	defer client.Exit()

	Send(server, "set", "hello", "world")
	Call(client, server, func(ack Ack) {
		client.Logger.Infof("received server's ack: %v", ack)
	}, "get", "hello")

	client.Logger.Debug("-----------------------------------")

	for i := 0; i < 10; i++ {
		Send(server, "set", i, i*i)
		Call(client, server, func(ack Ack) {
			client.Logger.Infof("received server's ack: %v", ack)
		}, "get", i)
	}

	time.Sleep(time.Second)
}

func TestTimerOnce(t *testing.T) {
	actor := UniqueActor("test_timer", nil, func(actor *Actor) {
		actor.Logger.Info("started")
	})
	defer actor.Exit()

	actor.Run(func() {
		actor.Logger.Debug("Run")
	})

	actor.Timer.Once(time.Second*5, func() {
		actor.Logger.Debug("TimerOnce")
	})

	time.Sleep(time.Second)
}

func TestTimerForever(t *testing.T) {
	actor := UniqueActor("test_timer", nil, func(actor *Actor) {
		actor.Logger.Info("started")
	})
	defer actor.Exit()

	wg := sync.WaitGroup{}
	wg.Add(10)

	actor.Timer.Forever(time.Second*2, func() {
		wg.Done()
		actor.Logger.Debug("TimerForever")
	})

	wg.Wait()
}

func TestTimerOnceForever(t *testing.T) {
	actor := UniqueActor("test_timer", nil, func(actor *Actor) {
		actor.Logger.Info("started")
	})
	defer actor.Exit()

	wg := sync.WaitGroup{}
	wg.Add(5)

	actor.Timer.OnceForever(time.Second, time.Second*3, func() {
		wg.Done()
		actor.Logger.Debug("TimerOnceForever")
	})

	wg.Wait()
}
