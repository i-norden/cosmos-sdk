package baseapp

import (
	"fmt"
	"strings"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// StreamingServiceConstructor is used to construct and load a WriteListener onto the provided BaseApp and expose it over a streaming service
type StreamingServiceConstructor func(bApp *BaseApp, opts servertypes.AppOptions, keys []sdk.StoreKey) error

// StreamingServiceType enum for specifying the type of StreamingService
type StreamingServiceType int

const (
	Unknown StreamingServiceType = iota
	File
	GRPC
)

// NewStreamingServiceType returns the StreamingServiceType corresponding to the provided name
func NewStreamingServiceType(name string) StreamingServiceType {
	switch strings.ToLower(name) {
	case "file", "f":
		return File
	case "grpc":
		return GRPC
	default:
		return Unknown
	}
}

// String returns the string name of a StreamingServiceType
func (sst StreamingServiceType) String() string {
	switch sst {
	case File:
		return "file"
	case GRPC:
		return "grpc"
	default:
		return ""
	}
}

// StreamingServiceConstructorLookupTable is a mapping of StreamingServiceTypes to StreamingServiceConstructors
var StreamingServiceConstructorLookupTable = map[StreamingServiceType]StreamingServiceConstructor{
	File: FileStreamingConstructor,
	GRPC: GRPCStreamingConstructor,
}

// NewStreamingServiceConstructor returns the StreamingServiceConstructor corresponding to the provided name
func NewStreamingServiceConstructor(name string) (StreamingServiceConstructor, error) {
	ssType := NewStreamingServiceType(name)
	if ssType == Unknown {
		return nil, fmt.Errorf("unrecognized streaming service name %s", name)
	}
	if constructor, ok := StreamingServiceConstructorLookupTable[ssType]; ok {
		return constructor, nil
	}
	return nil, fmt.Errorf("streaming service constructor of type %s not found", ssType.String())
}

// FileStreamingConstructor is the StreamingServiceConstructor function for writing out to a file
func FileStreamingConstructor(bApp *BaseApp, opts servertypes.AppOptions, keys []sdk.StoreKey) error {
	// TODO: implement me
	panic("implement me")
}

// GRPCStreamingConstructor is the StreamingServiceConstructor function for writing out to a gRPC stream
func GRPCStreamingConstructor(bApp *BaseApp, opts servertypes.AppOptions, keys []sdk.StoreKey) error {
	// TODO: implement me
	panic("implement me")
}