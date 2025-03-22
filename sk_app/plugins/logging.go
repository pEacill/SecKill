package plugins

import (
	"time"

	"github.com/go-kit/kit/log"
	"github.com/pEacill/SecKill/sk_app/model"
	"github.com/pEacill/SecKill/sk_app/service"
)

type skAppLoggingMiddleware struct {
	service.Service
	logger log.Logger
}

func SkAppLoggingMiddleware(logger log.Logger) service.ServiceMiddleware {
	return func(next service.Service) service.Service {
		return skAppLoggingMiddleware{next, logger}
	}
}

func (mw skAppLoggingMiddleware) HealthCheck() (result bool) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"function", "HealthChcek",
			"result", result,
			"took", time.Since(begin),
		)
	}(time.Now())

	result = mw.Service.HealthCheck()
	return
}

func (mw skAppLoggingMiddleware) SecInfo(productId int) map[string]interface{} {

	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"function", "SecInfo",
			"took", time.Since(begin),
		)
	}(time.Now())

	ret := mw.Service.SecInfo(productId)
	return ret
}

func (mw skAppLoggingMiddleware) SecInfoList() ([]map[string]interface{}, int, error) {

	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"function", "SecInfoList",
			"took", time.Since(begin),
		)
	}(time.Now())

	data, num, error := mw.Service.SecInfoList()
	return data, num, error
}

func (mw skAppLoggingMiddleware) SecKill(req *model.SecRequest) (map[string]interface{}, int, error) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"function", "SecKill",
			"took", time.Since(begin),
		)
	}(time.Now())

	result, num, error := mw.Service.SecKill(req)
	return result, num, error
}
