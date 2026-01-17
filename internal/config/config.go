package config

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/mrvin/anti-bruteforce/internal/grpcserver"
	"github.com/mrvin/anti-bruteforce/internal/logger"
	"github.com/mrvin/anti-bruteforce/internal/ratelimiting/leakybucket"
	"github.com/mrvin/anti-bruteforce/internal/storage/sqlite"
)

type Config struct {
	Buckets leakybucket.Conf
	DB      sqlite.Conf
	GRPC    grpcserver.Conf
	Logger  logger.Conf
}

// LoadFromEnv will load configuration solely from the environment.
func (c *Config) LoadFromEnv() {
	if strLimitLogin := os.Getenv("REQ_PER_MINUTE_LOGIN"); strLimitLogin != "" {
		if limitLogin, err := strconv.ParseUint(strLimitLogin, 10, 64); err != nil {
			slog.Warn("invalid limit login: " + strLimitLogin)
		} else {
			c.Buckets.LimitLogin = limitLogin
		}
	} else {
		slog.Warn("Empty limit login")
	}
	if strLimitPassword := os.Getenv("REQ_PER_MINUTE_PASSWORD"); strLimitPassword != "" {
		if limitPassword, err := strconv.ParseUint(strLimitPassword, 10, 64); err != nil {
			slog.Warn("invalid limit password: " + strLimitPassword)
		} else {
			c.Buckets.LimitPassword = limitPassword
		}
	} else {
		slog.Warn("Empty limit password")
	}
	if strLimitIP := os.Getenv("REQ_PER_MINUTE_IP"); strLimitIP != "" {
		if limitIP, err := strconv.ParseUint(strLimitIP, 10, 64); err != nil {
			slog.Warn("invalid limit ip: " + strLimitIP)
		} else {
			c.Buckets.LimitIP = limitIP
		}
	} else {
		slog.Warn("Empty limit ip")
	}
	if strMaxLifetimeIdle := os.Getenv("MAX_LIFETIME_IDLE"); strMaxLifetimeIdle != "" {
		if maxLifetimeIdle, err := strconv.ParseUint(strMaxLifetimeIdle, 10, 64); err != nil {
			slog.Warn("invalid limit ip: " + strMaxLifetimeIdle)
		} else {
			c.Buckets.MaxLifetimeIdle = maxLifetimeIdle
		}
	} else {
		slog.Warn("Empty max lifetime idle")
	}

	if storagePath := os.Getenv("SQLITE_STORAGE_PATH"); storagePath != "" {
		c.DB.StoragePath = storagePath
	} else {
		slog.Warn("Empty sqlite storage path")
	}

	if host := os.Getenv("GRPC_HOST"); host != "" {
		c.GRPC.Host = host
	} else {
		slog.Warn("Empty server http host")
	}
	if port := os.Getenv("GRPC_PORT"); port != "" {
		c.GRPC.Port = port
	} else {
		slog.Warn("Empty server http port")
	}

	if logFilePath := os.Getenv("LOGGER_FILEPATH"); logFilePath != "" {
		c.Logger.FilePath = logFilePath
	} else {
		slog.Warn("Empty log file path")
	}
	if logLevel := os.Getenv("LOGGER_LEVEL"); logLevel != "" {
		c.Logger.Level = logLevel
	} else {
		slog.Warn("Empty log level")
	}
}
