---
title: (翻译)取消多个Goroutines
date: 2017-06-28 13:06:04
tags:
    - Goroutine
    - channel
    - context
    - 并发控制
category:
    - Go学习小记
---

原文 https://chilts.org/2017/06/12/cancelling-multiple-goroutines

# Cancelling Multiple Goroutines 取消多协程
When Go was first released, there was a way to do some things in concurrency. As time has gone on, various things have changed. The Context package for one thing. :)

在Go刚开始发行的时候，只有一种并发的方法。随着时间流逝，很多事情都发生了变化。Context包就是其中之一。

This article doesn’t go into all of the ways of doing concurrency but will focus on one problem and take you through a few different solutions so you can see how things have evolved.

这篇文章不会探究所有并发的方法，但是会研究一个问题，并带你熟悉一些不同的解决方法，让你了解方法是如何推论出来的。

## The Problem
The problem I’d like to address here is being able to cancel multiple goroutines. There are many blog posts out there (I curate @CuratedGo, please follow) which show how to cancel just one goroutine, but my use-case was slightly more complicated. The rest of this article summarises my progress through getting this to work.

我在这里想提出的问题是取消对协程。有许多blog写了如何取消一个协程，但是我的用例是更加复杂的。下面的文章总结了我实现的过程。

The way we’re going decide when to quit is by listening for a C-c keypress. Of course at that point, we want to make sure we tidy up things nicely at that point. For example, if we’re currently streaming tweets from Twitter, we’d rather we told them we’re finished than just drop the connection.

我们要决定何时退出的方式是监听C-c按键。当然，我们确保我们在这一点上很好地处理事情。例如，如果我们目前正在推送Twitter的推文，我们要告诉他们我们已经完成，而不是放弃了解。

Let’s get started.
让我们开始吧

## A Main without Tidying up
``` go
package main

import (
    "fmt"
    "time"
)

func main() {
    ticker := time.NewTicker(3 * time.Second)

    for {
        select {
        case now := <-ticker.C:
            fmt.Printf("tick %s\n", now.UTC().Format("20060102-150405.000000000"))
        }
    }
}
```
And let’s run it and C-c it.
``` go
$ go run 01/tidy.go 
tick 20170612-213112.045887655
tick 20170612-213115.045986150
tick 20170612-213118.045993591
^Csignal: interrupt
```
Here you can see we have sent the interrupt signal. Make a mental note of that name. However, we haven’t actually tidied up the timer. There are a few ways we could do it, and the easiest for this program is to defer ticker.Stop() so it gets run at the end of main().

在这里你可以看到，我们发送了中断信号。在心里记住这个名字。但是我们还没有正确的整理好定时器。有许多方法可以做到，其中最简单的是defer tick.Stop(),他在main()的结尾运行。
``` go
package main

import (
    "fmt"
    "time"
)

func main() {
    ticker := time.NewTicker(3 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case now := <-ticker.C:
            fmt.Printf("tick %s\n", now.UTC().Format("20060102-150405.000000000"))
        }
    }
}

```
There is no discernable difference in the output, however you are being a good citizen. :)

在输出的结果中并没有明显的区别，但是你现在是一个良好市民.:)
``` go
$ go run 02/tidy.go 
tick 20170612-213456.385205269
tick 20170612-213459.385180852
tick 20170612-213502.385222563
^Csignal: interrupt
```
We said earlier that we want to run multiple goroutines and we want to listen for C-c, so let’s do the C-c first.
我们之前说过，我们想运行多协程并且我们想监听C-c，所以我们先完成C-c。

Using the os/signal package, we can tell Go to listen for (you guessed it) OS Signals such as os.Interrupt and os.Kill. Let’s see what that looks like:

使用 os/signal包，我们可以告诉Go监听操作系统的信号，例如os.Interrupt和os.Kill。让我们看一下吧。

``` go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "time"
)

func main() {
    ticker := time.NewTicker(3 * time.Second)
    defer ticker.Stop()

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)

    for {
        select {
        case now := <-ticker.C:
            fmt.Printf("tick %s\n", now.UTC().Format("20060102-150405.000000000"))
        case <-c:
            fmt.Println("Received C-c - shutting down")
            return
        }
    }
}
```
And when we run it, instead of seeing the default message Go provides when it receives an interrupt signal, we can see our own message:
``` go
$ go run 03/tidy.go 
tick 20170612-214602.313917282
tick 20170612-214605.313950300
tick 20170612-214608.313950904
^CReceived C-c - shutting down
```
Excellent, so let’s start moving the program closer to what we want - running multiple goroutines and stopping them cleanly

