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

	kitzipkin "github.com/go-kit/kit/tracing/zipkin"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	localconfig "github.com/pEacill/SecKill/oauth-service/config"
	"github.com/pEacill/SecKill/oauth-service/endpoint"
	"github.com/pEacill/SecKill/oauth-service/service"
	"github.com/pEacill/SecKill/oauth-service/transport"
	"github.com/pEacill/SecKill/pb"
	"github.com/pEacill/SecKill/pkg/bootstrap"
	conf "github.com/pEacill/SecKill/pkg/config"
	register "github.com/pEacill/SecKill/pkg/discovery"
	"github.com/pEacill/SecKill/pkg/mysql"
	"github.com/pEacill/SecKill/pkg/ratelimiter"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	var (
		servicePort = flag.String("service.port", bootstrap.HttpConfig.Port, "Service Port")
		grpcAddr    = flag.String("grpc", ":9002", "gRpc listen address")
	)
	flag.Parse()

	ctx := context.Background()
	errChan := make(chan error)

	ratebucket := rate.NewLimiter(rate.Every(time.Second*1), 100)

	tokenEnhancer := service.NewJwtTokenEnhancer("secret")
	tokenStore := service.NewJwtTokenStore(tokenEnhancer.(*service.JwtTokenEnhancer))
	tokenService := service.NewTokenService(tokenStore, tokenEnhancer)
	userDetailsService := service.NewRemoteUserDetailService()
	clientDetailsService := service.NewMysqlClientDetailsService()
	heanlthCheckService := service.NewCommonService()

	tokenGranter := service.NewComposeTokenGranter(map[string]service.TokenGranter{
		"password":      service.NewUsernamePasswordTokenGranter("password", userDetailsService, tokenService),
		"refresh_token": service.NewRefreshTokenGranter("refresh_token", userDetailsService, tokenService),
	})

	tokenEndpoint := endpoint.MakeTokenEndpoint(tokenGranter, clientDetailsService)
	tokenEndpoint = endpoint.MakeClientAuthorizationMiddleware(localconfig.Logger)(tokenEndpoint)
	tokenEndpoint = ratelimiter.NewTokenBucketLimitterWithBuildIn(ratebucket)(tokenEndpoint)
	tokenEndpoint = kitzipkin.TraceEndpoint(localconfig.ZipkinTracer, "token-endpoint")(tokenEndpoint)

	checkTokenEndpoint := endpoint.MakeCheckTokenEndpoint(tokenService)
	checkTokenEndpoint = endpoint.MakeClientAuthorizationMiddleware(localconfig.Logger)(checkTokenEndpoint)
	checkTokenEndpoint = ratelimiter.NewTokenBucketLimitterWithBuildIn(ratebucket)(checkTokenEndpoint)
	checkTokenEndpoint = kitzipkin.TraceEndpoint(localconfig.ZipkinTracer, "check-token-endpoint")(checkTokenEndpoint)

	gRPCCheckTokenEndpoint := endpoint.MakeCheckTokenEndpoint(tokenService)
	gRPCCheckTokenEndpoint = ratelimiter.NewTokenBucketLimitterWithBuildIn(ratebucket)(gRPCCheckTokenEndpoint)
	gRPCCheckTokenEndpoint = kitzipkin.TraceEndpoint(localconfig.ZipkinTracer, "grpc-check-endpoint")(gRPCCheckTokenEndpoint)

	healthEndpoint := endpoint.MakeHealthCheckEndpoint(heanlthCheckService)
	healthEndpoint = kitzipkin.TraceEndpoint(localconfig.ZipkinTracer, "health-endpoint")(healthEndpoint)

	endpts := endpoint.OAuth2Endpoints{
		TokenEndpoint:          tokenEndpoint,
		CheckTokenEndpoint:     checkTokenEndpoint,
		HealthCheckEndpoint:    healthEndpoint,
		GRPCCheckTokenEndpoint: gRPCCheckTokenEndpoint,
	}

	r := transport.MakeHttpHandler(ctx, endpts, tokenService, clientDetailsService, localconfig.ZipkinTracer, localconfig.Logger)

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
		handler := transport.NewGRPCServer(ctx, endpts, serverTracer)
		gRPCServer := grpc.NewServer()
		pb.RegisterOAuthServiceServer(gRPCServer, handler)
		errChan <- gRPCServer.Serve(listener)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	error := <-errChan
	register.Deregister()
	localconfig.Logger.Log("OAuth-service quit with error:", error)
}
