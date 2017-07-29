package stake

import (
	"encoding/binary"
	"fmt"

	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
)

var (
	queueKeyPrefix = "stake/u/"
	headKey        = []byte(queueKeyPrefix + "head")
	tailKey        = []byte(queueKeyPrefix + "tail")
)

func queueKey(n uint64) []byte {
	return []byte(queueKeyPrefix + fmt.Sprintf("%x", n))
}

type UnbondQueue struct {
	head  uint64
	tail  uint64
	store types.KVStore
}

func (uq *UnbondQueue) incrementHead() {
	uq.head += 1
	headBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(headBytes, uq.head)
	uq.store.Set(headKey, headBytes)
}

func (uq *UnbondQueue) incrementTail() {
	uq.tail += 1
	tailBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tailBytes, uq.tail)
	uq.store.Set(tailKey, tailBytes)
}

func (uq UnbondQueue) Length() uint64 {
	return uq.tail - uq.head
}

func (uq *UnbondQueue) Push(unbond Unbond) {
	pushKey := queueKey(uq.tail)
	unbondBytes := wire.BinaryBytes(unbond)
	uq.store.Set(pushKey, unbondBytes)
	uq.incrementTail()
}

func (uq *UnbondQueue) Pop() {
	if uq.Length() == 0 {
		return
	}
	popKey := queueKey(uq.head)
	uq.store.Set(popKey, nil) // TODO: remove
	uq.incrementHead()
}

func (uq UnbondQueue) Peek() (unbond *Unbond) {
	if uq.Length() == 0 {
		return nil
	}
	peekKey := queueKey(uq.head)
	unbondBytes := uq.store.Get(peekKey)
	wire.ReadBinaryBytes(unbondBytes, unbond)
	return
}

func loadUnbondQueue(store types.KVStore) UnbondQueue {
	uq := UnbondQueue{}
	uq.store = store
	headBytes := store.Get(headKey)
	if headBytes == nil {
		uq.head = 0
	} else {
		uq.head = binary.BigEndian.Uint64(headBytes)
	}
	tailBytes := store.Get(tailKey)
	if tailBytes == nil {
		uq.tail = 0
	} else {
		uq.tail = binary.BigEndian.Uint64(tailBytes)
	}
	return uq
}
