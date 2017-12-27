package stake

import (
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/rational"
)

// Tick - called at the end of every block
func Tick(ctx sdk.Context, store state.SimpleDB) (change []*abci.Validator, err error) {

	// retrieve params
	params := loadParams(store)
	gs := loadGlobalState(store)
	height := ctx.BlockHeight()

	// Process Validator Provisions
	// XXX right now just process every 5 blocks, in new SDK make hourly
	if gs.InflationLastTime+5 <= height {
		gs.InflationLastTime = height
		processProvisions(store, gs, params)
	}

	return UpdateValidatorSet(store, gs, params)
}

// XXX test processProvisions
func processProvisions(store state.SimpleDB, gs *GlobalState, params Params) {

	// get hourly, and save annual inflation
	hourly, annual := getInflation(gs, params)
	gs.Inflation = annual

	// Because the validators hold a relative bonded share (`GlobalStakeShare`), when
	// more bonded tokens are added proportionally to all validators the only term
	// which needs to be updated is the `BondedPool`. So for each previsions cycle:

	hourlyProvisions := hourly.Mul(rational.New(gs.TotalSupply)).Evaluate()
	gs.BondedPool += hourlyProvisions
	gs.TotalSupply += hourlyProvisions

	// XXX XXX XXX XXX XXX XXX XXX XXX XXX
	// XXX Mint them to the hold account
	// XXX XXX XXX XXX XXX XXX XXX XXX XXX

	// save the params
	saveGlobalState(store, gs)
}

func getInflation(gs *GlobalState, params Params) (hourly, annual rational.Rat) {

	// The target annual inflation rate is recalculated for each previsions cycle. The
	// inflation is also subject to a rate change (positive of negative) depending or
	// the distance from the desired ratio (67%). The maximum rate change possible is
	// defined to be 13% per year, however the annual inflation is capped as between
	// 7% and 20%.

	bondedRatio := rational.New(gs.BondedPool, gs.TotalSupply)
	annualInflationRateChange := rational.New(1).Sub(bondedRatio.Quo(params.GoalBonded)).Mul(params.InflationRateChange)
	annualInflation := gs.Inflation.Add(annualInflationRateChange)
	if annualInflation.GT(params.InflationMax) {
		annualInflation = params.InflationMax
	}
	if annualInflation.LT(params.InflationMin) {
		annualInflation = params.InflationMin
	}

	hoursPerYear := rational.New(876582, 100)
	hourlyInflation := annualInflation.Quo(hoursPerYear)

	return hourlyInflation, annualInflation
}
