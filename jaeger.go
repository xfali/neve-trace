// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package nevetrace

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/transport"
	"github.com/xfali/fig"
	"github.com/xfali/neve-core/bean"
	"github.com/xfali/neve-utils/neverror"
	"io"
	"strconv"
	"strings"
)

const (
	keyServiceName   = "neve.trace.serviceName"
	keySamplerName   = "neve.trace.sampler.type"
	keySamplerValue  = "neve.trace.sampler.value"
	keyReporterName  = "neve.trace.reporter.type"
	keyReporterValue = "neve.trace.reporter.value"

	keyGinEnable        = "neve.trace.gin.enable"
	keyRestClientEnable = "neve.trace.restclient.enable"
)

type jeagerOpts struct {
	tracer      opentracing.Tracer
	sampler     jaeger.Sampler
	reporter    jaeger.Reporter
	serviceName string

	closer io.Closer
}

type jaegerProcessor struct {
	jeagerOpts
}

type JaegerOpt func(*jeagerOpts)

func NewJaegerProcessor(opts ...JaegerOpt) *jaegerProcessor {
	ret := &jaegerProcessor{}
	for _, opt := range opts {
		opt(&ret.jeagerOpts)
	}
	return ret
}

// 初始化对象处理器
func (p *jaegerProcessor) Init(conf fig.Properties, container bean.Container) error {
	if p.tracer == nil {
		tracer, closer, err := p.initTrace(conf)
		if err != nil {
			cfg, err := config.FromEnv()
			if err != nil {
				// parsing errors might happen here, such as when we get a string where we expect a number
				return fmt.Errorf("Could not parse Jaeger env vars: %s. ", err.Error())
			}
			tracer, closer, err = cfg.NewTracer()
			if err != nil {
				return err
			}
		}
		p.tracer = tracer
		p.closer = closer
	}

	// register trace
	neverror.PanicError(container.Register(p.tracer))

	// gin trace is enable, register default GinTrace
	if strings.ToLower(conf.Get(keyGinEnable, "")) == "true" {
		neverror.PanicError(container.Register(NewGinTrace(GinOpts.WithTracer(p.tracer))))
	}
	// gin trace is enable, register default RestClientTrace
	if strings.ToLower(conf.Get(keyRestClientEnable, "")) == "true" {
		neverror.PanicError(container.Register(NewRestClientTraceFilter(RestClientOpts.WithTracer(p.tracer))))
	}
	return nil
}

// 对象分类，判断对象是否实现某些接口，并进行相关归类。为了支持多协程处理，该方法应线程安全。
// 注意：该方法建议只做归类，具体处理使用Process，不保证Processor的实现在此方法中做了相关处理。
// 该方法在Bean Inject注入之后调用
// return: bool 是否能够处理对象， error 处理是否有错误
func (p *jaegerProcessor) Classify(o interface{}) (bool, error) {
	return false, nil
}

// 对已分类对象做统一处理，注意如果存在耗时操作，请使用其他协程处理。
// 该方法在Classify及BeanAfterSet后调用。
// 成功返回nil，失败返回error
func (p *jaegerProcessor) Process() error {
	return nil
}

func (p *jaegerProcessor) BeanDestroy() error {
	if p.closer != nil {
		return p.closer.Close()
	}
	return nil
}

func (p *jaegerProcessor) initTrace(conf fig.Properties) (opentracing.Tracer, io.Closer, error) {
	serviceName := conf.Get(keyServiceName, p.serviceName)
	if serviceName == "" {
		return nil, nil, fmt.Errorf("Neve trace: %s missing, please set it in application. ", keyServiceName)
	}

	var err error
	if p.sampler == nil {
		sn := conf.Get(keySamplerName, "")
		if sn == "" {
			return nil, nil, fmt.Errorf("Neve trace: %s missing, please set it in application. ", keySamplerName)
		}
		sv := conf.Get(keySamplerValue, "")
		if sv == "" {
			return nil, nil, fmt.Errorf("Neve trace: %s missing, please set it in application. ", keySamplerValue)
		}

		p.sampler, err = selectSampler(sn, sv)
		if err != nil {
			return nil, nil, err
		}
	}

	if p.reporter == nil {
		rn := conf.Get(keyReporterName, "")
		if rn == "" {
			return nil, nil, fmt.Errorf("Neve trace: %s missing, please set it in application. ", keyReporterName)
		}
		rv := conf.Get(keyReporterValue, "")
		//if rv == "" {
		//	return nil, nil, fmt.Errorf("Neve trace: %s missing, please set it in application. ", keyReporterValue)
		//}

		p.reporter, err = selectReporter(rn, rv)
		if err != nil {
			return nil, nil, err
		}
	}

	tracer, closer := jaeger.NewTracer(serviceName, p.sampler, p.reporter)
	return tracer, closer, nil
}

func selectReporter(name string, value string) (jaeger.Reporter, error) {
	switch name {
	case "remote":
		sender := transport.NewHTTPTransport(value)
		return jaeger.NewRemoteReporter(sender, jaeger.ReporterOptions.Logger(NewLogger())), nil
	case "inmemory":
		return jaeger.NewInMemoryReporter(), nil
	}
	return nil, fmt.Errorf("Reporter type: %s value: %s not support. ", name, value)
}

func selectSampler(name string, value string) (jaeger.Sampler, error) {
	switch name {
	case jaeger.SamplerTypeConst:
		v := true
		if strings.ToLower(value) == "false" {
			v = false
		}
		return jaeger.NewConstSampler(v), nil
	case jaeger.SamplerTypeProbabilistic:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		return jaeger.NewProbabilisticSampler(v)
	case jaeger.SamplerTypeRemote:
		return jaeger.NewRemotelyControlledSampler(value), nil
	case jaeger.SamplerTypeRateLimiting:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		return jaeger.NewRateLimitingSampler(v), nil
	}
	return nil, fmt.Errorf("Sampler type: %s value: %s not support. ", name, value)
}

type jaegarOpts struct{}

var JaegarOpts jaegarOpts

func (opt jaegarOpts) OptSetJaegerTracer(tracer opentracing.Tracer, closer io.Closer) JaegerOpt {
	return func(opts *jeagerOpts) {
		opts.tracer = tracer
		opts.closer = closer
	}
}

func (opt jaegarOpts) OptSetJaegerSampler(sampler jaeger.Sampler) JaegerOpt {
	return func(opts *jeagerOpts) {
		opts.sampler = sampler
	}
}

func (opt jaegarOpts) OptSetJaegerReporter(reporter jaeger.Reporter) JaegerOpt {
	return func(opts *jeagerOpts) {
		opts.reporter = reporter
	}
}

func (opt jaegarOpts) OptSetServiceName(serviceName string) JaegerOpt {
	return func(opts *jeagerOpts) {
		opts.serviceName = serviceName
	}
}
