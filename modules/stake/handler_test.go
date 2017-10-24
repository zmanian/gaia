package stake

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/state"
)

func dummyTransferFn(store map[string]int64) transferFn {
	return func(from, to sdk.Actor, coins coin.Coins) abci.Result {
		store[string(from.Address)] -= int64(coins[0].Amount)
		store[string(to.Address)] += int64(coins[0].Amount)
		return abci.OK
	}
}

func initAccounts(n int, amount int64) ([]sdk.Actor, map[string]int64) {
	accStore := map[string]int64{}
	senders := newActors(n)
	for _, sender := range senders {
		accStore[string(sender.Address)] = amount
	}
	return senders, accStore
}

func newTxDeclareCandidacy(amt int64, pubKey crypto.PubKey) TxDeclareCandidacy {
	return TxDeclareCandidacy{BondUpdate{
		PubKey: pubKey,
		Bond:   coin.Coin{"fermion", amt},
	}}
}

func newTxDelegate(amt int64, pubKey crypto.PubKey) TxDelegate {
	return TxDelegate{BondUpdate{
		PubKey: pubKey,
		Bond:   coin.Coin{"fermion", amt},
	}}
}

func newTxUndelegate(amt int64, pubKey crypto.PubKey) TxUndelegate {
	return TxUndelegate{BondUpdate{
		PubKey: pubKey,
		Bond:   coin.Coin{"fermion", amt},
	}}
}

func newPubKey(pk string) crypto.PubKey {
	pkBytes, _ := hex.DecodeString(pk)
	var pkEd crypto.PubKeyEd25519
	copy(pkEd[:], pkBytes[:])
	return pkEd.Wrap()
}

//dummy public keys used for testing
var (
	pk1 = newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB57")
	pk2 = newPubKey("E2CB355FD7965D70627AB279016D713CAF82612216D833372187520E264C3588")
	pk3 = newPubKey("03E6F86BC3B9BDF9E6FEAB434BE1A3A2817BF4AE15CE0D4054C6EA7BFC2A30BE")
)

func TestDuplicatesTxDeclareCandidacy(t *testing.T) {
	assert := assert.New(t)

	store := state.NewMemKVStore() // for bonds
	initSender := int64(1000)
	senders, accStore := initAccounts(2, initSender) // for accounts
	sender, sender2 := senders[0], senders[1]

	bondAmount := int64(10)
	txDeclareCandidacy := newTxDeclareCandidacy(bondAmount, pk1)
	//panic(fmt.Sprintf("debug txDeclareCandidacy: %v\n", txDeclareCandidacy))
	got := runTxDeclareCandidacy(store, sender, dummyTransferFn(accStore), txDeclareCandidacy)
	assert.Equal(got, abci.OK, "expected no error on runTxDeclareCandidacy")

	// one sender can bond to two different pubKeys
	txDeclareCandidacy.PubKey = pk2
	err := checkTxDeclareCandidacy(txDeclareCandidacy, sender, store)
	assert.Nil(err, "didn't expected error on checkTx")

	// two senders cant bond to the same pubkey
	txDeclareCandidacy.PubKey = pk1
	err = checkTxDeclareCandidacy(txDeclareCandidacy, sender2, store)
	assert.NotNil(err, "expected error on checkTx")
}

func TestIncrementsTxDelegate(t *testing.T) {
	assert := assert.New(t)

	store := state.NewMemKVStore() // for bonds
	initSender := int64(1000)
	senders, accStore := initAccounts(1, initSender) // for accounts
	sender := senders[0]
	holder := defaultParams().HoldAccount

	// first declare ca
	bondAmount := int64(10)
	txDeclareCandidacy := newTxDeclareCandidacy(bondAmount, pk1)
	got := runTxDeclareCandidacy(store, sender, dummyTransferFn(accStore), txDeclareCandidacy)
	assert.True(got.IsOK(), "expected declare candidacy tx to be ok, got %v", got)
	expectedBond := bondAmount // 1 since we send 1 at the start of loop,

	// just send the same txbond multiple times
	txDeligate := newTxDelegate(bondAmount, pk1)
	for i := 0; i < 5; i++ {
		got := runTxDelegate(store, sender, dummyTransferFn(accStore), txDeligate)
		assert.True(got.IsOK(), "expected tx %d to be ok, got %v", i, got)

		//Check that the accounts and the bond account have the appropriate values
		candidates := LoadCandidates(store)
		expectedBond += bondAmount
		expectedSender := initSender - expectedBond
		gotBonded := int64(candidates[0].Shares)
		gotHolder := accStore[string(holder.Address)]
		gotSender := accStore[string(sender.Address)]
		assert.Equal(expectedBond, gotBonded, "%v, %v", expectedBond, gotBonded)
		assert.Equal(expectedBond, gotHolder, "%v, %v", expectedBond, gotHolder)
		assert.Equal(expectedSender, gotSender, "%v, %v", expectedSender, gotSender)
	}
}

