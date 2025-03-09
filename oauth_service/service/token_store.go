package service

import (
	"github.com/pEacill/SecKill/oauth-service/errors"
	"github.com/pEacill/SecKill/oauth-service/model"
)

type TokenStore interface {
	StoreAccessToken(oauth2Token *model.OAuth2Token, oauth2Details *model.OAuth2Details)

	ReadAccessToken(tokenValue string) (*model.OAuth2Token, error)

	ReadOAuth2Details(tokenValue string) (*model.OAuth2Details, error)

	GetAccessToken(oauth2Details *model.OAuth2Details) (*model.OAuth2Token, error)

	RemoveAccessToken(tokenValue string)

	StoreRefreshToken(oauth2Token *model.OAuth2Token, oauth2Details *model.OAuth2Details)

	RemoveRefreshToken(tokenValue string)

	ReadRefreshToken(tokenValue string) (*model.OAuth2Token, error)

	ReadOAuth2DetailsForRefreshToken(tokenValue string) (*model.OAuth2Details, error)
}

type JwtTokenStore struct {
	jwtTokenEnhancer *JwtTokenEnhancer
}

func NewJwtTokenStore(jwtTokenEnhancer *JwtTokenEnhancer) TokenStore {
	return &JwtTokenStore{
		jwtTokenEnhancer: jwtTokenEnhancer,
	}
}

func (tokenStore *JwtTokenStore) StoreAccessToken(oauth2Token *model.OAuth2Token, oauth2Details *model.OAuth2Details) {

}

func (tokenStore *JwtTokenStore) ReadAccessToken(tokenValue string) (*model.OAuth2Token, error) {
	oauth2Token, _, err := tokenStore.jwtTokenEnhancer.Extract(tokenValue)
	return oauth2Token, err
}

func (tokenStore *JwtTokenStore) ReadOAuth2Details(tokenValue string) (*model.OAuth2Details, error) {
	_, oauth2Details, err := tokenStore.jwtTokenEnhancer.Extract(tokenValue)
	return oauth2Details, err
}

func (tokenStore *JwtTokenStore) GetAccessToken(oauth2Details *model.OAuth2Details) (*model.OAuth2Token, error) {
	return nil, errors.ErrNotSupportOperation
}

func (tokenStore *JwtTokenStore) RemoveAccessToken(tokenValue string) {

}

func (tokenStore *JwtTokenStore) StoreRefreshToken(oauth2Token *model.OAuth2Token, oauth2Details *model.OAuth2Details) {

}

func (tokenStore *JwtTokenStore) RemoveRefreshToken(tokenValue string) {

}

func (tokenStore *JwtTokenStore) ReadRefreshToken(tokenValue string) (*model.OAuth2Token, error) {
	oauth2Token, _, err := tokenStore.jwtTokenEnhancer.Extract(tokenValue)
	return oauth2Token, err
}

func (tokenStore *JwtTokenStore) ReadOAuth2DetailsForRefreshToken(tokenValue string) (*model.OAuth2Details, error) {
	_, oauth2Details, err := tokenStore.jwtTokenEnhancer.Extract(tokenValue)
	return oauth2Details, err
}
