package model

import "time"

type OAuth2Token struct {
	RefreshToken *OAuth2Token
	TokenType    string
	TokenValue   string
	ExpiresTime  *time.Time
}

type OAuth2Details struct {
	Client *ClientDetails
	User   *UserDetails
}

func (token *OAuth2Token) IsExpired() bool {
	return token.ExpiresTime != nil && token.ExpiresTime.Before(time.Now())
}
