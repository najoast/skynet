# 简介
一个用 Go 实现的简易 actor 框架，受[云风 skynet](https://github.com/cloudwu/skynet) 的启发。

# 为什么做这个框架
实现这个框架的原因是使用了很多年云风的 skynet，已经很熟悉这种并发方式了，在 Go 里没有找到很接近的实现，所以动手实现了这个简单但够用的框架。

云风的 skynet 里的 actor 可以看作是在一个独立线程里跑的，其中的代码是完全隔离的，其他 actor 即不能直接访问其内存，也不能直接调用其接口。actor 之间想要通信，必须通过发消息来进行。

而 Go 自身的 goroutine 并不能做到这种程度的隔离，每个 goroutine 比较类似 C++ 的线程，它们之间即可以直接访问内存，也可以直接调函数。这带来问题是代码的边界变得很模糊，可能一个文件内有的函数在这 A goroutine 跑，而其他函数在 B goroutine 跑。当代码变得复杂后，彼此之间加锁变成一件很困难的事情。

而 Actor 模式可以完全解决这类问题，每个 Actor 可以有清晰的代码边界。当然由于 Go 自身的灵活性，哪怕是使用本 Actor 框架，仍然可以写出边界模糊的代码，这需要写代码的人自己对自己做好约束，并在写之前就规划好每个 Actor 代码的位置。

# 该简易框架实现的功能
* Actor 的创建
* 往 Actor 内投递消息
* Actor 之间的 RPC
* 在 Actor 主协程内执行任意函数
* 创建在主协程内运行的定时器（包括单次执行和永久执行）

# Actor 主协程
每个 Actor 并不限定于只能有一个协程，但它们仍然有一个自己的主协程，该主协程处理所有外部发来的消息处理。这样虽然牺牲了一定的并发性，但大部分在主协程内运行的代码不用再加锁了。

在主协程之外仍然可以像写传统 Go 代码那样随意开协程，这就回到旧的书写代码的方式了，写代码的人自己要管理代码边界，在该加锁的地方加锁。

# Actor 消息分发函数
每个 Actor 都有一个消息分发函数，定义:
```go
// Dispatcher is the message dispatch function.
type Dispatcher func(msg *Message) Ack
```
当 Actor 收到消息后，会使用这个函数处理。
也可以不设置 Dispatcher，没有 Dispatcher 的 Actor 除了不能处理别人发给他的消息外，其他功能都是正常的，比如：
* Call 另一个 Actor
* 在 Actor 主协程内执行函数
* 创建在 Actor 主协程内执行回调的定时器

# Actor 入口函数
每个 Actor 在创建时都必须指定一个入口函数，该入口会在创建它的协程内立即执行。
一般需要在该函数内设置 `Dispatcher`。
当然也可以在该函数内直接调用 `Actor.Exit` 立即退出协程，这种就是一次性 Actor 了。

# Unique Actor
Unique Actor 是受框架管理的一种唯一 Actor，一个进程内同一个名字的 Actor 只能存在一个。
```go
actor := skynet.UniqueActor("actor_name", nil, func(actor *Actor) {
    actor.Logger.Info("main started!")
})
```

相对应的，有一种不受框架管理的 Actor，相同名字可以有多个。
```go
actor := skynet.NewActor("actor_name", nil, func(actor *Actor) {
    actor.Logger.Info("main started!")
})
```
哪怕是这种 Actor，也是有名字的，这主要是用于输出日志，因为 Go 并不能方便的获取到 goroutine id，所以只能通过这种方式来区分不同的 Actor。当然也可以给每个 Actor 生成一个唯一 ID，目前没这么做的原因是我认为这种 ID 没有可读性。由于 Actor 名字的主要目的是输出日志用的，所以一个不具有可读性的 ID 暂时先不考虑了。

# Actor RPC
为了简化实现，目前实现的 RPC 有两个限制：
1. 只能发生在 Actor 之间
2. 只能使用异步回调方式处理返回值

其实只要接收方提供一个 `chan` 用于接收返回值，是可以做到任意协程 RPC Actor 的，但这会增加使用负担，这就违背了我想实现一个 KISS 框架的本意。

# 使用方法
```go
// 创建一个简易数据库服务器 Actor
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

// 创建一个用来访问 simpledb:server 的 actor
client := UniqueActor("simpledb:client", nil, func(actor *Actor) {
    actor.Logger.Info("started")
})
defer client.Exit()

// 往server发一条消息，设置 key=hello, value=world
Send(server, "set", "hello", "world")

// RPC get("hello")，并输出返回结果
Call(client, server, func(ack Ack) {
    client.Logger.Infof("received server's ack: %v", ack)
}, "get", "hello")
```

更多使用方法请参考 [skynet_test.go](skynet_test.go)。
