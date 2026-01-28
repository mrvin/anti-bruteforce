package ratelimiting

import (
	"errors"
	"time"
)

type Conf struct {
	LimitLogin    uint64
	LimitPassword uint64
	LimitIP       uint64
	TTLBucket     time.Duration
	Interval      time.Duration
}

var (
	ErrBucketNotFound = errors.New("bucket not found")
)

type Ratelimiter interface {
	Allow(ip, password, login string) bool
	CleanBucketIP(ip string) error
	CleanBucketPassword(password string) error
	CleanBucketLogin(login string) error
	Stop()
}
