package types

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/cast"

	"github.com/cosmos/cosmos-sdk/server/types"
)

var (
	writeDirPathTOMLBinding = "fileStreamer.writeDirPath"
	newLineBy               = []byte("\n")
)

type Streamer interface {
	// Init initializes the streaming service using the provided AppOptions
	Init(opts types.AppOptions) error
	// AddSource adds a stream source for the path
	AddSource(path string, src Emitter) error
	// AddDestination adds a stream destination for the path
	AddDestination(path string, dst io.WriteCloser) error
	// CreateDestination creates a new stream destination for the path
	CreateDestination(path string) error
	// Stream begins the streaming process, the wait group will block until the streaming has
	// completed due to error or call to Close()
	Stream(wg *sync.WaitGroup) error
	// Close shuts down the streaming service
	Close() error
}

// FileStreamer is a concrete Streamer implementation
// that streams data out by writing out to a file
type FileStreamer struct {
	writeDirPath string
	quitChan     chan struct{}
	destinations map[string][]io.WriteCloser
	sources      map[string][]Emitter
}

// NewStateStreamer returns a new FileStreamer
func NewFileStreamer() Streamer {
	return &FileStreamer{
		quitChan:     make(chan struct{}),
		sources:      make(map[string][]Emitter),
		destinations: make(map[string][]io.WriteCloser),
	}
}

// Init initializes the Streamer using the provided AppOptions
func (fs *FileStreamer) Init(opts types.AppOptions) (err error) {
	fs.writeDirPath, err = cast.ToStringE(opts.Get(writeDirPathTOMLBinding))
	return
}

func (fs *FileStreamer) AddSource(path string, src Emitter) error {
	if src != nil {
		fs.sources[path] = append(fs.sources[path], src)
	}
	return nil
}

// CreateDestination creates a new output destination for the provided path
// the nature of the destination created will depend on the Streamer implementaiton and/or initialization configuration
// in the case of the FileStreamer the provided path is used to create a file we write to
// this path could, for example, instead represent an endpoint URL in the case of a Websocket server
func (fs *FileStreamer) CreateDestination(path string) error {
	f, err := os.OpenFile(filepath.Join(fs.writeDirPath, path), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	fs.destinations[path] = append(fs.destinations[path], f)
	return nil
}

// AddDestination adds a new output destination for the provided path
// this is used to configure/add output destinations outside
func (fs *FileStreamer) AddDestination(path string, dst io.WriteCloser) error {
	if dst != nil {
		fs.destinations[path] = append(fs.destinations[path], dst)
	}
	return nil
}

// Stream begins the streaming process
// the provided wait group will be locked until Close() is called or the stream crashes unexpectedly
func (fs *FileStreamer) Stream(wg *sync.WaitGroup) error {
	for path, dsts := range fs.destinations {
		aggregateChan := make(chan []byte)
		go fs.writeLoop(wg, path, aggregateChan, dsts)
		for _, source := range fs.sources[path] {
			go fs.listenLoop(wg, aggregateChan, source.Output())
		}
	}
	return nil
}

// writeLoop write data off of the aggregating channel to the destination file
func (fs *FileStreamer) writeLoop(wg *sync.WaitGroup, path string, aggregateChan <-chan []byte, dsts []io.WriteCloser) {
	wg.Add(1)
	defer wg.Done()
	for {
		select {
		case by := <-aggregateChan:
			for _, dst := range dsts {
				if _, err := dst.Write(append(by, newLineBy...)); err != nil { // this could be any writer
					log.Printf("error writing to destination: path (%s) err (%v)", path, err)
				}
			}
		case <-fs.quitChan:
			return
		}
	}
}

// listenLoop listens for bytes off a source channel and then forwards it to the provided aggregating channel
func (fs *FileStreamer) listenLoop(wg *sync.WaitGroup, aggregateChan chan<- []byte, source <-chan []byte) {
	wg.Add(1)
	defer wg.Done()
	for {
		select {
		case aggregateChan <- <-source:
		case <-fs.quitChan:
			return
		}
	}
}

// Close is used to shutdown the Streaming
func (fs *FileStreamer) Close() error {
	close(fs.quitChan)
	for path, dsts := range fs.destinations {
		for _, dst := range dsts {
			if err := dst.Close(); err != nil {
				return fmt.Errorf("error closing destination: path (%s) err (%v)", path, err)
			}
		}
	}
	return nil
}
