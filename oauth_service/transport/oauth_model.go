package transport

import (
	"context"

	"github.com/pEacill/SecKill/oauth-service/endpoint"
	"github.com/pEacill/SecKill/oauth-service/model"
	"github.com/pEacill/SecKill/pb"
)

func EncodeGRPCCheckTokenRequest(_ context.Context, r interface{}) (interface{}, error) {
	req := r.(*endpoint.CheckTokenRequest)
	return &pb.CheckTokenRequest{
		Token: req.Token,
	}, nil
}

func DecodeGRPCCheckTokenRequest(_ context.Context, r interface{}) (interface{}, error) {
	req := r.(*pb.CheckTokenRequest)
	return &endpoint.CheckTokenRequest{
		Token: req.Token,
	}, nil
}

func EncodeGRPCCheckTokenResponse(_ context.Context, r interface{}) (interface{}, error) {
	resp := r.(endpoint.ChecktokenResponse)

	if resp.Error != "" {
		return &pb.CheckTokenResponse{
			IsValidToken: false,
			Err:          resp.Error,
		}, nil
	} else {
		return &pb.CheckTokenResponse{
			UserDetails: &pb.UserDetails{
				UserId:      resp.Oauth2Details.User.UserId,
				Username:    resp.Oauth2Details.User.Username,
				Authorities: resp.Oauth2Details.User.Autorities,
			},
			ClientDetails: &pb.ClientDetails{
				ClientId:                    resp.Oauth2Details.Client.ClientId,
				AccessTokenValiditySeconds:  int32(resp.Oauth2Details.Client.AccessTokenValiditySeconds),
				RefreshTokenValiditySeconds: int32(resp.Oauth2Details.Client.RefreshTokenValiditySeconds),
				AuthorizedGrantTypes:        resp.Oauth2Details.Client.AuthorizedGrantTypes,
			},
			IsValidToken: true,
			Err:          "",
		}, nil
	}
}

func DecodeGRPCCheckTokenResponse(_ context.Context, r interface{}) (interface{}, error) {
	resp := r.(*pb.CheckTokenResponse)

	if resp.Err != "" {
		return endpoint.ChecktokenResponse{
			Error: resp.Err,
		}, nil
	} else {
		return endpoint.ChecktokenResponse{
			Oauth2Details: &model.OAuth2Details{
				User: &model.UserDetails{
					UserId:     resp.UserDetails.UserId,
					Username:   resp.UserDetails.Username,
					Autorities: resp.UserDetails.Authorities,
				},
				Client: &model.ClientDetails{
					ClientId:                    resp.ClientDetails.ClientId,
					AccessTokenValiditySeconds:  int(resp.ClientDetails.AccessTokenValiditySeconds),
					RefreshTokenValiditySeconds: int(resp.ClientDetails.RefreshTokenValiditySeconds),
					AuthorizedGrantTypes:        resp.ClientDetails.AuthorizedGrantTypes,
				},
			},
		}, nil
	}
}
