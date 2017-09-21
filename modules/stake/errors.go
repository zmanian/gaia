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
	errNotEnoughTokens     = fmt.Errorf("Insufficient bond tokens")
	errCommissionNegative  = fmt.Errorf("Commission must be positive")
	errCommissionHuge      = fmt.Errorf("Commission cannot be more than 100%")

	resBadDelegateeAddr      = abci.ErrBaseUnknownAddress.AppendLog("Delegatee does not exist for that address")
	resMissingSignature      = abci.ErrBaseInvalidSignature.AppendLog("Missing signature")
	resBondNotNominated      = abci.ErrBaseInvalidOutput.AppendLog("Cannot bond to non-nominated account")
	resNoDelegateeForAddress = abci.ErrBaseUnknownAddress.AppendLog("Delegatee does not exist for that address")
	resNoDelegatorForAddress = abci.ErrBaseInvalidInput.AppendLog("Delegator does not contain delegatee bond")
	resInsufficientFunds     = abci.ErrBaseInsufficientFunds.AppendLog("Insufficient bond tokens")

	invalidInput = abci.CodeType_BaseInvalidInput
)

func resErrLoadingDelegatees(err error) {
	return abci.ErrBaseEncodingError.AppendLog("Error loading delegatees: " + err.Error()) //should never occur
}

func resErrLoadingDelegators(err error) {
	return abci.ErrBaseEncodingError.AppendLog("Error loading delegators: " + err.Error()) //should never occur
}

func resErrLoadingDelegators(key []byte) {
	return abci.ErrBaseEncodingError.AppendLog(fmt.Sprintf("error loading delegator with key: %v", key))
}

func resErrLoadingQueue(err error) {
	return abci.ErrBaseEncodingError.AppendLog("error loading queue" + err.Error()) //should never occur
}

///////////////////////////////////////

func ErrValidatorEmpty() errors.TMError {
	return errors.WithCode(errValidatorEmpty, invalidInput)
}
func ErrBadBondingDenom() errors.TMError {
	return errors.WithCode(errBadBondingDenom, invalidInput)
}
func ErrBadBondingAmount() errors.TMError {
	return errors.WithCode(errBadBondingAmount, invalidInput)
}
func ErrBadBondingValidator() errors.TMError {
	return errors.WithCode(errBadBondingValidator, invalidInput)
}
func ErrNoBondingAcct() errors.TMError {
	return errors.WithCode(errNoBondingAcct, invalidInput)
}
func ErrNotEnoughTokens() errors.TMError {
	return errors.WithCode(errNotEnoughTokens, invalidInput)
}
