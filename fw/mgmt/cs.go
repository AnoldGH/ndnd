/* YaNFD - Yet another NDN Forwarding Daemon
 *
 * Copyright (C) 2020-2022 Eric Newberry.
 *
 * This file is licensed under the terms of the MIT License, as found in LICENSE.md.
 */

package mgmt

import (
	"github.com/named-data/ndnd/fw/core"
	"github.com/named-data/ndnd/fw/dispatch"
	"github.com/named-data/ndnd/fw/fw"
	"github.com/named-data/ndnd/fw/table"
	enc "github.com/named-data/ndnd/std/encoding"
	mgmt "github.com/named-data/ndnd/std/ndn/mgmt_2022"
	spec "github.com/named-data/ndnd/std/ndn/spec_2022"
)

// ContentStoreModule is the module that handles Content Store Management.
type ContentStoreModule struct {
	manager            *Thread
	nextDatasetVersion uint64
}

func (c *ContentStoreModule) String() string {
	return "ContentStoreMgmt"
}

func (c *ContentStoreModule) registerManager(manager *Thread) {
	c.manager = manager
}

func (c *ContentStoreModule) getManager() *Thread {
	return c.manager
}

func (c *ContentStoreModule) handleIncomingInterest(interest *spec.Interest, pitToken []byte, inFace uint64) {
	// Only allow from /localhost
	if !LOCAL_PREFIX.IsPrefix(interest.Name()) {
		core.LogWarn(c, "Received CS management Interest from non-local source - DROP")
		return
	}

	// Dispatch by verb
	verb := interest.Name()[len(LOCAL_PREFIX)+1].String()
	switch verb {
	case "config":
		c.config(interest, pitToken, inFace)
	case "erase":
		// TODO
		//c.erase(interest, pitToken, inFace)
	case "info":
		c.info(interest, pitToken, inFace)
	default:
		core.LogWarn(c, "Received Interest for non-existent verb '", verb, "'")
		response := makeControlResponse(501, "Unknown verb", nil)
		c.manager.sendResponse(response, interest, pitToken, inFace)
		return
	}
}

func (c *ContentStoreModule) config(interest *spec.Interest, pitToken []byte, inFace uint64) {
	var response *mgmt.ControlResponse

	if len(interest.Name()) < len(LOCAL_PREFIX)+3 {
		// Name not long enough to contain ControlParameters
		core.LogWarn(c, "Missing ControlParameters in ", interest.Name())
		response = makeControlResponse(400, "ControlParameters is incorrect", nil)
		c.manager.sendResponse(response, interest, pitToken, inFace)
		return
	}

	params := decodeControlParameters(c, interest)
	if params == nil {
		response = makeControlResponse(400, "ControlParameters is incorrect", nil)
		c.manager.sendResponse(response, interest, pitToken, inFace)
		return
	}

	if (params.Flags == nil && params.Mask != nil) || (params.Flags != nil && params.Mask == nil) {
		core.LogWarn(c, "Flags and Mask fields must either both be present or both be not present")
		response = makeControlResponse(409, "ControlParameters are incorrect", nil)
		c.manager.sendResponse(response, interest, pitToken, inFace)
		return
	}

	if params.Capacity != nil {
		core.LogInfo(c, "Setting CS capacity to ", *params.Capacity)
		table.SetCsCapacity(int(*params.Capacity))
	}

	if params.Mask != nil && params.Flags != nil {
		if *params.Mask&CsFlagEnableAdmit > 0 {
			val := *params.Flags&CsFlagEnableAdmit > 0
			core.LogInfo(c, "Setting CS admit flag to ", val)
			table.SetCsAdmit(val)
		}

		if *params.Mask&CsFlagEnableServe > 0 {
			val := *params.Flags&CsFlagEnableServe > 0
			core.LogInfo(c, "Setting CS serve flag to ", val)
			table.SetCsServe(val)
		}
	}

	responseParams := map[string]any{
		"Capacity": uint64(table.CsCapacity()),
		"Flags":    c.getFlags(),
	}

	response = makeControlResponse(200, "OK", responseParams)
	c.manager.sendResponse(response, interest, pitToken, inFace)
}

func (c *ContentStoreModule) info(interest *spec.Interest, pitToken []byte, _ uint64) {
	if len(interest.Name()) > len(LOCAL_PREFIX)+2 {
		// Ignore because contains version and/or segment components
		return
	}

	// Generate new dataset
	status := mgmt.CsInfoMsg{
		CsInfo: &mgmt.CsInfo{
			Capacity:   uint64(table.CsCapacity()),
			Flags:      c.getFlags(),
			NCsEntries: 0,
		},
	}
	for threadID := 0; threadID < fw.NumFwThreads; threadID++ {
		thread := dispatch.GetFWThread(threadID)
		status.CsInfo.NCsEntries += uint64(thread.GetNumCsEntries())
	}

	name := append(LOCAL_PREFIX,
		enc.NewStringComponent(enc.TypeGenericNameComponent, "cs"),
		enc.NewStringComponent(enc.TypeGenericNameComponent, "info"),
	)
	segments := makeStatusDataset(name, c.nextDatasetVersion, status.Encode())
	c.manager.transport.Send(segments, pitToken, nil)

	core.LogTrace(c, "Published forwarder status dataset version=", c.nextDatasetVersion,
		", containing ", len(segments), " segments")
	c.nextDatasetVersion++
}

func (c *ContentStoreModule) getFlags() uint64 {
	flags := uint64(0)
	if table.CsAdmit() {
		flags |= CsFlagEnableAdmit
	}
	if table.CsServe() {
		flags |= CsFlagEnableServe
	}
	return flags
}
