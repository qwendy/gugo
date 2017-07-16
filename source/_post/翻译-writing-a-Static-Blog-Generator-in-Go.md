---
title: (翻译)writing a Static Blog Generator in Go
date: 2017-07-04 11:06:08
tags:
    - 静态博客
category:
    - Go学习小记
---

翻译原文：https://zupzup.org/static-blog-generator-go/?utm_source=golangweekly&utm_medium=emai

# Writing a Static Blog Generator in Go 使用Go写一个静态博客生成器

A static-site-generator is a tool which, given some input (e.g. markdown) generates a fully static website using HTML, CSS and JavaScript.

一个静态网站生成器是一个使用HTML, CSS and JavaScript,将给予的输入文件（例如markdown）生成完整的静态网站

Why is this cool? Well, for one it’s a lot easier to host a static site and it’s also (usually) quite a bit faster and resource-friendly. Static sites aren’t the best choice for all use-cases, but for mostly non-interactive websites such as blogs they are great.

为什么这很Cool？是的，对于一个人来说，主持一个静态的网站要容易得多，而且(通常)也会更快，而且资源更友好。静态网站不是对所有情况都是最好的选择，但是对于大多数非交互网站比如博客，是非常好的。

In this post, I will describe the static blog generator I wrote in Go, which powers this blog.

在本文中，我将描述我使用Go编写的静态博客生成器，它为这个博客提供了强大的功能。
<!-- more -->
## Motivation 动机

You might be familiar with static-site-generators like the great Hugo, which has just about all the features one could hope for in regards to static site generation.

你可能对像Hugo这样的静态站点生成器很熟悉，它拥有关于静态站点生成的所有特性。

So why would I write another tool just like it with fewer capabilities? The reason was twofold.

所以我为什么写另外一个像它一样但功能更少的工具？原因有两个。

One reason was that I wanted to dive deeper into Go and a command-line-based static-site-generator seemed like a great playground to hone my skills.

一个原因是我想要深入学习Go，一个基于命令行的静态博客生成器似乎是一个磨练我技能的好地方。

The second reason was simply that I had never done it before. I’ve done my fair share of web-development, but I never created a static-site-generator.

第二个原因是，我以前从未做过这个。我做过我的web开发的分享，但我从未创建过静态站点生成器。

This made it intriguing because I theoretically had all the prerequisites and skills to build such a tool with my background in web-development, but I had never tried doing it.

这很有趣，因为我在理论上具备了在web开发背景下构建这样一个工具的所有先决条件和技能，但我从未尝试过这样做。

I was hooked, implemented it within about 2 weeks and had a great time doing it. I used my blog-generator for creating this blog and it worked great so far. :)

我被迷住了，在大约两周的时间内实现了它，在实现的时候我很高兴。我用我的博客生成器创建了这个博客，到目前为止效果很好。:)

## Concept 概念

Early on, I decided to write my blog posts in markdown and keep them in a GitHub Repo. The posts are structured in folders, which represent the url of the blog post.

早些时候，我决定使用markdown写我的博客文章，并将他们放在github的仓库（repo）中。文章以文件夹为结构，代表了博客文章的url

For metadata, such as publication date, tags, title and subtitle I decided on keeping a meta.yml file with each post.md file with the following format:

至于metadata，例如发布日期，标签，标题和副标题，我决定使用meta.yml，每个帖子都有以下的格式：

```
title: Playing Around with BoltDB 
short: "Looking for a simple key/value store for your Go applications? Look no further!"
date: 20.04.2017
tags:
    - golang
    - go
    - boltdb
    - bolt
```

This allowed me to separate the content from the metadata, but still keep everything in the same place, where I’d find it later.

这允许我将内容从metadata中分离，但是需要将所有东西放在同一个地方，让我能待会儿能找到它。

The GitHub Repo was my data source. The next step was to think about features and I came up with this List:

github repo是我的数据来源。下一步是考虑特性，我想到了以下内容：

