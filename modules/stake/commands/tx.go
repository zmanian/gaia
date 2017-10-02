package commands

import (
	"encoding/hex"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	cmn "github.com/tendermint/tmlibs/common"

	"github.com/cosmos/cosmos-sdk"
	sdkcmd "github.com/cosmos/cosmos-sdk/client/commands"
	txcmd "github.com/cosmos/cosmos-sdk/client/commands/txs"
	"github.com/cosmos/cosmos-sdk/modules/coin"

	"github.com/cosmos/gaia/modules/stake"
)

//nolint
const (
	FlagAmount    = "amount"
	FlagValidator = ""
)

//nolint
var (
	CmdBond = &cobra.Command{
		Use:   "bond",
		Short: "bond some coins to give voting power to a validator/validator",
		RunE:  cmdBond,
	}
	CmdUnbond = &cobra.Command{
		Use:   "unbond",
		Short: "unbond your coins from a validator/validator",
		RunE:  cmdUnbond,
	}
)

func init() {
	//Add Flags
	fsDelegation := flag.NewFlagSet("", flag.ContinueOnError)
	fsDelegation.String(FlagValidator, "", "Validator Address") //TODO make string once decimal integrated with coin
	fsDelegation.Int(FlagAmount, 0, "Amount of Atoms")          //TODO make string once decimal integrated with coin

	CmdBond.Flags().AddFlagSet(fsDelegation)
	CmdUnbond.Flags().AddFlagSet(fsDelegation)
}

func cmdBond(cmd *cobra.Command, args []string) error {
	return cmdBonding(stake.NewTxBond)
}

func cmdUnbond(cmd *cobra.Command, args []string) error {
	return cmdBonding(stake.NewTxUnbond)
}

func cmdBonding(NewTx func(validator sdk.Actor, amount coin.Coin) sdk.Tx) error {

	validator, err := getValidator()
	if err != nil {
		return err
	}
	amount, err := coin.ParseCoin(viper.GetString(FlagAmount))
	if err != nil {
		return err
	}

	tx := NewTx(validator, amount)
	return txcmd.DoTx(tx)
}

func getValidator() (validator sdk.Actor, err error) {
	var valAddr []byte
	valAddr, err = hex.DecodeString(cmn.StripHex(viper.GetString(FlagValidator)))
	if err != nil {
		err = errors.Errorf("Validator is invalid hex: %v\n", err)
		return
	}
	validator = sdk.Actor{
		ChainID: sdkcmd.GetChainID(),
		App:     stake.Name(),
		Address: valAddr,
	}
	return
}
