package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/pEacill/SecKill/sk_app/model"
	"github.com/pEacill/SecKill/sk_app/service"
)

type SkAppEndpoints struct {
	SecKillEndpoint        endpoint.Endpoint
	HealthCheckEndPoint    endpoint.Endpoint
	GetSecInfoEndpoint     endpoint.Endpoint
	GetSecInfoListEndpoint endpoint.Endpoint
	TestEndpoint           endpoint.Endpoint
}

func (s SkAppEndpoints) HealthCheck() bool {
	return true
}

type SecInfoRequest struct {
	ProductId int `json:"id"`
}

type Response struct {
	Result map[string]interface{} `json:"result"`
	Error  error                  `json:"error"`
	Code   int                    `json:"code"`
}

type SecInfoListRequest struct{}

type SecInfoListResPonse struct {
	Result []map[string]interface{} `json:"result"`
	Num    int                      `json:"num"`
	Error  error                    `json:"error"`
}

type HealthCheckRequest struct{}

type HealthCheckResponse struct {
	Status bool `json:"statuc"`
}

func MakeSecInfoEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(SecInfoRequest)

		res := svc.SecInfo(req.ProductId)

		return Response{
			Result: res,
			Error:  nil,
		}, nil
	}
}

func MakeSecInfoListEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		res, num, err := svc.SecInfoList()
		return SecInfoListResPonse{
			Result: res,
			Num:    num,
			Error:  err,
		}, nil
	}
}

func MakeSecKillEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(model.SecRequest)
		res, code, err := svc.SecKill(&req)
		return Response{
			Result: res,
			Code:   code,
			Error:  err,
		}, nil
	}
}

func MakeHealthCheckEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		status := svc.HealthCheck()
		return HealthCheckResponse{
			Status: status,
		}, nil
	}
}

func MakeTestEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return Response{Result: nil, Code: 1, Error: nil}, nil
	}
}
