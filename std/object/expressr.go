package object

import (
	enc "github.com/named-data/ndnd/std/encoding"
	"github.com/named-data/ndnd/std/log"
	"github.com/named-data/ndnd/std/ndn"
	spec "github.com/named-data/ndnd/std/ndn/spec_2022"
	"github.com/named-data/ndnd/std/utils"
)

// Express a single interest with reliability
func (c *Client) ExpressR(args ndn.ExpressRArgs) {
	ExpressR(c.engine, args)
}

// Express a single interest with reliability
func ExpressR(engine ndn.Engine, args ndn.ExpressRArgs) {
	finalizeError := func(err error) {
		args.Callback(ndn.ExpressCallbackArgs{
			Result: ndn.InterestResultError,
			Error:  err,
		})
	}

	// Try local store if available
	if args.TryStore != nil {
		bytes, err := args.TryStore.Get(args.Name, args.Config.CanBePrefix)
		if bytes != nil && err == nil {
			wire := enc.Wire{bytes}
			data, sigCov, err := spec.Spec{}.ReadData(enc.NewWireView(wire))
			if err == nil {
				// Callback should always be called on the engine goroutine
				engine.Post(func() {
					args.Callback(ndn.ExpressCallbackArgs{
						Result:     ndn.InterestResultData,
						Data:       data,
						RawData:    wire,
						SigCovered: sigCov,
						IsLocal:    true,
					})
				})
				return
			}
		}

		// Try the store only once
		args.TryStore = nil
	}

	// New nonce for each transmitted interest
	args.Config.Nonce = utils.ConvertNonce(engine.Timer().Nonce())

	// Create interest packet
	interest, err := engine.Spec().MakeInterest(args.Name, args.Config, args.AppParam, args.Signer)
	if err != nil {
		finalizeError(err)
		return
	}

	// Send the interest
	// TODO: reexpress faster than lifetime
	err = engine.Express(interest, func(res ndn.ExpressCallbackArgs) {
		if res.Result == ndn.InterestResultTimeout {
			log.Debug(nil, "ExpressR Interest timeout", "name", args.Name)

			// Check if retries are exhausted
			if args.Retries == 0 {
				args.Callback(res)
				return
			}

			// Retry on timeout
			args.Retries--
			ExpressR(engine, args)
			return
		} else {
			// All other results / errors are final
			args.Callback(res)
			return
		}
	})
	if err != nil {
		finalizeError(err)
		return
	}
}
