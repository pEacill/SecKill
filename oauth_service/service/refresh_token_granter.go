package service

import (
	"context"
	"net/http"

	"github.com/pEacill/SecKill/oauth-service/errors"
	"github.com/pEacill/SecKill/oauth-service/model"
)

type RefreshTokenGranter struct {
	supportGrantType string
	tokenService     TokenService
}

func NewRefreshTokenGranter(grantType string, userDetailsService UserDetailsService, tokenService TokenService) TokenGranter {
	return &RefreshTokenGranter{
		supportGrantType: grantType,
		tokenService:     tokenService,
	}
}

func (tokenGranter *RefreshTokenGranter) Grant(ctx context.Context, grantType string, client *model.ClientDetails, reader *http.Request) (*model.OAuth2Token, error) {
	if grantType != tokenGranter.supportGrantType {
		return nil, errors.ErrNotSupportGrantType
	}

	refreshTokenValue := reader.URL.Query().Get("refresh_token")

	if refreshTokenValue == "" {
		return nil, errors.ErrInvalidTokenRequest
	}

	return tokenGranter.tokenService.RefreshAccessToken(refreshTokenValue)
}
