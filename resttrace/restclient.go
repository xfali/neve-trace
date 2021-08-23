// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package resttrace

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

func NewRestClientTraceFilter(tracer opentracing.Tracer) *restClientTraceFilter {
	return &restClientTraceFilter{
		logger: xlog.GetLogger(),
		tracer: tracer,
	}
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
