package fixedwindow

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mrvin/anti-bruteforce/internal/ratelimiting"
)

type Conf struct {
	LimitLogin    uint64
	LimitPassword uint64
	LimitIP       uint64
	TTLBucket     time.Duration
	Interval      time.Duration
}

type Bucket struct {
	count      atomic.Uint64
	lastAccess atomic.Int64 // time.Unix()
}

type Buckets struct {
	mBucketsLogin    sync.Map // map[string]*Bucket
	mBucketsPassword sync.Map
	mBucketsIP       sync.Map

	limitLogin    uint64
	limitPassword uint64
	limitIP       uint64

	TTLBucket time.Duration
	Interval  time.Duration

	done     chan struct{}
	doneOnce sync.Once
}

func New(conf *Conf) *Buckets {
	buckets := &Buckets{
		mBucketsLogin:    sync.Map{},
		mBucketsPassword: sync.Map{},
		mBucketsIP:       sync.Map{},

		limitLogin:    conf.LimitLogin,
		limitPassword: conf.LimitPassword,
		limitIP:       conf.LimitIP,

		TTLBucket: conf.TTLBucket,
		Interval:  conf.Interval,

		done:     make(chan struct{}),
		doneOnce: sync.Once{},
	}

	buckets.startCleanup()

	return buckets
}

func cleanAndDeleteBucket(m *sync.Map, ttl time.Duration) {
	toDelete := make([]string, 0)
	now := time.Now().Unix()
	ttlSeconds := int64(ttl.Seconds())

	m.Range(func(key, value any) bool {
		bucket := value.(*Bucket) //nolint:forcetypeassert

		if bucket.count.Load() > 0 {
			bucket.count.Store(0)
			return true
		}

		lastAccess := bucket.lastAccess.Load()
		if now-lastAccess > ttlSeconds {
			toDelete = append(toDelete, key.(string)) //nolint:forcetypeassert
		}

		return true
	})

	for _, key := range toDelete {
		m.Delete(key)
	}
}

func (b *Buckets) Allow(ip, password, login string) bool {
	if !allow(ip, &b.mBucketsIP, b.limitIP) {
		return false
	}
	if !allow(password, &b.mBucketsPassword, b.limitPassword) {
		return false
	}
	if !allow(login, &b.mBucketsLogin, b.limitLogin) {
		return false
	}

	return true
}

func allow(keyBucket string, m *sync.Map, limit uint64) bool {
	val, loaded := m.LoadOrStore(keyBucket, &Bucket{}) //nolint:exhaustruct
	bucket := val.(*Bucket)                            //nolint:forcetypeassert

	// Инициализировать lastAccess при первом создании бакета
	if !loaded {
		bucket.lastAccess.Store(time.Now().Unix())
	}

	// Попытки увеличить count на 1, если не превышен лимит
	// Если превышен, вернуть false
	// Иначе вернуть true
	for {
		currentCount := bucket.count.Load()
		if currentCount >= limit {
			return false
		}
		if bucket.count.CompareAndSwap(currentCount, currentCount+1) {
			bucket.lastAccess.Store(time.Now().Unix())
			return true
		}
	}
}

func cleanBucket(keyBucket string, m *sync.Map) error {
	val, ok := m.Load(keyBucket)
	if !ok {
		return fmt.Errorf("%w: %s", ratelimiting.ErrBucketNotFound, keyBucket)
	}
	bucket := val.(*Bucket) //nolint:forcetypeassert
	bucket.count.Store(0)
	bucket.lastAccess.Store(time.Now().Unix())

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

func (b *Buckets) startCleanup() {
	ticker := time.NewTicker(b.Interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				slog.Debug("Start cleaning buckets")
				cleanAndDeleteBucket(&b.mBucketsIP, b.TTLBucket)
				cleanAndDeleteBucket(&b.mBucketsPassword, b.TTLBucket)
				cleanAndDeleteBucket(&b.mBucketsLogin, b.TTLBucket)
			case <-b.done:
				return
			}
		}
	}()
}
