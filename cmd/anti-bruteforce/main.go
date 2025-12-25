//go:generate protoc -I=api/ --go_out=pkg/api --go-grpc_out=require_unimplemented_servers=false:pkg/api api/anti_bruteforce_service.proto
package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/mrvin/anti-bruteforce/internal/config"
	"github.com/mrvin/anti-bruteforce/internal/logger"
	"github.com/mrvin/anti-bruteforce/internal/ratelimiting/leakybucket"
	grpcserver "github.com/mrvin/anti-bruteforce/internal/server/grpc"
	sqlstorage "github.com/mrvin/anti-bruteforce/internal/storage/sql"
)

func main() {
	// init config
	var conf config.Config
	conf.LoadFromEnv()

	// init logger
	logFile, err := logger.Init(&conf.Logger)
	if err != nil {
		log.Printf("Init logger: %v", err)
		return
	}
	slog.Info("Init logger", slog.String("level", conf.Logger.Level))
	defer func() {
		if err := logFile.Close(); err != nil {
			slog.Error("Close log file: " + err.Error())
		}
	}()

	// init storage
	ctx := context.Background()
	storage, err := sqlstorage.New(ctx, &conf.DB)
	if err != nil {
		slog.Error("Failed to init storage: " + err.Error())
		return
	}
	slog.Info("Connected to database")
	defer func() {
		if err := storage.Close(); err != nil {
			slog.Error("Failed to close storage: " + err.Error())
		} else {
			slog.Info("Closing the database connection")
		}
	}()

	buckets := leakybucket.New(&conf.Buckets)

	server, err := grpcserver.New(&conf.GRPC, buckets, storage)
	if err != nil {
		slog.Error("New gRPC server: " + err.Error())
		return
	}

	// Start server
	server.Run(ctx)
}
