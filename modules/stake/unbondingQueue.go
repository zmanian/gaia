package stake

import (
	"encoding/binary"
	"fmt"

	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/go-wire"
)

const (
	queueKeyPrefix = "q/"
	headKey        = []byte(queueKeyPrefix + "head")
	tailKey        = []byte(queueKeyPrefix + "tail")
)

// getQueueKey - get the key for the queue'd record at position 'n'
func getQueueKey(n uint64) []byte {
	return []byte(queueKeyPrefix + fmt.Sprintf("%x", n))
}

// UnbondQueue - the Queue of the bonded tokens waiting to be unbonded
type UnbondQueue struct {
	tail  uint64
	head  uint64
	store state.KVStore
}

func (uq *UnbondQueue) incrementHead() {
	uq.head++
	headBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(headBytes, uq.head)
	uq.store.Set(headKey, headBytes)
}

func (uq *UnbondQueue) incrementTail() {
	uq.tail++
	tailBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tailBytes, uq.tail)
	uq.store.Set(tailKey, tailBytes)
}

func (uq UnbondQueue) length() uint64 {
	return uq.tail - uq.head
}

// Push - Add to the back of the Queue
func (uq *UnbondQueue) Push(unbond Unbond) {
	pushKey := getQueueKey(uq.tail)
	unbondBytes := wire.BinaryBytes(unbond)
	uq.store.Set(pushKey, unbondBytes)
	uq.incrementTail()
}

// Pop - Remove from the top of Queue
func (uq *UnbondQueue) Pop() {
	if uq.length() == 0 {
		return
	}
	popKey := getQueueKey(uq.head)
	uq.store.Set(popKey, nil) // TODO: remove
	uq.incrementHead()
}

// Peek - Get the top record on the stack
func (uq UnbondQueue) Peek() (unbond *Unbond) {
	if uq.length() == 0 {
		return nil
	}
	peekKey := getQueueKey(uq.head)
	unbondBytes := uq.store.Get(peekKey)
	wire.ReadBinaryBytes(unbondBytes, unbond)
	return
}

func loadUnbondQueue(store state.KVStore) UnbondQueue {
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
