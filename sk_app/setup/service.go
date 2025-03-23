package setup

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	kitzipkin "github.com/go-kit/kit/tracing/zipkin"
	conf "github.com/pEacill/SecKill/pkg/config"
	register "github.com/pEacill/SecKill/pkg/discovery"
	"github.com/pEacill/SecKill/pkg/ratelimiter"
	"github.com/pEacill/SecKill/sk_app/config"
	"github.com/pEacill/SecKill/sk_app/endpoint"
	"github.com/pEacill/SecKill/sk_app/plugins"
	"github.com/pEacill/SecKill/sk_app/service"
	"github.com/pEacill/SecKill/sk_app/transport"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
)

func InitServer(host string, servicePort string) {
	log.Printf("sk_app server port: %v", servicePort)
	flag.Parse()
	errChan := make(chan error)

	fieldKeys := []string{"method"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "Seckil",
		Subsystem: "sk_app",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)

	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "Seckill",
		Subsystem: "sk_app",
		Name:      "request_latency",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)

	ratebucket := rate.NewLimiter(rate.Every(time.Second*1), 5000)

	var skAppService service.Service
	skAppService = service.SkAppService{}

	skAppService = plugins.SkAppLoggingMiddleware(config.Logger)(skAppService)
	skAppService = plugins.SkAppMetrics(requestCount, requestLatency)(skAppService)

	healthCheckEnd := endpoint.MakeHealthCheckEndpoint(skAppService)
	healthCheckEnd = ratelimiter.NewTokenBucketLimitterWithBuildIn(ratebucket)(healthCheckEnd)
	healthCheckEnd = kitzipkin.TraceEndpoint(conf.ZipkinTracer, "health-check")(healthCheckEnd)

	GetSecInfoEnd := endpoint.MakeSecInfoEndpoint(skAppService)
	GetSecInfoEnd = ratelimiter.NewTokenBucketLimitterWithBuildIn(ratebucket)(GetSecInfoEnd)
	GetSecInfoEnd = kitzipkin.TraceEndpoint(conf.ZipkinTracer, "sec-info")(GetSecInfoEnd)

	GetSecInfoListEnd := endpoint.MakeSecInfoListEndpoint(skAppService)
	GetSecInfoListEnd = ratelimiter.NewTokenBucketLimitterWithBuildIn(ratebucket)(GetSecInfoListEnd)
	GetSecInfoListEnd = kitzipkin.TraceEndpoint(conf.ZipkinTracer, "sec-info-list")(GetSecInfoListEnd)

	secRateBucket := rate.NewLimiter(rate.Every(time.Microsecond*100), 1000)

	SecKillEnd := endpoint.MakeSecInfoEndpoint(skAppService)
	SecKillEnd = ratelimiter.NewTokenBucketLimitterWithBuildIn(secRateBucket)(SecKillEnd)
	SecKillEnd = kitzipkin.TraceEndpoint(conf.ZipkinTracer, "sec-kill")(SecKillEnd)

	testEnd := endpoint.MakeTestEndpoint(skAppService)
	testEnd = kitzipkin.TraceEndpoint(conf.ZipkinTracer, "test")(testEnd)

	endpts := endpoint.SkAppEndpoints{
		SecKillEndpoint:        SecKillEnd,
		HealthCheckEndPoint:    healthCheckEnd,
		GetSecInfoEndpoint:     GetSecInfoEnd,
		GetSecInfoListEndpoint: GetSecInfoListEnd,
		TestEndpoint:           testEnd,
	}
	ctx := context.Background()

	r := transport.MakeHttpHandler(ctx, endpts, conf.ZipkinTracer, conf.Logger)

	go func() {
		fmt.Println("sk_app Http Server start at port:" + servicePort)
		register.Register()
		handler := r
		errChan <- http.ListenAndServe(":"+servicePort, handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	error := <-errChan
	register.Deregister()
	fmt.Printf("sk_app Server down, with error: %V", error)
}
