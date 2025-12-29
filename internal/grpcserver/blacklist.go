package grpcserver

import (
	"context"
	"net"

	"github.com/mrvin/anti-bruteforce/pkg/api"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) AddNetworkToBlacklist(ctx context.Context, req *api.ReqNetwork) (*emptypb.Empty, error) {
	_, network, err := net.ParseCIDR(req.GetNetwork())
	if err != nil {
		return &emptypb.Empty{}, err
	}
	if err := s.storage.Blacklist.AddNetwork(ctx, network); err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) DeleteNetworkFromBlacklist(ctx context.Context, req *api.ReqNetwork) (*emptypb.Empty, error) {
	_, network, err := net.ParseCIDR(req.GetNetwork())
	if err != nil {
		return &emptypb.Empty{}, err
	}

	if err := s.storage.Blacklist.DeleteNetwork(ctx, network); err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) Blacklist(ctx context.Context, _ *emptypb.Empty) (*api.ResListNetworks, error) {
	strNetworks, err := s.storage.Blacklist.Items(ctx)
	if err != nil {
		return nil, err
	}

	return &api.ResListNetworks{Networks: strNetworks}, nil
}