func TestIncrementsTxUndelegate(t *testing.T) {
	assert := assert.New(t)

	store := state.NewMemKVStore() // for bonds
	initSender := int64(0)
	senders, accStore := initAccounts(1, initSender) // for accounts
	sender := senders[0]
	holder := defaultParams().HoldAccount

	// set initial bond
	initBond := int64(1000)
	accStore[string(sender.Address)] = initBond
	got := runTxDeclareCandidacy(store, sender, dummyTransferFn(accStore), newTxDeclareCandidacy(initBond, pk1))
	assert.True(got.IsOK(), "expected initial bond tx to be ok, got %v", got)

	// just send the same txunbond multiple times
	unbondAmount := int64(10)
	txUndeligate := newTxUndelegate(unbondAmount, pk1)
	for i := 0; i < 5; i++ {
		got := runTxUndelegate(store, sender, dummyTransferFn(accStore), txUndeligate)
		assert.True(got.IsOK(), "expected tx %d to be ok, got %v", i, got)

		//Check that the accounts and the bond account have the appropriate values
		candidates := LoadCandidates(store)
		expectedBond := initBond - int64(i+1)*unbondAmount // +1 since we send 1 at the start of loop
		expectedSender := initSender + (initBond - expectedBond)
		gotBonded := int64(candidates[0].Shares)
		gotHolder := accStore[string(holder.Address)]
		gotSender := accStore[string(sender.Address)]

		assert.Equal(expectedBond, gotBonded, "%v, %v", expectedBond, gotBonded)
		assert.Equal(expectedBond, gotHolder, "%v, %v", expectedBond, gotHolder)
		assert.Equal(expectedSender, gotSender, "%v, %v", expectedSender, gotSender)
	}
}

func TestMultipleTxDeclareCandidacy(t *testing.T) {
	assert := assert.New(t)

	store := state.NewMemKVStore()
	initSender := int64(1000)
	senders, accStore := initAccounts(3, initSender)
	pubKeys := []crypto.PubKey{pk1, pk2, pk3}

	// bond them all
	for i, sender := range senders {
		txDeclareCandidacy := newTxDeclareCandidacy(10, pubKeys[i])
		got := runTxDeclareCandidacy(store, sender, dummyTransferFn(accStore), txDeclareCandidacy)
		assert.True(got.IsOK(), "expected tx %d to be ok, got %v", i, got)

		//Check that the account is bonded
		candidates := LoadCandidates(store)
		val := candidates[i]
		balanceGot, balanceExpd := accStore[string(val.Owner.Address)], initSender-10
		assert.Equal(i+1, len(candidates), "expected %d candidates got %d, candidates: %v", i+1, len(candidates), candidates)
		assert.Equal(10, int(val.Shares), "expected %d shares, got %d", 10, val.Shares)
		assert.Equal(balanceExpd, balanceGot, "expected account to have %d, got %d", balanceExpd, balanceGot)
	}

	// unbond them all
	for i, sender := range senders {
		candidatePre := LoadCandidate(store, pubKeys[i])
		txUndeligate := newTxUndelegate(10, pubKeys[i])
		//panic(len(LoadCandidates(store)))
		got := runTxUndelegate(store, sender, dummyTransferFn(accStore), txUndeligate)
		assert.True(got.IsOK(), "expected tx %d to be ok, got %v", i, got)

		//Check that the account is unbonded
		candidates := LoadCandidates(store)
		assert.Equal(len(senders)-(i+1), len(candidates), "expected %d candidates got %d", len(senders)-(i+1), len(candidates))

		candidatePost := LoadCandidate(store, pubKeys[i])
		balanceGot, balanceExpd := accStore[string(candidatePre.Owner.Address)], initSender
		assert.Nil(candidatePost, "expected nil candidate retrieve, got %d", 0, candidatePost)
		assert.Equal(balanceExpd, balanceGot, "expected account to have %d, got %d", balanceExpd, balanceGot)
	}
}
