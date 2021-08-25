// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package test

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/xfali/neve-core"
	"github.com/xfali/neve-core/processor"
	"github.com/xfali/neve-trace"
	"github.com/xfali/neve-utils/neverror"
	"github.com/xfali/neve-web/gineve"
	"github.com/xfali/neve-web/gineve/midware/loghttp"
	"github.com/xfali/neve-web/result"
	"github.com/xfali/restclient"
	"github.com/xfali/xlog"
	"net/http"
	"testing"
	"time"
)

type webBean struct {
	V          string
	client     restclient.RestClient
	HttpLogger loghttp.HttpLogger         `inject:""`
	Trace      nevetrace.GinTracer         `inject:""`
	Filter     nevetrace.RestClientTracer `inject:""`

	count int
}

func (b *webBean) BeanAfterSet() error {
	b.client = restclient.New(restclient.AddIFilter(
		restclient.NewLog(xlog.GetLogger(), ""),
		b.Filter))
	return nil
}

func (b *webBean) HttpRoutes(engine gin.IRouter) {
	if b.V == "" {
		b.V = "hello world"
	}
	engine.GET("test", b.HttpLogger.LogHttp(), b.Trace.Trace("/test"), func(ctx *gin.Context) {
		sp := nevetrace.GetSpan(ctx)
		if sp == nil {
			xlog.Fatalln("sp is nil")
		}
		status, err := b.client.GetContext(nevetrace.ContextWithSpan(context.Background(), ctx), nil, "http://localhost:8080/api", nil)
		if err != nil {
			xlog.Fatalln("get api failed: ", err)
		}
		ctx.JSON(status, result.Ok(b.V))
	})

	engine.GET("test2", b.HttpLogger.LogHttp(), b.Trace.Trace("/test2"), func(ctx *gin.Context) {
		sp := nevetrace.GetSpan(ctx)
		if sp == nil {
			xlog.Fatalln("sp is nil")
		}
		status, err := b.client.GetContext(nevetrace.ContextWithSpan(context.Background(), ctx), nil, "http://localhost:8079/api", nil)
		if err != nil {
			xlog.Fatalln("get api failed: ", err)
		}
		ctx.JSON(status, result.Ok(b.V))
	})

	engine.GET("api", b.HttpLogger.LogHttp(), b.Trace.Trace("/api"), func(context *gin.Context) {
		sp := nevetrace.GetSpan(context)
		if sp == nil {
			xlog.Fatalln("sp is nil")
		}
		time.Sleep(1 * time.Second)
		b.count++
		if b.count % 2 == 0 {
			context.AbortWithStatus(http.StatusBadRequest)
			return
		}
		context.JSON(http.StatusOK, result.Ok(b.V))
	})
}

func TestClientAndServer(t *testing.T) {
	app := neve.NewFileConfigApplication("assets/application-test.yaml")
	neverror.PanicError(app.RegisterBean(gineve.NewProcessor()))
	neverror.PanicError(app.RegisterBean(processor.NewValueProcessor()))
	neverror.PanicError(app.RegisterBean(nevetrace.NewJaegerProcessor()))

	neverror.PanicError(app.RegisterBean(&webBean{}))
	err := app.Run()
	if err != nil {
		t.Fatal(err)
	}
}

func TestClientAnd2Server(t *testing.T) {
	go func() {
		app := neve.NewFileConfigApplication("assets/application-test2.yaml")
		neverror.PanicError(app.RegisterBean(gineve.NewProcessor()))
		neverror.PanicError(app.RegisterBean(processor.NewValueProcessor()))
		neverror.PanicError(app.RegisterBean(nevetrace.NewJaegerProcessor()))

		neverror.PanicError(app.RegisterBean(&webBean{}))
		err := app.Run()
		if err != nil {
			t.Fatal(err)
		}
	}()

	app := neve.NewFileConfigApplication("assets/application-test.yaml")
	neverror.PanicError(app.RegisterBean(gineve.NewProcessor()))
	neverror.PanicError(app.RegisterBean(processor.NewValueProcessor()))
	neverror.PanicError(app.RegisterBean(nevetrace.NewJaegerProcessor()))

	neverror.PanicError(app.RegisterBean(&webBean{}))
	err := app.Run()
	if err != nil {
		t.Fatal(err)
	}
}
