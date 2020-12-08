package types

import (
	"encoding/binary"
	"io"
	"math"
)

// WriteListener interface for streaming data out from a listenkv.Store
type WriteListener interface {
	// if value is nil then it was deleted
	OnWrite(key []byte, value []byte)
}

// PrefixWriteListener is used to configure listening to a KVStore by writing out big endian length-prefixed
// key-value pairs to an io.Writer
type PrefixWriteListener struct {
	writer io.Writer
	prefixBuf [6]byte
}

// NewPrefixWriteListener wraps a PrefixWriteListener around an io.Writer
func NewPrefixWriteListener(w io.Writer) *PrefixWriteListener {
	return &PrefixWriteListener{
		writer: w,
	}
}

// OnWrite satisfies the WriteListener interface by writing out big endian length-prefixed key-value pairs
// to an underlying io.Writer
// The first two bytes of the prefix encode the length of the key
// The last four bytes of the prefix encode the length of the value
// This WriteListener makes two assumptions
// 1) The key is no longer than 1<<16 - 1
// 2) The value is no longer than 1<<32 - 1
func(swl *PrefixWriteListener) OnWrite(key []byte, value []byte) {
	keyLen := len(key)
	valLen := len(key)
	if keyLen > math.MaxUint16 || valLen > math.MaxUint32 {
		return
	}
	binary.BigEndian.PutUint16(swl.prefixBuf[:2], uint16(keyLen))
	binary.BigEndian.PutUint32(swl.prefixBuf[2:], uint32(valLen))
	swl.writer.Write(swl.prefixBuf[:])
	swl.writer.Write(key)
	swl.writer.Write(value)
}

// NewlineWriteListener is used to configure listening to a KVStore by writing out big endian key-length-prefixed and
// newline delineated key-value pairs to an io.Writer
type NewlineWriteListener struct {
	writer              io.Writer
	keyLenBuf           [2]byte
}

// NewNewlineWriteListener wraps a StockWriteListener around an io.Writer
func NewNewlineWriteListener(w io.Writer) *NewlineWriteListener {
	return &NewlineWriteListener{
		writer: w,
	}
}

var newline = []byte("\n")

// OnWrite satisfies WriteListener interface by writing out newline delineated big endian key-length-prefixed key-value
// pairs to the underlying io.Writer
// The first two bytes encode the length of the key
// Separate key-value pairs are newline delineated
// This WriteListener makes three assumptions
// 1) The key is no longer than 1<<16 - 1
// 2) The value and keys contain no newline characters
func (nwl *NewlineWriteListener) OnWrite(key []byte, value []byte) {
	keyLen := len(key)
	if keyLen > math.MaxUint16 {
		return
	}
	binary.BigEndian.PutUint16(nwl.keyLenBuf[:], uint16(keyLen))
	nwl.writer.Write(nwl.keyLenBuf[:])
	nwl.writer.Write(key)
	nwl.writer.Write(value)
	nwl.writer.Write(newline)
}