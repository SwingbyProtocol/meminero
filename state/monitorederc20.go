package state

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/barnbridge/meminero/utils"
)

func (m *Manager) loadAllERC20(ctx context.Context) error {
	rows, err := m.db.Query(ctx, `
		select a.address
		from monitored_erc20 a
		union
		select pool_address
		from smart_yield.pools
		union
		select etoken_address
		from smart_exposure.tranches
		union
		select junior_token_address
		from smart_alpha.pools
		union
		select senior_token_address
		from smart_alpha.pools
	`)
	if err != nil {
		return errors.Wrap(err, "could not query database for monitored erc20")
	}

	m.monitoredERC20 = make(map[string]bool)
	for rows.Next() {
		var a string
		err := rows.Scan(&a)
		if err != nil {
			return errors.Wrap(err, "could no scan monitored erc20 from database")
		}
		a = utils.NormalizeAddress(a)
		m.monitoredERC20[a] = true
	}
	fmt.Println(m.monitoredERC20)

	return nil
}

func (m *Manager) IsMonitoredERC20(addr string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.monitoredERC20[utils.NormalizeAddress(addr)] {
		return true
	}

	return false
}
