package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/mrvin/anti-bruteforce/internal/grpcserver"
	"github.com/mrvin/anti-bruteforce/internal/logger"
	"github.com/mrvin/anti-bruteforce/internal/ratelimiting/fixedwindow"
	"github.com/mrvin/anti-bruteforce/internal/storage/sqlite"
)

type Config struct {
	Buckets fixedwindow.Conf
	DB      sqlite.Conf
	GRPC    grpcserver.Conf
	Logger  logger.Conf
}

// LoadFromEnv will load configuration solely from the environment.
//
//nolint:gocognit
func (c *Config) LoadFromEnv() {
	if strLimitLogin := os.Getenv("REQ_PER_INTERVAL_LOGIN"); strLimitLogin != "" {
		if limitLogin, err := strconv.ParseUint(strLimitLogin, 10, 64); err != nil {
			slog.Warn("invalid limit login: " + strLimitLogin)
		} else {
			c.Buckets.LimitLogin = limitLogin
		}
	} else {
		slog.Warn("Empty limit login")
	}
	if strLimitPassword := os.Getenv("REQ_PER_INTERVAL_PASSWORD"); strLimitPassword != "" {
		if limitPassword, err := strconv.ParseUint(strLimitPassword, 10, 64); err != nil {
			slog.Warn("invalid limit password: " + strLimitPassword)
		} else {
			c.Buckets.LimitPassword = limitPassword
		}
	} else {
		slog.Warn("Empty limit password")
	}
	if strLimitIP := os.Getenv("REQ_PER_INTERVAL_IP"); strLimitIP != "" {
		if limitIP, err := strconv.ParseUint(strLimitIP, 10, 64); err != nil {
			slog.Warn("invalid limit ip: " + strLimitIP)
		} else {
			c.Buckets.LimitIP = limitIP
		}
	} else {
		slog.Warn("Empty limit ip")
	}
	if strInterval := os.Getenv("INTERVAL"); strInterval != "" {
		if interval, err := strconv.ParseInt(strInterval, 10, 64); err != nil {
			slog.Warn("invalid interval: " + strInterval)
		} else {
			c.Buckets.Interval = time.Duration(interval) * time.Millisecond
		}
	} else {
		slog.Warn("Empty interval")
	}
	if strTTLBucket := os.Getenv("TTL_BUCKET"); strTTLBucket != "" {
		if ttlBucket, err := strconv.ParseInt(strTTLBucket, 10, 64); err != nil {
			slog.Warn("invalid ttl bucket: " + strTTLBucket)
		} else {
			c.Buckets.TTLBucket = time.Duration(ttlBucket) * time.Minute
		}
	} else {
		slog.Warn("Empty ttl bucket")
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
