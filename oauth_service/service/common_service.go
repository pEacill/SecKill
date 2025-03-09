package service

type Service interface {
	HealthCheck() bool
}

type CommonService struct{}

func NewCommonService() *CommonService {
	return &CommonService{}
}

func (c *CommonService) HealthCheck() bool {
	return true
}