- Very Lean (landing page should be 1 Request < 10K gzipped)非常轻量。（登录页面应该1个请求小于10k）
- Listing for the landing page and an Archive 展示
- Possibility to use syntax-highlighted Code and Images within Blog Posts 在文章中能够使用语法高亮代码，图片
- Tags 标签
- RSS feed (index.xml) RSS订阅
- Optional Static Pages (e.g. About) 可选的静态页面（例如：About） 
- High Maintainability - use the least amount of templates possible 高可维护性-使尽可能少的模板
- sitemap.xml for SEO 用于SEO的sitemap.xml
- Local preview of the whole blog (a simple run.sh script)本地的整个博客的预览(简单的run.sh脚本)

Quite a healthy feature set. What was very important for me from the start, was to keep everything simple, fast and clean - without any third-party trackers or ads, which compromise privacy and slow everything down.

相当健壮的特性集.对我来说最开始最重要的，是保持所有的东西都要简单，快速和干净 -- 没有任何第三方的跟踪器和广告，保证隐私和使所有事情都放缓。

Based on these ideas, I started making a rough plan of the architecture and started coding.

基于这些想法，我开始对架构进行粗略的规划，并开始编写代码。

## Architectural Overview 架构概述

The application is simple enough. The high-level elements are:

程序足够简单。高等级的元素如下：

- CLI 命令行界面
- DataSource 数据源
- Generators 生成器

The CLI in this case is very simple, as I didn’t add any features in terms of configurability. It basically just fetches data from the DataSource and runs the Generators on it.

在这里CLI非常简单，因为我没有在可配置性方面添加任何特性。仅仅是从DataSource中获取数据 ，并使用生成器运行它。

The DataSource interface looks like this:

DotaSource接口长这样：

``` go
type DataSource interface {
    Fetch(from, to string) ([]string, error)
}
```
The Generator interface looks like this:
生成器接口长这样：

``` go
type Generator interface {
    Generate() error
}
```
Pretty simple. Each Generator also receives a configuration struct, which contains all the necessary data for generation.

相当简单。每一个生成器都接受一个配置的结构体，包含了所有需要的信息。

There are 7 Generators at the time of writing this post:

写这篇文章的时候，有了7个生成器。

- SiteGenerator
- ListingGenerator
- PostGenerator
- RSSGenerator
- SitemapGenerator
- StaticsGenerator
- TagsGenerator
Where the SiteGenerator is the meta-generator, which calls all other generators and outputs the whole static website.

SiteGenerator是meta-generator调用其他所有的生成器并输出整个静态网站的地方。

The generation is based on HTML templates using Go’s html/template package.

生成器基于HTML模板，使用Go的html/template package.

## Implementation Details 实现细节

In this section I will just cover a few selected parts I think might be interesting, such as the git DataSource and the different Generators.

在本节中，我将介绍一些我认为可能很有趣的部分，如git数据源和不同的生成器。

### DataSource

First up, we need some data to generate our blog from. This data, as mentioned above has the form of a git repository. The following Fetch function captures most of what the DataSource implementation does:

首先，我们需要一些数据来生成我们的博客。如上所述，这些数据具有git存储库的形式。下面的Fetch函数展示了大多数数据源实现的功能:

``` go
func (ds *GitDataSource) Fetch(from, to string) ([]string, error) {
    fmt.Printf("Fetching data from %s into %s...\n", from, to)
    if err := createFolderIfNotExist(to); err != nil {
        return nil, err
    }
    if err := clearFolder(to); err != nil {
        return nil, err
    }
    if err := cloneRepo(to, from); err != nil {
        return nil, err
    }
    dirs, err := getContentFolders(to)
    if err != nil {
        return nil, err
    }
    fmt.Print("Fetching complete.\n")
    return dirs, nil
}
```
Fetch is called with two parameters from, which is a repository URL and to, which is the destination folder. The function creates and clears the destination folder, clones the repository using os/exec plus a git command and finally reads the folder, returning a list of paths for all the files within the repository.

