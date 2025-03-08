package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
	"github.com/pEacill/SecKill/user-service/service"
)

var (
	ErrInvalidRequestType = errors.New("invalid username, password")
)

type UserRequest struct {
	Username string `json:"user_name"`
	Password string `json:"password"`
}

type UserResponse struct {
	Result bool  `json:"result"`
	UserId int64 `json:"user_id"`
	Error  error `json:"error"`
}

type HealthCheckRequest struct{}

type HealthCheckResponse struct {
	Status bool `json:"status"`
}

type UserEndpoints struct {
	UserEndpoint        endpoint.Endpoint
	HealthCheckEndpoint endpoint.Endpoint
}

func MakeUserEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (responser interface{}, err error) {
		req := request.(UserRequest)

		username := req.Username
		password := req.Password
		userId, err := svc.Check(ctx, username, password)
		if err != nil {
			return UserResponse{Result: false, Error: err}, nil
		}

		return UserResponse{Result: true, UserId: userId, Error: err}, nil
	}
}

func MakeHealthCheckEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		status := svc.HealthCheck()
		return HealthCheckResponse{Status: status}, nil
	}
}

func (u *UserEndpoints) Check(ctx context.Context, username, password string) (int64, error) {
	resp, err := u.UserEndpoint(ctx, UserRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return 0, err
	}
	response := resp.(UserResponse)
	return response.UserId, nil
}

func (u *UserEndpoints) HealthCheck() bool {
	return true
}
