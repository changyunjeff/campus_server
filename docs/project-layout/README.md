## Go 目录

### `/cmd`

本项目的主干。

每个应用程序的目录名应该与你想要的可执行文件的名称相匹配(例如，`/cmd/myapp`)。

不要在这个目录中放置太多代码。如果你认为代码可以导入并在其他项目中使用，那么它应该位于 `/pkg` 目录中。如果代码不是可重用的，或者你不希望其他人重用它，请将该代码放到 `/internal` 目录中。你会惊讶于别人会怎么做，所以要明确你的意图!

通常有一个小的 `main` 函数，从 `/internal` 和 `/pkg` 目录导入和调用代码，除此之外没有别的东西。

有关示例，请参阅 [`/cmd`](cmd/README.md) 目录。

### `/internal`

私有应用程序和库代码。这是你不希望其他人在其应用程序或库中导入代码。请注意，这个布局模式是由 Go 编译器本身执行的。有关更多细节，请参阅Go 1.4 [`release notes`](https://golang.org/doc/go1.4#internalpackages) 。注意，你并不局限于顶级 `internal` 目录。在项目树的任何级别上都可以有多个内部目录。

你可以选择向 internal 包中添加一些额外的结构，以分隔共享和非共享的内部代码。这不是必需的(特别是对于较小的项目)，但是最好有可视化的线索来显示预期的包的用途。你的实际应用程序代码可以放在 `/internal/app` 目录下(例如 `/internal/app/myapp`)，这些应用程序共享的代码可以放在 `/internal/pkg` 目录下(例如 `/internal/pkg/myprivlib`)。

### `/pkg`

外部应用程序可以使用的库代码(例如 `/pkg/mypubliclib`)。其他项目会导入这些库，希望它们能正常工作，所以在这里放东西之前要三思:-)注意，`internal` 目录是确保私有包不可导入的更好方法，因为它是由 Go 强制执行的。`/pkg` 目录仍然是一种很好的方式，可以显式地表示该目录中的代码对于其他人来说是安全使用的好方法。由 Travis Jeffery  撰写的 [`I'll take pkg over internal`](https://travisjeffery.com/b/2019/11/i-ll-take-pkg-over-internal/) 博客文章提供了 `pkg` 和 `internal` 目录的一个很好的概述，以及什么时候使用它们是有意义的。

当根目录包含大量非 Go 组件和目录时，这也是一种将 Go 代码分组到一个位置的方法，这使得运行各种 Go 工具变得更加容易（正如在这些演讲中提到的那样: 来自 GopherCon EU 2018 的 [`Best Practices for Industrial Programming`](https://www.youtube.com/watch?v=PTE4VJIdHPg) , [GopherCon 2018: Kat Zien - How Do You Structure Your Go Apps](https://www.youtube.com/watch?v=oL6JBUk6tj0) 和 [GoLab 2018 - Massimiliano Pippi - Project layout patterns in Go](https://www.youtube.com/watch?v=3gQa1LWwuzk) ）。

如果你想查看哪个流行的 Go 存储库使用此项目布局模式，请查看 [`/pkg`](pkg/README.md) 目录。这是一种常见的布局模式，但并不是所有人都接受它，一些 Go 社区的人也不推荐它。

如果你的应用程序项目真的很小，并且额外的嵌套并不能增加多少价值(除非你真的想要:-)，那就不要使用它。当它变得足够大时，你的根目录会变得非常繁琐时(尤其是当你有很多非 Go 应用组件时)，请考虑一下。

### `/vendor`

应用程序依赖项(手动管理或使用你喜欢的依赖项管理工具，如新的内置 [`Go Modules`](https://go.dev/wiki/Modules) 功能)。`go mod vendor` 命令将为你创建 `/vendor` 目录。请注意，如果未使用默认情况下处于启用状态的 Go 1.14，则可能需要在 `go build` 命令中添加 `-mod=vendor` 标志。

如果你正在构建一个库，那么不要提交你的应用程序依赖项。

注意，自从 [`1.13`](https://golang.org/doc/go1.13#modules) 以后，Go 还启用了模块代理功能(默认使用 [`https://proxy.golang.org`](https://proxy.golang.org) 作为他们的模块代理服务器)。在[`here`](https://blog.golang.org/module-mirror-launch) 阅读更多关于它的信息，看看它是否符合你的所有需求和约束。如果需要，那么你根本不需要 `vendor` 目录。

国内模块代理功能默认是被墙的，七牛云有维护专门的的[`模块代理`](https://github.com/goproxy/goproxy.cn/blob/master/README.zh-CN.md) 。


## 服务应用程序目录

### `/api`

OpenAPI/Swagger 规范，JSON 模式文件，协议定义文件。

有关示例，请参见 [`/api`](api/README.md) 目录。

## Web 应用程序目录

### `/web`

特定于 Web 应用程序的组件:静态 Web 资源、服务器端模板和 SPAs。

## 通用应用目录


### `/configs`

配置文件模板或默认配置。

将你的 `confd` 或 `consul-template` 模板文件放在这里。

### `/init`

System init（systemd，upstart，sysv）和 process manager/supervisor（runit，supervisor）配置。

### `/scripts`

执行各种构建、安装、分析等操作的脚本。

这些脚本保持了根级别的 Makefile 变得小而简单(例如， [`https://github.com/hashicorp/terraform/blob/main/Makefile`](https://github.com/hashicorp/terraform/blob/main/Makefile) )。

有关示例，请参见  [`/scripts`](scripts/README.md) 目录。

### `/build`

打包和持续集成。

将你的云( AMI )、容器( Docker )、操作系统( deb、rpm、pkg )包配置和脚本放在 `/build/package` 目录下。

将你的 CI (travis、circle、drone)配置和脚本放在 `/build/ci` 目录中。请注意，有些 CI 工具(例如 Travis CI)对配置文件的位置非常挑剔。尝试将配置文件放在 `/build/ci` 目录中，将它们链接到 CI 工具期望它们的位置(如果可能的话)。

### `/deployments`

IaaS、PaaS、系统和容器编排部署配置和模板(docker-compose、kubernetes/helm、mesos、terraform、bosh)。注意，在一些存储库中(特别是使用 kubernetes 部署的应用程序)，这个目录被称为 `/deploy`。

### `/test`

额外的外部测试应用程序和测试数据。你可以随时根据需求构造 `/test` 目录。对于较大的项目，有一个数据子目录是有意义的。例如，你可以使用 `/test/data` 或 `/test/testdata` (如果你需要忽略目录中的内容)。请注意，Go 还会忽略以“.”或“_”开头的目录或文件，因此在如何命名测试数据目录方面有更大的灵活性。

有关示例，请参见  [`/test`](test2/README.md) 目录。

## 其他目录

### `/docs`

设计和用户文档(除了 godoc 生成的文档之外)。

有关示例，请参阅 [`/docs`](docs/README.md) 目录。

### `/tools`

这个项目的支持工具。注意，这些工具可以从 `/pkg` 和 `/internal` 目录导入代码。

有关示例，请参见 [`/tools`](tools/README.md) 目录。

### `/examples`

你的应用程序和/或公共库的示例。

有关示例，请参见 [`/examples`](examples/README.md) 目录。

### `/third_party`

外部辅助工具，分叉代码和其他第三方工具(例如 Swagger UI)。

### `/githooks`

Git hooks。

### `/assets`

与存储库一起使用的其他资源(图像、徽标等)。

### `/website`

如果你不使用 Github 页面，则在这里放置项目的网站数据。

有关示例，请参见 [`/website`](website/README.md) 目录。

