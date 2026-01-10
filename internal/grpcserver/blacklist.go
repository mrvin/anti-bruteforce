package grpcserver

import (
	"context"

	"github.com/mrvin/anti-bruteforce/pkg/api"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) AddNetworkToBlacklist(ctx context.Context, req *api.ReqNetwork) (*emptypb.Empty, error) {
	return addNetwork(ctx, s.storage.Blacklist, req.GetNetwork())
}

func (s *Server) DeleteNetworkFromBlacklist(ctx context.Context, req *api.ReqNetwork) (*emptypb.Empty, error) {
	return deleteNetwork(ctx, s.storage.Blacklist, req.GetNetwork())
}

func (s *Server) Blacklist(ctx context.Context, _ *emptypb.Empty) (*api.ResListNetworks, error) {
	return list(ctx, s.storage.Blacklist)
}
