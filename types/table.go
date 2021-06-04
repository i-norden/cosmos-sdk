package types

import "github.com/gogo/protobuf/proto"

// TableSchema contains a set of TableInfo
type TableSchema struct {
	Tables []TableInfo
}

// TableInfo contains information for constructing relational tables from protobuf values
type TableInfo struct {
	Name string

	// Type is the protobuf type of this table
	Type proto.Message

	// PrimaryKeyFields are the fields in the protobuf type which compose the primary key
	PrimaryKeyFields []string
}

// TableDecoder contains methods for accessing a TableSchema and decoding a key-value pair into a TableUpdate
type TableDecoder interface {
	// Schema returns the underlying TableSchema
	Schema() TableSchema

	// Decode decodes a key-value pair into a set of TableUpdates.
	// If value is set to nil this indicates that the key-value pair was deleted from storage.
	Decode(key, value []byte) ([]TableUpdate, error)
}

// TableUpdate contains the updates to a table caused by a key-value pair
type TableUpdate struct {
	Table string

	// true for update mode indicating that this is a non-overwriting patch update, false for replace mode indicating that this update replaces a whole row
	// this is needed to distinguish what to do with "zero" values, whether we want to overwrite with the zero values or if they are zero because we only want to update the non-zero values
	UpdateOrReplace bool

	// in update mode Updated contains only changed fields and ClearedFields indicates fields to be cleared, in replace mode the value of Updated replaces the whole table row
	Updated proto.Message

	// ClearedFields contains a list of fields to clear and should only be set in replace mode
	ClearedFields []string
}
