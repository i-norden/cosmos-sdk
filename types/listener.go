package types

import (
	"errors"
	"io"
)

// Emitter interface
type Emitter interface {
	Output() <-chan []byte
}

// Transformer interface
type Transformer interface {
	AddTransformation(func(b []byte) []byte)
}

// Listener interface
type Listener interface {
	io.Writer
	Emitter
}

// TransformingListener interface
type TransformingListener interface {
	Listener
	Transformer
}

// StateListener is a concrete TransformingListener implementation
type StateListener struct {
	c chan []byte
}

// NewStateListener returns a new Listener
func NewStateListener() Listener {
	return &StateListener{
		c: make(chan []byte),
	}
}

// Write satisfies the io.Write interface
// it writes the bytes to the Output channel
// after applying any transformations on the bytes
func (sl *StateListener) Write(b []byte) (int, error) {
	select {
	case sl.c <- b:
		return len(b), nil
	default:
		return 0, errors.New("unable to write bytes to channel; no listener available")
	}
}

// Output satisfies the Emitter interface
// it returns the channel being written to in Write()
func (sl *StateListener) Output() <-chan []byte {
	return sl.c
}

// TransformingStateListener is a concrete TransformingListener implementation
type TransformingStateListener struct {
	c chan []byte
	transformers []func(b []byte) []byte
}

// NewTransformingStateListener returns a new Listener
func NewTransformingStateListener() TransformingListener {
	return &TransformingStateListener{
		c: make(chan []byte),
	}
}

// Write satisfies the io.Write interface
// it writes a copy of the bytes to the Output channel
// after applying any transformations on the bytes
func (sl *TransformingStateListener) Write(b []byte) (int, error) {
	// copy the passed byte slice so that we do not modify the original data
	bb := make([]byte, len(b))
	copy(bb, b)
	for _, t := range sl.transformers {
		bb = t(bb)
	}
	select {
	case sl.c <- bb:
		return len(bb), nil
	default:
		return 0, errors.New("unable to write bytes to channel; no listener available")
	}
}

// Output satisfies the Emitter interface
// it returns the channel being written to in Write()
func (sl *TransformingStateListener) Output() <-chan []byte {
	return sl.c
}

// AddTransformation adds a transformation function which will be applied to
// any bytes this Writer Writes
// Transformations are applied in the order they were added
func (sl *TransformingStateListener) AddTransformation(t func(b []byte) []byte) {
	if t != nil {
		sl.transformers = append(sl.transformers, t)
	}
}
