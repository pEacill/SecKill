package service

import (
	"context"

	"github.com/pEacill/SecKill/oauth-service/errors"
	"github.com/pEacill/SecKill/oauth-service/model"
)

type ClientDetailsService interface {
	GetClientDetailsByClientId(ctx context.Context, clientId string, clientSecret string) (*model.ClientDetails, error)
}

type MysqlClientDetailsService struct{}

func NewMysqlClientDetailsService() ClientDetailsService {
	return &MysqlClientDetailsService{}
}

func (sservice *MysqlClientDetailsService) GetClientDetailsByClientId(ctx context.Context, clientId string, clientSecret string) (*model.ClientDetails, error) {
	clientDetailModel := model.NewClientDetailsModel()
	clientDetails, err := clientDetailModel.GetClientDetailsByClientId(clientId)
	if err != nil {
		return nil, err
	}

	if clientDetails.ClientSecret == clientSecret {
		return clientDetails, nil
	}

	return nil, errors.ErrClientMessage
}
