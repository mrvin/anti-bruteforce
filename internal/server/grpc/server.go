package grpcserver

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/mrvin/anti-bruteforce/internal/ratelimiting/leakybucket"
	sqlstorage "github.com/mrvin/anti-bruteforce/internal/storage/sql"
	"github.com/mrvin/anti-bruteforce/pkg/api"
	"google.golang.org/grpc"
)

type Conf struct {
	Host string
	Port string
}

type Server struct {
	serv    *grpc.Server
	conn    net.Listener
	buckets *leakybucket.Buckets
	addr    string
	storage *sqlstorage.Storage
}

func New(conf *Conf, buckets *leakybucket.Buckets, storage *sqlstorage.Storage) (*Server, error) {
	var server Server

	server.buckets = buckets
	server.storage = storage

	var err error
	server.addr = net.JoinHostPort(conf.Host, conf.Port)
	server.conn, err = net.Listen("tcp", server.addr)
	if err != nil {
		return nil, fmt.Errorf("establish tcp connection: %w", err)
	}

	server.serv = grpc.NewServer(
		grpc.ChainUnaryInterceptor(),
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
