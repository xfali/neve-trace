// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package nevetrace

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/xfali/goutils/idUtil"
	"github.com/xfali/neve-core/appcontext"
	"github.com/xfali/xlog"
	"net/http"
)

const (
	GinContextTraceKey = "_neve_trace_gin_ctx_key_"
)

type GinTracer interface {
	Trace(name string) gin.HandlerFunc
}

type ginTraceFilter struct {
	logger xlog.Logger
	tracer opentracing.Tracer

	errFunc func(httpStatus int) bool
}

type GinOpt func(f *ginTraceFilter)

func NewGinTrace(opts ...GinOpt) *ginTraceFilter {
	ret := &ginTraceFilter{
		logger:  xlog.GetLogger(),
		tracer:  opentracing.GlobalTracer(),
		errFunc: defaultErrFunc,
	}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

func (f *ginTraceFilter) RegisterFunction(registry appcontext.InjectFunctionRegistry) error {
	return registry.RegisterInjectFunction(func(tracer opentracing.Tracer) {
		f.tracer = tracer
	})
}

func (f *ginTraceFilter) Trace(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		wireContext, err := f.tracer.Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(c.Request.Header))
		if err != nil {
			f.logger.Warnln(err)
		}
		sp := f.tracer.StartSpan(name, ext.RPCServerOption(wireContext))
		defer sp.Finish()

		sp.SetTag("jaeger-debug-id", idUtil.RandomId(16))
		sp.SetTag("http.url", c.Request.RequestURI)
		sp.SetTag("http.method", c.Request.Method)
		c.Set(GinContextTraceKey, sp)
		c.Next()
		sp.SetTag("http.status_code", c.Writer.Status())
		if f.errFunc(c.Writer.Status()) {
			sp.SetTag("error", true)
		}
	}
}

func GetSpan(c *gin.Context) opentracing.Span {
	if v, ok := c.Get(GinContextTraceKey); ok {
		return v.(opentracing.Span)
	}
	return nil
}

func ContextWithSpan(ctx context.Context, c *gin.Context) context.Context {
	return opentracing.ContextWithSpan(ctx, GetSpan(c))
}

func defaultErrFunc(status int) bool {
	return status >= http.StatusBadRequest
}

type ginOpts struct{}

var GinOpts ginOpts

func (opt ginOpts) WithTracer(tracer opentracing.Tracer) GinOpt {
	return func(f *ginTraceFilter) {
		f.tracer = tracer
	}
}

func (opt ginOpts) WithLogger(logger xlog.Logger) GinOpt {
	return func(f *ginTraceFilter) {
		f.logger = logger
	}
}

func (opt ginOpts) WithErrFunc(errfunc func(httpStatus int) bool) GinOpt {
	return func(f *ginTraceFilter) {
		f.errFunc = errfunc
	}
}
