package governance

import (
	"context"
	"time"

	"github.com/barnbridge/smartbackend/config"
	"github.com/barnbridge/smartbackend/ethtypes"
	"github.com/barnbridge/smartbackend/types"
	"github.com/barnbridge/smartbackend/utils"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type GovStorable struct {
	block  *types.Block
	logger *logrus.Entry

	Processed struct {
		proposals                      []Proposal
		proposalsActions               []ProposalActions
		abrogationProposals            []ethtypes.GovernanceAbrogationProposalStartedEvent
		abrogationProposalsDescription map[string]string
		proposalEvents                 []ProposalEvent
		votes                          []ethtypes.GovernanceVoteEvent
		canceledVotes                  []ethtypes.GovernanceVoteCanceledEvent
		abrogationVotes                []ethtypes.GovernanceAbrogationProposalVoteEvent
		abrogationCanceledVotes        []ethtypes.GovernanceAbrogationProposalVoteCancelledEvent
	}
}

func New(block *types.Block) *GovStorable {
	return &GovStorable{
		block:  block,
		logger: logrus.WithField("module", "storable(governance)"),
	}
}

func (g *GovStorable) Execute(ctx context.Context) error {
	g.logger.Trace("executing")
	start := time.Now()
	defer func() {
		g.logger.WithField("duration", time.Since(start)).
			Trace("done")
	}()

	var govLogs []gethtypes.Log
	for _, data := range g.block.Txs {
		for _, log := range data.LogEntries {
			if utils.NormalizeAddress(log.Address.String()) == utils.NormalizeAddress(config.Store.Storable.Governance.Address) {
				govLogs = append(govLogs, log)
			}
		}
	}

	if len(govLogs) == 0 {
		log.Debug("no events found")
		return nil
	}

	err := g.handleProposals(ctx, govLogs)
	if err != nil {
		return err
	}

	err = g.handleAbrogationProposal(ctx, govLogs)
	if err != nil {
		return err
	}

	err = g.handleEvents(govLogs)
	if err != nil {
		return err
	}

	err = g.handleVotes(govLogs)
	if err != nil {
		return err
	}

	err = g.handleAbrogationProposalVotes(govLogs)
	if err != nil {
		return err
	}

	return nil
}

func (g *GovStorable) Rollback(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `delete from governance.proposals where included_in_block = $1`, g.block.Number)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `delete from governance.abrogation_proposals where included_in_block = $1`, g.block.Number)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `delete from governance.proposal_events where included_in_block = $1`, g.block.Number)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `delete from governance.votes where included_in_block = $1`, g.block.Number)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `delete from governance.votes_canceled where included_in_block = $1`, g.block.Number)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `delete from governance.abrogation_votes where included_in_block = $1`, g.block.Number)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `delete from governance.abrogation_votes_canceled where included_in_block = $1`, g.block.Number)
	if err != nil {
		return err
	}
	return err
}

func (g *GovStorable) SaveToDatabase(ctx context.Context, tx pgx.Tx) error {
	err := g.storeProposals(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "could not store proposals")
	}

	err = g.storeAbrogrationProposals(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "could not store abrogration proposals")
	}

	err = g.storeEvents(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "could not store proposals events")
	}

	err = g.storeProposalVotes(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "could not store proposal's votes")
	}

	err = g.storeProposalCanceledVotes(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "could not store proposal's  canceled votes")
	}

	err = g.storeProposalAbrogationVotes(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "could not store abrogation proposal's votes")
	}

	err = g.storeAbrogationProposalCanceledVotes(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "could not store abrogation proposal's  canceled votes")
	}

	return nil
}

func (g *GovStorable) Result() interface{} {
	return g.Processed
}
