package config

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/mrvin/anti-bruteforce/internal/grpcserver"
	"github.com/mrvin/anti-bruteforce/internal/logger"
	"github.com/mrvin/anti-bruteforce/internal/ratelimiting/leakybucket"
	sqlstorage "github.com/mrvin/anti-bruteforce/internal/storage/sql"
)

type Config struct {
	Buckets leakybucket.Conf
	DB      sqlstorage.Conf
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
		if maxLifetimeIdle, err := strconv.ParseUint(strMaxLifetimeIdle, 10, 32); err != nil {
			slog.Warn("invalid limit ip: " + strMaxLifetimeIdle)
		} else {
			c.Buckets.MaxLifetimeIdle = uint32(maxLifetimeIdle)
		}
	} else {
		slog.Warn("Empty max lifetime idle")
	}

	if host := os.Getenv("POSTGRES_HOST"); host != "" {
		c.DB.Host = host
	} else {
		slog.Warn("Empty postgres host")
	}
	if port := os.Getenv("POSTGRES_PORT"); port != "" {
		c.DB.Port = port
	} else {
		slog.Warn("Empty postgres port")
	}
	if user := os.Getenv("POSTGRES_USER"); user != "" {
		c.DB.User = user
	} else {
		slog.Warn("Empty postgres user")
	}
	if password := os.Getenv("POSTGRES_PASSWORD"); password != "" {
		c.DB.Password = password
	} else {
		slog.Warn("Empty postgres password")
	}
	if name := os.Getenv("POSTGRES_DB"); name != "" {
		c.DB.Name = name
	} else {
		slog.Warn("Empty postgres db name")
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
