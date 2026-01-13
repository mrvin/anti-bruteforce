package grpcserver

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/mrvin/anti-bruteforce/internal/logger"
	"github.com/mrvin/anti-bruteforce/internal/ratelimiting"
	"github.com/mrvin/anti-bruteforce/internal/storage"
	"github.com/mrvin/anti-bruteforce/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Conf struct {
	Host string
	Port string
}

type Server struct {
	serv      *grpc.Server
	conn      net.Listener
	addr      string
	ratelimit ratelimiting.Ratelimiter
	storage   storage.Storage
}

func New(ctx context.Context, conf *Conf, ratelimit ratelimiting.Ratelimiter, storage storage.Storage) (*Server, error) {
	var server Server

	server.ratelimit = ratelimit
	server.storage = storage

	var err error
	lc := net.ListenConfig{} //nolint:exhaustruct
	server.addr = net.JoinHostPort(conf.Host, conf.Port)
	server.conn, err = lc.Listen(ctx, "tcp", server.addr)
	if err != nil {
		return nil, fmt.Errorf("establish tcp connection: %w", err)
	}

	server.serv = grpc.NewServer(
		grpc.ChainUnaryInterceptor(loggingInterceptor),
	)
	api.RegisterAntiBruteForceServiceServer(server.serv, &server)

	return &server, nil
}

func (s *Server) Run(ctx context.Context) {
	ctx, cancel := signal.NotifyContext(
		ctx,
		os.Interrupt,    // SIGINT, (Control-C)
		syscall.SIGTERM, // systemd
		syscall.SIGQUIT,
	)

	go func() {
		defer cancel()
		if err := s.serv.Serve(s.conn); err != nil {
			slog.Error("Failed to start gRPC server: " + err.Error())
			return
		}
	}()
	slog.Info("Start gRPC server: " + s.addr)

	<-ctx.Done()

	s.serv.GracefulStop()
	s.conn.Close()

	slog.Info("Stop gRPC server")
}

func loggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (any, error) {
	timeStart := time.Now()
	requestID := uuid.New().String()

	addr := "unknown"
	if p, ok := peer.FromContext(ctx); ok && p != nil && p.Addr != nil {
		addr = p.Addr.String()
	}
	method := filepath.Base(info.FullMethod)

	logReq := slog.Default().With(
		slog.String("requestID", requestID),
	)
	switch val := req.(type) {
	case *api.ReqAllowAuthorization:
		logReq = logReq.With(
			slog.String("login", val.GetLogin()),
			slog.String("ip", val.GetIp()),
		)
	case *api.ReqNetwork:
		logReq = logReq.With(
			slog.String("network", val.GetNetwork()),
		)
	case *emptypb.Empty:
	default:
		logger.Warnf("invalid request type %T", val)
	}

	defer func() {
		slog.Info("Request grpc",
			slog.String("method", method),
			slog.String("requestID", requestID),
			slog.String("addr", addr),
			slog.String("duration", time.Since(timeStart).String()),
		)
	}()

	res, err := handler(ctx, req)
	if err != nil {
		logReq.Error(method, slog.String("error", err.Error()))
		return res, err
	}

	switch val := res.(type) {
	case *api.ResAllowAuthorization:
		logReq = logReq.With(
			slog.Bool("allow", val.GetAllow()),
		)
	case *api.ResListNetworks:
	default:
		logger.Warnf("invalid response type %T", val)
	}

	logReq.Debug(method)

	return res, err
}
