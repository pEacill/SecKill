package plugins

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/pEacill/SecKill/user-service/service"
)

type metricMiddleware struct {
	service.Service
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
}

func Metrics(requestCount metrics.Counter, requestLatency metrics.Histogram) service.ServiceMiddleware {
	return func(next service.Service) service.Service {
		return metricMiddleware{
			next,
			requestCount,
			requestLatency,
		}
	}
}

func (m metricMiddleware) Check(ctx context.Context, username, password string) (res int64, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "Check"}
		m.requestCount.With(lvs...).Add(1)
		m.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	res, err = m.Service.Check(ctx, username, password)
	return
}

func (m metricMiddleware) HealthCheck() (result bool) {

	defer func(begin time.Time) {
		lvs := []string{"method", "HealthCheck"}
		m.requestCount.With(lvs...).Add(1)
		m.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	result = m.Service.HealthCheck()
	return
}
