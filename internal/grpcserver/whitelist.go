package grpcserver

import (
	"context"
	"net"

	"github.com/mrvin/anti-bruteforce/pkg/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) AddNetworkToWhitelist(ctx context.Context, req *api.ReqNetwork) (*emptypb.Empty, error) {
	_, network, err := net.ParseCIDR(req.GetNetwork())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid network")
	}
	if err := s.storage.Whitelist.AddNetwork(ctx, network); err != nil {
		return nil, status.Error(codes.Internal, "failed add network")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) DeleteNetworkFromWhitelist(ctx context.Context, req *api.ReqNetwork) (*emptypb.Empty, error) {
	_, network, err := net.ParseCIDR(req.GetNetwork())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid network")
	}

	if err := s.storage.Whitelist.DeleteNetwork(ctx, network); err != nil {
		return nil, status.Error(codes.Internal, "failed delete network")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) Whitelist(ctx context.Context, _ *emptypb.Empty) (*api.ResListNetworks, error) {
	strNetworks, err := s.storage.Whitelist.Items(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed get whitelist")
	}

	return &api.ResListNetworks{Networks: strNetworks}, nil
}
