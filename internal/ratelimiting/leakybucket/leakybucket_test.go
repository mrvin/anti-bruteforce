package leakybucket

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const numGoroutine = 20

func TestBuckets(t *testing.T) {
	confBucketsTest := Conf{
		LimitLogin:      10,
		LimitPassword:   100,
		LimitIP:         1000,
		MaxLifetimeIdle: 2,
	}
	timeInterval = time.Second
	buckets := New(&confBucketsTest)
	t.Run("Smoke test", func(t *testing.T) {
		t.Parallel()

		ip := "127.0.0.1"
		password := "qwerty"
		login := "Bob"
		var wg sync.WaitGroup
		var allowedRequests atomic.Uint64
		wantAllowedRequests := min(confBucketsTest.LimitIP, confBucketsTest.LimitPassword, confBucketsTest.LimitLogin)
		for _ = range numGoroutine {
			wg.Go(func() {
				if got := buckets.Allow(ip, password, login); got {
					allowedRequests.Add(1)
				}
			})
		}
		wg.Wait()

		if gotAllowedRequests := allowedRequests.Load(); gotAllowedRequests != wantAllowedRequests {
			t.Errorf("Allowed requests: got: %d want: %d", gotAllowedRequests, wantAllowedRequests)
		}
	})
}