Fetch使用2个参数来调用，一个是repo的url ，一个是目标的文件夹。这个函数创建并清除目标文件夹，使用os/exec添加一个git命令克隆repo，最后读取文件夹，返回repo中的所有文件的路径列表

As mentioned above, the repository contains only folders, which represent the different blog posts. The array with these folder paths is then passed to the generators, which can then do their thing for each of the blog posts within the repository.

如上所说，这个repo仅仅包含文件夹，代表着不同博客的文章。将这些文件夹路径的数据传递给生成器，

### Kicking it all off 开始完成所有的目标

After the Fetch comes the Generate phase. When the blog-generator is executed, the following code is executed on the highest level:

在Fetch之后，接下来是Generate阶段。当blog-generator执行的时候，下面的代码会最先执行。

``` go
ds := datasource.New()
dirs, err := ds.Fetch(RepoURL, TmpFolder)
if err != nil {
    log.Fatal(err)
}
g := generator.New(&generator.SiteConfig{
    Sources:     dirs,
    Destination: DestFolder,
})
err = g.Generate()
if err != nil {
    log.Fatal(err)
}
```
The generator.New function creates a new SiteGenerator which is basically a generator, which calls other generators. It’s passed a destination folder and the directories for the blog posts within the repository.

generate.New函数创建一个新的SiteGenerate，SiteGenerate是一个基本的生成器，它调用其他生成器。需要传递目标文件夹和生成在repo中博客文章的目录。

As every Generator implementing the interface mentioned above, the SiteGenerator has a Generate method, which returns an error. The Generate method of the SiteGenerator prepares the destination folder, reads in templates, prepares data structures for the blog posts, registers the other generators and concurrently runs them.

正如上面提到的每一个生成器接口的实现，SiteGenerate有一个返回error的Generate方法。SiteGenerate的Generate的方法准备了目标文件夹，读取模板，准备博客文章的数据结构，注册其他生成器并且并发执行他们。

The SiteGenerator also registers some settings for the blog like the URL, Language, Date Format etc. These settings are simply global constants, which is certainly not the prettiest solution or the most scalable, but it’s simple and that was the highest goal here.

SiteGenerate同样注册了一些文章的设置，例如URL，语言，数据格式等等。这些设置是简单的全局常量，这当然不是最好的解决办法也不是最有扩展性的，但是它很简单，并且在这里，这是我们的最高目标。

### Posts

The most important concept on a blog are - surprise, surprise - blog posts! In the context of this blog-generator, they are represented by the following data-structure:

博客最重要的概念就是博客的文章。在这个blog-生成器的上下文中，它们由以下数据结构表示:

``` go
type Post struct {
    Name      string
    HTML      []byte
    Meta      *Meta
    ImagesDir string
    Images    []string
}
```
These posts are created by iterating over the folders in the repository, reading the meta.yml file, converting the post.md file to HTML and by adding images, if there are any.

这些文章是通过在存储库中的文件夹中迭代来创建的，读取meta.yml文件，转换post.md文件到HTML，如果有图像的话添加图像。

Quite a bit of work, but once we have the posts represented as a data structure, the generation of posts is quite simple and looks like this:

做了很多工作，但是一旦我们的post被表示为数据结构，文章的生成就很简单了，看起来就像这样:

``` go
func (g *PostGenerator) Generate() error {
    post := g.Config.Post
    destination := g.Config.Destination
    t := g.Config.Template
    staticPath := fmt.Sprintf("%s%s", destination, post.Name)
    if err := os.Mkdir(staticPath, os.ModePerm); err != nil {
      return fmt.Errorf("error creating directory at %s: %v", staticPath, err)
    }
    if post.ImagesDir != "" {
      if err := copyImagesDir(post.ImagesDir, staticPath); err != nil {
          return err
      }
    }
    if err := writeIndexHTML(staticPath, post.Meta.Title, template.HTML(string(post.HTML)), t); err != nil {
      return err
    }
    return nil
}
```
First, we create a directory for the post, then we copy the images in there and finally create the post’s index.html file using templating. The PostGenerator also implements syntax-highlighting, which I described in this post.

