package grpcserver

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/mrvin/hw-otus-go/anti-bruteforce/internal/ratelimiting/leakybucket"
	sqlstorage "github.com/mrvin/hw-otus-go/anti-bruteforce/internal/storage/sql"
	"github.com/mrvin/hw-otus-go/anti-bruteforce/pkg/api"
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

func (s *Server) Start() error {
	slog.Info("Start gRPC server: " + s.addr)
	if err := s.serv.Serve(s.conn); err != nil {
		return fmt.Errorf("start grpc server: %w", err)
	}

	return nil
}

func (s *Server) Stop() {
	slog.Info("Stop gRPC server")
	s.serv.GracefulStop()
	s.conn.Close()
}
