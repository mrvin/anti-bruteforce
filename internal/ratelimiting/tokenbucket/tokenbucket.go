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
	mBuckets sync.Map // map[string]*Bucket

	limits []uint64

	ttlBucket time.Duration
	interval  time.Duration

	refillPeriods []time.Duration

	done       chan struct{}
	doneOnce   sync.Once
	wgDeleting sync.WaitGroup
}

func New(conf *ratelimiting.Conf) *Limiter {
	limiter := &Limiter{
		mBuckets: sync.Map{},

		limits: []uint64{conf.LimitIP, conf.LimitPassword, conf.LimitLogin},

		ttlBucket: conf.TTLBucket,
		interval:  conf.Interval,

		// Необходимо подбирать подходящие значения для периода пополнения и размера пополнения корзины.
		refillPeriods: []time.Duration{
			time.Duration(uint64(conf.Interval.Nanoseconds()) / conf.LimitIP),       //nolint:gosec
			time.Duration(uint64(conf.Interval.Nanoseconds()) / conf.LimitPassword), //nolint:gosec
			time.Duration(uint64(conf.Interval.Nanoseconds()) / conf.LimitLogin),    //nolint:gosec
		},

		done:       make(chan struct{}),
		doneOnce:   sync.Once{},
		wgDeleting: sync.WaitGroup{},
	}

	limiter.startDeleting()

	return limiter
}

func (l *Limiter) Allow(ip, password, login string) bool {
	if !l.allow(ratelimiting.TypeIP, ip) {
		return false
	}
	if !l.allow(ratelimiting.TypePassword, password) {
		return false
	}
	if !l.allow(ratelimiting.TypeLogin, login) {
		return false
	}

	return true
}

func (l *Limiter) CleanBucketIP(ip string) error {
	return l.cleanBucket(ratelimiting.TypeIP, ip)
}

func (l *Limiter) CleanBucketPassword(password string) error {
	return l.cleanBucket(ratelimiting.TypePassword, password)
}

func (l *Limiter) CleanBucketLogin(login string) error {
	return l.cleanBucket(ratelimiting.TypeLogin, login)
}

func (l *Limiter) Stop() {
	l.doneOnce.Do(func() {
		close(l.done)
		l.wgDeleting.Wait()
	})
}

func (l *Limiter) allow(bType ratelimiting.BucketType, bucketKey string) bool {
	now := time.Now().UnixNano()
	limit := l.limits[bType]
	refillPeriod := l.refillPeriods[bType]

	key := ratelimiting.BucketKey{BType: bType, Key: bucketKey}
	val, _ := l.mBuckets.LoadOrStore(key, &Bucket{}) //nolint:exhaustruct
	bucket := val.(*Bucket)                          //nolint:forcetypeassert

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

func (l *Limiter) cleanBucket(bType ratelimiting.BucketType, bucketKey string) error {
	now := time.Now().UnixNano()

	key := ratelimiting.BucketKey{BType: bType, Key: bucketKey}
	val, ok := l.mBuckets.Load(key)
	if !ok {
		return fmt.Errorf("%w: %s", ratelimiting.ErrBucketNotFound, bucketKey)
	}
	bucket := val.(*Bucket) //nolint:forcetypeassert

	limit := l.limits[bType]

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	bucket.lastRefill = now
	bucket.tokens = limit

	return nil
}

func (l *Limiter) startDeleting() {
	ticker := time.NewTicker(l.interval)
	l.wgDeleting.Go(func() {
		for {
			select {
			case <-ticker.C:
				slog.Debug("Start delete old buckets")
				l.deleteOldBuckets()
			case <-l.done:
				ticker.Stop()
				return
			}
		}
	})
}

func (l *Limiter) deleteOldBuckets() {
	toDelete := make([]string, 0)
	now := time.Now().UnixNano()

	l.mBuckets.Range(func(key, value any) bool {
		bucketKey := key.(ratelimiting.BucketKey) //nolint:forcetypeassert
		refillPeriod := l.refillPeriods[bucketKey.BType]

		bucket := value.(*Bucket) //nolint:forcetypeassert

		bucket.mu.Lock()
		lastAccess := bucket.lastRefill + refillPeriod.Nanoseconds()
		bucket.mu.Unlock()

		if now-lastAccess > l.ttlBucket.Nanoseconds() {
			toDelete = append(toDelete, key.(string)) //nolint:forcetypeassert
		}

		return true
	})

	for _, key := range toDelete {
		l.mBuckets.Delete(key)
	}
}
