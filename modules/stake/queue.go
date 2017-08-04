package stake

import (
	"encoding/binary"
	"fmt"

	"github.com/tendermint/basecoin/state"
)

// Queue - Abstract queue implementation object
type Queue struct {
	tail    uint64         //Start position of the queue
	head    uint64         //End position of the queue
	store   state.SimpleDB //Queue store
	name    string         //Queue name in the store
	headKey []byte         //Store-key of the record which holds the head
	tailKey []byte         //Store-key of the record which holds the tail
}

// NewQueue - create a new generic queue under the namespace prefixed by name
func NewQueue(name string, store state.SimpleDB) (Queue, error) {
	q := Queue{
		tail:    0,
		head:    0,
		store:   store,
		name:    name,
		headKey: []byte(name + "head"),
		tailKey: []byte(name + "tail"),
	}

	// Test to make sure that the Queue doesn't already exist
	headBytes := store.Get(q.headKey)
	if headBytes != nil {
		return q, fmt.Errorf("cannot create a Queue under the name %v, Queue already exists")
	}

	// Set the position bytes
	positionBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(positionBytes, 0)
	q.store.Set(q.headKey, bytes)
	q.store.Set(q.tailKey, bytes)

	return q, nil
}

// LoadQueue - load an existing namespace
func LoadQueue(store state.KVStore) Queue {
	q := Queue{}
	q.store = store
	headBytes := store.Get(headKey)
	if headBytes == nil {
		q.head = 0
	} else {
		q.head = binary.BigEndian.Uint64(headBytes)
	}
	tailBytes := store.Get(tailKey)
	if tailBytes == nil {
		q.tail = 0
	} else {
		q.tail = binary.BigEndian.Uint64(tailBytes)
	}
	return q
}

// getQueueKey - get the key for the queue'd record at position 'n'
func (q Queue) getQueueKey(n uint64) []byte {
	return []byte(q.name + fmt.Sprintf("%x", n))
}

func (q *Queue) incrementHead() {
	q.head++
	headBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(headBytes, q.head)
	q.store.Set(headKey, headBytes)
}

func (q *Queue) incrementTail() {
	q.tail++
	tailBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tailBytes, q.tail)
	q.store.Set(tailKey, tailBytes)
}

func (q Queue) length() uint64 {
	return q.tail - q.head
}

// Push - Add to the beginning/tail of the queue
func (q *Queue) Push(bytes []byte) {
	pushKey := getQueueKey(q.tail)
	q.store.Set(pushKey, bytes)
	q.incrementTail()
}

// Pop - Remove from the end/head of queue
func (q *Queue) Pop() {
	if q.length() == 0 {
		return
	}
	popKey := getQueueKey(q.head)
	q.store.Set(popKey, nil) // TODO: remove
	q.incrementHead()
}

// Peek - Get the end/head record on the queue
func (q Queue) Peek() []byte {
	if q.length() == 0 {
		return nil
	}
	peekKey := getQueueKey(q.head)
	return q.store.Get(peekKey)
}
