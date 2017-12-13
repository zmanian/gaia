package stake

import (
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/state"

	abci "github.com/tendermint/abci/types"
	wire "github.com/tendermint/go-wire"
)

// Tick - called at the end of every block
func Tick(ctx sdk.Context, store state.SimpleDB) (change []*abci.Validator, err error) {

	// Process Validator Provisions
	processProvisions(store)

	return UpdateValidatorSet(store)
}

func processProvisions(store state.SimpleDB) {

	//The target annual inflation rate is recalculated for each previsions cycle. The
	//inflation is also subject to a rate change (positive of negative) depending or
	//the distance from the desired ratio (67%). The maximum rate change possible is
	//defined to be 13% per year, however the annual inflation is capped as between
	//7% and 20%.

	//```
	//inflationRateChange(0) = 0
	//annualInflation(0) = 0.07

	//bondedRatio = bondedTokenPool / totalTokenSupply
	//AnnualInflationRateChange = (1 - bondedRatio / 0.67) * 0.13

	//annualInflation += AnnualInflationRateChange

	//if annualInflation > 0.20 then annualInflation = 0.20
	//if annualInflation < 0.07 then annualInflation = 0.07

	//provisionTokensHourly = totalTokenSupply * annualInflation / (365.25*24)
	//```

	//Because the validators hold a relative bonded share (`GlobalStakeShare`), when
	//more bonded tokens are added proportionally to all validators the only term
	//which needs to be updated is the `BondedTokenPool`. So for each previsions cycle:

	//```
	//params.BondedTokenPool += provisionTokensHourly
}
