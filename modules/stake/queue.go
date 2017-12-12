package stake

//TODO  move to tmlibs

import (
	"encoding/binary"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
	crypto "github.com/tendermint/go-crypto"
)

// Queue - interface for a queue
type Queue interface {
	Push(bytes []byte)
	Pop()
	Peek() []byte
}

//______________________________________________________________________________

// MerkleQueue - Abstract queue implementation object
type MerkleQueue struct {
	slot  byte           //MerkleQueue name in the store
	store state.SimpleDB //MerkleQueue store
	tail  uint64         //Start position of the queue
	head  uint64         //End position of the queue
}

var _ Queue = &MerkleQueue{} //enforce interface at compile time

func (q MerkleQueue) headPositionKey() []byte { return []byte{q.slot, 0x00} }
func (q MerkleQueue) tailPositionKey() []byte { return []byte{q.slot, 0x01} }

// queueKey - get the key for the queue'd record at position 'n'
func (q MerkleQueue) queueKey(n uint64) []byte {
	b := make([]byte, 9)
	b[0] = q.slot //add prepended byte
	binary.BigEndian.PutUint64(b[1:], n)
	return b
}

// NewMerkleQueue - create a new generic queue under the designate slot
func NewMerkleQueue(slot byte, store state.SimpleDB) (*MerkleQueue, error) {
	q := &MerkleQueue{
		slot:  slot,
		store: store,
		tail:  0,
		head:  0,
	}

	// Test to make sure that the MerkleQueue doesn't already exist
	headBytes := store.Get(q.headPositionKey())
	if headBytes != nil {
		return q, fmt.Errorf("cannot create a MerkleQueue under the slot %v, MerkleQueue already exists", slot)
	}

	// Set the position bytes
	positionBytes := make([]byte, 8)
	q.store.Set(q.headPositionKey(), positionBytes)
	q.store.Set(q.tailPositionKey(), positionBytes)

	return q, nil
}

// LoadQueue - load an existing queue for the slot
func LoadQueue(slot byte, store state.SimpleDB) (*MerkleQueue, error) {

	q := &MerkleQueue{
		slot:  slot,
		store: store,
		tail:  0,
		head:  0,
	}

	headBytes := store.Get(q.headPositionKey())
	if headBytes == nil {
		//Create a new queue if the head information doesn't exist
		return NewMerkleQueue(slot, store)

		//return q, fmt.Errorf("cannot load MerkleQueue under the slot %v, head does not exists", slot)
	}
	q.head = binary.BigEndian.Uint64(headBytes)

	tailBytes := store.Get(q.tailPositionKey())
	if tailBytes == nil {
		return q, fmt.Errorf("cannot load MerkleQueue under the slot %v, tail does not exists", slot)
	}
	q.tail = binary.BigEndian.Uint64(tailBytes)

	return q, nil
}

func (q *MerkleQueue) incrementHead() {
	q.head++
	headBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(headBytes, q.head)
	q.store.Set(q.headPositionKey(), headBytes)
}

func (q *MerkleQueue) incrementTail() {
	q.tail++
	tailBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tailBytes, q.tail)
	q.store.Set(q.tailPositionKey(), tailBytes)
}

func (q MerkleQueue) length() uint64 {
	return (q.tail - q.head)
}

// Push - Add to the beginning/tail of the queue
func (q *MerkleQueue) Push(bytes []byte) {
	pushKey := q.queueKey(q.tail)
	q.store.Set(pushKey, bytes)
	q.incrementTail()
}

// Pop - Remove from the end/head of queue
func (q *MerkleQueue) Pop() {
	if q.length() == 0 {
		return
	}
	popKey := q.queueKey(q.head)
	q.store.Remove(popKey)
	q.incrementHead()
}

// Peek - Get the end/head record on the queue
func (q MerkleQueue) Peek() []byte {
	if q.length() == 0 {
		return nil
	}
	peekKey := q.queueKey(q.head)
	return q.store.Get(peekKey)
}

//______________________________________________________________________________

// QueueElem - TODO
type QueueElem struct {
	Candidate  crypto.PubKey
	InitHeight uint64 // when the queue was initiated
}

// QueueElemUnbondDelegation - TODO
type QueueElemUnbondDelegation struct {
	QueueElem
	Payout          sdk.Actor // account to pay out to
	Amount          uint64    // amount of shares which are unbonding
	StartSlashRatio uint64    // old candidate slash ratio at start of re-delegation
}

// QueueElemReDelegate - TODO
type QueueElemReDelegate struct {
	QueueElem
	Payout       sdk.Actor     // account to pay out to
	Shares       uint64        // amount of shares which are unbonding
	NewCandidate crypto.PubKey // validator to bond to after unbond
}

// QueueElemUnbondCandidate - TODO
type QueueElemUnbondCandidate struct {
	QueueElem
}
