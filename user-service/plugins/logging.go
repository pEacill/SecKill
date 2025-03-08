package plugins

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/pEacill/SecKill/user-service/service"
)

type loggingMiddleware struct {
	service.Service
	logger log.Logger
}

func LoggingMiddleware(logger log.Logger) service.ServiceMiddleware {
	return func(next service.Service) service.Service {
		return loggingMiddleware{next, logger}
	}
}

func (l loggingMiddleware) Check(ctx context.Context, username, password string) (res int64, err error) {
	defer func(begin time.Time) {
		_ = l.logger.Log(
			"Function", "Check",
			"Username", username,
			"Password", password,
			"result", res,
			"took", time.Since(begin),
		)
	}(time.Now())

	res, err = l.Service.Check(ctx, username, password)
	return res, err
}

func (l loggingMiddleware) HealthChcek() (result bool) {
	defer func(begin time.Time) {
		l.logger.Log(
			"Function", "Check",
			"Result", result,
			"took", time.Since(begin),
		)
	}(time.Now())
	result = l.Service.HealthCheck()
	return
}
