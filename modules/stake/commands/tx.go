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
	CmdDelegate = &cobra.Command{
		Use:   "delegate",
		Short: "delegate coins to an existing validator/candidate",
		RunE:  cmdDelegate,
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
	CmdDelegate.Flags().AddFlagSet(fs)
	CmdUnbond.Flags().AddFlagSet(fs)
}

// MakeTx function type for cmd consolidation
type MakeTx func(coin.Coin, crypto.PubKey) sdk.Tx

func cmdDeclareCandidacy(cmd *cobra.Command, args []string) error {
	return cmdBondUpdate(cmd, args, stake.NewTxDeclareCandidacy)
}

func cmdDelegate(cmd *cobra.Command, args []string) error {
	return cmdBondUpdate(cmd, args, stake.NewTxDelegate)
}

func cmdBondUpdate(cmd *cobra.Command, args []string, makeTx MakeTx) error {
	amount, err := coin.ParseCoin(viper.GetString(FlagAmount))
	if err != nil {
		return err
	}

	pk, err := GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	tx := makeTx(amount, pk)
	return txcmd.DoTx(tx)
}

func cmdUnbond(cmd *cobra.Command, args []string) error {

	sharesRaw := viper.GetInt64(FlagAmount)
	if sharesRaw <= 0 {
		return fmt.Errorf("shares must be positive interger")
	}
	shares := uint64(sharesRaw)

	pk, err := GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	tx := stake.NewTxUnbond(shares, pk)
	return txcmd.DoTx(tx)
}

// GetPubKey - create the pubkey from a pubkey string
func GetPubKey(pubKeyStr string) (pk crypto.PubKey, err error) {

	if len(pubKeyStr) == 0 {
		err = fmt.Errorf("must use --pubkey flag")
		return
	}
	if len(pubKeyStr) != 64 { //if len(pkBytes) != 32 {
		err = fmt.Errorf("pubkey must be Ed25519 hex encoded string which is 64 characters long")
		return
	}
	var pkBytes []byte
	pkBytes, err = hex.DecodeString(pubKeyStr)
	if err != nil {
		return
	}
	var pkEd crypto.PubKeyEd25519
	copy(pkEd[:], pkBytes[:])
	pk = pkEd.Wrap()
	return
}
