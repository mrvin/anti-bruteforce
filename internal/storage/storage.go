package storage

import (
	"context"
	"net"
)

type List interface {
	AddNetwork(ctx context.Context, network *net.IPNet) error
	DeleteNetwork(ctx context.Context, network *net.IPNet) error
	Contains(ip net.IP) bool
	Items(ctx context.Context) ([]string, error)
	Close()
}

type Storage struct {
	Whitelist List
	Blacklist List
}
