package events

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

func (s *Storable) Rollback(ctx context.Context, tx pgx.Tx) error {
	b := &pgx.Batch{}
	tables := []string{
		"pool_epoch_info",
		"pool_state",
	}
	for _, t := range tables {
		query := fmt.Sprintf(`delete from smart_alpha.%s where included_in_block = $1`, t)
		b.Queue(query, s.block.Number)
	}

	br := tx.SendBatch(ctx, b)
	_, err := br.Exec()
	if err != nil {
		return err
	}

	return br.Close()
}
