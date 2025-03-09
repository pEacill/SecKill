package service

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/pEacill/SecKill/oauth-service/errors"
	"github.com/pEacill/SecKill/oauth-service/model"
	uuid "github.com/satori/go.uuid"
)

type TokenGranter interface {
	Grant(ctx context.Context, grantType string, client *model.ClientDetails, reader *http.Request) (*model.OAuth2Token, error)
}

type ComposeTokenGranter struct {
	TokenGrantDict map[string]TokenGranter
}

func NewComposeTokenGranter(tokenGrantDict map[string]TokenGranter) TokenGranter {
	return &ComposeTokenGranter{
		TokenGrantDict: tokenGrantDict,
	}
}

func (c *ComposeTokenGranter) Grant(ctx context.Context, grantType string, client *model.ClientDetails, reader *http.Request) (*model.OAuth2Token, error) {
	dispatchGranter := c.TokenGrantDict[grantType]

	if dispatchGranter == nil {
		return nil, errors.ErrNotSupportGrantType
	}

	return dispatchGranter.Grant(ctx, grantType, client, reader)
}

type TokenService interface {
	GetOAuth2DetailsByAccessToken(tokenValue string) (*model.OAuth2Details, error)

	CreateAccessToken(oauth2Details *model.OAuth2Details) (*model.OAuth2Token, error)

	RefreshAccessToken(refreshTokenValue string) (*model.OAuth2Token, error)

	GetAccessToken(details *model.OAuth2Details) (*model.OAuth2Token, error)

	ReadAccessToken(tokenValue string) (*model.OAuth2Token, error)
}

type DefaultTokenService struct {
	tokenStore    TokenStore
	tokenEnhancer TokenEnhancer
}

func NewTokenService(tokenStore TokenStore, tokenEnhancer TokenEnhancer) TokenService {
	return &DefaultTokenService{
		tokenStore:    tokenStore,
		tokenEnhancer: tokenEnhancer,
	}
}

func (tokenService *DefaultTokenService) GetOAuth2DetailsByAccessToken(tokenValue string) (*model.OAuth2Details, error) {
	accessToken, err := tokenService.tokenStore.ReadAccessToken(tokenValue)
	if err == nil {
		if accessToken.IsExpired() {
			return nil, errors.ErrExpiredToken
		}
		return tokenService.tokenStore.ReadOAuth2Details(tokenValue)
	}
	return nil, err
}

func (tokenService *DefaultTokenService) CreateAccessToken(oauth2Details *model.OAuth2Details) (*model.OAuth2Token, error) {
	existToken, err := tokenService.tokenStore.GetAccessToken(oauth2Details)
	var refreshToken *model.OAuth2Token
	if err == nil {
		if !existToken.IsExpired() {
			tokenService.tokenStore.StoreAccessToken(existToken, oauth2Details)
			return existToken, nil
		}

		tokenService.tokenStore.RemoveAccessToken(existToken.TokenValue)
		if existToken.RefreshToken != nil {
			refreshToken = existToken.RefreshToken
			tokenService.tokenStore.RemoveRefreshToken(refreshToken.TokenValue)
		}
	}

	if refreshToken == nil || refreshToken.IsExpired() {
		refreshToken, err = tokenService.createRefreshToken(oauth2Details)
		if err != nil {
			return nil, err
		}
	}

	accessToken, err := tokenService.createAccessToken(refreshToken, oauth2Details)
	if err == nil {
		tokenService.tokenStore.StoreAccessToken(accessToken, oauth2Details)
		tokenService.tokenStore.StoreRefreshToken(refreshToken, oauth2Details)
	}

	return accessToken, err
}

func (tokenService *DefaultTokenService) createRefreshToken(oauth2Details *model.OAuth2Details) (*model.OAuth2Token, error) {
	validitySeconds := oauth2Details.Client.RefreshTokenValiditySeconds
	s, _ := time.ParseDuration(strconv.Itoa(validitySeconds) + "s")
	expireTime := time.Now().Add(s)
	refreshToken := &model.OAuth2Token{
		ExpiresTime: &expireTime,
		TokenValue:  uuid.NewV4().String(),
	}

	if tokenService.tokenEnhancer != nil {
		return tokenService.tokenEnhancer.Enhance(refreshToken, oauth2Details)
	}

	return refreshToken, nil
}

func (tokenService *DefaultTokenService) createAccessToken(refreshToken *model.OAuth2Token, oauth2Details *model.OAuth2Details) (*model.OAuth2Token, error) {
	validitySeconds := oauth2Details.Client.AccessTokenValiditySeconds
	s, _ := time.ParseDuration(strconv.Itoa(validitySeconds) + "s")
	expiredTime := time.Now().Add(s)
	accessToken := &model.OAuth2Token{
		RefreshToken: refreshToken,
		ExpiresTime:  &expiredTime,
		TokenValue:   uuid.NewV4().String(),
	}

	if tokenService.tokenEnhancer != nil {
		return tokenService.tokenEnhancer.Enhance(accessToken, oauth2Details)
	}

	return accessToken, nil
}

func (tokenService *DefaultTokenService) RefreshAccessToken(refreshTokenValue string) (*model.OAuth2Token, error) {
	refreshToken, err := tokenService.tokenStore.ReadRefreshToken(refreshTokenValue)

	if err == nil {
		if refreshToken.IsExpired() {
			return nil, errors.ErrExpiredToken
		}

		oauth2Details, err := tokenService.tokenStore.ReadOAuth2DetailsForRefreshToken(refreshTokenValue)
		if err == nil {
			oauth2Token, err := tokenService.tokenStore.GetAccessToken(oauth2Details)
			if err == nil {
				tokenService.tokenStore.ReadAccessToken(oauth2Token.TokenValue)
			}

			tokenService.tokenStore.RemoveRefreshToken(refreshTokenValue)
			refreshToken, err = tokenService.createRefreshToken(oauth2Details)
			if err == nil {
				accessToken, err := tokenService.createAccessToken(refreshToken, oauth2Details)
				if err == nil {
					tokenService.tokenStore.StoreAccessToken(accessToken, oauth2Details)
					tokenService.tokenStore.StoreRefreshToken(refreshToken, oauth2Details)
				}
				return accessToken, err
			}
		}
	}

	return nil, err
}

func (tokenService *DefaultTokenService) GetAccessToken(details *model.OAuth2Details) (*model.OAuth2Token, error) {
	return tokenService.tokenStore.GetAccessToken(details)
}

func (tokenService *DefaultTokenService) ReadAccessToken(tokenValue string) (*model.OAuth2Token, error) {
	return tokenService.tokenStore.ReadAccessToken(tokenValue)
}
