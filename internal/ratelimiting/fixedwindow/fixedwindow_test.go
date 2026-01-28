package fixedwindow

import (
	"sync"
	"sync/atomic"
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

const numGoroutine = 20

func TestAllow(t *testing.T) {
	const numRepetition = 2

	limiter := New(&confLimiterTest)
	defer limiter.Stop()

	ip := "127.0.0.1"
	password := "qwerty"
	login := "Bob"
	wantAllowedRequests := min(confLimiterTest.LimitIP, confLimiterTest.LimitPassword, confLimiterTest.LimitLogin)
	for i := range numRepetition {
		var wg sync.WaitGroup
		var allowedRequests atomic.Uint64
		for _ = range numGoroutine {
			wg.Go(func() {
				if got := limiter.Allow(ip, password, login); got {
					allowedRequests.Add(1)
				}
			})
		}
		wg.Wait()

		if gotAllowedRequests := allowedRequests.Load(); gotAllowedRequests != wantAllowedRequests {
			t.Errorf("Allowed requests: got: %d want: %d", gotAllowedRequests, wantAllowedRequests)
		}

		if i != numRepetition-1 {
			time.Sleep(confLimiterTest.Interval)
		}
	}
}

func TestCleanBucket(t *testing.T) {
	const numRepetition = 5

	limiter := New(&confLimiterTest)
	defer limiter.Stop()

	ip := "127.0.0.1"
	password := "qwerty"
	login := "Bob"
	wantAllowedRequests := min(confLimiterTest.LimitIP, confLimiterTest.LimitPassword, confLimiterTest.LimitLogin)
	for _ = range numRepetition {
		var wg sync.WaitGroup
		var allowedRequests atomic.Uint64
		for _ = range numGoroutine {
			wg.Go(func() {
				if got := limiter.Allow(ip, password, login); got {
					allowedRequests.Add(1)
				}
			})
		}
		wg.Wait()

		if gotAllowedRequests := allowedRequests.Load(); gotAllowedRequests != wantAllowedRequests {
			t.Errorf("Allowed requests: got: %d want: %d", gotAllowedRequests, wantAllowedRequests)
		}

		limiter.CleanBucketLogin(login)

	}
}
