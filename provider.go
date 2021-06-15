package giris

import (
	"context"
	"fmt"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/recover"
	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/glacier/infra"
	"github.com/mylxsw/graceful"
)

type Option func(resolver infra.Resolver, irisApp *iris.Application)

type serviceProvider struct {
	listenerBuilder infra.ListenerBuilder
	options         []Option
}

func Provider(listenerBuilder infra.ListenerBuilder, options ...Option) infra.DaemonProvider {
	return serviceProvider{
		listenerBuilder: listenerBuilder,
		options:         options,
	}
}

func (p serviceProvider) Register(app infra.Binder) {
	app.MustSingletonOverride(func() *iris.Application {
		app := iris.New()
		app.Use(recover.New())
		app.Configure(iris.WithoutStartupLog)
		app.Logger().Install(logger{})

		return app
	})
}

func (p serviceProvider) Boot(app infra.Resolver) {
	app.MustResolve(func(irisApp *iris.Application) {
		for _, opt := range p.options {
			opt(app, irisApp)
		}
	})
}

func (p serviceProvider) Daemon(ctx context.Context, app infra.Resolver) {
	app.MustResolve(func(gf graceful.Graceful, irisApp *iris.Application) {
		listener, err := p.listenerBuilder.Build(app)
		if err != nil {
			panic(err)
		}

		gf.AddShutdownHandler(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if log.DebugEnabled() {
				log.Debugf("prepare to shutdown http server...")
			}

			if err := irisApp.Shutdown(ctx); err != nil {
				log.Errorf("shutdown http server failed: %s", err)
			}

			if log.WarningEnabled() {
				log.Warning("http server has been shutdown")
			}
		})

		if log.DebugEnabled() {
			log.Debugf("http server started, listening on %s", listener.Addr())
		}

		if err := irisApp.Run(iris.Listener(listener)); err != nil {
			if log.DebugEnabled() {
				log.Debugf("http server stopped: %v", err)
			}

			if err != iris.ErrServerClosed {
				gf.Shutdown()
			}
		}
	})
}

type logger struct{}

func (l logger) Print(i ...interface{}) {
	log.Debug(i...)
}

func (l logger) Println(i ...interface{}) {
	log.Debugf(fmt.Sprintf("%v", i[0]), i[1:]...)
}

func (l logger) Error(i ...interface{}) {
	log.Error(i...)
}

func (l logger) Warn(i ...interface{}) {
	log.Warning(i...)
}

func (l logger) Info(i ...interface{}) {
	log.Info(i...)
}

func (l logger) Debug(i ...interface{}) {
	log.Debug(i...)
}
