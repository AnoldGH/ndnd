package dv

import (
	"time"

	"github.com/named-data/ndnd/dv/tlv"
	enc "github.com/named-data/ndnd/std/encoding"
	"github.com/named-data/ndnd/std/log"
	"github.com/named-data/ndnd/std/object"
	"github.com/named-data/ndnd/std/utils"
)

func (dv *Router) advertGenerateNew() {
	dv.mutex.Lock()
	defer dv.mutex.Unlock()

	// Increment sequence number
	dv.advertSyncSeq++

	// Produce the advertisement
	name, err := dv.client.Produce(object.ProduceArgs{
		Name: dv.config.AdvertisementDataPrefix().Append(
			enc.NewTimestampComponent(dv.advertBootTime),
		),
		Content:         dv.rib.Advert().Encode(),
		Version:         utils.IdPtr(dv.advertSyncSeq),
		FreshnessPeriod: 10 * time.Second,
	})
	if err != nil {
		log.Errorf("advert-data: failed to produce advertisement: %+v", err)
	}
	dv.advertDir.Push(name)
	dv.advertDir.Evict(dv.client)

	// Notify neighbors with sync for new advertisement
	go dv.advertSyncSendInterest()
}

func (dv *Router) advertDataFetch(nName enc.Name, bootTime uint64, seqNo uint64) {
	// debounce; wait before fetching, then check if this is still the latest
	// sequence number known for this neighbor
	time.Sleep(10 * time.Millisecond)
	if ns := dv.neighbors.Get(nName); ns == nil || ns.AdvertBoot != bootTime || ns.AdvertSeq != seqNo {
		return
	}

	// Fetch the advertisement
	advName := enc.LOCALHOP.Append(nName.Append(
		enc.NewStringComponent(enc.TypeKeywordNameComponent, "DV"),
		enc.NewStringComponent(enc.TypeKeywordNameComponent, "ADV"),
		enc.NewTimestampComponent(bootTime),
		enc.NewVersionComponent(seqNo),
	)...)

	dv.client.Consume(advName, func(state *object.ConsumeState) bool {
		if !state.IsComplete() {
			return true
		}

		go func() {
			fetchErr := state.Error()
			if fetchErr != nil {
				log.Warnf("advert-data: failed to fetch advertisement %s: %+v", state.Name(), fetchErr)
				time.Sleep(1 * time.Second) // wait on error
				dv.advertDataFetch(nName, bootTime, seqNo)
				return
			}

			// Process the advertisement
			dv.advertDataHandler(nName, seqNo, state.Content())
		}()

		return true
	})
}

// Received advertisement Data
func (dv *Router) advertDataHandler(nName enc.Name, seqNo uint64, data []byte) {
	// Lock DV state
	dv.mutex.Lock()
	defer dv.mutex.Unlock()

	// Check if this is the latest advertisement
	ns := dv.neighbors.Get(nName)
	if ns == nil {
		log.Warnf("advert-data: unknown advertisement %s", nName)
		return
	}
	if ns.AdvertSeq != seqNo {
		log.Debugf("advert-data: old advertisement for %s (%d != %d)", nName, ns.AdvertSeq, seqNo)
		return
	}

	// Parse the advertisement
	advert, err := tlv.ParseAdvertisement(enc.NewBufferReader(data), false)
	if err != nil {
		log.Errorf("advert-data: failed to parse advertisement: %+v", err)
		return
	}

	// Update the local advertisement list
	ns.Advert = advert
	go dv.ribUpdate(ns)
}
