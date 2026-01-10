package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"slices"
	"sync"
)

type Networks struct {
	sync.RWMutex

	Items []*net.IPNet
}

type List struct {
	insertNetwork *sql.Stmt
	deleteNetwork *sql.Stmt
	list          *sql.Stmt

	Cache Networks
}

func NewList(ctx context.Context, db *sql.DB, listType string) (*List, error) {
	var l List
	var err error
	fmtStrErr := `prepare "%s" query: %w`

	switch listType {
	case "Whitelist":
		// Whitelist query prepare
		sqlInsertNetworkToWhitelist := `
		INSERT INTO whitelist (
			ip_range
		)
		VALUES ($1)`
		l.insertNetwork, err = db.PrepareContext(ctx, sqlInsertNetworkToWhitelist)
		if err != nil {
			return nil, fmt.Errorf(fmtStrErr, "insertNetworkToWhitelist", err)
		}
		sqlDeleteNetworkFromWhitelist := `DELETE FROM whitelist WHERE ip_range = $1`
		l.deleteNetwork, err = db.PrepareContext(ctx, sqlDeleteNetworkFromWhitelist)
		if err != nil {
			return nil, fmt.Errorf(fmtStrErr, "deleteNetworkFromWhitelist", err)
		}
		sqlWhitelist := `SELECT ip_range FROM whitelist`
		l.list, err = db.PrepareContext(ctx, sqlWhitelist)
		if err != nil {
			return nil, fmt.Errorf(fmtStrErr, "whitelist", err)
		}
	case "Blacklist":
		// Blacklist query prepare
		sqlInsertNetworkToBlacklist := `
		INSERT INTO blacklist (
			ip_range
		)
		VALUES ($1)`
		l.insertNetwork, err = db.PrepareContext(ctx, sqlInsertNetworkToBlacklist)
		if err != nil {
			return nil, fmt.Errorf(fmtStrErr, "insertNetworkToBlacklist", err)
		}
		sqlDeleteNetworkFromBlacklist := `DELETE FROM blacklist WHERE ip_range = $1`
		l.deleteNetwork, err = db.PrepareContext(ctx, sqlDeleteNetworkFromBlacklist)
		if err != nil {
			return nil, fmt.Errorf(fmtStrErr, "deleteNetworkFromBlacklist", err)
		}
		sqlBlacklist := `SELECT ip_range FROM blacklist`
		l.list, err = db.PrepareContext(ctx, sqlBlacklist)
		if err != nil {
			return nil, fmt.Errorf(fmtStrErr, "blacklist", err)
		}
	default:
		panic("invalid list type")
	}

	strNetworks, err := l.Items(ctx)
	if err != nil {
		return nil, fmt.Errorf("get list: %w", err)
	}
	items, err := strNetToNet(strNetworks)
	if err != nil {
		return nil, fmt.Errorf("convert networks: %w", err)
	}
	l.Cache.Lock()
	l.Cache.Items = items
	l.Cache.Unlock()

	return &l, nil
}

func strNetToNet(strNetworks []string) ([]*net.IPNet, error) {
	networks := make([]*net.IPNet, 0, len(strNetworks))
	for _, strNetwork := range strNetworks {
		_, network, err := net.ParseCIDR(strNetwork)
		if err != nil {
			return nil, fmt.Errorf("invalid network: %w", err)
		}
		networks = append(networks, network)
	}
	return networks, nil
}

func (l *List) AddNetwork(ctx context.Context, network *net.IPNet) error {
	if _, err := l.insertNetwork.ExecContext(ctx, network.String()); err != nil {
		return fmt.Errorf("add network to list: %w", err)
	}

	l.Cache.Lock()
	l.Cache.Items = append(l.Cache.Items, network)
	l.Cache.Unlock()

	return nil
}

func (l *List) DeleteNetwork(ctx context.Context, network *net.IPNet) error {
	if _, err := l.deleteNetwork.ExecContext(ctx, network.String()); err != nil {
		return fmt.Errorf("delete network from list: %w", err)
	}

	l.Cache.Lock()
	for i := 0; i < len(l.Cache.Items); i++ {
		if network.String() == l.Cache.Items[i].String() {
			l.Cache.Items = slices.Delete(l.Cache.Items, i, i+1)
			break
		}
	}
	l.Cache.Unlock()

	return nil
}

func (l *List) Items(ctx context.Context) ([]string, error) {
	networks := make([]string, 0)

	rows, err := l.list.QueryContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return networks, nil
		}
		return nil, fmt.Errorf("can't get all networks: %w", err)
	}
	defer rows.Close()
	var network string
	for rows.Next() {
		err = rows.Scan(&network)
		if err != nil {
			return nil, fmt.Errorf("can't scan next row: %w", err)
		}
		networks = append(networks, network)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return networks, nil
}

func (l *List) Contains(ip net.IP) bool {
	l.Cache.RLock()
	defer l.Cache.RUnlock()
	for _, network := range l.Cache.Items {
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

func (l *List) Close() {
	l.insertNetwork.Close()
	l.deleteNetwork.Close()
	l.list.Close()
}
