package stake

import (
	"fmt"

	"github.com/tendermint/basecoin/errors"
)

var (
	errInvalidCounter = fmt.Errorf("Counter Tx marked invalid")
)

// ErrInvalidCounter - custom error class
func ErrInvalidCounter() error {
	return errors.WithCode(errInvalidCounter, abci.CodeType_BaseInvalidInput)
}

// IsInvalidCounterErr - custom error class check
func IsInvalidCounterErr(err error) bool {
	return errors.IsSameError(errInvalidCounter, err)
}

// ErrDecoding - This is just a helper function to return a generic "internal error"
func ErrDecoding() error {
	return errors.ErrInternal("Error decoding state")
}
