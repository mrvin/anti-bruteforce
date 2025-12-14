//go:generate protoc -I=api/ --go_out=internal/api --go-grpc_out=require_unimplemented_servers=false:internal/api api/anti_bruteforce_service.proto
package main

import (
	"context"
	"flag"
	"log"
	"log/slog"

	"github.com/mrvin/hw-otus-go/anti-bruteforce/internal/config"
	"github.com/mrvin/hw-otus-go/anti-bruteforce/internal/logger"
	"github.com/mrvin/hw-otus-go/anti-bruteforce/internal/ratelimiting/leakybucket"
	grpcserver "github.com/mrvin/hw-otus-go/anti-bruteforce/internal/server/grpc"
	sqlstorage "github.com/mrvin/hw-otus-go/anti-bruteforce/internal/storage/sql"
)

type Config struct {
	Buckets leakybucket.Conf `yaml:"buckets"`
	DB      sqlstorage.Conf  `yaml:"db"`
	GRPC    grpcserver.Conf  `yaml:"grpc"`
	Logger  logger.Conf      `yaml:"logger"`
}

func main() {
	ctx := context.Background()

	configFile := flag.String("config", "/etc/anti-bruteforce/anti-bruteforce.yml", "path to configuration file")
	flag.Parse()

	var conf Config
	if err := config.Parse(*configFile, &conf); err != nil {
		log.Printf("Parse config: %v", err)
		return
	}

	logFile, err := logger.Init(&conf.Logger)
	if err != nil {
		log.Printf("Init logger: %v\n", err)
		return
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			slog.Error("Close log file: " + err.Error())
		}
	}()
	slog.Info("Init logger", slog.String("Logging level", conf.Logger.Level))

	storage, err := sqlstorage.New(ctx, &conf.DB)
	if err != nil {
		slog.Error("Failed to init storage: " + err.Error())
		return
	}
	defer func() {
		if err := storage.Close(); err != nil {
			slog.Error("Failed to close storage: " + err.Error())
		} else {
			slog.Info("Closing the database connection")
		}
	}()
	slog.Info("Connected to database")

	buckets := leakybucket.New(&conf.Buckets)

	serverGRPC, err := grpcserver.New(&conf.GRPC, buckets, storage)
	if err != nil {
		slog.Error("New gRPC server: " + err.Error())
		return
	}

	if err := serverGRPC.Start(); err != nil {
		slog.Error("Failed to start gRPC server: " + err.Error())
		return
	}
}
