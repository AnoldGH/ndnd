package object

import (
	"fmt"
	"sync/atomic"
	"time"

	enc "github.com/named-data/ndnd/std/encoding"
	"github.com/named-data/ndnd/std/log"
	"github.com/named-data/ndnd/std/ndn"
	rdr "github.com/named-data/ndnd/std/ndn/rdr_2024"
	"github.com/named-data/ndnd/std/types/optional"
)

// maximum number of segments in an object (for safety)
const maxObjectSeg = 1e8

// arguments for the consume callback
type ConsumeState struct {
	// original arguments
	args ndn.ConsumeExtArgs
	// error that occurred during fetching
	err error
	// raw data contents.
	content enc.Wire
	// fetching is completed
	complete atomic.Bool
	// fetched metadata
	meta *rdr.MetaData
	// versioned object name
	fetchName enc.Name

	// fetching window
	// - [0] is the position till which the user has already consumed the fetched buffer
	// - [1] is the position till which the buffer is valid (window start)
	// - [2] is the end of the current fetching window
	//
	// content[0:wnd[0]] is invalid (already used and freed)
	// content[wnd[0]:wnd[1]] is valid (not used yet)
	// content[wnd[1]:wnd[2]] is currently being fetched
	// content[wnd[2]:] will be fetched in the future
	wnd [3]int

	// segment count from final block id (-1 if unknown)
	segCnt int
}

// returns the name of the object being consumed
func (a *ConsumeState) Name() enc.Name {
	return a.fetchName
}

// returns the version of the object being consumed
func (a *ConsumeState) Version() uint64 {
	if ver := a.fetchName.At(-1); ver.IsVersion() {
		return ver.NumberVal()
	}
	return 0
}

// returns the error that occurred during fetching
func (a *ConsumeState) Error() error {
	return a.err
}

// returns true if the content has been completely fetched
func (a *ConsumeState) IsComplete() bool {
	return a.complete.Load()
}

// returns the currently available buffer in the content
// any subsequent calls to Content() will return data after the previous call
func (a *ConsumeState) Content() enc.Wire {
	// return valid range of buffer (can be empty)
	wire := make(enc.Wire, a.wnd[1]-a.wnd[0])

	// free buffers
	for i := a.wnd[0]; i < a.wnd[1]; i++ {
		wire[i-a.wnd[0]] = a.content[i] // retain
		a.content[i] = nil              // gc
	}

	a.wnd[0] = a.wnd[1]
	return wire
}

// get the progress counter
func (a *ConsumeState) Progress() int {
	return a.wnd[1]
}

// get the max value for the progress counter (-1 for unknown)
func (a *ConsumeState) ProgressMax() int {
	return a.segCnt
}

// cancel the consume operation
func (a *ConsumeState) Cancel() {
	if !a.complete.Swap(true) {
		a.err = ndn.ErrCancelled
	}
}

// send a fatal error to the callback
func (a *ConsumeState) finalizeError(err error) {
	if !a.complete.Swap(true) {
		a.err = err
		a.args.Callback(a)
	}
}

// Consume an object with a given name
func (c *Client) Consume(name enc.Name, callback func(status ndn.ConsumeState)) {
	c.ConsumeExt(ndn.ConsumeExtArgs{Name: name, Callback: callback})
}

// ConsumeExt is a more advanced consume API that allows for more control
// over the fetching process.
func (c *Client) ConsumeExt(args ndn.ConsumeExtArgs) {
	// clone the name for good measure
	args.Name = args.Name.Clone()

	// create new consume state
	c.consumeObject(&ConsumeState{
		args:      args,
		err:       nil,
		content:   make(enc.Wire, 0), // just in case
		complete:  atomic.Bool{},
		meta:      nil,
		fetchName: args.Name,
		wnd:       [3]int{0, 0},
		segCnt:    -1,
	})
}

func (c *Client) consumeObject(state *ConsumeState) {
	name := state.fetchName

	// will segfault if name is empty
	if len(name) == 0 {
		state.finalizeError(fmt.Errorf("%w: consume name cannot be empty", ndn.ErrProtocol))
		return
	}

	// fetch object metadata if the last name component is not a version
	if !name.At(-1).IsVersion() {
		// when called with metadata, call with versioned name.
		// state will always have the original object name.
		if state.meta != nil {
			state.finalizeError(fmt.Errorf("%w: metadata does not have version component", ndn.ErrProtocol))
			return
		}

		// if metadata fetching is disabled, just attempt to fetch one segment
		// with the prefix, then get the versioned name from the segment.
		if state.args.NoMetadata {
			c.fetchDataByPrefix(name, func(data ndn.Data, err error) {
				if err != nil {
					state.finalizeError(err)
					return
				}
				meta, err := extractSegMetadata(data)
				if err != nil {
					state.finalizeError(err)
					return
				}
				c.consumeObjectWithMeta(state, meta)
			})
			return
		}

		// fetch RDR metadata for this object
		c.fetchMetadata(name, func(meta *rdr.MetaData, err error) {
			if err != nil {
				state.finalizeError(err)
				return
			}
			c.consumeObjectWithMeta(state, meta)
		})
		return
	}

	// passes ownership of state and callback to fetcher
	c.fetcher.add(state)
}

