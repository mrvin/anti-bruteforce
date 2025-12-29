package grpcserver

import (
	"context"
	"net"

	"github.com/mrvin/anti-bruteforce/pkg/api"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) AddNetworkToWhitelist(ctx context.Context, req *api.ReqNetwork) (*emptypb.Empty, error) {
	_, network, err := net.ParseCIDR(req.GetNetwork())
	if err != nil {
		return &emptypb.Empty{}, err
	}
	if err := s.storage.Whitelist.AddNetwork(ctx, network); err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) DeleteNetworkFromWhitelist(ctx context.Context, req *api.ReqNetwork) (*emptypb.Empty, error) {
	_, network, err := net.ParseCIDR(req.GetNetwork())
	if err != nil {
		return &emptypb.Empty{}, err
	}

	if err := s.storage.Whitelist.DeleteNetwork(ctx, network); err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) Whitelist(ctx context.Context, _ *emptypb.Empty) (*api.ResListNetworks, error) {
	strNetworks, err := s.storage.Whitelist.Items(ctx)
	if err != nil {
		return nil, err
	}

	return &api.ResListNetworks{Networks: strNetworks}, nil
}
