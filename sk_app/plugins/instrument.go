package plugins

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/pEacill/SecKill/sk_app/model"
	"github.com/pEacill/SecKill/sk_app/service"
)

type skAppMetricMiddleware struct {
	service.Service
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
}

func SkAppMetrics(requestCount metrics.Counter, requestLatency metrics.Histogram) service.ServiceMiddleware {
	return func(next service.Service) service.Service {
		return skAppMetricMiddleware{
			next,
			requestCount,
			requestLatency}
	}
}

func (mw skAppMetricMiddleware) HealthCheck() (result bool) {

	defer func(begin time.Time) {
		lvs := []string{"method", "HealthCheck"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	result = mw.Service.HealthCheck()
	return
}

func (mw skAppMetricMiddleware) SecInfo(productId int) map[string]interface{} {

	defer func(begin time.Time) {
		lvs := []string{"method", "SecInfo"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	ret := mw.Service.SecInfo(productId)
	return ret
}

func (mw skAppMetricMiddleware) SecInfoList() ([]map[string]interface{}, int, error) {

	defer func(begin time.Time) {
		lvs := []string{"method", "SecInfoList"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	data, num, error := mw.Service.SecInfoList()
	return data, num, error
}

func (mw skAppMetricMiddleware) SecKill(req *model.SecRequest) (map[string]interface{}, int, error) {

	defer func(begin time.Time) {
		lvs := []string{"method", "SecKill"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	result, num, error := mw.Service.SecKill(req)
	return result, num, error
}
