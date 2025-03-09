package model

import (
	"testing"

	"github.com/pEacill/SecKill/pkg/mysql"
)

var (
	hostMysql = "localhost"
	portMysql = "3306"
	userMysql = "root"
	pwdMysql  = "root"
	dbMysql   = "oauth"
)

func TestClient(t *testing.T) {
	mysql.InitMysql(hostMysql, portMysql, userMysql, pwdMysql, dbMysql)
	if mysql.DB() == nil {
		t.Fatal("Database connection failed")
	}

	clientDetailsModel := NewClientDetailsModel()

	t.Run("CreateClientDetails", func(t *testing.T) {
		client := &ClientDetails{
			ClientId:                    "test_client_id",
			ClientSecret:                "test_client_secret",
			AccessTokenValiditySeconds:  3600,
			RefreshTokenValiditySeconds: 7200,
			RegisteredRedirectUri:       "",
			AuthorizedGrantTypes:        []string{"authorization_code", "refresh_token"},
		}

		err := clientDetailsModel.CreateClientDetails(client)
		if err != nil {
			t.Errorf("Failed to create client details: %v", err)
		}
	})

	t.Run("GetClientDetailsByClientId", func(t *testing.T) {
		clientId := "test_client_id"
		client, err := clientDetailsModel.GetClientDetailsByClientId(clientId)
		if err != nil {
			t.Errorf("Failed to get client details: %v", err)
		}

		if client == nil {
			t.Errorf("Client details not found")
		}

		if client.ClientId != clientId {
			t.Errorf("Client ID mismatch")
		}

		if client.ClientSecret != "test_client_secret" {
			t.Errorf("Client secret mismatch")
		}

		if client.AccessTokenValiditySeconds != 3600 {
			t.Errorf("Access token validity seconds mismatch")
		}

		if client.RefreshTokenValiditySeconds != 7200 {
			t.Errorf("Refresh token validity seconds mismatch")
		}

		if client.RegisteredRedirectUri != "" {
			t.Errorf("Registered redirect URI mismatch")
		}

		if len(client.AuthorizedGrantTypes) != 2 || client.AuthorizedGrantTypes[0] != "authorization_code" || client.AuthorizedGrantTypes[1] != "refresh_token" {
			t.Errorf("Authorized grant types mismatch")
		}
	})
}
