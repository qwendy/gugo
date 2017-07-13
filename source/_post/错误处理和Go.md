---
title: Error handling and Go (翻译)
date: 2016-12-05 19:32:38
tag:
  - 翻译
  - 错误处理
category:
  - Go学习小记
---
本篇文章来自《The Go Blog》。文章地址为https://blog.golang.org/error-handling-and-go
下面是在下的翻译。因为看英文看完就忘，所以翻译翻译，聊以自慰。

## 错误处理和Go
2011年7月12日

### 引言
如果你写过Go的代码，你可能已经遇到了内置的error类型。Go的代码使用error表示异常状态。例如os.Open函数打开文件失败的时候，返回一个不为nil的error值。
``` golang
func Open(name string) (file *File, err error)
```
下面的代码使用os.Open打开一个文件。如果出现一个error,它会调用log.Fatal来打印出错误结束。
``` golang
f, err := os.Open("filename.ext")
if err != nil {
    log.Fatal(err)
}
// do something with the open *File f
```
在Go中，你仅需要知道这些关于error类型的知识，就能做很多事情了，但是在这篇文章中我们将仔细研究error，并讨论在Go中的一些好的错误处理的实例
<!-- more -->

### error 类型
error类型是一个接口(interface)类型。一个error变量代表任意一个能将自己描绘成字符串的值。
这个是接口的定义
``` golang
type error interface {
    Error() string
}
```
error类型如同其他所有内置类型是在universe block([The universe block encompasses all Go source text.](https://golang.org/ref/spec#Blocks))中预声明（[predeclare](https://golang.org/ref/spec#Predeclared_identifiers)）的。
最常使用的error实例是error包中未导出（export）的errorString类型。
``` golang
// errorString is a trivial implementation of error.
type errorString struct {
    s string
}

func (e *errorString) Error() string {
    return e.s
}
```
你可以使用errors.New函数来构建一个这样的值。它接受一个字符串，将之转成errors.errorString，并返回error值
``` golang
// New returns an error that formats as the given text.
func New(text string) error {
    return &errorString{text}
}
```
这是你可能会使用的errors.New的情况：
``` golang
func Sqrt(f float64) (float64, error) {
    if f < 0 {
        return 0, errors.New("math: square root of negative number")
    }
    // implementation
}
```
使用一个负数参数来调用Sqrt，会得到一个非nil的error值（具体的表示是一个errors.errorString值）。调用者能通过`error`类型的Error方法获取错误字符串("math: square root of...")，或者就直接输出它：
``` golang
f, err := Sqrt(-1)
if err != nil {
    fmt.Println(err)
}
```
fmt包通过调用它的Error()方法格式化error类型的值。
这个是错误的实现总结上下文的方法。错误应该由os.Open格式化返回为"open /etc/passwd: permission denied," ,而不仅仅是"permission denied."由Sqrt返回的错误缺少了关于无效参数的信息。
在fmt包中有个有用的函数Errorf,用于增加这个信息。它根据 `Printf`规则来格式化字符串，并返回一个由errors.New创建的error类型。
``` golang
if f < 0 {
    return 0, fmt.Errorf("math: square root of negative number %g", f)
}
```
在许多情况下fmt.Errorf已经足够 ，但是既然error是一个interface，你可以使用任意的数据结构作为error，让调用者来检查错误的详细信息。
举个例子，我们假想的调用也许想重新使用这个无效参数来调用Sqrt。我们能通过定义一个新的error的实现完成这个目标，而不是使用errors.errorString：
``` golang
type NegativeSqrtError float64

func (f NegativeSqrtError) Error() string {
    return fmt.Sprintf("math: square root of negative number %g", float64(f))
}
```
复杂的调用可以使用类型断言（[type assertion](https://golang.org/ref/spec#Type_assertions)）来检查NegativeSqrtError并特别处理，当调用者将error传递给fmt.Println或者log.Fatal的时候，不会有任何行为上的变化。
再举个例子，当json.Decode函数解析一个JSON blob遇到语法错误时，返回一个json包中指定的SyntaxError类型。
``` golang
type SyntaxError struct {
    msg    string // description of error
    Offset int64  // error occurred after reading Offset bytes
}

func (e *SyntaxError) Error() string { return e.msg }
```
Offset字段甚至不会在error默认的格式化中出现，但是调用者可以使用它，添加文件和行信息到到error消息中。
``` golang
if err := dec.Decode(&val); err != nil {
    if serr, ok := err.(*json.SyntaxError); ok {
        line, col := findLine(f, serr.Offset)
        return fmt.Errorf("%s:%d:%d: %v", f.Name(), line, col, err)
    }
    return err
}
```
（这是来自Camlistore项目的真实代码的省略的简化版本）
error接口仅仅需要Error方法；特殊的error实现可能需要其他的方法。例如，net包中返回error类型的错误，遵循了通常的惯例，但是一些error的实现通过net.Error接口定义了其他方法：
``` golang
package net

type Error interface {
    error
    Timeout() bool   // Is the error a timeout?
    Temporary() bool // Is the error temporary?
}
```
客户端代码可以使用类型断言（type assertion）来测试，然后从永久性错误中区分短暂的的网络错误。例如，一个网络爬虫也许会休眠并重试当它遇到一个零时性错误并放弃其他的。
``` golang
if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
    time.Sleep(1e9)
    continue
}
if err != nil {
    log.Fatal(err)
}
```
### 简化重复的错误处理
在Go中，错误处理非常重要。语言的设计和约定鼓励你在发生错误的时候明确检查错误（和其他语言使用抛出错误并获取错误（ throwing exceptions and sometimes catching them ）的惯例不一样）。在有些情况下，这使你的Go代码冗余，但是幸运的是，你可以使用一些技巧来最少的重复错误处理。
思考一个使用HTTP处理器的App Engine应用，从数据库中检索一条记录并使用模板格式化。
``` golang
func init() {
    http.HandleFunc("/view", viewRecord)
}

func viewRecord(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    key := datastore.NewKey(c, "Record", r.FormValue("id"), 0, nil)
    record := new(Record)
    if err := datastore.Get(c, key, record); err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    if err := viewTemplate.Execute(w, record); err != nil {
        http.Error(w, err.Error(), 500)
    }
}
```
这个函数处理由datastore.Get函数和`viewTemplate`的Execute方法产生的错误。在这两个错误中，它用HTTP状态码500("Internal Server Error")来向用户展现一个简单的错误信息。这段代码看起来是可管理的，但是添加了多余的HTTP处理器，并以许多相同的错误处理代码结束。
为了减少重复，我们可以定义我们自己的HTTP appHandler类型来包含error并返回值。
``` golang
type appHandler func(http.ResponseWriter, *http.Request) error
```
然后，我们修改我们的viewRecord 函数来返回错误：
``` golang
func viewRecord(w http.ResponseWriter, r *http.Request) error {
    c := appengine.NewContext(r)
    key := datastore.NewKey(c, "Record", r.FormValue("id"), 0, nil)
    record := new(Record)
    if err := datastore.Get(c, key, record); err != nil {
        return err
    }
    return viewTemplate.Execute(w, record)
}
```
这和原始的版本相似，但是http包不能理解返回error的函数。为了修正，我们在appHandler上实现了http.Handler接口的ServeHTTP方法：
``` golang
func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if err := fn(w, r); err != nil {
        http.Error(w, err.Error(), 500)
    }
}
```
ServeHTTP方法调用appHandler函数并且向用户显示错误（如果有的话）。注意这个方法的接收者，fn，是一个函数。（Go可以这样做！）方法调用函数通过调用表达式fn(w, r)中的接收者。
现在，当使用http包注册viewRecord的时候，我们使用Handle函数（而不是HandleFunc）appHandler作为http.Handler（而不是http.HandlerFunc）
``` golang
func init() {
    http.Handle("/view", appHandler(viewRecord))
}
```
通过基本架构中的错误处理，我们可以使它更加的用户友好。相比仅仅展示错误的字符串，更好的方法是给用户简单的错误消息并附加适合的HTTP状态码，同时在App Engine开发者控制台记录完整的错误用于调试。
为了做到这点，我们创建一个appError 结构体，包含一个error类型和一些其他字段
``` golang
type appError struct {
    Error   error
    Message string
    Code    int
}
```
接下来，我们修改appHandler类型以便返回*appError值：
``` golang
type appHandler func(http.ResponseWriter, *http.Request) *appError
```
（通常，不使用error，而是使用具体的error类型进行传递的做法是错误的，原因会在[Go FAQ](https://golang.org/doc/faq#nil_error)中讨论，不过在这里是正确的，因为ServeHTTP是唯一看到这个值并且使用其内容的地方。）
 为了使`appHandler`的ServeHTTP方法向用户显示`appError`的消息和正确的HTTP状态码，并向开发者记录完整的Error：
 ``` golang
 func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if e := fn(w, r); e != nil { // e is *appError, not os.Error.
        c := appengine.NewContext(r)
        c.Errorf("%v", e.Error)
        http.Error(w, e.Message, e.Code)
    }
}
 ```
 最后，我们更新viewRecord函数，并使它遇到错误的时候返回更多的上下文：
 ``` golang
 func viewRecord(w http.ResponseWriter, r *http.Request) *appError {
    c := appengine.NewContext(r)
    key := datastore.NewKey(c, "Record", r.FormValue("id"), 0, nil)
    record := new(Record)
    if err := datastore.Get(c, key, record); err != nil {
        return &appError{err, "Record not found", 404}
    }
    if err := viewTemplate.Execute(w, record); err != nil {
        return &appError{err, "Can't display record", 500}
    }
    return nil
}
 ```
 这个版本的viewRecord和原始版本的程度相同，但是现在每一行都有特别的意思，并且，我们提供更加友好的用户体验。
 这还没结束；我们将会在我们的应用中优化错误处理。一些想法：
 - 给予每个错误处理一个漂亮的HTML模板
 - 当用户是管理员时，将stack trace写到HTTP回应中，使debug更加简单
 - 为appError写一个构建函数，用于存储stack trace，以便调试
 - 在appHandler中发生panic时recover，将错误作为“危险”写入开发者控制台，并用户“发生了一个严重的错误”。 这是避免向用户暴露由于编程错误引起的不可预测的错误的信息的一个很好的想法。参看[Defer, Panic, and Recover](http://golang.org/doc/articles/defer_panic_recover.html) 文章获取更多细节。

### 总结
适合的错误处理是一个好的软件的基本需求。根据本文所讨论的技术，你应该能够写出更可靠且更简洁的Go代码

Andrew Gerrand 著
