package sqlstorage

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"sync"
	"time"

	// Add pure Go Postgres driver for the database/sql package.
	_ "github.com/lib/pq"
	"github.com/mrvin/hw-otus-go/anti-bruteforce/pkg/retry"
)

const retriesConnect = 5

const maxOpenConns = 25
const maxIdleConns = 25
const connMaxLifetime = 5 * time.Minute

type Conf struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type ListIP struct {
	List []*net.IPNet
	sync.RWMutex
}

type Storage struct {
	db *sql.DB

	// Whitelist
	insertNetworkToWhitelist   *sql.Stmt
	deleteNetworkFromWhitelist *sql.Stmt
	whitelist                  *sql.Stmt

	// Blacklist
	insertNetworkToBlacklist   *sql.Stmt
	deleteNetworkFromBlacklist *sql.Stmt
	blacklist                  *sql.Stmt

	CacheWhitelist ListIP
	CacheBlacklist ListIP

	conf *Conf
}

func New(ctx context.Context, conf *Conf) (*Storage, error) {
	var st Storage

	st.conf = conf

	if err := st.RetryConnect(ctx, retriesConnect); err != nil {
		return nil, fmt.Errorf("new database connection: %w", err)
	}

	if err := st.prepareQuery(ctx); err != nil {
		return nil, fmt.Errorf("prepare query: %w", err)
	}

	var err error
	st.CacheWhitelist.Lock()
	st.CacheWhitelist.List, err = st.Whitelist(ctx)
	st.CacheWhitelist.Unlock()
	if err != nil {
		return nil, fmt.Errorf("get whitelist: %w", err)
	}

	st.CacheBlacklist.Lock()
	st.CacheBlacklist.List, err = st.Blacklist(ctx)
	st.CacheBlacklist.Unlock()
	if err != nil {
		return nil, fmt.Errorf("get blacklist: %w", err)
	}

	return &st, nil
}

func (s *Storage) Connect(ctx context.Context) error {
	var err error
	dbConfStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		s.conf.Host, s.conf.Port, s.conf.User, s.conf.Password, s.conf.Name)
	s.db, err = sql.Open("postgres", dbConfStr)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	// Setting db connections pool.
	s.db.SetMaxOpenConns(maxOpenConns)
	s.db.SetMaxIdleConns(maxIdleConns)
	s.db.SetConnMaxLifetime(connMaxLifetime)

	return nil
}

func (s *Storage) RetryConnect(ctx context.Context, retries int) error {
	retryConnect := retry.Retry(s.Connect, retries)
	if err := retryConnect(ctx); err != nil {
		return fmt.Errorf("connection db: %w", err)
	}

	return nil
}

func (s *Storage) prepareQuery(ctx context.Context) error {
	var err error
	fmtStrErr := "prepare \"%s\" query: %w"

	// Whitelist query prepare
	sqlInsertNetworkToWhitelist := `
		INSERT INTO whitelist (
			ip_range
		)
		VALUES ($1)`
	s.insertNetworkToWhitelist, err = s.db.PrepareContext(ctx, sqlInsertNetworkToWhitelist)
	if err != nil {
		return fmt.Errorf(fmtStrErr, "insertNetworkToWhitelist", err)
	}
	sqlDeleteNetworkFromWhitelist := `DELETE FROM whitelist WHERE ip_range = $1`
	s.deleteNetworkFromWhitelist, err = s.db.PrepareContext(ctx, sqlDeleteNetworkFromWhitelist)
	if err != nil {
		return fmt.Errorf(fmtStrErr, "deleteNetworkFromWhitelist", err)
	}
	sqlWhitelist := `SELECT ip_range FROM whitelist`
	s.whitelist, err = s.db.PrepareContext(ctx, sqlWhitelist)
	if err != nil {
		return fmt.Errorf(fmtStrErr, "whitelist", err)
	}

	// Blacklist query prepare
	sqlInsertNetworkToBlacklist := `
		INSERT INTO blacklist (
			ip_range
		)
		VALUES ($1)`
	s.insertNetworkToBlacklist, err = s.db.PrepareContext(ctx, sqlInsertNetworkToBlacklist)
	if err != nil {
		return fmt.Errorf(fmtStrErr, "insertNetworkToBlacklist", err)
	}
	sqlDeleteNetworkFromBlacklist := `DELETE FROM blacklist WHERE ip_range = $1`
	s.deleteNetworkFromBlacklist, err = s.db.PrepareContext(ctx, sqlDeleteNetworkFromBlacklist)
	if err != nil {
		return fmt.Errorf(fmtStrErr, "deleteNetworkFromBlacklist", err)
	}
	sqlBlacklist := `SELECT ip_range FROM blacklist`
	s.blacklist, err = s.db.PrepareContext(ctx, sqlBlacklist)
	if err != nil {
		return fmt.Errorf(fmtStrErr, "blacklist", err)
	}

	return nil
}

func (s *Storage) Close() error {
	// Whitelist
	s.insertNetworkToWhitelist.Close()
	s.deleteNetworkFromWhitelist.Close()
	s.whitelist.Close()

	// Blacklist
	s.insertNetworkToBlacklist.Close()
	s.deleteNetworkFromBlacklist.Close()
	s.blacklist.Close()

	return s.db.Close() //nolint:wrapcheck
}
