package ratelimiting

import (
	"context"
	"errors"
)

var (
	ErrBucketNotFound = errors.New("bucket not found")
)

type Ratelimiter interface {
	Allow(ctx context.Context, ip, password, login string) bool
	CleanBucketIP(ip string) error
	CleanBucketPassword(password string) error
	CleanBucketLogin(login string) error
	Stop()
}
