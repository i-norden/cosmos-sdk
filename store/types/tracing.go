package types

// Operation represents an IO operation
type Operation string

const (
	WriteOp     Operation = "write"
	ReadOp      Operation = "read"
	DeleteOp    Operation = "delete"
	IterKeyOp   Operation = "iterKey"
	IterValueOp Operation = "iterValue"
)

// AllOperations contains all operations
var AllOperations = map[Operation]struct{}{
	WriteOp:     {},
	ReadOp:      {},
	DeleteOp:    {},
	IterKeyOp:   {},
	IterValueOp: {},
}

// TraceOperation implements a traced KVStore operation
type TraceOperation struct {
	Operation Operation              `json:"operation"`
	Key       string                 `json:"key"`
	Value     string                 `json:"value"`
	Metadata  map[string]interface{} `json:"metadata"`
}
