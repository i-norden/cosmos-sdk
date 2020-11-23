package types

import (
	"bytes"
	"io"
)

// operation represents an IO operation
type Operation string

const (
	WriteOp     Operation = "write"
	ReadOp      Operation = "read"
	DeleteOp    Operation = "delete"
	IterKeyOp   Operation = "iterKey"
	IterValueOp Operation = "iterValue"
)

// TraceOperation implements a traced KVStore operation
type TraceOperation struct {
	Operation Operation              `json:"operation"`
	Key       string                 `json:"key"`
	Value     string                 `json:"value"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// Listener struct used to configure listening to specific keys of a KVStore
type Listener struct {
	Writer              io.Writer
	Context             TraceContext
	WhitelistedKeys     [][]byte // Keys explicitly allowed to be listened to
	BlacklistedKeys     [][]byte // Keys explicitly disallowed to be listened to
	WhitelistedPrefixes [][]byte // Key prefixes explicitly allowed to be listened to
	BlacklistedPrefixes [][]byte // Key prefixes explicitly disallowed to be listened to
}

// Allowed returns whether or not the Listener is allowed to listen to the provided key
func (l Listener) Allowed(key []byte) bool {
	// if there are no keys or prefixes in the whitelists then every key is allowed (unless explicity disallowed below)
	allowed := true
	if len(l.WhitelistedKeys) > 0 {
		allowed = byteSliceContains(l.WhitelistedKeys, key, bytes.Equal)
	}
	if len(l.WhitelistedPrefixes) > 0 {
		allowed = byteSliceContains(l.WhitelistedPrefixes, key, bytes.HasPrefix)
	}
	// only keys/prefixes in the blacklists are considered disallowed
	disallowed := byteSliceContains(l.BlacklistedKeys, key, bytes.Equal)
	disallowed = byteSliceContains(l.BlacklistedPrefixes, key, bytes.HasPrefix)
	return allowed && !disallowed
}

// byteSliceContains returns whether or not the provided slice of byte slices contains an element that matches
// the provided key according to the provided matching function
func byteSliceContains(slice [][]byte, key []byte, doesMatch func(key, sliceElement []byte) bool) bool {
	for _, el := range slice {
		if doesMatch(key, el) {
			return true
		}
	}
	return false
}
