package stake

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// this makes sure that txs are rejected with invalid data or permissions
// TestQueue - test the queue!
func TestQueue(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	store := state.NewMemKVStore()

	//records for the queue
	rec1, rec2, rec3 := []byte("record1"), []byte("record2"), []byte("record3")
	var slot1 byte = 0x01

	// Create new Queue, make sure empty
	queueNew, err := NewQueue(slot1, store)
	require.Nil(err)
	assert.Nil(queueNew.Peek())

	// Load the new queue
	queueLoad, err := LoadQueue(slot1, store)
	require.Nil(err)

	// Push a record to the queue
	queueLoad.Push(rec1)

	// Reload the queue to make sure loaded queue has the newly added record
	queueLoad2, err := LoadQueue(slot1, store)
	require.Nil(err)
	assert.Equal(rec1, queueLoad2.Peek())

	// Add a two more records
	// Pop and check the peek of all the records
	queueLoad.Push(rec2)
	queueLoad.Push(rec3)
	assert.Equal(rec1, queueLoad.Peek())
	queueLoad.Pop()
	assert.Equal(rec2, queueLoad.Peek())
	queueLoad.Pop()
	assert.Equal(rec3, queueLoad.Peek())
	queueLoad.Pop()
	assert.Nil(queueLoad.Peek())

	// Load the queue again to make sure is empty
	queueLoad3, err := LoadQueue(slot1, store)
	require.Nil(err)
	assert.Nil(queueLoad3.Peek())
}
