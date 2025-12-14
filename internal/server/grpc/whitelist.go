package grpcserver

import (
	"context"
	"net"
	"slices"

	"github.com/mrvin/hw-otus-go/anti-bruteforce/internal/api"
	sqlstorage "github.com/mrvin/hw-otus-go/anti-bruteforce/internal/storage/sql"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) AddNetworkToWhitelist(ctx context.Context, req *api.ReqNetwork) (*emptypb.Empty, error) {
	_, network, err := net.ParseCIDR(req.GetNetwork())
	if err != nil {
		return &emptypb.Empty{}, err
	}
	if err := s.storage.AddNetworkToWhitelist(ctx, network); err != nil {
		return &emptypb.Empty{}, err
	}
	s.storage.CacheWhitelist.Lock()
	s.storage.CacheWhitelist.List = append(s.storage.CacheWhitelist.List, network)
	s.storage.CacheWhitelist.Unlock()

	return &emptypb.Empty{}, nil
}

func (s *Server) DeleteNetworkFromWhitelist(ctx context.Context, req *api.ReqNetwork) (*emptypb.Empty, error) {
	_, network, err := net.ParseCIDR(req.GetNetwork())
	if err != nil {
		return &emptypb.Empty{}, err
	}

	if err := s.storage.DeleteNetworkFromWhitelist(ctx, network); err != nil {
		return &emptypb.Empty{}, err
	}

	s.storage.CacheWhitelist.RLock()
	defer s.storage.CacheWhitelist.RUnlock()
	for i := 0; i < len(s.storage.CacheWhitelist.List); i++ {
		if network.String() == s.storage.CacheWhitelist.List[i].String() {
			s.storage.CacheWhitelist.Lock()
			s.storage.CacheWhitelist.List = slices.Delete(s.storage.CacheWhitelist.List, i, i+1)
			s.storage.CacheWhitelist.Unlock()
			break
		}
	}

	return &emptypb.Empty{}, nil
}

func list(listNetwork *sqlstorage.ListIP) *api.ResListNetworks {
	pbNetworks := make([]string, len(listNetwork.List))
	listNetwork.RLock()
	defer listNetwork.Unlock()
	for i, network := range listNetwork.List {
		pbNetworks[i] = network.String()
	}

	return &api.ResListNetworks{Networks: pbNetworks}
}

func (s *Server) Whitelist(_ context.Context, _ *emptypb.Empty) (*api.ResListNetworks, error) {
	return list(&s.storage.CacheWhitelist), nil
}
