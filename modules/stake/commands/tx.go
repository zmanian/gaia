package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"

	"github.com/cosmos/cosmos-sdk/client/commands/keys"
	txcmd "github.com/cosmos/cosmos-sdk/client/commands/txs"
	"github.com/cosmos/cosmos-sdk/modules/coin"

	"github.com/cosmos/gaia/modules/stake"
)

// nolint
const (
	FlagAmount = "amount"
	FlagPubKey = "pubkey"
)

// nolint
var (
	CmdBond = &cobra.Command{
		Use:   "bond",
		Short: "bond coins to your validator bond account",
		RunE:  cmdBond,
	}
	CmdUnbond = &cobra.Command{
		Use:   "unbond",
		Short: "unbond coins from your validator bond account",
		RunE:  cmdUnbond,
	}
)

func init() {
	// Add Flags
	fsDelegation := flag.NewFlagSet("", flag.ContinueOnError)
	fsDelegation.String(FlagAmount, "1atom", "Amount of Atoms")
	fsDelegation.String(FlagPubKey, "", "PubKey of the Validator")

	CmdBond.Flags().AddFlagSet(fsDelegation)
	CmdUnbond.Flags().AddFlagSet(fsDelegation)
}

func cmdBond(cmd *cobra.Command, args []string) error {
	amount, err := coin.ParseCoin(viper.GetString(FlagAmount))
	if err != nil {
		return err
	}

	var pubkey crypto.PubKey
	pubkeyStr := viper.GetString(FlagPubKey)
	if len(pubkeyStr) != 0 {
		pkBytes, err := hex.DecodeString(pubkeyStr)
		if err != nil {
			return err
		}

		if len(pkBytes) != 32 { //if len(pubkeyStr) != 64 {
			return fmt.Errorf("pubkey must be hex encoded string which is 64 characters long")
		}
		var pkEd crypto.PubKeyEd25519
		copy(pkEd[:], pkBytes[:])
		pubkey = pkEd.Wrap()

	} else { // if pubkey flag is not used get the pubkey of the signer
		name := viper.GetString(txcmd.FlagName)
		if len(name) == 0 {
			return fmt.Errorf("must use --name flag")
		}

		info, err := keys.GetKeyManager().Get(name)
		if err != nil {
			return err
		}
		pubkey = info.PubKey
	}

	tx := stake.NewTxBond(amount, wire.BinaryBytes(pubkey))
	return txcmd.DoTx(tx)
}

func cmdUnbond(cmd *cobra.Command, args []string) error {
	amount, err := coin.ParseCoin(viper.GetString(FlagAmount))
	if err != nil {
		return err
	}

	tx := stake.NewTxUnbond(amount)
	return txcmd.DoTx(tx)
}
