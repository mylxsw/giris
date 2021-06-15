# giris

Iris 框架适配 Glacier 框架

使用示例

```go
package main

import (
	"os"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/mylxsw/giris"
	"github.com/mylxsw/glacier/infra"
	"github.com/mylxsw/glacier/listener"
	"github.com/mylxsw/glacier/starter/application"
	"github.com/urfave/cli"
	"github.com/urfave/cli/altsrc"
)

func main() {
	app := application.Create("0.1")
	app.AddFlags(altsrc.NewStringFlag(cli.StringFlag{
		Name:   "listen",
		Usage:  "HTTP listen address",
		Value:  "127.0.0.1:19921",
	}))

	app.Singleton(func() DemoService { return DemoService{} })

	app.Provider(giris.Provider(
		listener.FlagContext("listen"),
		giris.SetDefaultRequestLogHandlerOption(),
		giris.SetRouteOption(func(resolver infra.Resolver, inject giris.Inject, irisApp *iris.Application) {
			v1 := irisApp.Party("/v1")
			{
				v1.Get("/", inject(func(ctx context.Context, srv DemoService) string {
					return srv.Hello(ctx.URLParam("name"))
				}))
				v1.Get("/{name}", inject(func(ctx context.Context, srv DemoService) string {
					return srv.Hello(ctx.Params().Get("name"))
				}))
			}
		}),
		giris.SetIrisInitOption(func(resolver infra.Resolver, irisApp *iris.Application) {
			irisApp.Logger().SetLevel("debug")
		}),
	))

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

type DemoService struct{}

func (srv DemoService) Hello(name string) string {
	return "hello, " + name
}
```
