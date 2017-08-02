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

	tx := stake.NewTx(validator, amount)
	return txcmd.DoTx(tx)
}
