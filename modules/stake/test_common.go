package stake

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk"
)

func newActors(n int) (actors []sdk.Actor) {
	for i := 0; i < n; i++ {
		actors = append(actors, sdk.Actor{
			"testChain", "testapp", []byte(fmt.Sprintf("addr%d", i))})
	}

	return
}

func newPubKey(pk string) crypto.PubKey {
	pkBytes, _ := hex.DecodeString(pk)
	var pkEd crypto.PubKeyEd25519
	copy(pkEd[:], pkBytes[:])
	return pkEd.Wrap()
}

// dummy pubkeys used for testing
var pks = []crypto.PubKey{
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB51"),
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB52"),
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB53"),
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB54"),
	newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB55"),
}

// NOTE: PubKey is supposed to be the binaryBytes of the crypto.PubKey
// instead this is just being set the address here for testing purposes
func candidatesFromActors(actors []sdk.Actor, amts []int64) (candidates Candidates) {
	for i := 0; i < len(actors); i++ {
		c := &Candidate{
			PubKey:      pks[i],
			Owner:       actors[i],
			Assets:      NewFraction(amts[i]),
			Liabilities: NewFraction(amts[i]),
			VotingPower: NewFraction(amts[i]),
		}
		candidates = append(candidates, c)
	}

	return
}

// helper function test if Candidate is changed asabci.Validator
func testChange(t *testing.T, val Validator, chg *abci.Validator) {
	assert := assert.New(t)
	assert.Equal(val.PubKey.Bytes(), chg.PubKey)
	assert.Equal(val.VotingPower.Evaluate(), chg.Power)
}

// helper function test if Candidate is removed as abci.Validator
func testRemove(t *testing.T, val Validator, chg *abci.Validator) {
	assert := assert.New(t)
	assert.Equal(val.PubKey.Bytes(), chg.PubKey)
	assert.Equal(int64(0), chg.Power)
}