首先，我们给post创建一个目录，然后复制图片到这里，最后使用模板创建post的index.html。我在本文中说了，PostGenerate实现了语法高亮。

### Listing Creation 清单创建

When a user comes to the landing page of the blog, she sees the latest posts with information like the reading time of the article and a short description. For this feature and for the archive, I implemented the ListingGenerator, which takes the following config:

当一个用户访问博客登录界面的时候，她会看到最近的文章的信息，例如阅读文章需要的时间和简单的描述。为了这个特性并且为了存档，我们实现了ListingGenerator，使用了如下的配置：

``` go
type ListingConfig struct {
    Posts                  []*Post
    Template               *template.Template
    Destination, PageTitle string
}
```
The Generate method of this generator iterates over the post, assembles their metadata and creates short blocks based on the given template. Then these blocks are appended and written to the index template.

这个生成器的生成方法在post上迭代，组装它们的元数据，并根据给定的模板创建短块。然后将这些块附加到索引模板中。

I liked medium’s feature to approximate the time to read an article, so I implemented my own version of it, based on the assumption that an average human reads about 200 words per minute. Images also count towards the overall reading time with a constant 12 seconds added for each img tag in the post. This will obviously not scale for arbitrary content, but should be a fine approximation for my usual articles:

我喜欢阅读文章的近似时间的媒体特性，所以我实现了自己的版本。假设人类凭借每分钟阅读200个单词。图片同样计入总阅读时间，以每张图12秒计算。这显然不适合任意内容，但很适合我的文章。

``` go
func calculateTimeToRead(input string) string {
    // an average human reads about 200 wpm
    var secondsPerWord = 60.0 / 200.0
    // multiply with the amount of words
    words := secondsPerWord * float64(len(strings.Split(input, " ")))
    // add 12 seconds for each image
    images := 12.0 * strings.Count(input, "<img")
    result := (words + float64(images)) / 60.0
    if result < 1.0 {
        result = 1.0
    }
    return fmt.Sprintf("%.0fm", result)
}
```
### Tags 标签

Next, to have a way to categorize and filter the posts by topic, I opted to implement a simple tagging mechanism. Posts have a list of tags in their meta.yml file. These tags should be listed on a separate Tags Page and upon clicking on a tag, the user is supposed to see a listing of posts with the selected tag.

接下来，为了分类和通过主题过滤文章，我选择实现一个简单的标签机制。文章在meta.yml中有标签列表。这些标签应该在独立的标签页中，根据点击一个标签，用户应该能看到这个标签的文章列表。

First up, we create a map from tag to Post:

首先，我们创建标签到文章的映射表。

``` go
func createTagPostsMap(posts []*Post) map[string][]*Post {
result := make(map[string][]*Post)
    for _, post := range posts {
        for _, tag := range post.Meta.Tags {
            key := strings.ToLower(tag)
             if result[key] == nil {
                 result[key] = []*Post{post}
             } else {
                 result[key] = append(result[key], post)
             }
        }
    }
    return result
}
```
Then, there are two tasks to implement:

然后，有两个任务要去实现

- Tags Page 标签页
- List of Posts for a selected Tag 展示选择的标签的文章啊
The data structure of a Tag looks like this:

数据结构如下
```go
type Tag struct {
    Name  string
    Link  string
    Count int
}
```
So, we have the actual tag (Name), the Link to the tag’s listing page and the amount of posts with this tag. These tags are created from the tagPostsMap and then sorted by Count descending:

所以，我们有实际的标签（名），标签列表页的连接和这个标签的文章。这些标签由tagPostMap创建，然后根据数目降序排列。
``` go
tags := []*Tag{}
for tag, posts := range tagPostsMap {
    tags = append(tags, &Tag{Name: tag, Link: getTagLink(tag), Count: len(posts)})
}
sort.Sort(ByCountDesc(tags))
```
The Tags Page basically just consists of this list rendered into the tags/index.html file.

