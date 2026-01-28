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
	mWindowsLogin    sync.Map // map[string]*Window
	mWindowsPassword sync.Map
	mWindowsIP       sync.Map

	limitLogin    uint64
	limitPassword uint64
	limitIP       uint64

	TTLBucket time.Duration
	Interval  time.Duration

	done     chan struct{}
	doneOnce sync.Once
}

func New(conf *ratelimiting.Conf) *Limiter {
	limiter := &Limiter{
		mWindowsLogin:    sync.Map{},
		mWindowsPassword: sync.Map{},
		mWindowsIP:       sync.Map{},

		limitLogin:    conf.LimitLogin,
		limitPassword: conf.LimitPassword,
		limitIP:       conf.LimitIP,

		TTLBucket: conf.TTLBucket,
		Interval:  conf.Interval,

		done:     make(chan struct{}),
		doneOnce: sync.Once{},
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
		if now-(window.startTime+interval.Nanoseconds()) > ttl.Nanoseconds() {
			toDelete = append(toDelete, key.(string)) //nolint:forcetypeassert
		}
		window.mu.Unlock()

		return true
	})

	for _, key := range toDelete {
		m.Delete(key)
	}
}

func (l *Limiter) Allow(ip, password, login string) bool {
	if !allow(ip, &l.mWindowsIP, l.limitIP, l.Interval) {
		return false
	}
	if !allow(password, &l.mWindowsPassword, l.limitPassword, l.Interval) {
		return false
	}
	if !allow(login, &l.mWindowsLogin, l.limitLogin, l.Interval) {
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

	if window.startTime == 0 || now-window.startTime >= interval.Nanoseconds() {
		window.startTime = now
		window.count = 0
	}

	if window.count >= limit {
		return false
	}

	window.count++

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
	window.count = 0

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
	})
}

func (l *Limiter) startDeleting() {
	ticker := time.NewTicker(l.Interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				slog.Debug("Start delete old windows")
				deleteOldWindows(&l.mWindowsIP, l.TTLBucket, l.Interval)
				deleteOldWindows(&l.mWindowsPassword, l.TTLBucket, l.Interval)
				deleteOldWindows(&l.mWindowsLogin, l.TTLBucket, l.Interval)
			case <-l.done:
				return
			}
		}
	}()
}
