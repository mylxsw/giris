package giris

import (
	"fmt"
	"time"

	"github.com/kataras/iris/v12"
	requestLogger "github.com/kataras/iris/v12/middleware/logger"
	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/glacier/infra"
)

// SetRequestLogHandlerOption set a log handler option
func SetRequestLogHandlerOption(handler func(endTime time.Time, latency time.Duration, status, ip, method, path string, message interface{}, headerMessage interface{})) Option {
	return func(resolver infra.Resolver, irisApp *iris.Application) {
		reqLoggerConf := requestLogger.DefaultConfig()
		reqLoggerConf.LogFunc = handler

		irisApp.Use(requestLogger.New(reqLoggerConf))
	}
}

// SetDefaultRequestLogHandlerOption set default log handler
func SetDefaultRequestLogHandlerOption() Option {
	return SetRequestLogHandlerOption(func(endTime time.Time, latency time.Duration, status, ip, method, path string, message interface{}, headerMessage interface{}) {
		log.WithFields(log.Fields{
			"latency": latency.String(),
			"status":  status,
			"ip":      ip,
			"method":  method,
			"path":    path,
		}).Debugf("%s %s [%s] [%s]", method, path, status, latency)
	})
}

// SetIrisInitOption set iris framework init option
func SetIrisInitOption(fn func(resolver infra.Resolver, irisApp *iris.Application)) Option {
	return fn
}

// Inject is a wrap function for iris request handler
type Inject func(handler interface{}) func(ctx iris.Context)

// SetRouteOption set route option
func SetRouteOption(fn func(resolver infra.Resolver, inject Inject, irisApp *iris.Application)) Option {
	return func(resolver infra.Resolver, irisApp *iris.Application) {
		inject := func(handler interface{}) func(ctx iris.Context) {
			return func(ctx iris.Context) {
				results, err := resolver.CallWithProvider(handler, resolver.Provider(func() iris.Context { return ctx }))
				if err != nil {
					panic(err)
				}

				if len(results) == 0 {
					return
				}

				if len(results) > 1 {
					if err, ok := results[1].(error); ok {
						if err != nil {
							panic(err)
						}
					}
				}

				switch results[0].(type) {
				case string:
					if _, err := ctx.HTML(results[0].(string)); err != nil {
						panic(err)
					}
				case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
					if _, err := ctx.HTML(fmt.Sprintf("%d", results[0])); err != nil {
						panic(err)
					}
				case float32, float64:
					if _, err := ctx.HTML(fmt.Sprintf("%f", results[0])); err != nil {
						panic(err)
					}
				case error:
					if results[0] != nil {
						panic(results[0])
					}
				default:
					if _, err := ctx.JSON(results[0]); err != nil {
						panic(err)
					}
				}
			}
		}
		fn(resolver, inject, irisApp)
	}
}
