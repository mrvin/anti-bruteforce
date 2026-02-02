// Package slidingwindowcounter реализует счетчик скользящих интервалов (sliding window counter) с ленивым сбросом счетчика (при запросе).
package slidingwindowcounter

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/mrvin/anti-bruteforce/internal/ratelimiting"
)

type Window struct {
	prevCount uint64
	currCount uint64
	startTime int64 // time.UnixNano()
	mu        sync.Mutex
}

type Limiter struct {
	mWindowsLogin    sync.Map // map[string]*Window
	mWindowsPassword sync.Map
	mWindowsIP       sync.Map

	limitLogin    uint64
	limitPassword uint64
	limitIP       uint64

	ttlBucket time.Duration
	interval  time.Duration

	done       chan struct{}
	doneOnce   sync.Once
	wgDeleting sync.WaitGroup
}

func New(conf *ratelimiting.Conf) *Limiter {
	limiter := &Limiter{
		mWindowsLogin:    sync.Map{},
		mWindowsPassword: sync.Map{},
		mWindowsIP:       sync.Map{},

		limitLogin:    conf.LimitLogin,
		limitPassword: conf.LimitPassword,
		limitIP:       conf.LimitIP,

		ttlBucket: conf.TTLBucket,
		interval:  conf.Interval,

		done:       make(chan struct{}),
		doneOnce:   sync.Once{},
		wgDeleting: sync.WaitGroup{},
	}

	limiter.startDeleting()

	return limiter
}

func deleteOldWindows(m *sync.Map, ttl time.Duration, interval time.Duration) {
	toDelete := make([]string, 0)
	now := time.Now().UnixNano()

	m.Range(func(key, value any) bool {
		window := value.(*Window) //nolint:forcetypeassert

		window.mu.Lock()
		lastAccess := window.startTime + interval.Nanoseconds()
		window.mu.Unlock()

		if now-lastAccess > ttl.Nanoseconds() {
			toDelete = append(toDelete, key.(string)) //nolint:forcetypeassert
		}

		return true
	})

	for _, key := range toDelete {
		m.Delete(key)
	}
}

func (l *Limiter) Allow(ip, password, login string) bool {
	if !allow(ip, &l.mWindowsIP, l.limitIP, l.interval) {
		return false
	}
	if !allow(password, &l.mWindowsPassword, l.limitPassword, l.interval) {
		return false
	}
	if !allow(login, &l.mWindowsLogin, l.limitLogin, l.interval) {
		return false
	}

	return true
}

func allow(keyBucket string, m *sync.Map, limit uint64, interval time.Duration) bool {
	now := time.Now().UnixNano()
	val, _ := m.LoadOrStore(keyBucket, &Window{}) //nolint:exhaustruct
	window := val.(*Window)                       //nolint:forcetypeassert

	window.mu.Lock()
	defer window.mu.Unlock()

	if window.startTime+interval.Nanoseconds() < now {
		window.prevCount = window.currCount
		window.currCount = 0
		window.startTime += interval.Nanoseconds()
		if window.startTime+(interval.Nanoseconds()*2) < now { //nolint:mnd
			window.prevCount = 0
			window.startTime = now
		}
	}

	fInterval := float64(interval.Nanoseconds())
	fCurrCount := float64(window.currCount)
	fPrevCount := float64(window.prevCount)
	fElapsed := float64(now - window.startTime)
	count := uint64((fPrevCount * (fInterval - fElapsed) / fInterval) + fCurrCount)

	if count >= limit {
		return false
	}
	window.currCount++

	return true
}

func cleanWindow(keyBucket string, m *sync.Map) error {
	now := time.Now().UnixNano()
	val, ok := m.Load(keyBucket)
	if !ok {
		return fmt.Errorf("%w: %s", ratelimiting.ErrBucketNotFound, keyBucket)
	}
	window := val.(*Window) //nolint:forcetypeassert

	window.mu.Lock()
	defer window.mu.Unlock()

	window.startTime = now
	window.currCount = 0
	window.prevCount = 0

	return nil
}

func (l *Limiter) CleanBucketIP(ip string) error {
	return cleanWindow(ip, &l.mWindowsIP)
}

func (l *Limiter) CleanBucketPassword(password string) error {
	return cleanWindow(password, &l.mWindowsPassword)
}

func (l *Limiter) CleanBucketLogin(login string) error {
	return cleanWindow(login, &l.mWindowsLogin)
}

func (l *Limiter) Stop() {
	l.doneOnce.Do(func() {
		close(l.done)
		l.wgDeleting.Wait()
	})
}

func (l *Limiter) startDeleting() {
	ticker := time.NewTicker(l.interval)
	l.wgDeleting.Go(func() {
		for {
			select {
			case <-ticker.C:
				slog.Debug("Start delete old windows")
				deleteOldWindows(&l.mWindowsIP, l.ttlBucket, l.interval)
				deleteOldWindows(&l.mWindowsPassword, l.ttlBucket, l.interval)
				deleteOldWindows(&l.mWindowsLogin, l.ttlBucket, l.interval)
			case <-l.done:
				ticker.Stop()
				return
			}
		}
	})
}
