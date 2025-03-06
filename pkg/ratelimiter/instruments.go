package ratelimiter

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/pEacill/SecKill/pkg/errors"
	"golang.org/x/time/rate"
)

func NewTokenBucketLimitterWithBuildIn(bkt *rate.Limiter) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			if !bkt.Allow() {
				return nil, errors.ErrLimitExceed
			}

			return next(ctx, request)
		}
	}
}

func NewDynamicLimitter(interval int, burst int) endpoint.Middleware {
	bucket := rate.NewLimiter(rate.Every(time.Second*time.Duration(interval)), burst)
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			if !bucket.Allow() {
				return nil, errors.ErrLimitExceed
			}
			return next(ctx, request)
		}
	}
}
