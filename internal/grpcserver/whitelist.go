package grpcserver

import (
	"context"

	"github.com/mrvin/anti-bruteforce/pkg/api"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) AddNetworkToWhitelist(ctx context.Context, req *api.ReqNetwork) (*emptypb.Empty, error) {
	return addNetwork(ctx, s.storage.Whitelist, req.GetNetwork())
}

func (s *Server) DeleteNetworkFromWhitelist(ctx context.Context, req *api.ReqNetwork) (*emptypb.Empty, error) {
	return deleteNetwork(ctx, s.storage.Whitelist, req.GetNetwork())
}

func (s *Server) Whitelist(ctx context.Context, _ *emptypb.Empty) (*api.ResListNetworks, error) {
	return list(ctx, s.storage.Whitelist)
}
