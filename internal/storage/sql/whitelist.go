package sqlstorage //nolint:dupl

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
)

func (s *Storage) AddNetworkToWhitelist(ctx context.Context, network *net.IPNet) error {
	if _, err := s.insertNetworkToWhitelist.ExecContext(ctx, network.String()); err != nil {
		return fmt.Errorf("add network to whitelist: %w", err)
	}

	return nil
}

func (s *Storage) DeleteNetworkFromWhitelist(ctx context.Context, network *net.IPNet) error {
	if _, err := s.deleteNetworkFromWhitelist.ExecContext(ctx, network.String()); err != nil {
		return fmt.Errorf("delete network from whitelist: %w", err)
	}

	return nil
}

func (s *Storage) Whitelist(ctx context.Context) ([]*net.IPNet, error) {
	networks := make([]*net.IPNet, 0)

	rows, err := s.whitelist.QueryContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return networks, nil
		}
		return nil, fmt.Errorf("can't get all networks: %w", err)
	}
	defer rows.Close()
	var strNetwork string
	for rows.Next() {
		err = rows.Scan(&strNetwork)
		if err != nil {
			return nil, fmt.Errorf("can't scan next row: %w", err)
		}
		_, network, err := net.ParseCIDR(strNetwork)
		if err != nil {
			return nil, err
		}
		networks = append(networks, network)
	}
	if err := rows.Err(); err != nil {
		return networks, fmt.Errorf("rows error: %w", err)
	}

	return networks, nil
}