// consumeObjectWithMeta consumes an object with a given metadata
func (c *Client) consumeObjectWithMeta(state *ConsumeState, meta *rdr.MetaData) {
	state.meta = meta
	state.fetchName = meta.Name
	c.consumeObject(state)
}

// fetchMetadata gets the RDR metadata for an object with a given name
func (c *Client) fetchMetadata(
	name enc.Name,
	callback func(meta *rdr.MetaData, err error),
) {
	log.Debug(c, "Fetching object metadata", "name", name)
	c.ExpressR(ndn.ExpressRArgs{
		Name: name.Append(enc.NewKeywordComponent(rdr.MetadataKeyword)),
		Config: &ndn.InterestConfig{
			CanBePrefix: true,
			MustBeFresh: true,
			Lifetime:    optional.Some(time.Millisecond * 1000),
		},
		Retries: 3,
		Callback: func(args ndn.ExpressCallbackArgs) {
			if args.Result == ndn.InterestResultError {
				callback(nil, fmt.Errorf("%w: fetch metadata failed: %w", ndn.ErrNetwork, args.Error))
				return
			}

			if args.Result != ndn.InterestResultData {
				callback(nil, fmt.Errorf("%w: fetch metadata failed with result: %s", ndn.ErrNetwork, args.Result))
				return
			}

			c.Validate(args.Data, args.SigCovered, func(valid bool, err error) {
				// validate with trust config
				if !valid {
					callback(nil, fmt.Errorf("%w: validate metadata failed: %w", ndn.ErrSecurity, err))
					return
				}

				// parse metadata
				metadata, err := rdr.ParseMetaData(enc.NewWireView(args.Data.Content()), false)
				if err != nil {
					callback(nil, fmt.Errorf("%w: failed to parse object metadata: %w", ndn.ErrProtocol, err))
					return
				}

				// clone fields for lifetime
				metadata.Name = metadata.Name.Clone()
				metadata.FinalBlockID = append([]byte{}, metadata.FinalBlockID...)
				callback(metadata, nil)
			})
		},
	})
}

// fetchWithPrefix gets any fresh data with a given prefix
func (c *Client) fetchDataByPrefix(
	name enc.Name,
	callback func(data ndn.Data, err error),
) {
	log.Debug(c, "Fetching data with prefix", "name", name)
	c.ExpressR(ndn.ExpressRArgs{
		Name: name,
		Config: &ndn.InterestConfig{
			CanBePrefix: true,
			MustBeFresh: true,
			Lifetime:    optional.Some(time.Millisecond * 1000),
		},
		Retries: 3,
		Callback: func(args ndn.ExpressCallbackArgs) {
			if args.Result == ndn.InterestResultError {
				callback(nil, fmt.Errorf("%w: fetch by prefix failed: %w", ndn.ErrNetwork, args.Error))
				return
			}

			if args.Result != ndn.InterestResultData {
				callback(nil, fmt.Errorf("%w: fetch by prefix failed with result: %s", ndn.ErrNetwork, args.Result))
				return
			}

			c.Validate(args.Data, args.SigCovered, func(valid bool, err error) {
				if !valid {
					callback(nil, fmt.Errorf("%w: validate by prefix failed: %w", ndn.ErrSecurity, err))
					return
				}

				callback(args.Data, nil)
			})
		},
	})
}

// extractSegMetadata constructs partial metadata from a given data segment
// returns (metadata, error)
func extractSegMetadata(data ndn.Data) (*rdr.MetaData, error) {
	// check if the object has segment and version components
	name := data.Name()
	if len(name) < 2 {
		return nil, fmt.Errorf("%w: data has no version or segment", ndn.ErrProtocol)
	}

	// get segment component
	if !name.At(-1).IsSegment() {
		return nil, fmt.Errorf("%w: data has no segment", ndn.ErrProtocol)
	}

	// get version component
	if !name.At(-2).IsVersion() {
		return nil, fmt.Errorf("%w: data has no version", ndn.ErrProtocol)
	}

	// construct metadata
	return &rdr.MetaData{Name: name.Prefix(-1)}, nil
}
