// Package leakybucket реализует алгоритм дырявого ведра на основе счетчика (leaking bucket) с ленивым вытиканием (при запросе).
package leakybucket

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/mrvin/anti-bruteforce/internal/ratelimiting"
)

type Bucket struct {
	count    uint64
	lastLeak int64 // time.Unix()
	mu       sync.Mutex
}

type Limiter struct {
	mBuckets sync.Map // map[string]*Bucket

	limits []uint64

	ttlBucket time.Duration
	interval  time.Duration

	leakPeriods []time.Duration

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
		leakPeriods: []time.Duration{
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
	leakPeriod := l.leakPeriods[bType]

	key := ratelimiting.BucketKey{BType: bType, Key: bucketKey}
	val, _ := l.mBuckets.LoadOrStore(key, &Bucket{lastLeak: now}) //nolint:exhaustruct
	bucket := val.(*Bucket)                                       //nolint:forcetypeassert

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	if bucket.lastLeak != now {
		elapsed := now - bucket.lastLeak
		leakReq := uint64(elapsed / leakPeriod.Nanoseconds()) //nolint:gosec
		if leakReq > 0 {
			if leakReq >= bucket.count {
				bucket.count = 0
			} else {
				bucket.count -= leakReq
			}
			bucket.lastLeak = now
		}
	}

	if bucket.count < limit {
		bucket.count++
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

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	bucket.lastLeak = now
	bucket.count = 0

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
	toDelete := make([]ratelimiting.BucketKey, 0)
	now := time.Now().UnixNano()

	l.mBuckets.Range(func(key, value any) bool {
		bucketKey := key.(ratelimiting.BucketKey) //nolint:forcetypeassert
		leakPeriod := l.leakPeriods[bucketKey.BType]

		bucket := value.(*Bucket) //nolint:forcetypeassert

		bucket.mu.Lock()
		lastAccess := bucket.lastLeak + leakPeriod.Nanoseconds()
		bucket.mu.Unlock()

		if now-lastAccess > l.ttlBucket.Nanoseconds() {
			toDelete = append(toDelete, bucketKey)
		}

		return true
	})

	for _, key := range toDelete {
		l.mBuckets.Delete(key)
	}
}
