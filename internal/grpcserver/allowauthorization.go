package grpcserver

import (
	"context"
	"fmt"
	"net"

	"github.com/mrvin/anti-bruteforce/pkg/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) AllowAuthorization(_ context.Context, req *api.ReqAllowAuthorization) (*api.ResAllowAuthorization, error) {
	ip := net.ParseIP(req.GetIp())
	if ip == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid ip") //nolint:wrapcheck
	}

	if s.storage.Whitelist.Contains(ip) {
		return &api.ResAllowAuthorization{Allow: true}, nil
	}
	if s.storage.Blacklist.Contains(ip) {
		return &api.ResAllowAuthorization{Allow: false}, nil
	}

	if !s.ratelimit.Allow(req.GetIp(), req.GetPassword(), req.GetLogin()) {
		return &api.ResAllowAuthorization{Allow: false}, nil
	}

	return &api.ResAllowAuthorization{Allow: true}, nil
}

func (s *Server) CleanBucket(_ context.Context, req *api.ReqCleanBucket) (*emptypb.Empty, error) {
	if err := s.ratelimit.CleanBucketIP(req.GetIp()); err != nil {
		return nil, fmt.Errorf("%w: %w", status.Error(codes.Internal, "failed clean bucket ip"), err)
	}
	if err := s.ratelimit.CleanBucketPassword(req.GetPassword()); err != nil {
		return nil, fmt.Errorf("%w: %w", status.Error(codes.Internal, "failed clean bucket password"), err)
	}
	if err := s.ratelimit.CleanBucketLogin(req.GetLogin()); err != nil {
		return nil, fmt.Errorf("%w: %w", status.Error(codes.Internal, "failed clean bucket login"), err)
	}

	return &emptypb.Empty{}, nil
}
