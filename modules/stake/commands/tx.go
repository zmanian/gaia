package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	wire "github.com/tendermint/go-wire"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/client/commands/keys"
	txcmd "github.com/cosmos/cosmos-sdk/client/commands/txs"
	"github.com/cosmos/cosmos-sdk/modules/coin"

	"github.com/cosmos/gaia/modules/stake"
)

//nolint
const (
	FlagAmount = "amount"
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
	fsDelegation.String(FlagAmount, "1atom", "Amount of Atoms") //TODO make string once decimal integrated with coin

	CmdBond.Flags().AddFlagSet(fsDelegation)
	CmdUnbond.Flags().AddFlagSet(fsDelegation)
}

func cmdBond(cmd *cobra.Command, args []string) error {
	amount, err := coin.ParseCoin(viper.GetString(FlagAmount))
	if err != nil {
		return err
	}

	name := viper.GetString(txcmd.FlagName)
	if len(name) == 0 {
		return fmt.Errorf("must use --name flag")
	}
	info, err := keys.GetKeyManager().Get(name)
	if err != nil {
		return err
	}
	tx := stake.NewTxBond(amount, wire.BinaryBytes(info.PubKey))
	return txcmd.DoTx(tx)
}

func cmdUnbond(cmd *cobra.Command, args []string) error {
	return cmdBonding(stake.NewTxUnbond)
}

func cmdBonding(NewTx func(amount coin.Coin) sdk.Tx) error {

	amount, err := coin.ParseCoin(viper.GetString(FlagAmount))
	if err != nil {
		return err
	}

	tx := NewTx(amount)
	return txcmd.DoTx(tx)
}
