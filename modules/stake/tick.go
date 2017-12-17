package stake

import (
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"

	abci "github.com/tendermint/abci/types"
)

// Tick - called at the end of every block
func Tick(ctx sdk.Context, store state.SimpleDB) (change []*abci.Validator, err error) {

	// retrieve params
	params := loadParams(store)
	gs := loadGlobalState(store)
	height := ctx.BlockHeight()

	// Process Validator Provisions
	// TODO right now just process every 5 blocks, in new SDK make hourly
	if InflationLastHeight+5 <= height {
		params.InflationLastHeight = height
		processProvisions(store, params)
	}

	return UpdateValidatorSet(store, params)
}

// XXX test processProvisions
func processProvisions(store state.SimpleDB, params Params, gs GlobalState) {

	//The target annual inflation rate is recalculated for each previsions cycle. The
	//inflation is also subject to a rate change (positive of negative) depending or
	//the distance from the desired ratio (67%). The maximum rate change possible is
	//defined to be 13% per year, however the annual inflation is capped as between
	//7% and 20%.

	bondedRatio := NewFraction(params.BondedPool, params.TotalSupply)
	annualInflationRateChange := One.Sub(bondedRatio.Div(params.GoalBonded)).Mul(params.InflationRateChange)
	annualInflation := params.Inflation.Add(annualInflationRateChange)
	if annualInflation.Sub(params.InflationMax).Positive() {
		annualInflation = params.InflationMax
	}
	if annualInflation.Sub(params.InflationMin).Negative() {
		annualInflation = params.InflationMin
	}
	hoursPerYear := NewFraction(876582, 100)
	provisionTokensHourly := annualInflation.Div(hoursPerYear).MulInt(params.TotalSupply)

	// save the new inflation for the next tick
	gs.Inflation = annualInflation

	//Because the validators hold a relative bonded share (`GlobalStakeShare`), when
	//more bonded tokens are added proportionally to all validators the only term
	//which needs to be updated is the `BondedPool`. So for each previsions cycle:

	gs.BondedPool += provisionTokensHourly.Evaluate()

	//XXX XXX XXX XXX XXX XXX XXX XXX XXX
	//XXX Mint them to the hold account
	//XXX XXX XXX XXX XXX XXX XXX XXX XXX

	// save the params
	saveGlobalState(store, gs)
}