非常好，我们将程序改成更接近我们想要的--运行多协程并干净的结束他们

## Signalling a Goroutine to Stop 发送信号让协程停止
Even though we only have one task at the moment, we will put it into it’s own goroutine and signal it to stop when we have received the C-c. I’m going to use the first half of a post called “Stopping Goroutines” by the excellent Mat Ryer as the basis for this process. Note when this post was written - 2015 - and be sure we’ll change a few things by the time we’ve finished this article.

尽管我们此刻只有一个任务，但我们我们会将他放在他自己的协程中，并且当他接受到C-c信号的时候停止。

The next example shows the ticker in it’s own goroutine. Notice that instead of keeping the signal receiver in the for select case <-c we’ll just change it to <-c since that’s the only thing we’re going to leave in main(). I will prefix the messages with either main or tick so you can see what’s going on.

下一个例子展示了在它自己单独的协程里的ticker。注意，我们将改用<-c, 而不是将信号接受放在for select case <-c中。因为这是我们唯一放在main()中的语句.我将会在消息中加上main或者tick的前缀，让你看到哪一个正在执行。
``` go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "time"
)

func main() {
    // a channel to tell `tick()` to stop, and one to tell us they've stopped
    stopChan := make(chan struct{})
    stoppedChan := make(chan struct{})
    go tick(stopChan, stoppedChan)

    // listen for C-c
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    <-c
    fmt.Println("main: received C-c - shutting down")

    // tell the goroutine to stop
    fmt.Println("main: telling goroutines to stop")
    close(stopChan)
    // and wait for them to reply back
    <-stoppedChan
    fmt.Println("main: goroutine has told us they've finished")
}

func tick(stop, stopped chan struct{}) {
    // tell the caller we've stopped
    defer close(stopped)

    ticker := time.NewTicker(3 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case now := <-ticker.C:
            fmt.Printf("tick: tick %s\n", now.UTC().Format("20060102-150405.000000000"))
        case <-stop:
            fmt.Println("tick: caller has told us to stop")
            return
        }
    }
}
```
Once you press C-c here, you can see the exchange of messages.

一旦你按下了C-c，你可以看到消息的交换。

``` go
$ go run 04/tidy.go 
tick: tick 20170612-220018.345218301
tick: tick 20170612-220021.345202622
tick: tick 20170612-220024.345147172
^Cmain: received C-c - shutting down
main: telling goroutines to stop
tick: caller has told us to stop
main: goroutine has told us they've finished
```
So far so good. It works.
到现在为止还挺好。它起作用了。

But I can see one problem on the horizon. When we add another goroutine, we’ll have to create another stopped channel for the second goroutine to tell us when they’ve stopped. (Side-note: I originally also created a new stop chan too, but we can re-use that channel for both goroutines.)

但是我预见了一个问题。当我添加另外一个协程时，我们将为第二个协程创建另外一个stoped channel用于告诉我们协程何时停止。（边注：我原来也创建了一个新的stop channel，但是我们可以重新使用该channel用于这两个goroutine）

Let’s see what the extra stopped channel looks like. In this example our second goroutine tock() is very similar to the first, except it tocks every 5s instead of ticks every 3s.

让我们看看额外的stopped channel。在这个例子中，我们的第二个goroutine tock()和第一个非常相似，不同点是第二个是每隔5s而第一个是每隔3s。

