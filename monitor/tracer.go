package monitor

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sirupsen/logrus"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/yuanzhangcai/chaos/common"
)

var (
	// defaultTransport 全局变量，用于配置长链接。
	defaultTransport *http.Transport
)

func init() {
	defaultTransport = &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		MaxIdleConnsPerHost: 1000,
		MaxIdleConns:        5000,
	}
}

// Tracer 链路跟踪中间件
func Tracer(serviceName string, jaegerHostPort string) func(c *gin.Context) {
	cfg := &jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  "const", //固定采样
			Param: 1,       //1=全采样、0=不采样
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: jaegerHostPort,
		},
		ServiceName: serviceName,
	}
	tracer, _, err := cfg.NewTracer()
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	opentracing.SetGlobalTracer(tracer)

	return func(c *gin.Context) {
		err := c.Request.ParseForm()
		if err != nil {
			logrus.Panic("parse from failed")
		}

		var opts = []opentracing.StartSpanOption{
			opentracing.Tag{Key: string(ext.Component), Value: "HTTP"},
			opentracing.Tag{Key: "nid", Value: c.Request.Form.Get("nid")},
			opentracing.Tag{Key: "act_id", Value: c.Request.Form.Get("act_id")},
			opentracing.Tag{Key: "flow_id", Value: c.Request.Form.Get("flow_id")},
			ext.SpanKindRPCServer,
		}
		spCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))
		if err == nil {
			opts = append(opts, opentracing.ChildOf(spCtx))
		}

		span := opentracing.GlobalTracer().StartSpan(c.Request.URL.Path, opts...)
		defer span.Finish()
		c.Set("span", span)
		c.Set("spanCtx", opentracing.ContextWithSpan(context.Background(), span))
		c.Next()
	}
}

// StartSpan 创建跟踪span
func StartSpan(ctx context.Context, name string, tags map[string]interface{}) (opentracing.Span, context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}

	var opts = []opentracing.StartSpanOption{
		opentracing.Tag{Key: string(ext.Component), Value: "func"},
		ext.SpanKindRPCServer,
	}

	for key, value := range tags {
		opts = append(opts, opentracing.Tag{Key: key, Value: value})
	}

	return opentracing.StartSpanFromContext(ctx, name, opts...)
}

// SpanHTTP 发送http请求
func SpanHTTP(ctx context.Context, params *common.HTTPParam) ([]byte, int, error) {
	t := defaultTransport
	if params.UseShort {
		t = &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: true,
		}
	}

	var resBody []byte

	body := strings.NewReader(params.Data)
	client := &http.Client{Transport: t, Timeout: time.Second * time.Duration(params.Timeout)}

	params.Method = strings.ToUpper(params.Method)
	if params.Method == "" || params.Method == "GET" {
		params.Method = "GET"
		if !strings.Contains(params.URL, "?") {
			params.URL += "?" + params.Data
		} else {
			params.URL += "&" + params.Data
		}
	}

	request, err := http.NewRequest(params.Method, params.URL, body)
	if err != nil {
		return resBody, 0, err
	}

	// 设置header
	if params.Headers != nil {
		for key, value := range params.Headers {
			strValue := common.ToString(value)
			request.Header.Set(key, strValue)
		}
	}

	// 设置cookie
	if params.Cookies != nil {
		for key, value := range params.Cookies {
			request.AddCookie(&http.Cookie{Name: key, Value: common.ToString(value), HttpOnly: true})
		}
	}

	span, _ := opentracing.StartSpanFromContext(ctx, params.URL)
	defer span.Finish()

	tracer := opentracing.GlobalTracer()
	injectErr := tracer.(opentracing.Tracer).Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(request.Header))
	if injectErr != nil {
		log.Fatalf("%s: Couldn't inject headers", err)
	}

	response, err := client.Do(request)
	if err != nil {
		return resBody, 0, err
	}
	defer response.Body.Close()

	resBody, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return resBody, 0, err
	}
	return resBody, response.StatusCode, nil
}
