package ratelimiter

import (
	"context"
	"sync"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/pEacill/SecKill/pkg/errors"
)

type LeakyBucket struct {
	capacity     int
	rate         time.Duration
	currentWater int
	lastLeakTime time.Time
	mu           sync.Mutex
}

func NewLeakyBucket(capacity int, rate time.Duration) *LeakyBucket {
	return &LeakyBucket{
		capacity:     capacity,
		rate:         rate,
		currentWater: 0,
		lastLeakTime: time.Now(),
	}
}

func (lb *LeakyBucket) Allow() bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(lb.lastLeakTime)

	leaks := int(elapsed / lb.rate)
	if leaks > 0 {
		lb.currentWater -= leaks
		if lb.currentWater < 0 {
			lb.currentWater = 0
		}
		lb.lastLeakTime = lb.lastLeakTime.Add(time.Duration(leaks) * lb.rate)
	}

	if lb.currentWater < lb.capacity {
		lb.currentWater++
		return true
	}

	return false
}

func NewLeakyBucketLimiter(capacity int, leakRate time.Duration) endpoint.Middleware {
	bucket := NewLeakyBucket(capacity, leakRate)
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			if !bucket.Allow() {
				return nil, errors.ErrLimitExceed
			}
			return next(ctx, request)
		}
	}
}
