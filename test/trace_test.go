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
	"github.com/xfali/neve-trace/gintrace"
	"github.com/xfali/neve-trace/resttrace"
	"github.com/xfali/neve-utils/neverror"
	"github.com/xfali/neve-web/gineve"
	"github.com/xfali/neve-web/gineve/midware/loghttp"
	"github.com/xfali/neve-web/result"
	"github.com/xfali/restclient"
	"github.com/xfali/xlog"
	"net/http"
	"testing"
)

type webBean struct {
	V          string
	client     restclient.RestClient
	HttpLogger loghttp.HttpLogger         `inject:""`
	Trace      gintrace.GinTracer         `inject:""`
	Filter     resttrace.RestClientTracer `inject:""`
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
		sp := gintrace.GetSpan(ctx)
		if sp == nil {
			xlog.Fatalln("sp is nil")
		}
		_, err := b.client.GetContext(gintrace.ContextWithSpan(context.Background(), ctx), nil, "http://localhost:8080/api", nil)
		if err != nil {
			xlog.Fatalln("get api failed: ", err)
		}
		ctx.JSON(http.StatusOK, result.Ok(b.V))
	})

	engine.GET("api", b.HttpLogger.LogHttp(), b.Trace.Trace("/api"), func(context *gin.Context) {
		sp := gintrace.GetSpan(context)
		if sp == nil {
			xlog.Fatalln("sp is nil")
		}
		context.JSON(http.StatusOK, result.Ok(b.V))
	})
}

func TestClientAndServer(t *testing.T) {
	app := neve.NewFileConfigApplication("assets/application-test.yaml")
	neverror.PanicError(app.RegisterBean(gineve.NewProcessor()))
	neverror.PanicError(app.RegisterBean(processor.NewValueProcessor()))
	neverror.PanicError(app.RegisterBean(trace.NewJaegerProcessor()))

	neverror.PanicError(app.RegisterBean(&webBean{}))
	err := app.Run()
	if err != nil {
		t.Fatal(err)
	}
}
