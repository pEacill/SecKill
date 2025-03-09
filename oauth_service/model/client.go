package model

import (
	"encoding/json"

	"github.com/pEacill/SecKill/pkg/mysql"
	"gorm.io/gorm"
)

type ClientDetails struct {
	ClientId                    string
	ClientSecret                string
	AccessTokenValiditySeconds  int
	RefreshTokenValiditySeconds int
	RegisteredRedirectUri       string
	AuthorizedGrantTypes        []string
}

func (clientDetails *ClientDetails) IsMatch(clientId string, clientSecret string) bool {
	return clientId == clientDetails.ClientId && clientSecret == clientDetails.ClientSecret
}

type ClientDetailsDB struct {
	ClientId                    string          `gorm:"column:client_id"`
	ClientSecret                string          `gorm:"column:client_secret"`
	AccessTokenValiditySeconds  int             `gorm:"column:access_token_validity_seconds"`
	RefreshTokenValiditySeconds int             `gorm:"column:refresh_token_validity_seconds"`
	RegisteredRedirectUri       string          `gorm:"column:registered_redirect_uri"`
	AuthorizedGrantTypesJSON    json.RawMessage `gorm:"column:authorized_grant_types"`
}

type ClientDetailsModel struct {
	DB *gorm.DB
}

func NewClientDetailsModel() *ClientDetailsModel {
	return &ClientDetailsModel{
		DB: mysql.DB(),
	}
}

func (c *ClientDetailsModel) getTableName() string {
	return "client_details"
}

func (p *ClientDetailsModel) GetClientDetailsByClientId(clientId string) (*ClientDetails, error) {
	var clientDetailsDB ClientDetailsDB
	if err := p.DB.Table(p.getTableName()).Where("client_id = ?", clientId).First(&clientDetailsDB).Error; err != nil {
		return nil, err
	}

	var authorizedGrantTypes []string
	err := json.Unmarshal(clientDetailsDB.AuthorizedGrantTypesJSON, &authorizedGrantTypes)
	if err != nil {
		return nil, err
	}

	return &ClientDetails{
		ClientId:                    clientDetailsDB.ClientId,
		ClientSecret:                clientDetailsDB.ClientSecret,
		AccessTokenValiditySeconds:  clientDetailsDB.AccessTokenValiditySeconds,
		RefreshTokenValiditySeconds: clientDetailsDB.RefreshTokenValiditySeconds,
		RegisteredRedirectUri:       clientDetailsDB.RegisteredRedirectUri,
		AuthorizedGrantTypes:        authorizedGrantTypes,
	}, nil
}

func (p *ClientDetailsModel) CreateClientDetails(clientDetails *ClientDetails) error {
	grantTypeString, _ := json.Marshal(clientDetails.AuthorizedGrantTypes)
	if err := p.DB.Table(p.getTableName()).Create(map[string]interface{}{
		"client_id":                      clientDetails.ClientId,
		"client_secret":                  clientDetails.ClientSecret,
		"access_token_validity_seconds":  clientDetails.AccessTokenValiditySeconds,
		"refresh_token_validity_seconds": clientDetails.RefreshTokenValiditySeconds,
		"registered_redirect_uri":        clientDetails.RegisteredRedirectUri,
		"authorized_grant_types":         grantTypeString,
	}).Error; err != nil {
		return err
	}
	return nil
}
