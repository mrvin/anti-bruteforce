package ratelimiting

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const numGoroutine = 20

//nolint:thelper
func RunTestAllow(t *testing.T, limiter Ratelimiter, conf *Conf) {
	const numRepetition = 2

	ip := "127.0.0.1"
	password := "qwerty"
	login := "Bob"
	wantAllowedRequests := min(conf.LimitIP, conf.LimitPassword, conf.LimitLogin)
	for i := range numRepetition {
		var wg sync.WaitGroup
		var allowedRequests atomic.Uint64
		for range numGoroutine {
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
			time.Sleep(conf.Interval)
		}
	}
}

//nolint:thelper
func RunTestCleanBucket(t *testing.T, limiter Ratelimiter, conf *Conf) {
	const numRepetition = 5

	ip := "127.0.0.1"
	password := "qwerty"
	login := "Bob"
	wantAllowedRequests := min(conf.LimitIP, conf.LimitPassword, conf.LimitLogin)
	for range numRepetition {
		var wg sync.WaitGroup
		var allowedRequests atomic.Uint64
		for range numGoroutine {
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

		if err := limiter.CleanBucketLogin(login); err != nil {
			t.Errorf("Clean Bucket return error: %v", err)
		}
	}
}
