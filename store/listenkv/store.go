package listenkv

import (
	"encoding/base64"
	"encoding/json"
	"io"

	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

// Store implements the KVStore interface with listening enabled.
// Operations are traced on each core KVStore call and written to any of the
// underlying io.writer for listeners with the proper key permissions
type Store struct {
	parent    types.KVStore
	listeners []types.Listener
}

// NewStore returns a reference to a new traceKVStore given a parent
// KVStore implementation and a buffered writer.
func NewStore(parent types.KVStore, listeners []types.Listener) *Store {
	return &Store{parent: parent, listeners: listeners}
}

// Get implements the KVStore interface. It traces a read operation and
// delegates a Get call to the parent KVStore.
func (tkv *Store) Get(key []byte) []byte {
	value := tkv.parent.Get(key)

	writeOperation(tkv.listeners, types.ReadOp, key, value)
	return value
}

// Set implements the KVStore interface. It traces a write operation and
// delegates the Set call to the parent KVStore.
func (tkv *Store) Set(key []byte, value []byte) {
	types.AssertValidKey(key)
	writeOperation(tkv.listeners, types.WriteOp, key, value)
	tkv.parent.Set(key, value)
}

// Delete implements the KVStore interface. It traces a write operation and
// delegates the Delete call to the parent KVStore.
func (tkv *Store) Delete(key []byte) {
	writeOperation(tkv.listeners, types.DeleteOp, key, nil)
	tkv.parent.Delete(key)
}

// Has implements the KVStore interface. It delegates the Has call to the
// parent KVStore.
func (tkv *Store) Has(key []byte) bool {
	return tkv.parent.Has(key)
}

// Iterator implements the KVStore interface. It delegates the Iterator call
// the to the parent KVStore.
func (tkv *Store) Iterator(start, end []byte) types.Iterator {
	return tkv.iterator(start, end, true)
}

// ReverseIterator implements the KVStore interface. It delegates the
// ReverseIterator call the to the parent KVStore.
func (tkv *Store) ReverseIterator(start, end []byte) types.Iterator {
	return tkv.iterator(start, end, false)
}

// iterator facilitates iteration over a KVStore. It delegates the necessary
// calls to it's parent KVStore.
func (tkv *Store) iterator(start, end []byte, ascending bool) types.Iterator {
	var parent types.Iterator

	if ascending {
		parent = tkv.parent.Iterator(start, end)
	} else {
		parent = tkv.parent.ReverseIterator(start, end)
	}

	return newTraceIterator(parent, tkv.listeners)
}

type traceIterator struct {
	parent    types.Iterator
	listeners []types.Listener
}

func newTraceIterator(parent types.Iterator, listeners []types.Listener) types.Iterator {
	return &traceIterator{parent: parent, listeners: listeners}
}

// Domain implements the Iterator interface.
func (ti *traceIterator) Domain() (start []byte, end []byte) {
	return ti.parent.Domain()
}

// Valid implements the Iterator interface.
func (ti *traceIterator) Valid() bool {
	return ti.parent.Valid()
}

// Next implements the Iterator interface.
func (ti *traceIterator) Next() {
	ti.parent.Next()
}

// Key implements the Iterator interface.
func (ti *traceIterator) Key() []byte {
	key := ti.parent.Key()

	writeOperation(ti.listeners, types.IterKeyOp, key, nil)
	return key
}

// Value implements the Iterator interface.
func (ti *traceIterator) Value() []byte {
	value := ti.parent.Value()

	writeOperation(ti.listeners, types.IterValueOp, nil, value)
	return value
}

// Close implements the Iterator interface.
func (ti *traceIterator) Close() error {
	return ti.parent.Close()
}

// Error delegates the Error call to the parent iterator.
func (ti *traceIterator) Error() error {
	return ti.parent.Error()
}

// GetStoreType implements the KVStore interface. It returns the underlying
// KVStore type.
func (tkv *Store) GetStoreType() types.StoreType {
	return tkv.parent.GetStoreType()
}

// CacheWrap implements the KVStore interface. It panics as a Store
// cannot be cache wrapped.
func (tkv *Store) CacheWrap() types.CacheWrap {
	panic("cannot CacheWrap a Store")
}

// CacheWrapWithTrace implements the KVStore interface. It panics as a
// Store cannot be cache wrapped.
func (tkv *Store) CacheWrapWithTrace(_ io.Writer, _ types.TraceContext) types.CacheWrap {
	panic("cannot CacheWrapWithTrace a Store")
}

// CacheWrapWithListeners implements the KVStore interface. It panics as a
// Store cannot be cache wrapped.
func (tkv *Store) CacheWrapWithListeners(_ []types.Listener) types.CacheWrap {
	panic("cannot CacheWrapWithListeners a Store")
}

// writeOperation writes a KVStore operation to the underlying io.Writer as
// JSON-encoded data where the key/value pair is base64 encoded.
func writeOperation(listeners []types.Listener, op types.Operation, key, value []byte) {
	traceOp := types.TraceOperation{
		Operation: op,
		Key:       base64.StdEncoding.EncodeToString(key),
		Value:     base64.StdEncoding.EncodeToString(value),
	}
	for _, l := range listeners {
		if !l.Allowed(op, key) {
			continue
		}
		traceOp.Metadata = l.Context
		raw, err := json.Marshal(traceOp)
		if err != nil {
			panic(errors.Wrap(err, "failed to serialize listen operation"))
		}
		if _, err := l.Writer.Write(raw); err != nil {
			panic(errors.Wrap(err, "failed to write listen operation"))
		}
		io.WriteString(l.Writer, "\n")
	}
}
