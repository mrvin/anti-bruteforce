// Package fixedwindow реализует счетчик фиксированных интервалов (fixed window counter) с ленивым сбросом счетчика (при запросе).
package fixedwindow

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/mrvin/anti-bruteforce/internal/ratelimiting"
)

type Window struct {
	count     uint64
	startTime int64 // time.UnixNano()
	mu        sync.Mutex
}

type Limiter struct {
	mWindows sync.Map // map[string]*Window

	limits []uint64

	ttlBucket time.Duration
	interval  time.Duration

	done       chan struct{}
	doneOnce   sync.Once
	wgDeleting sync.WaitGroup
}

func New(conf *ratelimiting.Conf) *Limiter {
	limiter := &Limiter{
		mWindows: sync.Map{},

		limits: []uint64{conf.LimitIP, conf.LimitPassword, conf.LimitLogin},

		ttlBucket: conf.TTLBucket,
		interval:  conf.Interval,

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
	return l.cleanWindow(ratelimiting.TypeIP, ip)
}

func (l *Limiter) CleanBucketPassword(password string) error {
	return l.cleanWindow(ratelimiting.TypePassword, password)
}

func (l *Limiter) CleanBucketLogin(login string) error {
	return l.cleanWindow(ratelimiting.TypeLogin, login)
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

	key := ratelimiting.BucketKey{BType: bType, Key: bucketKey}
	val, _ := l.mWindows.LoadOrStore(key, &Window{startTime: now}) //nolint:exhaustruct
	window := val.(*Window)                                        //nolint:forcetypeassert

	window.mu.Lock()
	defer window.mu.Unlock()

	if now-window.startTime >= l.interval.Nanoseconds() {
		window.startTime = now
		window.count = 0
	}

	if window.count >= limit {
		return false
	}

	window.count++

	return true
}

func (l *Limiter) cleanWindow(bType ratelimiting.BucketType, bucketKey string) error {
	now := time.Now().UnixNano()

	key := ratelimiting.BucketKey{BType: bType, Key: bucketKey}
	val, ok := l.mWindows.Load(key)
	if !ok {
		return fmt.Errorf("%w: %s", ratelimiting.ErrBucketNotFound, bucketKey)
	}
	window := val.(*Window) //nolint:forcetypeassert

	window.mu.Lock()
	defer window.mu.Unlock()

	window.startTime = now
	window.count = 0

	return nil
}

func (l *Limiter) startDeleting() {
	ticker := time.NewTicker(l.interval)
	l.wgDeleting.Go(func() {
		for {
			select {
			case <-ticker.C:
				slog.Debug("Start delete old windows")
				l.deleteOldWindows()
			case <-l.done:
				ticker.Stop()
				return
			}
		}
	})
}

func (l *Limiter) deleteOldWindows() {
	toDelete := make([]ratelimiting.BucketKey, 0)
	now := time.Now().UnixNano()

	l.mWindows.Range(func(key, value any) bool {
		window := value.(*Window) //nolint:forcetypeassert

		window.mu.Lock()
		lastAccess := window.startTime + l.interval.Nanoseconds()
		window.mu.Unlock()

		if now-lastAccess > l.ttlBucket.Nanoseconds() {
			toDelete = append(toDelete, key.(ratelimiting.BucketKey)) //nolint:forcetypeassert
		}

		return true
	})

	for _, key := range toDelete {
		l.mWindows.Delete(key)
	}
}
