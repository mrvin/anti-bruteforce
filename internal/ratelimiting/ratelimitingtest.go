package ratelimiting

import (
	"errors"
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

	var wg sync.WaitGroup
	var allowedRequests atomic.Uint64
	wantAllowedRequests := min(conf.LimitIP, conf.LimitPassword, conf.LimitLogin)
	for i := range numRepetition {
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
		allowedRequests.Store(0)

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

	var wg sync.WaitGroup
	var allowedRequests atomic.Uint64
	wantAllowedRequests := min(conf.LimitIP, conf.LimitPassword, conf.LimitLogin)
	for range numRepetition {
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
		allowedRequests.Store(0)

		if err := limiter.CleanBucketLogin(login); err != nil {
			t.Errorf("Clean Bucket return error: %v", err)
		}
	}
}

// RunCleanNonexistentBucket проверяет ошибку при очистке несуществующего bucket.
//
//nolint:thelper
func RunCleanNonexistentBucket(t *testing.T, limiter Ratelimiter) {
	if err := limiter.CleanBucketLogin("nonexistent_login"); !errors.Is(err, ErrBucketNotFound) {
		t.Error("CleanBucketLogin should return error for nonexistent bucket")
	}

	if err := limiter.CleanBucketPassword("nonexistent_pass"); !errors.Is(err, ErrBucketNotFound) {
		t.Error("CleanBucketPassword should return error for nonexistent bucket")
	}

	if err := limiter.CleanBucketIP("192.168.100.100"); !errors.Is(err, ErrBucketNotFound) {
		t.Error("CleanBucketIP should return error for nonexistent bucket")
	}
}
