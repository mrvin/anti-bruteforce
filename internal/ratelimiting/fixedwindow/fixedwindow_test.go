package fixedwindow

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

func TestAllowFixedWindow(t *testing.T) {
	limiter := New(&confLimiterTest)
	defer limiter.Stop()

	ratelimiting.RunTestAllow(t, limiter, &confLimiterTest)
}

func TestCleanBucketFixedWindow(t *testing.T) {
	limiter := New(&confLimiterTest)
	defer limiter.Stop()

	ratelimiting.RunTestCleanBucket(t, limiter, &confLimiterTest)
}

func TestCleanNonexistentBucketFixedWindow(t *testing.T) {
	limiter := New(&confLimiterTest)
	defer limiter.Stop()

	ratelimiting.RunCleanNonexistentBucket(t, limiter)
}
