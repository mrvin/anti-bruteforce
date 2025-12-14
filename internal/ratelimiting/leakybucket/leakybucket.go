package leakybucket

import (
	"fmt"
	"sync"
	"time"
)

var timeInterval = time.Minute

type Conf struct {
	LimitLogin      uint64 `yaml:"reqPerMinuteForLogin"`
	LimitPassword   uint64 `yaml:"reqPerMinuteForPassword"`
	LimitIP         uint64 `yaml:"reqPerMinuteForIP"`
	MaxLifetimeIdle uint32 `yaml:"maxLifetimeIdle"`
}

type Bucket struct {
	rate         uint64
	lifetimeIdle uint32
	sync.Mutex
}

type mBuckets struct {
	buckets map[string]*Bucket
	sync.RWMutex
}

type Buckets struct {
	mBucketsLogin    mBuckets
	mBucketsPassword mBuckets
	mBucketsIP       mBuckets

	limitLogin    uint64
	limitPassword uint64
	limitIP       uint64
}

func New(conf *Conf) *Buckets {
	allBuckets := Buckets{
		mBucketsLogin:    mBuckets{make(map[string]*Bucket), sync.RWMutex{}},
		mBucketsPassword: mBuckets{make(map[string]*Bucket), sync.RWMutex{}},
		mBucketsIP:       mBuckets{make(map[string]*Bucket), sync.RWMutex{}},
		limitLogin:       conf.LimitLogin,
		limitPassword:    conf.LimitPassword,
		limitIP:          conf.LimitIP,
	}

	ticker := time.NewTicker(timeInterval)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			cleanAndDeleteBucket(&allBuckets.mBucketsLogin, conf.MaxLifetimeIdle)

			cleanAndDeleteBucket(&allBuckets.mBucketsPassword, conf.MaxLifetimeIdle)

			cleanAndDeleteBucket(&allBuckets.mBucketsIP, conf.MaxLifetimeIdle)
		}
	}()

	return &allBuckets
}

func cleanAndDeleteBucket(b *mBuckets, maxLifetimeIdle uint32) {
	b.RLock()
	for login, bucket := range b.buckets {
		bucket.Lock()
		if bucket.rate == 0 {
			bucket.lifetimeIdle++
			if bucket.lifetimeIdle >= maxLifetimeIdle {
				b.Lock()
				delete(b.buckets, login)
				b.Unlock()
			}
		} else {
			bucket.rate = 0
			bucket.lifetimeIdle = 0
		}
		bucket.Unlock()
	}
	b.RUnlock()
}

func check(keyBucket string, b *mBuckets, limit uint64) bool {
	b.Lock()
	bucket, ok := b.buckets[keyBucket]
	if !ok {
		bucket = &Bucket{}
		b.buckets[keyBucket] = bucket
	}
	b.Unlock()

	bucket.Lock()
	defer bucket.Unlock()
	if bucket.rate >= limit {
		return false
	}
	bucket.rate++

	return true
}

func (b *Buckets) СheckLogin(keyBucket string) bool {
	return check(keyBucket, &b.mBucketsLogin, b.limitLogin)
}

func (b *Buckets) СheckPassword(keyBucket string) bool {
	return check(keyBucket, &b.mBucketsPassword, b.limitPassword)
}

func (b *Buckets) СheckIP(keyBucket string) bool {
	return check(keyBucket, &b.mBucketsIP, b.limitIP)
}

func cleanBucket(keyBucket string, b *mBuckets) error {
	b.RLock()
	bucket, ok := b.buckets[keyBucket]
	b.RUnlock()
	if !ok {
		return fmt.Errorf("bucket not found: %s", keyBucket)
	}
	bucket.Lock()
	bucket.rate = 0
	bucket.Unlock()

	return nil
}

func (b *Buckets) CleanBucketLogin(keyBucket string) error {
	return cleanBucket(keyBucket, &b.mBucketsLogin)
}

func (b *Buckets) CleanBucketIP(keyBucket string) error {
	return cleanBucket(keyBucket, &b.mBucketsIP)
}
