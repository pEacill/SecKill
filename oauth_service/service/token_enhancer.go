package service

import (
	"time"

	"github.com/form3tech-oss/jwt-go"
	"github.com/pEacill/SecKill/oauth-service/model"
)

type TokenEnhancer interface {
	Enhance(oauth2Token *model.OAuth2Token, oauth2Details *model.OAuth2Details) (*model.OAuth2Token, error)

	Extract(tokenValue string) (*model.OAuth2Token, *model.OAuth2Details, error)
}

func NewJwtTokenEnhancer(secretKey string) TokenEnhancer {
	return &JwtTokenEnhancer{
		secretKey: []byte(secretKey),
	}
}

type OAuth2TokenCustomClaims struct {
	UserDetails   model.UserDetails
	ClientDetails model.ClientDetails
	RefreshToken  model.OAuth2Token
	jwt.StandardClaims
}

type JwtTokenEnhancer struct {
	secretKey []byte
}

func (enhancer *JwtTokenEnhancer) Enhance(oauth2Token *model.OAuth2Token, oauth2Details *model.OAuth2Details) (*model.OAuth2Token, error) {
	return enhancer.sign(oauth2Token, oauth2Details)
}

func (enhancer *JwtTokenEnhancer) Extract(tokenValue string) (*model.OAuth2Token, *model.OAuth2Details, error) {
	token, err := jwt.ParseWithClaims(tokenValue, &OAuth2TokenCustomClaims{}, func(t *jwt.Token) (i interface{}, err error) {
		return enhancer.secretKey, nil
	})
	if err != nil {
		return nil, nil, err
	}

	claims := token.Claims.(*OAuth2TokenCustomClaims)
	expiretime := time.Unix(claims.ExpiresAt, 0)

	return &model.OAuth2Token{
			RefreshToken: &claims.RefreshToken,
			TokenValue:   tokenValue,
			ExpiresTime:  &expiretime,
		}, &model.OAuth2Details{
			User:   &claims.UserDetails,
			Client: &claims.ClientDetails,
		}, nil
}

func (enhancer *JwtTokenEnhancer) sign(oauth2Token *model.OAuth2Token, oauth2Details *model.OAuth2Details) (*model.OAuth2Token, error) {
	expiretime := oauth2Token.ExpiresTime
	clientDetails := *oauth2Details.Client
	userDetails := *oauth2Details.User
	clientDetails.ClientSecret = ""
	userDetails.Password = ""

	claims := OAuth2TokenCustomClaims{
		UserDetails:   userDetails,
		ClientDetails: clientDetails,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiretime.Unix(),
			Issuer:    "System",
		},
	}

	if oauth2Token.RefreshToken != nil {
		claims.RefreshToken = *oauth2Token.RefreshToken
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenValue, err := token.SignedString(enhancer.secretKey)
	if err != nil {
		return nil, err
	}

	oauth2Token.TokenValue = tokenValue
	oauth2Token.TokenType = "jwt"
	return oauth2Token, nil
}
