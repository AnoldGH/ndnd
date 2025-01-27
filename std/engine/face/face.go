package face

import enc "github.com/named-data/ndnd/std/encoding"

type Face interface {
	Open() error
	Close() error
	Send(pkt enc.Wire) error
	IsRunning() bool
	IsLocal() bool
	SetCallback(onPkt func(frame []byte) error,
		onError func(err error) error)
}
