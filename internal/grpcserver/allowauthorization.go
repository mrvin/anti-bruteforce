package grpcserver

import (
	"context"
	"errors"
	"net"

	"github.com/mrvin/anti-bruteforce/internal/ratelimiting"
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

func (s *Server) CleanBucketIP(_ context.Context, req *api.ReqCleanBucket) (*emptypb.Empty, error) {
	if err := s.ratelimit.CleanBucketIP(req.GetKeyBucket()); err != nil {
		if errors.Is(err, ratelimiting.ErrBucketNotFound) {
			return nil, status.Errorf(codes.NotFound, "failed clean bucket ip: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed clean bucket ip: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) CleanBucketPassword(_ context.Context, req *api.ReqCleanBucket) (*emptypb.Empty, error) {
	if err := s.ratelimit.CleanBucketPassword(req.GetKeyBucket()); err != nil {
		if errors.Is(err, ratelimiting.ErrBucketNotFound) {
			return nil, status.Errorf(codes.NotFound, "failed clean bucket password: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed clean bucket password: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) CleanBucketLogin(_ context.Context, req *api.ReqCleanBucket) (*emptypb.Empty, error) {
	if err := s.ratelimit.CleanBucketLogin(req.GetKeyBucket()); err != nil {
		if errors.Is(err, ratelimiting.ErrBucketNotFound) {
			return nil, status.Errorf(codes.NotFound, "failed clean bucket login: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed clean bucket login: %v", err)
	}
	return &emptypb.Empty{}, nil
}
