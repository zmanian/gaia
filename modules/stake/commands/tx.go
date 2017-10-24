package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk"
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
	CmdDeclareCandidacy = &cobra.Command{
		Use:   "declare-candidacy",
		Short: "create new validator/candidate account and delegate some coins to it",
		RunE:  cmdDeclareCandidacy,
	}
	CmdBond = &cobra.Command{
		Use:   "bond",
		Short: "delegate coins to an existing validator/candidate",
		RunE:  cmdBond,
	}
	CmdUnbond = &cobra.Command{
		Use:   "unbond",
		Short: "unbond coins from a validator/candidate",
		RunE:  cmdUnbond,
	}
)

func init() {

	//Add Flags to commands
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagPubKey, "", "PubKey of the validator-candidate")
	fs.String(FlagAmount, "1fermion", "Amount of coins to bond")

	CmdDeclareCandidacy.Flags().AddFlagSet(fs)
	CmdBond.Flags().AddFlagSet(fs)
	CmdUnbond.Flags().AddFlagSet(fs)
}

type makeTx func(coin.Coin, crypto.PubKey) sdk.Tx

func cmdDeclareCandidacy(cmd *cobra.Command, args []string) error {
	return cmdBondUpdate(cmd, args, stake.NewTxDeclareCandidacy)
}

func cmdBond(cmd *cobra.Command, args []string) error {
	return cmdBondUpdate(cmd, args, stake.NewTxDelegate)
}

func cmdUnbond(cmd *cobra.Command, args []string) error {
	return cmdBondUpdate(cmd, args, stake.NewTxUndelegate)
}

func cmdBondUpdate(cmd *cobra.Command, args []string, makeTx makeTx) error {
	amount, err := coin.ParseCoin(viper.GetString(FlagAmount))
	if err != nil {
		return err
	}

	pubKey, err := getPubKey()
	if err != nil {
		return err
	}

	tx := makeTx(amount, pubKey)
	return txcmd.DoTx(tx)
}

func getPubKey() (pk crypto.PubKey, err error) {
	// Get the pubkey
	pubkeyStr := viper.GetString(FlagPubKey)
	if len(pubkeyStr) == 0 {
		err = fmt.Errorf("must use --pubkey flag")
		return
	}
	if len(pubkeyStr) != 64 { //if len(pkBytes) != 32 {
		err = fmt.Errorf("pubkey must be Ed25519 hex encoded string which is 64 characters long")
		return
	}
	var pkBytes []byte
	pkBytes, err = hex.DecodeString(pubkeyStr)
	if err != nil {
		return
	}
	var pkEd crypto.PubKeyEd25519
	copy(pkEd[:], pkBytes[:])
	pk = pkEd.Wrap()
	return
}
