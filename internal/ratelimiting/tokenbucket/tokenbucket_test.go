package tokenbucket

import (
	"testing"
	"time"

	"github.com/mrvin/anti-bruteforce/internal/ratelimiting"
)

var confLimiterTest = ratelimiting.Conf{
	LimitLogin:    10,
	LimitPassword: 100,
	LimitIP:       1000,
	Interval:      100 * time.Millisecond, // 0,1 секунды
	TTLBucket:     1000 * time.Millisecond,
}

func TestAllowTokenBucket(t *testing.T) {
	limiter := New(&confLimiterTest)
	defer limiter.Stop()

	ratelimiting.RunTestAllow(t, limiter, &confLimiterTest)
}

func TestCleanBucketTokenBucket(t *testing.T) {
	limiter := New(&confLimiterTest)
	defer limiter.Stop()

	ratelimiting.RunTestCleanBucket(t, limiter, &confLimiterTest)
}
