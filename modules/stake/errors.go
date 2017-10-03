//nolint
package stake

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/errors"
	abci "github.com/tendermint/abci/types"
)

var (
	errValidatorEmpty      = fmt.Errorf("Cannot bond to an empty validator")
	errBadBondingDenom     = fmt.Errorf("Invalid coin denomination")
	errBadBondingAmount    = fmt.Errorf("Amount must be > 0")
	errBadBondingValidator = fmt.Errorf("Cannot bond to non-nominated account")
	errNoBondingAcct       = fmt.Errorf("No bond account for this (address, validator) pair")
	errCommissionNegative  = fmt.Errorf("Commission must be positive")
	errCommissionHuge      = fmt.Errorf("Commission cannot be more than 100%")

	resBadValidatorAddr      = abci.ErrBaseUnknownAddress.AppendLog("Validator does not exist for that address")
	resMissingSignature      = abci.ErrBaseInvalidSignature.AppendLog("Missing signature")
	resBondNotNominated      = abci.ErrBaseInvalidOutput.AppendLog("Cannot bond to non-nominated account")
	resNoValidatorForAddress = abci.ErrBaseUnknownAddress.AppendLog("Validator does not exist for that address")
	resNoDelegatorForAddress = abci.ErrBaseInvalidInput.AppendLog("Delegator does not contain validator bond")
	resInsufficientFunds     = abci.ErrBaseInsufficientFunds.AppendLog("Insufficient bond tokens")

	invalidInput = abci.CodeType_BaseInvalidInput
)

func resErrLoadingValidators(err error) abci.Result {
	return abci.ErrBaseEncodingError.AppendLog("Error loading validators: " + err.Error()) //should never occur
}

///////////////////////////////////////

func ErrValidatorEmpty() errors.TMError {
	return errors.WithCode(errValidatorEmpty, invalidInput)
}
