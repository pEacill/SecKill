package client

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/pEacill/SecKill/pb"
	"github.com/pEacill/SecKill/pkg/discovery"
	"github.com/pEacill/SecKill/pkg/loadbalance"
)

type UserClient interface {
	CheckUser(ctx context.Context, tracer opentracing.Tracer, request *pb.UserRequest) (*pb.UserResponse, error)
}

type UserClientImpl struct {
	manager     ClientManager
	serviceName string
	loadBalance loadbalance.LoadBalance
	tracer      opentracing.Tracer
}

func NewUserClient(serviceName string, lb loadbalance.LoadBalance, tracer opentracing.Tracer) (UserClient, error) {
	if serviceName == "" {
		serviceName = "user"
	}
	if lb == nil {
		lb = defaultLoadBalance
	}

	return &UserClientImpl{
		manager: &DefaultClientManager{
			serviceName:     serviceName,
			loadBalance:     lb,
			discoveryClient: discovery.DiscoverService,
			logger:          discovery.Logger,
		},
		serviceName: serviceName,
		loadBalance: lb,
		tracer:      tracer,
	}, nil
}

func (client *UserClientImpl) CheckUser(ctx context.Context, tracer opentracing.Tracer, request *pb.UserRequest) (*pb.UserResponse, error) {
	resp := new(pb.UserResponse)

	if err := client.manager.DecoratorInvoke("/pb.UserService/Check", "user_check", tracer, ctx, request, resp); err == nil {
		return resp, nil
	} else {
		return nil, err
	}
}
