// Package tokenbucket реализует алгоритм маркерной корзины (token bucket) с ленивым пополнением (при запросе).
package tokenbucket

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/mrvin/anti-bruteforce/internal/ratelimiting"
)

type Bucket struct {
	tokens     uint64
	lastRefill int64 // time.Unix()
	mu         sync.Mutex
}

type Limiter struct {
	mBucketsLogin    sync.Map // map[string]*Bucket
	mBucketsPassword sync.Map
	mBucketsIP       sync.Map

	limitLogin    uint64
	limitPassword uint64
	limitIP       uint64

	ttlBucket time.Duration
	interval  time.Duration

	refillPeriodLogin    time.Duration
	refillPeriodPassword time.Duration
	refillPeriodIP       time.Duration

	done     chan struct{}
	doneOnce sync.Once
}

func New(conf *ratelimiting.Conf) *Limiter {
	limiter := &Limiter{
		mBucketsLogin:    sync.Map{},
		mBucketsPassword: sync.Map{},
		mBucketsIP:       sync.Map{},

		limitLogin:    conf.LimitLogin,
		limitPassword: conf.LimitPassword,
		limitIP:       conf.LimitIP,

		ttlBucket: conf.TTLBucket,
		interval:  conf.Interval,

		// Необходимо подбирать подходящие значения для периода пополнения и размера пополнения корзины.
		refillPeriodLogin:    time.Duration(uint64(conf.Interval.Nanoseconds()) / conf.LimitLogin),    //nolint:gosec
		refillPeriodPassword: time.Duration(uint64(conf.Interval.Nanoseconds()) / conf.LimitPassword), //nolint:gosec
		refillPeriodIP:       time.Duration(uint64(conf.Interval.Nanoseconds()) / conf.LimitIP),       //nolint:gosec

		done:     make(chan struct{}),
		doneOnce: sync.Once{},
	}

	limiter.startDeleting()

	return limiter
}

func deleteOldBuckets(m *sync.Map, ttl time.Duration, refillPeriod time.Duration) {
	toDelete := make([]string, 0)
	now := time.Now().UnixNano()

	m.Range(func(key, value any) bool {
		bucket := value.(*Bucket) //nolint:forcetypeassert

		bucket.mu.Lock()
		if now-(bucket.lastRefill+refillPeriod.Nanoseconds()) > ttl.Nanoseconds() {
			toDelete = append(toDelete, key.(string)) //nolint:forcetypeassert
		}
		bucket.mu.Unlock()

		return true
	})

	for _, key := range toDelete {
		m.Delete(key)
	}
}

func (l *Limiter) Allow(ip, password, login string) bool {
	if !allow(ip, &l.mBucketsIP, l.limitIP, l.refillPeriodIP) {
		return false
	}
	if !allow(password, &l.mBucketsPassword, l.limitPassword, l.refillPeriodPassword) {
		return false
	}
	if !allow(login, &l.mBucketsLogin, l.limitLogin, l.refillPeriodLogin) {
		return false
	}

	return true
}

func allow(keyBucket string, m *sync.Map, limit uint64, refillPeriod time.Duration) bool {
	now := time.Now().UnixNano()
	val, _ := m.LoadOrStore(keyBucket, &Bucket{}) //nolint:exhaustruct
	bucket := val.(*Bucket)                       //nolint:forcetypeassert

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	if bucket.lastRefill == 0 {
		bucket.tokens = limit
	} else {
		elapsed := now - bucket.lastRefill
		refillTokens := uint64(elapsed / refillPeriod.Nanoseconds()) //nolint:gosec
		bucket.tokens += refillTokens
	}
	bucket.lastRefill = now

	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

func cleanBucket(keyBucket string, m *sync.Map, limit uint64) error {
	now := time.Now().UnixNano()
	val, ok := m.Load(keyBucket)
	if !ok {
		return fmt.Errorf("%w: %s", ratelimiting.ErrBucketNotFound, keyBucket)
	}
	bucket := val.(*Bucket) //nolint:forcetypeassert

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	bucket.lastRefill = now
	bucket.tokens = limit

	return nil
}

func (l *Limiter) CleanBucketIP(ip string) error {
	return cleanBucket(ip, &l.mBucketsIP, l.limitIP)
}

func (l *Limiter) CleanBucketPassword(password string) error {
	return cleanBucket(password, &l.mBucketsPassword, l.limitPassword)
}

func (l *Limiter) CleanBucketLogin(login string) error {
	return cleanBucket(login, &l.mBucketsLogin, l.limitLogin)
}

func (l *Limiter) Stop() {
	l.doneOnce.Do(func() {
		close(l.done)
	})
}

func (l *Limiter) startDeleting() {
	ticker := time.NewTicker(l.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				slog.Debug("Start delete old buckets")
				deleteOldBuckets(&l.mBucketsIP, l.ttlBucket, l.refillPeriodIP)
				deleteOldBuckets(&l.mBucketsPassword, l.ttlBucket, l.refillPeriodPassword)
				deleteOldBuckets(&l.mBucketsLogin, l.ttlBucket, l.refillPeriodLogin)
			case <-l.done:
				return
			}
		}
	}()
}
