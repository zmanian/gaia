package stake

//import (
//"encoding/json"
//"testing"

//"github.com/stretchr/testify/assert"
//"github.com/stretchr/testify/require"

//crypto "github.com/tendermint/go-crypto"
//"github.com/tendermint/tmlibs/log"

//sdk "github.com/cosmos/cosmos-sdk"
//"github.com/cosmos/cosmos-sdk/errors"
//"github.com/cosmos/cosmos-sdk/modules/auth"
//"github.com/cosmos/cosmos-sdk/stack"
//"github.com/cosmos/cosmos-sdk/state"
//)

//// this makes sure that txs are rejected with invalid data or permissions
//func TestHandlerValidation(t *testing.T) {
//assert := assert.New(t)

//// these are all valid, except for minusCoins
//addr1 := sdk.Actor{App: "coin", Address: []byte{1, 2}}
//addr2 := sdk.Actor{App: "role", Address: []byte{7, 8}}
//someCoins := Coins{{"atom", 123}}
//doubleCoins := Coins{{"atom", 246}}
//minusCoins := Coins{{"eth", -34}}

//cases := []struct {
//valid bool
//tx    sdk.Tx
//perms []sdk.Actor
//}{
//// auth works with different apps
//{true,
//NewSendTx(
//[]TxInput{NewTxInput(addr1, someCoins)},
//[]TxOutput{NewTxOutput(addr2, someCoins)}),
//[]sdk.Actor{addr1}},
//{true,
//NewSendTx(
//[]TxInput{NewTxInput(addr2, someCoins)},
//[]TxOutput{NewTxOutput(addr1, someCoins)}),
//[]sdk.Actor{addr1, addr2}},
//// check multi-input with both sigs
//{true,
//NewSendTx(
//[]TxInput{NewTxInput(addr1, someCoins), NewTxInput(addr2, someCoins)},
//[]TxOutput{NewTxOutput(addr1, doubleCoins)}),
//[]sdk.Actor{addr1, addr2}},
//// wrong permissions fail
//{false,
//NewSendTx(
//[]TxInput{NewTxInput(addr1, someCoins)},
//[]TxOutput{NewTxOutput(addr2, someCoins)}),
//[]sdk.Actor{}},
//{false,
//NewSendTx(
//[]TxInput{NewTxInput(addr1, someCoins)},
//[]TxOutput{NewTxOutput(addr2, someCoins)}),
//[]sdk.Actor{addr2}},
//{false,
//NewSendTx(
//[]TxInput{NewTxInput(addr1, someCoins), NewTxInput(addr2, someCoins)},
//[]TxOutput{NewTxOutput(addr1, doubleCoins)}),
//[]sdk.Actor{addr1}},
//// invalid input fails
//{false,
//NewSendTx(
//[]TxInput{NewTxInput(addr1, minusCoins)},
//[]TxOutput{NewTxOutput(addr2, minusCoins)}),
//[]sdk.Actor{addr2}},
//}

//for i, tc := range cases {
//ctx := stack.MockContext("base-chain", 100).WithPermissions(tc.perms...)
//err := checkTx(ctx, tc.tx.Unwrap().(SendTx))
//if tc.valid {
//assert.Nil(err, "%d: %+v", i, err)
//} else {
//assert.NotNil(err, "%d", i)
//}
//}
//}
