package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	kitzipkin "github.com/go-kit/kit/tracing/zipkin"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	"github.com/pEacill/SecKill/pb"
	"github.com/pEacill/SecKill/pkg/bootstrap"
	conf "github.com/pEacill/SecKill/pkg/config"
	register "github.com/pEacill/SecKill/pkg/discovery"
	"github.com/pEacill/SecKill/pkg/mysql"
	"github.com/pEacill/SecKill/pkg/ratelimiter"
	localconfig "github.com/pEacill/SecKill/user-service/config"
	"github.com/pEacill/SecKill/user-service/endpoint"
	"github.com/pEacill/SecKill/user-service/plugins"
	"github.com/pEacill/SecKill/user-service/service"
	"github.com/pEacill/SecKill/user-service/transport"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	var (
		servicePort = flag.String("service.port", bootstrap.HttpConfig.Port, "Service Port")
		grpcAddr    = flag.String("grpc", ":9000", "gRpc listen address")
	)
	flag.Parse()

	ctx := context.Background()
	errChan := make(chan error)

	fieldKeys := []string{"method"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "SecKill",
		Subsystem: "user_service",
		Name:      "request_count",
		Help:      "Nember of requests received.",
	}, fieldKeys)

	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "SecKill",
		Subsystem: "user_service",
		Name:      "request_latency",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)

	ratebucket := rate.NewLimiter(rate.Every(time.Second*1), 100)

	var svc service.Service
	svc = service.UserService{}

	// Add log middleware
	svc = plugins.LoggingMiddleware(localconfig.Logger)(svc)
	// Add prometheus middleware
	svc = plugins.Metrics(requestCount, requestLatency)(svc)

	userPorint := endpoint.MakeUserEndpoint(svc)
	userPorint = ratelimiter.NewTokenBucketLimitterWithBuildIn(ratebucket)(userPorint)
	userPorint = kitzipkin.TraceEndpoint(localconfig.ZipkinTracer, "user-endpoint")(userPorint)

	healthEndpoint := endpoint.MakeHealthCheckEndpoint(svc)
	healthEndpoint = kitzipkin.TraceEndpoint(localconfig.ZipkinTracer, "health-endpoint")(healthEndpoint)

	edpts := endpoint.UserEndpoints{
		UserEndpoint:        userPorint,
		HealthCheckEndpoint: healthEndpoint,
	}

	// Create HTTP handler
	r := transport.MakeHttpHandler(ctx, edpts, localconfig.ZipkinTracer, localconfig.Logger)

	// HTTP Server
	go func() {
		fmt.Println("HTTP Server start at Port: " + *servicePort)
		mysql.InitMysql(
			conf.MysqlConfig.Host,
			conf.MysqlConfig.Port,
			conf.MysqlConfig.User,
			conf.MysqlConfig.Pwd,
			conf.MysqlConfig.Db,
		)
		register.Register()
		handler := r
		errChan <- http.ListenAndServe(":"+*servicePort, handler)
	}()

	// GRpc Server
	go func() {
		fmt.Println("GRpc Server start at Port " + *grpcAddr)
		listener, err := net.Listen("tcp", *grpcAddr)
		if err != nil {
			localconfig.Logger.Log("GRpc Server start Fail!")
			errChan <- err
			return
		}

		serverTracer := kitzipkin.GRPCServerTrace(localconfig.ZipkinTracer, kitzipkin.Name("grpc-transport"))
		tr := localconfig.ZipkinTracer
		md := metadata.MD{}
		parentSpan := tr.StartSpan("test")

		b3.InjectGRPC(&md)(parentSpan.Context())

		ctx := metadata.NewIncomingContext(context.Background(), md)
		handler := transport.NewGRPCServer(ctx, edpts, serverTracer)
		gRPCServer := grpc.NewServer()
		pb.RegisterUserServiceServer(gRPCServer, handler)
		errChan <- gRPCServer.Serve(listener)
	}()

	// quit signal
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	error := <-errChan
	register.Deregister()
	localconfig.Logger.Log("User-service quit with error:", error)
}
