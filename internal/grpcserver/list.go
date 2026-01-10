package grpcserver

import (
	"context"
	"fmt"
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
		return nil, fmt.Errorf("%w: %w", status.Error(codes.InvalidArgument, "invalid network"), err)
	}
	if err := list.AddNetwork(ctx, network); err != nil {
		return nil, fmt.Errorf("%w: %w", status.Error(codes.Internal, "failed add network"), err)
	}

	return &emptypb.Empty{}, nil
}

func deleteNetwork(ctx context.Context, list storage.List, strNetwork string) (*emptypb.Empty, error) {
	_, network, err := net.ParseCIDR(strNetwork)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", status.Error(codes.InvalidArgument, "invalid network"), err)
	}

	if err := list.DeleteNetwork(ctx, network); err != nil {
		return nil, fmt.Errorf("%w: %w", status.Error(codes.Internal, "failed delete network"), err)
	}

	return &emptypb.Empty{}, nil
}

func list(ctx context.Context, list storage.List) (*api.ResListNetworks, error) {
	strNetworks, err := list.Items(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", status.Error(codes.Internal, "failed get list"), err)
	}

	return &api.ResListNetworks{Networks: strNetworks}, nil
}
