package commands

import (
	"encoding/hex"
	"flag"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/basecoin"
	txcmd "github.com/tendermint/basecoin/client/commands/txs"
	"github.com/tendermint/basecoin/modules/coin"

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
	CmdUnbond = &cobra.Command{
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
	fsDelegation.Int(FlagAmount, 0, "Amount of Atoms")

	fsNominate.AddFlagSet(fsDelegation)
	fsNominate.String(FlagCommission, "0.01", "Validator's commission rate")

	fsModComm.String(FlagValidator, "", "Validator's public key")
	fsModComm.String(FlagCommission, "0.01", "Validator's commission rate")

	CmdBond.Flags().AddFlagSet(fsDelegation)
	CmdUnbond.Flags().AddFlagSet(fsDelegation)
	CmdNominate.Flags().AddFlagSet(fsNominate)
	CmdModComm.Flags().AddFlagSet(fsModComm)
}

func cmdBond(cmd *cobra.Command, args []string) error {
	return cmdDelegation(stake.NewTxBond)
}

func cmdUnbond(cmd *cobra.Command, args []string) error {
	return cmdDelegation(stake.NewTxUnbond)
}

func cmdDelegation(NewTx func(validator basecoin.Actor, amount coin.Coin) basecoin.TxInner) error {
	// convert validator pubkey to bytes
	validator, err := hex.DecodeString(bcmd.StripHex(validatorFlag))
	if err != nil {
		return errors.Errorf("Validator is invalid hex: %v\n", err)
	}

	amount, err := coin.ParseCoin(viper.GetString(FlagAmount))
	if err != nil {
		return err
	}

	tx := NewTx(validator, amount)
	return txcmd.DoTx(tx)
}

func cmdNominate(cmd *cobra.Command, args []string) error {
	// convert validator pubkey to bytes
	validator, err := hex.DecodeString(bcmd.StripHex(validatorFlag))
	if err != nil {
		return errors.Errorf("Validator is invalid hex: %v\n", err)
	}
	amount, err := coin.ParseCoin(viper.GetString(FlagAmount))
	if err != nil {
		return err
	}
	commission := viper.GetString(FlagCommission)

	tx := stake.NewTxNominate(validator, amount, commission)
	return txcmd.DoTx(tx)
}

func cmdModComm(cmd *cobra.Command, args []string) error {
	validator, err := hex.DecodeString(bcmd.StripHex(validatorFlag))
	if err != nil {
		return errors.Errorf("Validator is invalid hex: %v\n", err)
	}
	commission := viper.GetString(FlagCommission)

	tx := stake.NewTxModComm(validator, commission)
	return txcmd.DoTx(tx)
}
