package grpcserver

import (
	"context"
	"net"

	"github.com/mrvin/anti-bruteforce/internal/storage"
	"github.com/mrvin/anti-bruteforce/pkg/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func addNetwork(ctx context.Context, list storage.List, strNetwork string) (*emptypb.Empty, error) {
	_, network, err := net.ParseCIDR(strNetwork)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid network: %v", err)
	}
	if err := list.AddNetwork(ctx, network); err != nil {
		return nil, status.Errorf(codes.Internal, "failed add network: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func deleteNetwork(ctx context.Context, list storage.List, strNetwork string) (*emptypb.Empty, error) {
	_, network, err := net.ParseCIDR(strNetwork)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid network: %v", err)
	}

	if err := list.DeleteNetwork(ctx, network); err != nil {
		return nil, status.Errorf(codes.Internal, "failed delete network: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func list(ctx context.Context, list storage.List) (*api.ResListNetworks, error) {
	strNetworks, err := list.Items(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed get list: %v", err)
	}

	return &api.ResListNetworks{Networks: strNetworks}, nil
}