``` go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "time"
)

func main() {
    // a channel to tell `tick()` and `tock()` to stop
    stopChan := make(chan struct{})

    // a channel for `tick()` to tell us they've stopped
    tickStoppedChan := make(chan struct{})
    go tick(stopChan, tickStoppedChan)

    // a channel for `tock()` to tell us they've stopped
    tockStoppedChan := make(chan struct{})
    go tock(stopChan, tockStoppedChan)

    // listen for C-c
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    <-c
    fmt.Println("main: received C-c - shutting down")

    // tell the goroutine to stop
    fmt.Println("main: telling goroutines to stop")
    close(stopChan)
    // and wait for them to reply back
    <-tickStoppedChan
    <-tockStoppedChan
    fmt.Println("main: all goroutines have told us they've finished")
}

func tick(stop, stopped chan struct{}) {
    // tell the caller we've stopped
    defer close(stopped)

    ticker := time.NewTicker(3 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case now := <-ticker.C:
            fmt.Printf("tick: tick %s\n", now.UTC().Format("20060102-150405.000000000"))
        case <-stop:
            fmt.Println("tick: caller has told us to stop")
            return
        }
    }
}

func tock(stop, stopped chan struct{}) {
    // tell the caller we've stopped
    defer close(stopped)

    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case now := <-ticker.C:
            fmt.Printf("tock: tock %s\n", now.UTC().Format("20060102-150405.000000000"))
        case <-stop:
            fmt.Println("tock: caller has told us to stop")
            return
        }
    }
}
```
It’s starting to look unwieldy. However, let’s take a look at the output for completeness:
``` go
$ go run 05/tidy.go 
tick: tick 20170612-220618.466725240
tock: tock 20170612-220620.466789888
tick: tick 20170612-220621.466756817
tick: tick 20170612-220624.466762771
^Cmain: received C-c - shutting down
main: telling goroutines to stop
tock: caller has told us to stop
tick: caller has told us to stop
main: all goroutines have told us they've finished
```
Even though it’s looking a bit nasty, it still works as it should.

## sync.WaitGroup
Let’s try and tidy-up and simplify a bit here. The reason to do this is because if we’d like to add another goroutine to this program - or indeed another 10, 20 or a hundred - we’re going to have a headache with all the channels we need to create.

让我们尝试、整理并简化。做这个的理由是因为如果我们想在这个程序中添加其他goroutine - 或者甚至其他10，20，或者100个 - 处理所有我们需要创建的channel会让我们头疼。

So instead of channels, let’s try another concurrency fundamental that Go provides, which is sync.WaitGroup. Here we create just one WaitGroup (instead of two channels) and use that for the goroutines to signal they’ve finished. Remember, once we create the WaitGroup we shouldn’t copy it, so we need to pass it by reference.

所以，除了channel，让我们尝试另外的Go提供的并发基本原则--sync.WaitGroup.在这里，我们仅仅创建一个WaitGroup(而不是两个channel)，用于当协程结束时发送信息。记住，一旦我们创建了WaitGroup，我们不能拷贝它，我们需要通过引用(指针？？？)传递。

``` go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "sync"
    "time"
)

func main() {
    // a channel to tell `tick()` and `tock()` to stop
    stopChan := make(chan struct{})

    // a WaitGroup for the goroutines to tell us they've stopped
    wg := sync.WaitGroup{}

    // a channel for `tick()` to tell us they've stopped
    wg.Add(1)
    go tick(stopChan, &wg)

    // a channel for `tock()` to tell us they've stopped
    wg.Add(1)
    go tock(stopChan, &wg)

    // listen for C-c
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    <-c
    fmt.Println("main: received C-c - shutting down")

    // tell the goroutine to stop
    fmt.Println("main: telling goroutines to stop")
    close(stopChan)
    // and wait for them both to reply back
    wg.Wait()
    fmt.Println("main: all goroutines have told us they've finished")
}

func tick(stop chan struct{}, wg *sync.WaitGroup) {
    // tell the caller we've stopped
    defer wg.Done()

    ticker := time.NewTicker(3 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case now := <-ticker.C:
            fmt.Printf("tick: tick %s\n", now.UTC().Format("20060102-150405.000000000"))
        case <-stop:
            fmt.Println("tick: caller has told us to stop")
            return
        }
    }
}

func tock(stop chan struct{}, wg *sync.WaitGroup) {
    // tell the caller we've stopped
    defer wg.Done()

    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case now := <-ticker.C:
            fmt.Printf("tock: tock %s\n", now.UTC().Format("20060102-150405.000000000"))
        case <-stop:
            fmt.Println("tock: caller has told us to stop")
            return
        }
    }
}

```
The output is exactly the same as the previous program, so we should be on the right lines. The program itself has a few lines removed, a few lines added and looks very similar, however adding new goroutines is a little simpler now. We just need call wg.Add(1) and pass both the stop channel and the waitgroup to it. As I said, it’s only a little simpler but that’s good, right?

