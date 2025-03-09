package service

import (
	"context"

	"github.com/pEacill/SecKill/oauth-service/errors"
	"github.com/pEacill/SecKill/oauth-service/model"
	"github.com/pEacill/SecKill/pb"
	"github.com/pEacill/SecKill/pkg/client"
)

type UserDetailsService interface {
	GetUserDetailByUsername(ctx context.Context, username, password string) (*model.UserDetails, error)
}

type RemoteUserService struct {
	userClient client.UserClient
}

func NewRemoteUserDetailService() *RemoteUserService {
	userClient, _ := client.NewUserClient("user", nil, nil)

	return &RemoteUserService{
		userClient: userClient,
	}
}

func (r *RemoteUserService) GetUserDetailByUsername(ctx context.Context, username, password string) (*model.UserDetails, error) {
	resp, err := r.userClient.CheckUser(ctx, nil, &pb.UserRequest{
		Username: username,
		Password: password,
	})
	if err == nil {
		if resp.UserId != 0 {
			return &model.UserDetails{
				UserId:   resp.UserId,
				Username: username,
				Password: password,
			}, nil
		} else {
			return nil, errors.ErrInvalidUserInfo
		}
	}

	return nil, err
}
