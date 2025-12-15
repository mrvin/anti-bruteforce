package grpcserver

import (
	"context"
	"net"
	"slices"

	"github.com/mrvin/hw-otus-go/anti-bruteforce/pkg/api"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) AddNetworkToBlacklist(ctx context.Context, req *api.ReqNetwork) (*emptypb.Empty, error) {
	_, network, err := net.ParseCIDR(req.GetNetwork())
	if err != nil {
		return &emptypb.Empty{}, err
	}
	if err := s.storage.AddNetworkToBlacklist(ctx, network); err != nil {
		return &emptypb.Empty{}, err
	}
	s.storage.CacheBlacklist.Lock()
	s.storage.CacheBlacklist.List = append(s.storage.CacheBlacklist.List, network)
	s.storage.CacheBlacklist.Unlock()

	return &emptypb.Empty{}, nil
}

func (s *Server) DeleteNetworkFromBlacklist(ctx context.Context, req *api.ReqNetwork) (*emptypb.Empty, error) {
	_, network, err := net.ParseCIDR(req.GetNetwork())
	if err != nil {
		return &emptypb.Empty{}, err
	}

	if err := s.storage.DeleteNetworkFromBlacklist(ctx, network); err != nil {
		return &emptypb.Empty{}, err
	}

	s.storage.CacheBlacklist.RLock()
	defer s.storage.CacheBlacklist.RUnlock()
	for i := 0; i < len(s.storage.CacheBlacklist.List); i++ {
		if network.String() == s.storage.CacheBlacklist.List[i].String() {
			s.storage.CacheBlacklist.Lock()
			s.storage.CacheBlacklist.List = slices.Delete(s.storage.CacheBlacklist.List, i, i+1)
			s.storage.CacheBlacklist.Unlock()
			break
		}
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) Blacklist(_ context.Context, _ *emptypb.Empty) (*api.ResListNetworks, error) {
	return list(&s.storage.CacheBlacklist), nil
}
