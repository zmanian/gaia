package commands

import (
	"encoding/hex"
	"flag"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	wire "github.com/tendermint/go-wire"

	"github.com/cosmos/gaia/modules/stake"
)

// CmdBond - cobra command to bond coins
var CmdBond = &cobra.Command{
	Use:   "bond",
	Short: "Bond some coins to give voting power to a validator",
	RunE:  cmdBond,
}

// CmdUnbond - cobra command to unbond coins
var CmdUnbond = &cobra.Command{
	Use:   "unbond",
	Short: "unbond your coins from a validator",
	RunE:  cmdBond,
}

//nolint
const (
	FlagAmount    = "amount"
	FlagValidator = "validator"
)

func init() {

	//Add Flags
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagValidator, "", "Validator's public key")
	fs.Int(FlagAmount, 0, "Amount of Atoms")
	CmdBond.Flags().AddFlagSet(fs)
	CmdUnbond.Flags().AddFlagSet(fs)

	bcmd.RegisterTxSubcommand(CmdBond)
	bcmd.RegisterStartPlugin("stake",
		func() types.Plugin {
			return stake.Plugin{UnbondingPeriod: 100}
		},
	)

	bcmd.RegisterStartPlugin("delegationGame",
		func() types.Plugin {
			return &dg.Plugin{
				BondDenom:             "atom",
				MinimumOwnCoinsBonded: 1000,
			}
		},
	)

	bcmd.RegisterStartPlugin("stake",
		func() types.Plugin {
			return &stake.Plugin{
				UnbondingPeriod:    0,
				CoinDenom:          "atom",
				DisableVotingPower: true,
			}
		},
	)

}

func cmdBond(cmd *cobra.Command, args []string) error {
	// convert validator pubkey to bytes
	validator, err := hex.DecodeString(bcmd.StripHex(viper.GetString(FlagValidator)))
	if err != nil {
		return errors.Errorf("Validator is invalid hex: %v\n", err)
	}

	bondTx := stake.BondTx{
		ValidatorPubKey: validator,
		Sequence:        0,
	}
	fmt.Println("BondTx:", string(wire.JSONBytes(bondTx)))
	bytes := wire.BinaryBytes(bondTx)
	return bcmd.AppTx("stake", bytes)
}

func cmdUnbond(cmd *cobra.Command, args []string) error {
	// convert validator pubkey to bytes
	validator, err := hex.DecodeString(bcmd.StripHex(validatorFlag))
	if err != nil {
		return errors.Errorf("Validator is invalid hex: %v\n", err)
	}

	bondTx := stake.BondTx{ValidatorPubKey: validator}
	fmt.Println("BondTx:", string(wire.JSONBytes(bondTx)))
	bytes := wire.BinaryBytes(bondTx)
	return bcmd.AppTx("stake", bytes)
}
