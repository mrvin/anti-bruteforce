//go:generate protoc -I=../../api/ --go_out=../../pkg/api --go-grpc_out=require_unimplemented_servers=false:../../pkg/api ../../api/anti_bruteforce_service.proto
package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/mrvin/anti-bruteforce/internal/config"
	"github.com/mrvin/anti-bruteforce/internal/grpcserver"
	"github.com/mrvin/anti-bruteforce/internal/logger"
	"github.com/mrvin/anti-bruteforce/internal/ratelimiting"
	"github.com/mrvin/anti-bruteforce/internal/ratelimiting/leakybucket"
	"github.com/mrvin/anti-bruteforce/internal/storage"
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
	db, err := sqlstorage.New(ctx, &conf.DB)
	if err != nil {
		slog.Error("Failed to init storage: " + err.Error())
		return
	}
	slog.Info("Connected to database")
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("Failed to close storage: " + err.Error())
		} else {
			slog.Info("Closing the database connection")
		}
	}()
	whitelist, err := sqlstorage.NewList(ctx, db, "Whitelist")
	if err != nil {
		slog.Error("Failed to init whitelist: " + err.Error())
		return
	}
	defer whitelist.Close()
	blacklist, err := sqlstorage.NewList(ctx, db, "Blacklist")
	if err != nil {
		slog.Error("Failed to init blacklist: " + err.Error())
		return
	}
	defer blacklist.Close()

	// init rate limiting
	var ratelimit ratelimiting.Ratelimiter = leakybucket.New(&conf.Buckets)
	defer ratelimit.Stop()

	server, err := grpcserver.New(ctx, &conf.GRPC, ratelimit, storage.Storage{Whitelist: whitelist, Blacklist: blacklist})
	if err != nil {
		slog.Error("New gRPC server: " + err.Error())
		return
	}

	// Start server
	server.Run(ctx)
}
