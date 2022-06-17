# Introduction
A simple actor framework implemented in Go, inspired by [cloudwu's skynet](https://github.com/cloudwu/skynet).

# Why do this framework
The reason for implementing this framework is that I have used [cloudwu's skynet](https://github.com/cloudwu/skynet) for many years. I am already familiar with this concurrency method. I have not found a very close implementation in Go, so I have implemented this simple but sufficient framework.

The actors in cloudwu's skynet can be regarded as running in an independent thread, the code in it is completely isolated, and other actors cannot directly access its memory or call its interface directly. For actors to communicate, they must send messages.

Go's own goroutines cannot achieve this level of isolation. Each goroutine is similar to a C++ thread, and can directly access memory or call functions between them. This brings the problem that the boundaries of the code become very blurred. It is possible that some functions in a file run in this A goroutine, while other functions run in B goroutine. When the code becomes complex, locking each other becomes a very difficult thing.

The Actor pattern can completely solve this kind of problem, and each actor can have a clear code boundary. Of course, due to the flexibility of Go itself, even using this Actor framework, it is still possible to write code with ambiguous boundaries. This requires the person who writes the code to constrain himself and plan the location of each Actor code before writing.

# Functions implemented by this simple framework
* Creation of Actor
* Deliver messages to Actors
* RPC between Actors
* Execute arbitrary functions within the Actor's main goroutine
* Create timers that run in the main goroutine (including single execution and permanent execution)

# Actor main goroutine
Each actor is not limited to only one goroutine, but they still have their own main goroutine, which handles all external message processing. Although this sacrifices a certain degree of concurrency, most of the code running in the main goroutine does not need to be locked again.

Outside the main goroutine, you can still open the goroutine at will like writing traditional Go code. This returns to the old way of writing code. The person who writes the code must manage the code boundary and lock the place where the lock should be locked.

# Actor message dispatch function
Each actor has a message dispatch function that defines:
````go
// Dispatcher is the message dispatch function.
type Dispatcher func(msg *Message) Ack
````
When the actor receives the message, it will use this function to process it.
You can also not set the Dispatcher. Actors without Dispatcher have normal functions except that they cannot process messages sent to them by others, such as:
* Call another Actor
* Execute function inside Actor main goroutine
* Create a timer that executes callbacks within the actor's main goroutine

# Actor entry function
Every Actor must specify an entry point function when it is created, which is executed immediately within the goroutine that created it.
Usually you need to set `Dispatcher` inside this function.
Of course, you can also directly call `Actor.Exit` in this function to exit the goroutine immediately, which is a one-time Actor.

# Unique Actor
Unique Actor is a unique Actor managed by the framework, and only one Actor with the same name can exist in a process.
````go
actor := skynet.UniqueActor("actor_name", nil, func(actor *Actor) {
    actor.Logger.Info("main started!")
})
````

Correspondingly, there is an Actor that is not managed by the framework, and there can be multiple actors with the same name.
````go
actor := skynet.NewActor("actor_name", nil, func(actor *Actor) {
    actor.Logger.Info("main started!")
})
````
Even this kind of actor has a name, which is mainly used for outputting logs, because Go cannot easily obtain the goroutine id, so it can only distinguish different actors in this way. Of course it is also possible to generate a unique ID for each actor, the reason for not doing this at the moment is that I don't think such IDs are readable. Since the main purpose of the actor name is to output the log, an unreadable ID will be ignored for the time being.

# Actor RPCs
To simplify implementation, the currently implemented RPC has two limitations:
1. Can only happen between Actors
2. Only use asynchronous callbacks to process return values

In fact, as long as the receiver provides a `chan` for receiving the return value, any goroutine RPC Actor can be implemented, but this will increase the usage burden, which goes against the original intention of implementing a KISS framework.

# Instructions
````go
// Create a simple database server Actor
server := UniqueActor("simpledb:server", nil, func(actor *Actor) {
    db := make(map[interface{}]interface{})
    actor.setDispatcher(func(msg *Message) Ack {
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

// Create an actor to access simpledb:server
client := UniqueActor("simpledb:client", nil, func(actor *Actor) {
    actor.Logger.Info("started")
})
defer client.Exit()

// Send a message to the server, set key=hello, value=world
Send(server, "set", "hello", "world")

// RPC get("hello"), and output the return result
Call(client, server, func(ack Ack) {
    client.Logger.Infof("received server's ack: %v", ack)
}, "get", "hello")
````

For more usage methods, please refer to [skynet_test.go](skynet_test.go).