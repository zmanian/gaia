//nolint
package stake

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/errors"
)

var (
	errValidatorEmpty      = fmt.Errorf("Cannot bond to an empty validator")
	errBadBondingDenom     = fmt.Errorf("Invalid coin denomination")
	errBadBondingAmount    = fmt.Errorf("Amount must be > 0")
	errBadBondingValidator = fmt.Errorf("Cannot bond to non-nominated account")
	errNoBondingAcct       = fmt.Errorf("No bond account for this (address, validator) pair")
	errNotEnoughTokens     = fmt.Errorf("Insufficient bond tokens")

	invalidInput = abci.CodeType_BaseInvalidInput
)

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
