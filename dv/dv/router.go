package dv

import (
	"sync"
	"time"

	"github.com/pulsejet/go-ndn-dv/config"
	"github.com/pulsejet/go-ndn-dv/nfdc"
	"github.com/pulsejet/go-ndn-dv/table"
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
	basic_engine "github.com/zjkmxy/go-ndn/pkg/engine/basic"
	ndn_sync "github.com/zjkmxy/go-ndn/pkg/engine/sync"
	mgmt "github.com/zjkmxy/go-ndn/pkg/ndn/mgmt_2022"
	"github.com/zjkmxy/go-ndn/pkg/utils"
)

type Router struct {
	// go-ndn app that this router is attached to
	engine *basic_engine.Engine
	// config for this router
	config *config.Config
	// nfd management thread
	nfdc *nfdc.NfdMgmtThread
	// single mutex for all operations
	mutex sync.Mutex

	// channel to stop the DV
	stop chan bool
	// heartbeat for outgoing Advertisements
	heartbeat *time.Ticker
	// deadcheck for neighbors
	deadcheck *time.Ticker

	// neighbor table
	neighbors *table.NeighborTable
	// routing information base
	rib *table.Rib
	// prefix table
	pfx *table.PrefixTable

	// advertisement sequence number for self
	advertSyncSeq uint64
	// prefix table svs instance
	pfxSvs *ndn_sync.SvSync
}

// Create a new DV router.
func NewRouter(config *config.Config, engine *basic_engine.Engine) (*Router, error) {
	// Validate configuration
	err := config.Parse()
	if err != nil {
		return nil, err
	}

	// Create the DV router
	dv := &Router{
		engine: engine,
		config: config,
		nfdc:   nfdc.NewNfdMgmtThread(engine),
		mutex:  sync.Mutex{},
	}

	// Create sync groups
	dv.pfxSvs = ndn_sync.NewSvSync(engine, config.PfxSyncPfxN, dv.onPfxSyncUpdate)

	// Set initial sequence numbers
	now := uint64(time.Now().UnixMilli())
	dv.advertSyncSeq = now
	dv.pfxSvs.SetSeqNo(dv.config.RouterPfxN, now)

	// Create tables
	dv.neighbors = table.NewNeighborTable(config, dv.nfdc)
	dv.rib = table.NewRib(config)
	dv.pfx = table.NewPrefixTable(config, engine, dv.pfxSvs)

	return dv, nil
}

// Start the DV router. Blocks until Stop() is called.
func (dv *Router) Start() (err error) {
	dv.stop = make(chan bool)

	// Start timers
	dv.heartbeat = time.NewTicker(dv.config.AdvertisementSyncInterval)
	dv.deadcheck = time.NewTicker(dv.config.RouterDeadInterval)
	defer dv.heartbeat.Stop()
	defer dv.deadcheck.Stop()

	// Start management thread
	go dv.nfdc.Start()
	defer dv.nfdc.Stop()

	// Configure face
	err = dv.configureFace()
	if err != nil {
		return err
	}

	// Register interest handlers
	err = dv.register()
	if err != nil {
		return err
	}

	// Start sync groups
	dv.pfxSvs.Start()
	defer dv.pfxSvs.Stop()

	// Add self to the RIB
	dv.rib.Set(dv.config.RouterPfxN, dv.config.RouterPfxN, 0)

	for {
		select {
		case <-dv.heartbeat.C:
			dv.advertSyncSendInterest()
			dv.pfxSvs.IncrSeqNo(dv.config.RouterPfxN) // TODO: remove
		case <-dv.deadcheck.C:
			dv.checkDeadNeighbors()
		case <-dv.stop:
			return nil
		}
	}
}

// Stop the DV router.
func (dv *Router) Stop() {
	dv.stop <- true
}

// Configure the face to forwarder.
func (dv *Router) configureFace() (err error) {
	// Enable local fields on face. This includes incoming face indication.
	dv.nfdc.Exec(nfdc.NfdMgmtCmd{
		Module: "faces",
		Cmd:    "update",
		Args: &mgmt.ControlArgs{
			Mask:  utils.IdPtr(uint64(0x01)),
			Flags: utils.IdPtr(uint64(0x01)),
		},
		Retries: -1,
	})

	return nil
}

// Register interest handlers for DV prefixes.
func (dv *Router) register() (err error) {
	// TODO: retry when these fail

	// Advertisement Sync
	err = dv.engine.AttachHandler(dv.config.AdvSyncPfxN, dv.advertSyncOnInterest)
	if err != nil {
		return err
	}

	// Advertisement Data
	err = dv.engine.AttachHandler(dv.config.AdvDataPfxN, dv.advertDataOnInterest)
	if err != nil {
		return err
	}

	// Prefix Data
	err = dv.engine.AttachHandler(dv.config.PfxDataPfxN, dv.pfx.OnDataInterest)
	if err != nil {
		return err
	}

	// Register routes to forwarder
	pfxs := []enc.Name{
		dv.config.AdvSyncPfxN,
		dv.config.AdvDataPfxN,
		dv.config.PfxSyncPfxN,
		dv.config.PfxDataPfxN,
	}
	for _, prefix := range pfxs {
		err = dv.engine.RegisterRoute(prefix)
		if err != nil {
			return err
		}
	}

	// Set strategy to multicast for sync prefixes
	mcast, _ := enc.NameFromStr(config.MulticastStrategy)
	pfxs = []enc.Name{
		dv.config.AdvSyncPfxN,
		dv.config.PfxSyncPfxN,
	}
	for _, prefix := range pfxs {
		dv.nfdc.Exec(nfdc.NfdMgmtCmd{
			Module: "strategy-choice",
			Cmd:    "set",
			Args: &mgmt.ControlArgs{
				Name: prefix,
				Strategy: &mgmt.Strategy{
					Name: mcast,
				},
			},
			Retries: -1,
		})
	}

	return nil
}
