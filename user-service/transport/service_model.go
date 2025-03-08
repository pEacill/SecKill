package transport

import (
	"context"

	"github.com/pEacill/SecKill/pb"
	"github.com/pEacill/SecKill/user-service/endpoint"
)

func EncodeGRPCUserRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(pb.UserRequest)
	return &pb.UserRequest{
		Username: string(req.Username),
		Password: string(req.Password),
	}, nil
}

func DecodeGRPCUserRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.UserRequest)
	return endpoint.UserRequest{
		Username: string(req.Username),
		Password: string(req.Password),
	}, nil
}

func EncodeGRPCUserResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(endpoint.UserResponse)

	if resp.Error != nil {
		return &pb.UserResponse{
			Result: bool(resp.Result),
			Err:    "error",
		}, nil
	}

	return &pb.UserResponse{
		Result: bool(resp.Result),
		UserId: resp.UserId,
		Err:    "",
	}, nil
}

func DecodeGRPCUserResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*pb.UserResponse)

	return pb.UserResponse{
		Result: bool(resp.Result),
		Err:    resp.Err,
	}, nil
}
