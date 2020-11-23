package types

import (
	"bytes"
	"io"
)

// Operation represents an IO operation
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

// Listener is used to configure listening on specific keys of a KVStore
type Listener struct {
	Writer              io.Writer
	Context             TraceContext
	AllowedOperations   map[Operation]struct{} // The operations which this listener is allowed to listen to
	WhitelistedKeys     [][]byte               // Keys explicitly allowed to be listened to
	BlacklistedKeys     [][]byte               // Keys explicitly disallowed to be listened to
	WhitelistedPrefixes [][]byte               // Key prefixes explicitly allowed to be listened to
	BlacklistedPrefixes [][]byte               // Key prefixes explicitly disallowed to be listened to
}

func (l *Listener) AllowOperation(op Operation) {
	if l.AllowedOperations == nil {
		l.AllowedOperations = make(map[Operation]struct{})
	}
	l.AllowedOperations[op] = struct{}{}
}

func (l *Listener) DisallowOperation(op Operation) {
	delete(l.AllowedOperations, op)
}

func (l *Listener) SetWriter(w io.Writer) {
	l.Writer = w
}

func (l *Listener) SetTraceContext(c TraceContext) {
	l.Context = c
}

func (l *Listener) AddKeyToWhitelist(key []byte) {
	l.WhitelistedKeys = append(l.WhitelistedKeys, key)
}

func (l *Listener) AddPrefixToWhitelist(prefix []byte) {
	l.WhitelistedPrefixes = append(l.WhitelistedPrefixes, prefix)
}

func (l *Listener) AddKeyToBlacklist(key []byte) {
	l.BlacklistedKeys = append(l.BlacklistedKeys, key)
}

func (l *Listener) AddPrefixToBlacklist(prefix []byte) {
	l.BlacklistedPrefixes = append(l.BlacklistedPrefixes, prefix)
}

// Allowed returns whether or not the Listener is allowed to listen to the provided operation at the provided key
func (l *Listener) Allowed(op Operation, key []byte) bool {
	// first check if the operation is allowed
	if _, ok := l.AllowedOperations[op]; !ok {
		return false
	}
	// if there are no keys or prefixes in the whitelists then every key is allowed (unless disallowed in blacklists)
	// if there are whitelisted keys or prefixes then only the keys which conform are allowed (unless disallowed in blacklists)
	allowed := true
	if len(l.WhitelistedKeys) > 0 || len(l.WhitelistedPrefixes) > 0 {
		allowed = listsContains(l.WhitelistedKeys, l.WhitelistedPrefixes, key)
	}
	return allowed && !listsContains(l.BlacklistedKeys, l.BlacklistedPrefixes, key)
}

func listsContains(keys, prefixes [][]byte, key []byte) bool {
	for _, k := range keys {
		if bytes.Equal(key, k) {
			return true
		}
	}
	for _, p := range prefixes {
		if bytes.HasPrefix(key, p) {
			return true
		}
	}
	return false
}
