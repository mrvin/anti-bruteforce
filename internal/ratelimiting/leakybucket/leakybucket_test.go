package leakybucket

import (
	"sync"
	"testing"
	"time"
)

const numGoroutine = 5

func TestBuckets(t *testing.T) {
	confBucketsTest := Conf{
		LimitLogin:    10,
		LimitPassword: 100,
		LimitIP:       1000,
	}
	timeInterval = 3 * time.Second
	buckets := New(&confBucketsTest)
	t.Run("Сheck login", func(t *testing.T) {
		t.Parallel()

		login := "Bob"
		wg := &sync.WaitGroup{}
		wg.Add(numGoroutine)
		for j := 0; j < numGoroutine; j++ {
			go func() {
				defer wg.Done()
				for i := uint64(0); i < confBucketsTest.LimitLogin/numGoroutine; i++ {
					if got := buckets.СheckLogin(login); !got {
						t.Errorf("СheckLogin(%q):\n\tgot: %t\n\twant: %t", login, got, true)
					}
				}
			}()
		}
		wg.Wait()

		if got := buckets.СheckLogin(login); got {
			t.Errorf("СheckLogin(%q):\n\tgot: %t\n\twant: %t", login, got, false)
		}

		time.Sleep(timeInterval + time.Millisecond*500)

		if got := buckets.СheckLogin(login); !got {
			t.Errorf("СheckLogin(%q):\n\tgot: %t\n\twant: %t", login, got, true)
		}
	})

	t.Run("Сheck password", func(t *testing.T) {
		t.Parallel()

		password := "qwerty"
		wg := &sync.WaitGroup{}
		wg.Add(numGoroutine)
		for j := 0; j < numGoroutine; j++ {
			go func() {
				defer wg.Done()
				for i := uint64(0); i < confBucketsTest.LimitPassword/numGoroutine; i++ {
					if got := buckets.СheckPassword(password); !got {
						t.Errorf("СheckPassword(%q):\n\tgot: %t\n\twant: %t", password, got, true)
					}
				}
			}()
		}
		wg.Wait()

		if got := buckets.СheckPassword(password); got {
			t.Errorf("СheckPassword(%q):\n\tgot: %t\n\twant: %t", password, got, false)
		}

		time.Sleep(timeInterval + time.Millisecond*500)

		if got := buckets.СheckPassword(password); !got {
			t.Errorf("СheckPassword(%q):\n\tgot: %t\n\twant: %t", password, got, true)
		}
	})

	t.Run("Сheck ip", func(t *testing.T) {
		t.Parallel()

		ip := "172.18.0.1"
		wg := &sync.WaitGroup{}
		wg.Add(numGoroutine)
		for j := 0; j < numGoroutine; j++ {
			go func() {
				defer wg.Done()
				for i := uint64(0); i < confBucketsTest.LimitIP/numGoroutine; i++ {
					if got := buckets.СheckIP(ip); !got {
						t.Errorf("СheckIP(%q):\n\tgot: %t\n\twant: %t", ip, got, true)
					}
				}
			}()

		}

		wg.Wait()
		if got := buckets.СheckIP(ip); got {
			t.Errorf("СheckIP(%q):\n\tgot: %t\n\twant: %t", ip, got, false)
		}

		time.Sleep(timeInterval + time.Millisecond*500)

		if got := buckets.СheckIP(ip); !got {
			t.Errorf("СheckIP(%q):\n\tgot: %t\n\twant: %t", ip, got, true)
		}
	})
}
