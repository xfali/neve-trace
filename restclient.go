// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package nevetrace

import (
	"github.com/opentracing/opentracing-go"
	"github.com/xfali/neve-core/appcontext"
	"github.com/xfali/restclient"
	"github.com/xfali/xlog"
	"net/http"
)

type RestClientTracer restclient.IFilter

type restClientTraceFilter struct {
	logger xlog.Logger
	tracer opentracing.Tracer
}

type RestClientOpt func(f *restClientTraceFilter)

func NewRestClientTraceFilter(opts ...RestClientOpt) *restClientTraceFilter {
	ret := &restClientTraceFilter{
		logger: xlog.GetLogger(),
		tracer: opentracing.GlobalTracer(),
	}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

func (f *restClientTraceFilter) RegisterFunction(registry appcontext.InjectFunctionRegistry) error {
	return registry.RegisterInjectFunction(func(tracer opentracing.Tracer) {
		f.tracer = tracer
	})
}

func (f *restClientTraceFilter) Filter(request *http.Request, fc restclient.FilterChain) (*http.Response, error) {
	if span := opentracing.SpanFromContext(request.Context()); span != nil {
		// Transmit the span's TraceContext as HTTP headers on our
		// outbound request.
		err := f.tracer.Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(request.Header))

		if err != nil {
			f.logger.Errorln(err)
		}
		return fc.Filter(request)
	}

	return fc.Filter(request)
}

type restclientOpts struct{}

var RestClientOpts restclientOpts

func (opt restclientOpts) WithTracer(tracer opentracing.Tracer) RestClientOpt {
	return func(f *restClientTraceFilter) {
		f.tracer = tracer
	}
}

func (opt restclientOpts) WithLogger(logger xlog.Logger) RestClientOpt {
	return func(f *restClientTraceFilter) {
		f.logger = logger
	}
}
