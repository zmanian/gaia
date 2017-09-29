package stake

//TODO  move to tmlibs

import (
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/state"
)

// Queue - Abstract queue implementation object
type Queue struct {
	slot  byte           //Queue name in the store
	store state.SimpleDB //Queue store
	tail  uint64         //Start position of the queue
	head  uint64         //End position of the queue
}

func (q Queue) headPositionKey() []byte { return []byte{q.slot, 0x00} }
func (q Queue) tailPositionKey() []byte { return []byte{q.slot, 0x01} }

// queueKey - get the key for the queue'd record at position 'n'
func (q Queue) queueKey(n uint64) []byte {
	b := make([]byte, 9)
	b[0] = q.slot //add prepended byte
	binary.BigEndian.PutUint64(b[1:], n)
	return b
}

// NewQueue - create a new generic queue under the designate slot
func NewQueue(slot byte, store state.SimpleDB) (*Queue, error) {
	q := &Queue{
		slot:  slot,
		store: store,
		tail:  0,
		head:  0,
	}

	// Test to make sure that the Queue doesn't already exist
	headBytes := store.Get(q.headPositionKey())
	if headBytes != nil {
		return q, fmt.Errorf("cannot create a Queue under the slot %v, Queue already exists", slot)
	}

	// Set the position bytes
	positionBytes := make([]byte, 8)
	q.store.Set(q.headPositionKey(), positionBytes)
	q.store.Set(q.tailPositionKey(), positionBytes)

	return q, nil
}

// LoadQueue - load an existing queue for the slot
func LoadQueue(slot byte, store state.SimpleDB) (*Queue, error) {

	q := &Queue{
		slot:  slot,
		store: store,
		tail:  0,
		head:  0,
	}

	headBytes := store.Get(q.headPositionKey())
	if headBytes == nil {
		//Create a new queue if the head information doesn't exist
		return NewQueue(slot, store)

		//return q, fmt.Errorf("cannot load Queue under the slot %v, head does not exists", slot)
	}
	q.head = binary.BigEndian.Uint64(headBytes)

	tailBytes := store.Get(q.tailPositionKey())
	if tailBytes == nil {
		return q, fmt.Errorf("cannot load Queue under the slot %v, tail does not exists", slot)
	}
	q.tail = binary.BigEndian.Uint64(tailBytes)

	return q, nil
}

func (q *Queue) incrementHead() {
	q.head++
	headBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(headBytes, q.head)
	q.store.Set(q.headPositionKey(), headBytes)
}

func (q *Queue) incrementTail() {
	q.tail++
	tailBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tailBytes, q.tail)
	q.store.Set(q.tailPositionKey(), tailBytes)
}

func (q Queue) length() uint64 {
	return (q.tail - q.head)
}

// Push - Add to the beginning/tail of the queue
func (q *Queue) Push(bytes []byte) {
	pushKey := q.queueKey(q.tail)
	q.store.Set(pushKey, bytes)
	q.incrementTail()
}

// Pop - Remove from the end/head of queue
func (q *Queue) Pop() {
	if q.length() == 0 {
		return
	}
	popKey := q.queueKey(q.head)
	q.store.Remove(popKey)
	q.incrementHead()
}

// Peek - Get the end/head record on the queue
func (q Queue) Peek() []byte {
	if q.length() == 0 {
		return nil
	}
	peekKey := q.queueKey(q.head)
	return q.store.Get(peekKey)
}

// GetAll - Return an array of all the elements inside the queue
func (q Queue) GetAll() [][]byte {
	//panic(fmt.Sprintf("length %v, head %v, tail %v", q.length() == 0, q.head, q.tail))
	if q.length() == 0 {
		return nil
	}
	var res [][]byte
	for i := q.head; i < q.tail; i++ {
		key := q.queueKey(i)
		res = append(res, q.store.Get(key))
		//panic(fmt.Sprintf("%v", q.store.Get(key)))
		//panic(fmt.Sprintf("%v", q.Peek()))
	}
	//panic(fmt.Sprintf("%v", res))
	return res
}