输出完全和之前程序一样，所以我们应该在正确的行上。这个程序本身有些行删除了，一些行添加了，看起来差不多，但是添加新的goroutine比之前简单了。我们仅仅需要调用wg.Add(1)，传递stop通道和waitgroup给它。正如我所说的，仅仅简单一点点，但是很不错，对吧？！

``` go
$ go run 06/tidy.go 
tick: tick 20170612-221717.992723221
tock: tock 20170612-221719.992700713
tick: tick 20170612-221720.992722592
tick: tick 20170612-221723.992745407
^Cmain: received C-c - shutting down
main: telling goroutines to stop
tock: caller has told us to stop
tick: caller has told us to stop
main: all goroutines have told us they've finished
```
So far, so good. However, there is another problem on the horizon. Let’s imagine we want to also create a webserver in a goroutine. In the past we used to create one using the following code. The problem here though is that the server blocks the goroutine until it has finished.

到现在为止，都很好，但是，又有一个新的问题。让我们想象一下我们也想在goroutine中创建web服务器。过去我们常常使用以下的代码创建协程。而现在的问题是server直到它结束前，一直阻塞着协程。

``` go
package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Hello, World!")
    })
    http.ListenAndServe(":8080", nil)
}
```
So the question is, how do we also tell the web server to stop?
所以问题是，我们应该怎样告诉web服务器去停止运行。