标签页仅仅由渲染到tags/index.html的文件组成

The List of Posts for a selected Tag can be achieved using the ListingGenerator described above. We just need to iterate the tags, create a folder for each tag, select the posts to display and generate a listing for them.

选择的标签的文章列表能被存档，使用上面的ListingGenerator。我们仅仅需要迭代标签，给每一个标签创建文件夹，选择选择文章展示并生成列表

### Sitemap & RSS

To improve searchability on the web, it’s a good idea to have a sitemap.xml which can be crawled by bots. Creating such a file is fairly straightforward and can be done using the Go standard library.

为了提高网站的搜索能力，有个好主意是拥有一个sitemap.xml，它能够被bots爬到。创建这样的一个文件相当简单，使用Go的标准库就可以了。

In this tool, however, I opted to use the great etree library, which provides a nice API for creating and reading XML.

但是在这个工具里，我选择使用[etree](https://github.com/beevik/etree)库，它提供创建个读取xml非常好用的API。

The SitemapGenerator uses this config:

SitemapGenerator使用这个配置：

``` go
type SitemapConfig struct {
    Posts       []*Post
    TagPostsMap map[string][]*Post
    Destination string
}
```

blog-generator takes a basic approach to the sitemap and just generates url and image locations by using the addURL function:

blog-generator对站点地图进行了基本的处理，并通过使用addURL函数生成url和图像位置。

``` go
func addURL(element *etree.Element, location string, images []string) {
    url := element.CreateElement("url")
     loc := url.CreateElement("loc")
     loc.SetText(fmt.Sprintf("%s/%s/", blogURL, location))

     if len(images) > 0 {
         for _, image := range images {
            img := url.CreateElement("image:image")
             imgLoc := img.CreateElement("image:loc")
             imgLoc.SetText(fmt.Sprintf("%s/%s/images/%s", blogURL, location, image))
         }
     }
}
```
After creating the XML document with etree, it’s just saved to a file and stored in the output folder.

使用etree创建xml文件之后，保存并存储文件到文件夹中。

RSS generation works the same way - iterate all posts and create XML entries for each post, then write to index.xml.

Rss生成器使用同样的方法生成 - 遍历所有文章并为每个post创建XML条目，然后将其写入index.XML。

### Handling Statics

The last concept I needed were entirely static assets like a favicon.ico or a static page like About. To do this, the tool runs the StaticsGenerator with this config:

我所需要的最后一个概念是完全静态的资源，比如favicon.ico或静态页面例如About。为了做到这一点，该工具使用这个配置来运行StaticsGenerator:

``` go
type StaticsConfig struct {
    FileToDestination map[string]string
    TemplateToFile    map[string]string
    Template          *template.Template
}
```

The FileToDestination-map represents static files like images or the robots.txt and TemplateToFile is a mapping from templates in the static folder to their designated output path.

FileToDestination-map表示像图片或者robot.txt的静态文件，TemplateToFiles是从静态文件夹中的模板到指定的输出路径的映射。

This configuration could look like this in practice:

实际上，配置文件是这样的：

``` go
fileToDestination := map[string]string{
    "static/favicon.ico": fmt.Sprintf("%s/favicon.ico", destination),
    "static/robots.txt":  fmt.Sprintf("%s/robots.txt", destination),
    "static/about.png":   fmt.Sprintf("%s/about.png", destination),
}
templateToFile := map[string]string{
    "static/about.html": fmt.Sprintf("%s/about/index.html", destination),
}
statg := StaticsGenerator{&StaticsConfig{
FileToDestination: fileToDestination,
   TemplateToFile:    templateToFile,
   Template:          t,
}}
```

The code for generating these statics is not particularly interesting - as you can imagine, the files are just iterated and copied to the given destination.

生成这些静态文件的代码不是很有趣 - 正如你想的那样，仅仅就是遍历文件，并复制到目标地点

### Parallel Execution 并行执行

For blog-generator to be fast, the generators are all run in parallel. For this purpose, they all follow the Generator interface - this way we can put them all inside a slice and concurrently call Generate for all of them.

为了让blog-generator更快，生成器都并行运行。为了这个目的，他们都遵循着Generator的接口 - 这样我可以将他们放在一个slice中，并发的调用生成器。

The generators all work independently of one another and don’t use any global state mutation, so parallelizing them was a simple exercise of using channels and a sync.WaitGroup like this:

生成器运行都相互独立，且不使用任何全局状态变化，所以并行是一个简单的使用channel和sync.WaitGroup的练习，就像这样：

``` go
func runTasks(posts []*Post, t *template.Template, destination string) error {
    var wg sync.WaitGroup
    finished := make(chan bool, 1)
    errors := make(chan error, 1)
    pool := make(chan struct{}, 50)
    generators := []Generator{}

    for _, post := range posts {
        pg := PostGenerator{&PostConfig{
            Post:        post,
             Destination: destination,
             Template:    t,
        }}
        generators = append(generators, &pg)
    }

    fg := ListingGenerator{&ListingConfig{
        Posts:       posts[:getNumOfPagesOnFrontpage(posts)],
        Template:    t,
        Destination: destination,
        PageTitle:   "",
    }}

    ...create the other generators...

    generators = append(generators, &fg, &ag, &tg, &sg, &rg, &statg)

    for _, generator := range generators {
        wg.Add(1)
        go func(g Generator) {
            defer wg.Done()
            pool <- struct{}{}
            defer func() { <-pool }()
            if err := g.Generate(); err != nil {
                errors <- err
            }
        }(generator)
    }

    go func() {
        wg.Wait()
        close(finished)
    }()

    select {
    case <-finished:
        return nil
    case err := <-errors:
        if err != nil {
           return err
        }
    }
    return nil
}
```
The runTasks function uses a pool of max. 50 goroutines, creates all generators, adds them to a slice and then runs them in parallel.

runTasks函数使用最大50协程的协程池，创建所有生成器，将他们添加到slice中，然后并行运行他们

These examples were just a short dive into the basic concepts used to write a static-site generator in Go.

这些例子仅仅初步深入基本概念用于使用Go写一个静态网站生成器

If you’re interested in the full implementation, you can find the code here.

如果你对整个实现感兴趣，你可以在[这里](https://github.com/zupzup/blog-generator)找到代码.

## Conclusion 总结

Writing my blog-generator was an absolute blast and a great learning experience. It’s also quite satisfying to use my own hand-crafted tool for creating my blog.

编写我的博客生成器是一种绝对的挑战，也是一种很好的学习体验。使用我自己的手工工具创建我的博客也很令人满意。

To publish my posts to AWS, I also created static-aws-deploy, another Go command-line tool, which I covered in this post.

为了发布我的文章到AWS，我同样创建了[static-aws-deploy](https://github.com/zupzup/static-aws-deploy)，额外的Go命令行工具，我在这篇[文章](https://zupzup.org/static-deploy-tool/)中会提到提到

If you want to use the tool yourself, just fork the repo and change the configuration. However, I didn’t put much time into customizability or configurability, as Hugo provides all that and more.

如果你想自己使用这个工具，仅仅fork这个repo，并修改配置就可以了。但是，我没有花太多时间来做可定制化和可配置化，因为[Hugo](https://gohugo.io/)提供的了所有你需要的。

Of course, one should strive not to re-invent the wheel all the time, but sometimes re-inventing a wheel or two can be rewarding and can help you learn quite a bit in the process. :)

当然，一个人应该努力不去重新发明轮子，但是有时候重新发明一两个轮子是值得的，可以帮助你在这个过程中学到很多东西。:)

### Resources
- [blog-generator on GitHub](https://github.com/zupzup/blog-generator)
- [Hugo](https://gohugo.io/)
- [etree](https://github.com/beevik/etree)