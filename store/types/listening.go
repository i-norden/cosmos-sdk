package types

import (
	"bytes"
	"io"
)

// Listener is used to configure listening on specific keys of a KVStore
type Listener struct {
	writer              io.Writer
	context             TraceContext
	allowedOperations   map[Operation]struct{} // The operations which this listener is allowed to listen to
	whitelistedKeys     [][]byte               // Keys explicitly allowed to be listened to
	blacklistedKeys     [][]byte               // Keys explicitly disallowed to be listened to
	whitelistedPrefixes [][]byte               // Key prefixes explicitly allowed to be listened to
	blacklistedPrefixes [][]byte               // Key prefixes explicitly disallowed to be listened to
}

// NewDefaultStateListener returns a Listener using the provided io.Writer and TraceContext
// it allows listening to all operations and all keys (empty whitelists and blacklists)
func NewDefaultStateListener(w io.Writer, c TraceContext) *Listener {
	return &Listener{
		writer:            w,
		context:           c,
		allowedOperations: AllOperations,
	}
}

// NewStateListener returns a Listener built using the provided params
func NewStateListener(w io.Writer, c TraceContext, ops map[Operation]struct{},
	whitelistedKeys, whitelistedPrefixes, blacklistedKeys, blacklistedPrefixes [][]byte) *Listener {
	return &Listener{
		writer:              w,
		context:             c,
		allowedOperations:   ops,
		whitelistedKeys:     whitelistedKeys,
		whitelistedPrefixes: whitelistedPrefixes,
		blacklistedKeys:     blacklistedKeys,
		blacklistedPrefixes: blacklistedPrefixes,
	}
}

// AllowOperation adds an operation to the allowed list
func (l *Listener) AllowOperation(op Operation) Listening {
	if l.allowedOperations == nil {
		l.allowedOperations = make(map[Operation]struct{})
	}

	l.allowedOperations[op] = struct{}{}
	return l
}

// DisallowOperation removes an operation from the allowed list
func (l *Listener) DisallowOperation(op Operation) {
	delete(l.allowedOperations, op)
}

// SetWriter sets the underlying io.Writer for this Listener
func (l *Listener) SetWriter(w io.Writer) {
	l.writer = w
}

// SetTraceContext sets the TraceContext for this Listener
func (l *Listener) SetTraceContext(c TraceContext) {
	l.context = c
}

// AddKeyToWhitelist adds a key to whitelist
// if not key or prefix is whitelisted, then all keys are considered whitelisted
// otherwise if any key or prefix is whitelisted then only those keys will be listened to
func (l *Listener) AddKeyToWhitelist(key []byte) {
	l.whitelistedKeys = append(l.whitelistedKeys, key)
}

// RemoveKeyFromWhitelist removes a key from the whitelist
func (l *Listener) RemoveKeyFromWhitelist(key []byte) {
	removeFromSlice(l.whitelistedKeys, key)
}

// AddPrefixToWhitelist adds a prefix to whitelist
func (l *Listener) AddPrefixToWhitelist(prefix []byte) {
	l.whitelistedPrefixes = append(l.whitelistedPrefixes, prefix)
}

// RemovePrefixFromWhitelist removes a prefix from the whitelist
func (l *Listener) RemovePrefixFromWhitelist(prefix []byte) {
	removeFromSlice(l.whitelistedPrefixes, prefix)
}

// AddKeyToBlacklist adds a key to blacklist
// blacklisted keys cannot be listened to
func (l *Listener) AddKeyToBlacklist(key []byte) {
	l.blacklistedKeys = append(l.blacklistedKeys, key)
}

// RemoveKeyFromBlacklist removes a key from the blacklist
func (l *Listener) RemoveKeyFromBlacklist(key []byte) {
	removeFromSlice(l.blacklistedKeys, key)
}

// AddPrefixToBlacklist adds a prefix to blacklist
// keys with the blacklisted prefix cannot be listened to
func (l *Listener) AddPrefixToBlacklist(prefix []byte) {
	l.blacklistedPrefixes = append(l.blacklistedPrefixes, prefix)
}

// RemovePrefixFromBlacklist removes a prefix from the whitelist
func (l *Listener) RemovePrefixFromBlacklist(prefix []byte) {
	removeFromSlice(l.blacklistedPrefixes, prefix)
}

// GetContext satisfies Listening interface
func (l *Listener) GetContext() TraceContext {
	return l.context
}

// Write satisfies Listening interface
// it wraps around the underlying writer interface
func (l *Listener) Write(b []byte) (int, error) {
	return l.writer.Write(b)
}

// Allowed satisfies Listening interface
// it returns whether or not the Listener is allowed to listen to the provided operation at the provided key
func (l *Listener) Allowed(op Operation, key []byte) bool {
	// first check if the operation is allowed
	if _, ok := l.allowedOperations[op]; !ok {
		return false
	}
	// if there are no keys or prefixes in the whitelists then every key is allowed (unless disallowed in blacklists)
	// if there are whitelisted keys or prefixes then only the keys which conform are allowed (unless disallowed in blacklists)
	allowed := true
	if len(l.whitelistedKeys) > 0 || len(l.whitelistedPrefixes) > 0 {
		allowed = listsContain(l.whitelistedKeys, l.whitelistedPrefixes, key)
	}
	return allowed && !listsContain(l.blacklistedKeys, l.blacklistedPrefixes, key)
}

func listsContain(keys, prefixes [][]byte, key []byte) bool {
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

func removeFromSlice(s [][]byte, rem []byte) [][]byte {
	res := make([][]byte, 0, len(s))
	for _, el := range s {
		if bytes.Equal(el, rem) {
			continue
		}
		res = append(res, el)
	}
	return res
}
