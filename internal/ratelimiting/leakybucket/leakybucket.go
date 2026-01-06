package leakybucket

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mrvin/anti-bruteforce/internal/ratelimiting"
)

var timeInterval = time.Minute

type Conf struct {
	LimitLogin      uint64
	LimitPassword   uint64
	LimitIP         uint64
	MaxLifetimeIdle uint64
}

type Bucket struct {
	rate         atomic.Uint64
	lifetimeIdle atomic.Uint64
}

type Buckets struct {
	mBucketsLogin    sync.Map // map[string]*Bucket
	mBucketsPassword sync.Map
	mBucketsIP       sync.Map

	limitLogin    uint64
	limitPassword uint64
	limitIP       uint64
	done          chan struct{}
	doneOnce      sync.Once
}

func New(conf *Conf) *Buckets {
	buckets := &Buckets{
		limitLogin:    conf.LimitLogin,
		limitPassword: conf.LimitPassword,
		limitIP:       conf.LimitIP,
		done:          make(chan struct{}),
	}

	// Если MaxLifetimeIdle == 0 — пропускаем очистку (без удаления)
	if conf.MaxLifetimeIdle != 0 {
		buckets.startCleanup(conf.MaxLifetimeIdle)
	}

	return buckets
}

func (b *Buckets) startCleanup(maxLifetimeIdle uint64) {
	ticker := time.NewTicker(timeInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				slog.Debug("Start cleaning buckets")
				cleanAndDeleteBucket(&b.mBucketsIP, maxLifetimeIdle)
				cleanAndDeleteBucket(&b.mBucketsPassword, maxLifetimeIdle)
				cleanAndDeleteBucket(&b.mBucketsLogin, maxLifetimeIdle)
			case <-b.done:
				return
			}
		}
	}()
}

func cleanAndDeleteBucket(m *sync.Map, maxLifetimeIdle uint64) {
	toDelete := make([]string, 0)

	m.Range(func(key, value interface{}) bool {
		bucket := value.(*Bucket)

		// Если rate == 0, увеличиваем lifetimeIdle на 1
		// Иначе сбрасываем lifetimeIdle в 0 и rate в 0
		for {
			rate := bucket.rate.Load()
			if rate == 0 {
				currentLifetimeIdle := bucket.lifetimeIdle.Load()
				if currentLifetimeIdle >= maxLifetimeIdle-1 {
					toDelete = append(toDelete, key.(string))
					break
				}
				if bucket.lifetimeIdle.CompareAndSwap(currentLifetimeIdle, currentLifetimeIdle+1) {
					break
				}
				continue
			}

			bucket.rate.Store(0)
			bucket.lifetimeIdle.Store(0)
			break
		}

		return true
	})

	for _, key := range toDelete {
		m.Delete(key)
	}
}

func (b *Buckets) Allow(ctx context.Context, ip, password, login string) bool {
	if !allow(ctx, ip, &b.mBucketsIP, b.limitIP) {
		return false
	}
	if !allow(ctx, password, &b.mBucketsPassword, b.limitPassword) {
		return false
	}
	if !allow(ctx, login, &b.mBucketsLogin, b.limitLogin) {
		return false
	}

	return true
}

func allow(ctx context.Context, keyBucket string, m *sync.Map, limit uint64) bool {
	val, _ := m.LoadOrStore(keyBucket, &Bucket{})
	bucket := val.(*Bucket)

	// Попытки увеличить rate на 1, если не превышен лимит
	// Если превышен, вернуть false
	// Иначе вернуть true
	for {
		select {
		case <-ctx.Done():
			return false
		default:
		}
		currentRate := bucket.rate.Load()
		if currentRate >= limit {
			return false
		}
		if bucket.rate.CompareAndSwap(currentRate, currentRate+1) {
			return true
		}
	}
}

func cleanBucket(keyBucket string, m *sync.Map) error {
	val, ok := m.Load(keyBucket)
	if !ok {
		return fmt.Errorf("%w: %s", ratelimiting.ErrBucketNotFound, keyBucket)
	}
	bucket := val.(*Bucket)
	bucket.rate.Store(0)
	bucket.lifetimeIdle.Store(0)

	return nil
}

func (b *Buckets) CleanBucketIP(ip string) error {
	return cleanBucket(ip, &b.mBucketsIP)
}

func (b *Buckets) CleanBucketPassword(password string) error {
	return cleanBucket(password, &b.mBucketsPassword)
}

func (b *Buckets) CleanBucketLogin(login string) error {
	return cleanBucket(login, &b.mBucketsLogin)
}

func (b *Buckets) Stop() {
	b.doneOnce.Do(func() {
		close(b.done)
	})
}
