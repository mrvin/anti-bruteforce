package fixedwindow

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const numGoroutine = 20

var confBucketsTest = Conf{
	LimitLogin:      10,
	LimitPassword:   100,
	LimitIP:         1000,
	MaxLifetimeIdle: 2,
}

func TestAllow(t *testing.T) {
	const numRepetition = 2

	timeInterval = 100 * time.Millisecond // 0,1 секунды
	buckets := New(&confBucketsTest)
	defer buckets.Stop()

	ip := "127.0.0.1"
	password := "qwerty"
	login := "Bob"
	for i := range numRepetition {
		timeStart := time.Now()
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

		timeTest := time.Since(timeStart)
		delta := 5 * time.Millisecond
		if i != numRepetition-1 {
			time.Sleep(timeInterval - timeTest + delta)
		}
	}
}

func TestCleanBucket(t *testing.T) {
	const numRepetition = 5

	buckets := New(&confBucketsTest)
	defer buckets.Stop()

	ip := "127.0.0.1"
	password := "qwerty"
	login := "Bob"
	for _ = range numRepetition {
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

		buckets.CleanBucketLogin(login)

	}
}
