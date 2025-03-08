package service

import (
	"context"
	"log"

	"github.com/pEacill/SecKill/user-service/model"
)

type Service interface {
	Check(ctx context.Context, username, password string) (int64, error)

	HealthCheck() bool
}

type UserService struct{}

func (s UserService) Check(ctx context.Context, username, password string) (int64, error) {
	model := model.NewUserModel()
	res, err := model.CheckUser(username, password)
	if err != nil {
		log.Printf("user-service.model.CheckUser Error: %v", err)
		return 0, err
	}
	return res.UserId, nil
}

func (s UserService) HealthCheck() bool {
	return true
}

type ServiceMiddleware func(Service) Service
