package slidingwindowcounter

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

func TestAllowSlidingWindowCounter(t *testing.T) {
	limiter := New(&confLimiterTest)
	defer limiter.Stop()

	ip := "127.0.0.1"
	password := "qwerty"
	login := "Bob"

	for i := 0; i < 9; i++ {
		if !limiter.Allow(ip, password, login) {
			t.Fatalf("Request %d in window 1 should be allowed", i+1)
		}
	}

	time.Sleep(confLimiterTest.Interval + confLimiterTest.Interval/2)

	for i := 0; i < 6; i++ {
		if !limiter.Allow(ip, password, login) {
			t.Fatalf("Request %d in window 1 should be allowed", i+1)
		}
	}

	if limiter.Allow(ip, password, login) {
		t.Error("16th request should be denied (exceeds limit)")
	}
}

func TestCleanBucketSlidingWindowCounter(t *testing.T) {
	limiter := New(&confLimiterTest)
	defer limiter.Stop()

	ratelimiting.RunTestCleanBucket(t, limiter, &confLimiterTest)
}

func TestCleanNonexistentBucketSlidingWindowCounter(t *testing.T) {
	limiter := New(&confLimiterTest)
	defer limiter.Stop()

	ratelimiting.RunCleanNonexistentBucket(t, limiter)
}
