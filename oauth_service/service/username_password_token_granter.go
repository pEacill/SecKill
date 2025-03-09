package service

import (
	"context"
	"net/http"

	"github.com/pEacill/SecKill/oauth-service/errors"
	"github.com/pEacill/SecKill/oauth-service/model"
)

type UsernamePasswordTokenGranter struct {
	supportGrantType   string
	userDetailsService UserDetailsService
	tokenService       TokenService
}

func NewUsernamePasswordTokenGranter(grantType string, userDetailsService UserDetailsService, tokenService TokenService) TokenGranter {
	return &UsernamePasswordTokenGranter{
		supportGrantType:   grantType,
		userDetailsService: userDetailsService,
		tokenService:       tokenService,
	}
}

func (tokenGranter *UsernamePasswordTokenGranter) Grant(ctx context.Context, grantType string, client *model.ClientDetails, reader *http.Request) (*model.OAuth2Token, error) {
	if grantType != tokenGranter.supportGrantType {
		return nil, errors.ErrNotSupportGrantType
	}

	username := reader.FormValue("username")
	password := reader.FormValue("password")
	if username == "" || password == "" {
		return nil, errors.ErrInvalidUsernameOrPasswordRequestEmpty
	}

	userDetails, err := tokenGranter.userDetailsService.GetUserDetailByUsername(ctx, username, password)
	if err != nil {
		return nil, errors.ErrInvalidUsernameOrPasswordRequest
	}

	return tokenGranter.tokenService.CreateAccessToken(&model.OAuth2Details{
		Client: client,
		User:   userDetails,
	})
}
