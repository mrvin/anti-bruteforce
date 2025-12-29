package grpcserver

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"time"

	"github.com/mrvin/anti-bruteforce/pkg/api"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) AllowAuthorization(_ context.Context, req *api.ReqAllowAuthorization) (*api.ResAllowAuthorization, error) {
	var res api.ResAllowAuthorization
	defer trace(req.GetIp(), &res)()

	ip := net.ParseIP(req.GetIp())
	if ip == nil {
		return &res, errors.New("parse ip")
	}

	if s.storage.Whitelist.Contains(ip) {
		res.Allow = true
		return &res, nil
	}
	if s.storage.Blacklist.Contains(ip) {
		res.Allow = false
		return &res, nil
	}

	if !s.buckets.СheckIP(req.GetIp()) {
		return &res, nil
	}
	if !s.buckets.СheckLogin(req.GetLogin()) {
		return &res, nil
	}
	if !s.buckets.СheckPassword(req.GetPassword()) {
		return &res, nil
	}

	res.Allow = true
	return &res, nil
}

func (s *Server) CleanBucket(_ context.Context, req *api.ReqCleanBucket) (*emptypb.Empty, error) {
	if err := s.buckets.CleanBucketLogin(req.GetLogin()); err != nil {
		return &emptypb.Empty{}, err
	}
	if err := s.buckets.CleanBucketIP(req.GetIp()); err != nil {
		return &emptypb.Empty{}, err
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
