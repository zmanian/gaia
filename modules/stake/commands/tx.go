package commands

import (
	"encoding/hex"
	"fmt"

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
	FlagAmount     = "amount"
	FlagValidator  = "validator"
	FlagCommission = "commission"
)

//nolint
var (
	CmdBond = &cobra.Command{
		Use:   "bond",
		Short: "bond some coins to give voting power to a delegatee/validator",
		RunE:  cmdBond,
	}
	CmdUnbond = &cobra.Command{
		Use:   "unbond",
		Short: "unbond your coins from a delegatee/validator",
		RunE:  cmdUnbond,
	}
	CmdNominate = &cobra.Command{
		Use:   "nominate",
		Short: "nominate yourself to become a delegatee/validator",
		RunE:  cmdNominate,
	}
	CmdModComm = &cobra.Command{
		Use:   "modify-commission",
		Short: "modify your commission rate if you are a delegatee/validator",
		RunE:  cmdModComm,
	}
)

func init() {

	//Add Flags
	fsDelegation := flag.NewFlagSet("", flag.ContinueOnError)
	fsNominate := flag.NewFlagSet("", flag.ContinueOnError)
	fsModComm := flag.NewFlagSet("", flag.ContinueOnError)

	fsDelegation.String(FlagValidator, "", "Validator's public key")
	fsDelegation.Int(FlagAmount, 0, "Amount of Atoms") //TODO make string once decimal integrated with coin

	fsNominate.AddFlagSet(fsDelegation)
	fsNominate.String(FlagCommission, "0", "Validator's commission rate")

	fsModComm.String(FlagValidator, "", "Validator's public key")
	fsModComm.String(FlagCommission, "0", "Validator's commission rate")

	CmdBond.Flags().AddFlagSet(fsDelegation)
	CmdUnbond.Flags().AddFlagSet(fsDelegation)
	CmdNominate.Flags().AddFlagSet(fsNominate)
	CmdModComm.Flags().AddFlagSet(fsModComm)
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

func cmdNominate(cmd *cobra.Command, args []string) error {

	validator, err := getValidator()
	if err != nil {
		return err
	}
	amount, err := coin.ParseCoin(viper.GetString(FlagAmount))
	if err != nil {
		return err
	}
	commission, err := getCommission()
	if err != nil {
		return err
	}

	tx := stake.NewTxNominate(validator, amount, commission)
	return txcmd.DoTx(tx)
}

func cmdModComm(cmd *cobra.Command, args []string) error {

	validator, err := getValidator()
	if err != nil {
		return err
	}
	commission, err := getCommission()
	if err != nil {
		return err
	}

	tx := stake.NewTxModComm(validator, commission)
	return txcmd.DoTx(tx)
}

////////////////////////////////////////////////////////////

func getCommission() (commission stake.Decimal, err error) {
	commissionStr := viper.GetString(FlagCommission)
	commission, err = stake.NewDecimalFromString(commissionStr)
	if err != nil {
		err = fmt.Errorf("Error parsing commission, must be in decimal format (eg 0.05), Error: ", err.Error())
		return
	}
	if commission.LT(stake.Zero) {
		err = errors.Errorf("Must use positive commission")
		return
	}
	return commission, nil
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
