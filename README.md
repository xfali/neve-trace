# neve-trace

## With Environment variables
当由application配置文件中读取配置失败时，尝试从环境变量中读取配置，配置见
[jaeger环境变量](https://github.com/jaegertracing/jaeger-client-go/blob/master/README.md)

## 内置tag
[opentracing](https://github.com/opentracing/specification/blob/master/semantic_conventions.md)

## DEBUG
1. 参考: [getting started](https://www.jaegertracing.io/docs/1.25/getting-started/)

2. 下载: [All-in-One](https://www.jaegertracing.io/download/)

3. 执行:
```
jaeger-all-in-one --collector.zipkin.host-port=:9411
```
4. 配置remote reporter直连地址：http://localhost:14268/api/traces