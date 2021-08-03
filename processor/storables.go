package processor

import (
	"github.com/sirupsen/logrus"

	"github.com/barnbridge/meminero/config"
	"github.com/barnbridge/meminero/processor/storables/accounterc20transfers"
	"github.com/barnbridge/meminero/processor/storables/dao/barn"
	"github.com/barnbridge/meminero/processor/storables/dao/governance"
	"github.com/barnbridge/meminero/processor/storables/erc20transfers"
	saEvents "github.com/barnbridge/meminero/processor/storables/smartalpha/events"
	saState "github.com/barnbridge/meminero/processor/storables/smartalpha/state"
	sePools "github.com/barnbridge/meminero/processor/storables/smartexposure/pool_state"
	seScrape "github.com/barnbridge/meminero/processor/storables/smartexposure/scrape"
	seTranches "github.com/barnbridge/meminero/processor/storables/smartexposure/tranche_state"
	syERC721 "github.com/barnbridge/meminero/processor/storables/smartyield/erc721"
	syEvents "github.com/barnbridge/meminero/processor/storables/smartyield/events"
	syRewards "github.com/barnbridge/meminero/processor/storables/smartyield/rewards"
	syState "github.com/barnbridge/meminero/processor/storables/smartyield/state"
	"github.com/barnbridge/meminero/processor/storables/tokenprices"
	"github.com/barnbridge/meminero/processor/storables/yieldfarming"
)

// registerStorables instantiates all the storables defined via code with the requested raw data
// Only the storables that are registered will be executed when the Store function is called
func (p *Processor) registerStorables() {
	if config.Store.Storable.AccountERC20Transfers.Enabled {
		p.storables = append(p.storables, accounterc20transfers.New(p.Block, p.state))
	}

	if config.Store.Storable.Erc20Transfers.Enabled {
		p.storables = append(p.storables, erc20transfers.New(p.Block, p.state))
	}

	if config.Store.Storable.TokenPrices.Enabled {
		p.storables = append(p.storables, tokenprices.New(p.Block, p.state))
	}

	if config.Store.Storable.YieldFarming.Enabled {
		p.storables = append(p.storables, yieldfarming.New(p.Block))
	}

	p.registerDAO()
	p.registerSmartYield()
	p.registerSmartExposure()
	p.registerSmartAlpha()
}

func (p *Processor) registerDAO() {
	if config.Store.Storable.Governance.Enabled {
		p.storables = append(p.storables, governance.New(p.Block))
	} else if config.Store.Storable.Barn.Enabled {
		logrus.Fatal("governance is disabled but other storables depend on it")
	}

	if config.Store.Storable.Barn.Enabled {
		p.storables = append(p.storables, barn.New(p.Block))
	} else if config.Store.Storable.Governance.Enabled {
		logrus.Fatal("barn is disabled but other storables depend on it")
	}
}

func (p *Processor) registerSmartYield() {
	if config.Store.Storable.SmartYield.Enabled {
		if !config.Store.Storable.Erc20Transfers.Enabled || !config.Store.Storable.TokenPrices.Enabled {
			logrus.Fatal("could not register smartYield storables because incomplete dependencies")
		}

		for _, pool := range p.state.SmartYield.Pools {
			if !p.state.CheckTokenExists(pool.UnderlyingAddress) {
				logrus.Fatalf("smart yield underlying token missing from tokens list: %s (%s)", pool.UnderlyingSymbol, pool.UnderlyingAddress)
			}
		}

		p.storables = append(p.storables, syEvents.New(p.Block, p.state))
		p.storables = append(p.storables, syERC721.New(p.Block, p.state))
		p.storables = append(p.storables, syRewards.New(p.Block, p.state))
		p.storables = append(p.storables, syState.New(p.Block, p.state))
	}
}

func (p *Processor) registerSmartExposure() {
	if config.Store.Storable.SmartExposure.Enabled {
		if !config.Store.Storable.Erc20Transfers.Enabled || !config.Store.Storable.TokenPrices.Enabled {
			logrus.Fatal("could not register smartExposure storables because incomplete dependencies")
		}

		for _, pool := range p.state.SmartExposure.Pools {
			if !p.state.CheckTokenExists(pool.TokenA.Address) {
				logrus.Fatalf("smart exposure underlying token missing from tokens list: %s (%s)", pool.TokenA.Symbol, pool.TokenA.Address)
			}

			if !p.state.CheckTokenExists(pool.TokenB.Address) {
				logrus.Fatalf("smart exposure underlying token missing from tokens list: %s (%s)", pool.TokenB.Symbol, pool.TokenB.Address)
			}
		}

		p.storables = append(p.storables, seScrape.New(p.Block, p.state))
		p.storables = append(p.storables, seTranches.New(p.Block, p.state))
		p.storables = append(p.storables, sePools.New(p.Block, p.state))
	}
}

func (p *Processor) registerSmartAlpha() {
	if config.Store.Storable.SmartAlpha.Enabled {
		if !config.Store.Storable.Erc20Transfers.Enabled || !config.Store.Storable.TokenPrices.Enabled {
			logrus.Fatal("could not register smartAlpha storables because incomplete dependencies")
		}

		for _, pool := range p.state.SmartAlpha.Pools {
			if !p.state.CheckTokenExists(pool.PoolToken.Address) {
				logrus.Fatalf("smart alpha underlying token missing from tokens list: %s (%s)", pool.PoolToken.Symbol, pool.PoolToken.Symbol)
			}
		}

		p.storables = append(p.storables, saEvents.New(p.Block, p.state))
		p.storables = append(p.storables, saState.New(p.Block, p.state))
	}
}
