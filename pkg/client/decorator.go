package client

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/pEacill/SecKill/pkg/bootstrap"
	conf "github.com/pEacill/SecKill/pkg/config"
	"github.com/pEacill/SecKill/pkg/discovery"
	"github.com/pEacill/SecKill/pkg/errors"
	"github.com/pEacill/SecKill/pkg/loadbalance"
	"google.golang.org/grpc"
)

var defaultLoadBalance loadbalance.LoadBalance = &loadbalance.RandomeLoadBalance{}

type ClientManager interface {
	DecoratorInvoke(path, hystrixName string, tracer opentracing.Tracer, ctx context.Context,
		inputVal, outVal interface{}) (err error)
}

type InvokerAfterFunc func() error
type InvokerBeforeFunc func() error

type DefaultClientManager struct {
	serviceName     string
	logger          *log.Logger
	discoveryClient discovery.DiscoveryClient
	loadBalance     loadbalance.LoadBalance
	before          []InvokerBeforeFunc
	after           []InvokerAfterFunc
}

func (manager *DefaultClientManager) DecoratorInvoke(path, hystrixName string, tracer opentracing.Tracer, ctx context.Context,
	inputVal, outVal interface{}) (err error) {

	for _, f := range manager.before {
		if err = f(); err != nil {
			return err
		}
	}

	if err = hystrix.Do(hystrixName, func() error {
		instances := manager.discoveryClient.DiscoverServices(manager.serviceName, manager.logger)
		if instance, err := manager.loadBalance.SelectService(instances); err == nil {
			if instance.GrpcPort > 0 {
				if conn, err := grpc.Dial(
					instance.Host+":"+strconv.Itoa(instance.GrpcPort),
					grpc.WithInsecure(),
					grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(genTracer(tracer), otgrpc.LogPayloads())),
					grpc.WithTimeout(time.Second*1),
				); err == nil {
					if err = conn.Invoke(ctx, path, inputVal, outVal); err != nil {
						return err
					}
				} else {
					return err
				}
			} else {
				return errors.ErrRPCService
			}
		} else {
			return err
		}
		return nil
	}, func(e error) error {
		return e
	}); err != nil {
		return err
	} else {
		for _, fn := range manager.after {
			if err = fn(); err != nil {
				return err
			}
		}
		return nil
	}
}

func genTracer(tracer opentracing.Tracer) opentracing.Tracer {
	if tracer != nil {
		return tracer
	}

	zipkinUrl := "http://" + conf.TraceConfig.Host + ":" + conf.TraceConfig.Port + conf.TraceConfig.Url

	reporter := zipkinhttp.NewReporter(zipkinUrl)

	localEndpoint, err := zipkin.NewEndpoint(bootstrap.DiscoverConfig.ServiceName, bootstrap.HttpConfig.Host+":"+bootstrap.HttpConfig.Port)
	if err != nil {
		log.Fatalf("zipkin.NewEndpoint err: %v", err)
	}

	nativeTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(localEndpoint))
	if err != nil {
		log.Fatalf("zipkin.NewTracer err: %v", err)
	}

	opentracingTracer := zipkinot.Wrap(nativeTracer)

	return opentracingTracer
}
