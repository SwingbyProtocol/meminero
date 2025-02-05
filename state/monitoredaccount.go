package state

import (
	"context"
	"fmt"

	"github.com/barnbridge/meminero/utils"
	"github.com/pkg/errors"
)

func (m *Manager) loadAllAccounts(ctx context.Context) error {
	rows, err := m.db.Query(ctx, `select address from monitored_accounts`)
	if err != nil {
		return errors.Wrap(err, "could not query database for monitored accounts")
	}

	m.monitoredAccounts = make(map[string]bool)
	for rows.Next() {
		var a string
		err := rows.Scan(&a)
		if err != nil {
			return errors.Wrap(err, "could no scan monitored accounts from database")
		}
		a = utils.NormalizeAddress(a)
		m.monitoredAccounts[a] = true
	}
	fmt.Println((m.monitoredAccounts))

	return nil
}

func (m *Manager) IsMonitoredAccount(addr string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.monitoredAccounts[utils.NormalizeAddress(addr)] {
		return true
	}

	return false
}