## Context
In Go v1.7, the context package was added and that is our next secret. The ability to tell a webserver to stop using a context was also added. Using a Context has become the swiss-army knife of concurrency control in Go over the past few years (it used to live at https://godoc.org/golang.org/x/net/context but was moved into the standard library).

在Go的1.7版本加入了context包，这就是我们下一个秘密武器。使用context来停止webservver的运行的能力同样也具有。在Go的过去的几年中，使用Context变成并发控制的瑞士军刀。（它以前在 https://godoc.org/golang.org/x/net/context,但是现在已经移入标准库了）

Let’s have a very quick look at how we can create and cancel a Context:
让我们快速地了解一下如何创建和取消上下文:

``` go
    // create a context that we can cancel
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // pass this ctx to our goroutines - each of which would select on `<-ctx.Done()`
    go tick(ctx, ...)

    // sometime later ... in our case after a `C-c`
    cancel()
```
(Side note: if you haven’t see JustForFunc by Francesc Campoy yet, you should watch it - Francesc talks about the Context package in episodes 9 and 10.)

One major advantage of using a Context over a stop channel is if any of the goroutines are also creating other goroutines to do the work for them. In the case of using stopped channels we’d have to create more stop channels to tell the child goroutines to finish. We’d also have to tie much of this together to make it work. When we use a Context however, each goroutine would derive a Context from the one it was given, and each of them would be told to cancel.

使用Context比使用stop channel最主要的优势是，协程是否也创建其他协程来工作。因为使用stopped channels，我们不得不创建更多的stop channe，来让子协程终止。我们同样不得不将这些整合起来来让它起作用。但是当我们使用Context，每一个协程都会继承传递给它的Context，并且每一个(继承者)都被告之要删除。

Before we try adding a webserver, let’s change our example above to use a Context. The first thing we’ll need to do is pass the context to each goroutine instead of the channel. Instead of selecting on the channel, it’ll select on <-ctx.Done() and still signal back to main() when it has tidied up.

在我们尝试添加一个webserver之前,让我们修改我们上面使用Context例子。我要做的第一个事情就是将传递channel改为传递Context给每一个goroutine，将监听channel改为监听ctx.Done()，并且仍发送信号，当整理（tidy up）好协程之后返回main()

``` go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "sync"
    "time"
)

func main() {
    // create a context that we can cancel
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // a WaitGroup for the goroutines to tell us they've stopped
    wg := sync.WaitGroup{}

    // a channel for `tick()` to tell us they've stopped
    wg.Add(1)
    go tick(ctx, &wg)

    // a channel for `tock()` to tell us they've stopped
    wg.Add(1)
    go tock(ctx, &wg)

    // listen for C-c
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    <-c
    fmt.Println("main: received C-c - shutting down")

    // tell the goroutines to stop
    fmt.Println("main: telling goroutines to stop")
    cancel()

    // and wait for them both to reply back
    wg.Wait()
    fmt.Println("main: all goroutines have told us they've finished")
}

func tick(ctx context.Context, wg *sync.WaitGroup) {
    // tell the caller we've stopped
    defer wg.Done()

    ticker := time.NewTicker(3 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case now := <-ticker.C:
            fmt.Printf("tick: tick %s\n", now.UTC().Format("20060102-150405.000000000"))
        case <-ctx.Done():
            fmt.Println("tick: caller has told us to stop")
            return
        }
    }
}

func tock(ctx context.Context, wg *sync.WaitGroup) {
    // tell the caller we've stopped
    defer wg.Done()

    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case now := <-ticker.C:
            fmt.Printf("tock: tock %s\n", now.UTC().Format("20060102-150405.000000000"))
        case <-ctx.Done():
            fmt.Println("tock: caller has told us to stop")
            return
        }
    }
}
```
There is very little difference between this program and the previous one, however we now have the ability to:

这个程序与之前的有点不一样，我们现在有能力来：

1. create a webserver that we can cancel with the Context 
创建一个能使用Context取消的webserver


2. pass the same context to sub goroutines which will also cancel their work when told
And again, the output is the same. We must be doing something right.

传递相同的context到子协程，并取消协程。输出结果又一次相同。我们做对了。
``` go
$ go run 07/tidy.go 
tick: tick 20170612-223954.341894561
tock: tock 20170612-223956.341886006
tick: tick 20170612-223957.341887182
tick: tick 20170612-224000.341927373
^Cmain: received C-c - shutting down
main: telling goroutines to stop
tock: caller has told us to stop
tick: caller has told us to stop
main: all goroutines have told us they've finished
```
Now let’s get onto the beast and tell our program to also serve HTTP requests.

现在要实现我们的野望了。让我们的程序同样提供HTTP请求

## The Webserver
Before we show the entire program, let’s take a look at what the webserver goroutine would look like. The magic here is that instead of calling http.ListenAndServe() we explicitly create the webserver and by doing this we can eventually signal to it to stop. We’re going to model this on the excellent HTTP server connection draining section of this article by Tyler Christensen.

在我们展示全部程序之前，让我们看一下webserver的协程是什么样的。这里神奇的地方是，不是调用 http.ListenAndServe()而是明确的创建webserver并且通过这样做，我们能最终使用信号终止它。我们将要在这优秀的HTTP服务连接上构建这个方法

``` go
func server(ctx context.Context, wg *sync.WaitGroup) {
    // tell the caller that we've stopped
    defer wg.Done()

    // create a new mux and handler
    mux := http.NewServeMux()
    mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("server: received request")
        time.Sleep(3 * time.Second)
        io.WriteString(w, "Finished!\n")
        fmt.Println("server: request finished")
    }))

    // create a server
    srv := &http.Server{Addr: ":8080", Handler: mux}

    go func() {
        // service connections
        if err := srv.ListenAndServe(); err != nil {
            fmt.Printf("Listen : %s\n", err)
        }
    }()

    <-ctx.Done()
    fmt.Println("server: caller has told us to stop")

    // shut down gracefully, but wait no longer than 5 seconds before halting
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // ignore error since it will be "Err shutting down server : context canceled"
    srv.Shutdown(shutdownCtx)

    fmt.Println("server gracefully stopped")
}
```
For this func, the only two lines we added in main() were:

对于这个函数来说，只需main()添加两行就行了。

``` go
    // run `server` in it's own goroutine
    wg.Add(1)
    go server(ctx, &wg)
```
For the output of this program, I will send a request to the server curl localhost:8080 after the first tick and you should see the request start and finish either side of the 2nd tick. And as usual we’ll just show three ticks (and one tock):

输出程序的结果，我将在第一个tick后面向服务器发送请求，你应该看到请求在开始和结束在第二个tick的两边。和通常的一样，我们仅展示3个tick（和一个tock）

``` go
$ go run 08/tidy.go 
tick: tick 20170612-230003.228960866
server: received request
tock: tock 20170612-230005.228893119
tick: tick 20170612-230006.228868513
server: request finished
tick: tick 20170612-230009.228863351
^Cmain: received C-c - shutting down
main: telling goroutines to stop
server: caller has told us to stop
tick: caller has told us to stop
server gracefully stopped
tock: caller has told us to stop
main: all goroutines have told us they've finished
```
And as we expected the server also shut down correctly. This time though, I’ll send a request after the 2nd tick but C-c the server before the 3rd tick to demonstrate the server graefully shutting down.

和我们预期的一样，正确的退出了。这一次，我们在第二个tick后发出请求，但是在第三个tick前发送C-c，以此来证明服务器安全退出了。

``` go
$ go run 08/tidy.go 
tick: tick 20170612-230408.026717601
tock: tock 20170612-230410.026710464
tick: tick 20170612-230411.026700385
server: received request
^Cmain: received C-c - shutting down
main: telling goroutines to stop
tick: caller has told us to stop
tock: caller has told us to stop
server: caller has told us to stop
Listen : http: Server closed
server: request finished
server gracefully stopped
main: all goroutines have told us they've finished
```
Notice that both tick() and tock() finished first, then we had a couple of seconds where we waited for the webserver to finish it’s request and then finally shut down. In the previous example the server shut down when it wasn’t servicing any requests and the srv.ListenAndServe() didn’t return any error. In this example the server was servicing a request and returned the http: Server closed error which appeared above - after which the request finished message appeared to prove the request was still in progress. However, it did finish, the client received the response and everything shut down as expected.、

注意：tick()和tock()先结束，然后我们我们等待了一会服务器完成请求之后，最终退出了。之前的例子，当服务器没有处理任何请求的时候，且 the srv.ListenAndServe()没有返回任何错误，服务器（立即）退出。在这个例子中，服务器处理请求并返回了上面出现的http: Server closed错误 --  请求完成之后，出现的“request finished”信息证明了请求仍然在执行。但是，它的确结束了，客户端收到了响应，一切都像期望的那样结束。

``` go
$ curl localhost:8080
Finished!
```
And that’s it! I hope you’ve enjoyed following along in this rather long article, but I hope we demonstrated not just how to use a Context to cancel multiple goroutines, but also how the way we write concurrent Go programs has changed over the years. As with everything, there are many ways to do all of this and I’m sure I’ve missed some but I hope that has given you a taster to play with more concurrency and Context.


# 我的读（翻译？）后感
使用stopped channel等待所有的协程结束
``` go
for i := 0; i < count; i++{
    <-stoppedChan
} 
```
来等待所有的协程结束。具体的代码如下
``` go
    // a channel to tell `tick()` and `tock()` to stop
    stopChan := make(chan struct{})
    stoppedChan := make(chan struct{})
    count := 0

    count++
    go tick(stopChan, stoppedChan)

    count++
    go tock(stopChan, stoppedChan)

    // listen for C-c
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    <-c
    fmt.Println("main: received C-c - shutting down")

    // tell the goroutine to stop
    fmt.Println("main: telling goroutines to stop")
    close(stopChan)
    // and wait for them both to reply back

    for i := 0; i < count; i++{
        <-stoppedChan
    } 
    fmt.Println("main: all goroutines have told us they've finished")
```
确定很明显，子协程仍需要像父线程一样，创建stop channel，麻烦不止一点点。。

我的翻译真是一坨屎。。。。

context.Backgroud()是一个非nil空Context。没有值，没有deadline。经常用于main函数初始化，测试，作为请求的最高等级的Context。

server中的shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)的第一个返回值就是新的带时间截止控制的context。第二个返回值是执行时间到达的时，执行的函数。例如这样：
``` go
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	cancel = func() {
		fmt.Println("斩斩斩")
	}
```
主动创建服务器的过程
``` go
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("server: received request")
		time.Sleep(3 * time.Second)
		io.WriteString(w, "Finished!\n")
		fmt.Println("server: request finished")
	}))

	srv := &http.Server{Addr: ":8080", Handler: mux}

    go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil {
			fmt.Printf("Listen : %s\n", err)
		}
	}()

```
服务器关闭
``` go
	srv.Shutdown(shutdownCtx)
```
注意，此Shutdown方法，只有go的版本大于等于1.8的才有。