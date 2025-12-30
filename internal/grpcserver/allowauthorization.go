package grpcserver

import (
	"context"
	"log/slog"
	"net"
	"time"

	"github.com/mrvin/anti-bruteforce/pkg/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) AllowAuthorization(_ context.Context, req *api.ReqAllowAuthorization) (*api.ResAllowAuthorization, error) {
	//defer trace(req.GetIp(), &res)()

	ip := net.ParseIP(req.GetIp())
	if ip == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid ip")
	}

	if s.storage.Whitelist.Contains(ip) {
		return &api.ResAllowAuthorization{Allow: true}, nil
	}
	if s.storage.Blacklist.Contains(ip) {
		return &api.ResAllowAuthorization{Allow: false}, nil
	}

	if !s.buckets.СheckIP(req.GetIp()) {
		return &api.ResAllowAuthorization{Allow: false}, nil
	}
	if !s.buckets.СheckLogin(req.GetLogin()) {
		return &api.ResAllowAuthorization{Allow: false}, nil
	}
	if !s.buckets.СheckPassword(req.GetPassword()) {
		return &api.ResAllowAuthorization{Allow: false}, nil
	}

	return &api.ResAllowAuthorization{Allow: true}, nil
}

func (s *Server) CleanBucket(_ context.Context, req *api.ReqCleanBucket) (*emptypb.Empty, error) {
	if err := s.buckets.CleanBucketLogin(req.GetLogin()); err != nil {
		return nil, status.Error(codes.Internal, "failed clean bucket login")
	}
	if err := s.buckets.CleanBucketIP(req.GetIp()); err != nil {
		return nil, status.Error(codes.Internal, "failed clean bucket ip")
	}

	return &emptypb.Empty{}, nil
}

func trace(ip string, res *api.ResAllowAuthorization) func() {
	start := time.Now()
	return func() {
		slog.Info("Request",
			slog.String("ip", ip),
			slog.Bool("result", res.GetAllow()),
			slog.Duration("duration", time.Since(start)),
		)
	}
}
