package governance

import (
	"context"

	"github.com/barnbridge/smartbackend/ethtypes"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

func (g *GovStorable) handleVotes(logs []gethtypes.Log) error {
	for _, log := range logs {
		if ethtypes.Governance.IsGovernanceVoteEvent(&log) {
			var vote ethtypes.GovernanceVoteEvent

			g.Processed.votes = append(g.Processed.votes, vote)
		}
		if ethtypes.Governance.IsGovernanceVoteCanceledEvent(&log){
			var vote ethtypes.GovernanceVoteCanceledEvent
			g.Processed.canceledVotes = append(g.Processed.canceledVotes, vote)
		}

	}

	return nil
}


func (g *GovStorable) storeProposalVotes(ctx context.Context,tx pgx.Tx) error {
	if len(g.Processed.votes) == 0{
		return nil
	}

	var rows [][]interface{}
	for _, v := range g.Processed.votes {
		power :=decimal.NewFromBigInt(v.Power,0)
		rows = append(rows, []interface{}{
			v.ProposalId.Int64(),
			v.User.String(),
			v.Support,
			power,
			g.block.BlockCreationTime,
			v.Raw.BlockNumber,
			v.Raw.TxHash.String(),
			v.Raw.TxIndex,
			v.Raw.Index,
		})
	}
	_, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"votes"},
		[]string{"proposal_id", "user_id","support","power","block_timestamp","included_in_block","tx_hash","tx_index","log_index"},
		pgx.CopyFromSlice(len(g.Processed.abrProposals), func(i int) ([]interface{}, error) {
			return []interface{}{rows}, nil
		}),
	)
	if err != nil {
		return errors.Wrap(err,"could not store proposal votes")
	}

	return nil
}

func (g *GovStorable) storeProposalCanceledVotes(ctx context.Context,tx pgx.Tx) error {
	if len(g.Processed.canceledVotes) == 0{
		return nil
	}

	var rows [][]interface{}
	for _, v := range g.Processed.canceledVotes {
		rows = append(rows, []interface{}{
			v.ProposalId.Int64(),
			v.User.String(),
			g.block.BlockCreationTime,
			v.Raw.BlockNumber,
			v.Raw.TxHash.String(),
			v.Raw.TxIndex,
			v.Raw.Index,
		})
	}
	_, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"votes_canceled"},
		[]string{"proposal_id", "user_id","block_timestamp","included_in_block","tx_hash","tx_index","log_index"},
		pgx.CopyFromSlice(len(g.Processed.abrProposals), func(i int) ([]interface{}, error) {
			return []interface{}{rows}, nil
		}),
	)
	if err != nil {
		return errors.Wrap(err,"could not store proposal canceled votes")
	}

	return nil
}
