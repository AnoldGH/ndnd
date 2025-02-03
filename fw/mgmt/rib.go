/* YaNFD - Yet another NDN Forwarding Daemon
 *
 * Copyright (C) 2020-2021 Eric Newberry.
 *
 * This file is licensed under the terms of the MIT License, as found in LICENSE.md.
 */

package mgmt

import (
	"strconv"
	"time"

	"github.com/named-data/ndnd/fw/core"
	"github.com/named-data/ndnd/fw/face"
	"github.com/named-data/ndnd/fw/table"
	enc "github.com/named-data/ndnd/std/encoding"
	mgmt "github.com/named-data/ndnd/std/ndn/mgmt_2022"
	spec "github.com/named-data/ndnd/std/ndn/spec_2022"
)

// RIBModule is the module that handles RIB Management.
type RIBModule struct {
	manager *Thread
}

func (r *RIBModule) String() string {
	return "mgmt-rib"
}

func (r *RIBModule) registerManager(manager *Thread) {
	r.manager = manager
}

func (r *RIBModule) getManager() *Thread {
	return r.manager
}

func (r *RIBModule) handleIncomingInterest(interest *Interest) {
	// Dispatch by verb
	verb := interest.Name()[len(LOCAL_PREFIX)+1].String()
	switch verb {
	case "register":
		r.register(interest)
	case "unregister":
		r.unregister(interest)
	case "announce":
		r.announce(interest)
	case "list":
		r.list(interest)
	default:
		r.manager.sendCtrlResp(interest, 501, "Unknown verb", nil)
		return
	}
}

func (r *RIBModule) register(interest *Interest) {
	if len(interest.Name()) < len(LOCAL_PREFIX)+3 {
		r.manager.sendCtrlResp(interest, 400, "ControlParameters is incorrect", nil)
		return
	}

	params := decodeControlParameters(r, interest)
	if params == nil {
		r.manager.sendCtrlResp(interest, 400, "ControlParameters is incorrect", nil)
		return
	}

	if params.Name == nil {
		r.manager.sendCtrlResp(interest, 400, "ControlParameters is incorrect (missing Name)", nil)
		return
	}

	faceID := interest.inFace.Unwrap()
	if fid, ok := params.FaceId.Get(); ok && fid != 0 {
		faceID = fid
		if face.FaceTable.Get(faceID) == nil {
			r.manager.sendCtrlResp(interest, 410, "Face does not exist", nil)
			return
		}
	}

	origin := params.Origin.GetOr(uint64(mgmt.RouteOriginApp))
	cost := params.Cost.GetOr(uint64(0))
	flags := params.Flags.GetOr(uint64(mgmt.RouteFlagChildInherit))

	expirationPeriod := (*time.Duration)(nil)
	if expiry, ok := params.ExpirationPeriod.Get(); ok {
		expirationPeriod = new(time.Duration)
		*expirationPeriod = time.Duration(expiry) * time.Millisecond
	}

	table.Rib.AddEncRoute(params.Name, &table.Route{
		FaceID:           faceID,
		Origin:           origin,
		Cost:             cost,
		Flags:            flags,
		ExpirationPeriod: expirationPeriod,
	})
	if expirationPeriod != nil {
		core.Log.Info(r, "Created route", "name", params.Name, "faceid", faceID, "origin", origin,
			"cost", cost, "flags", strconv.FormatUint(flags, 16), "expires", expirationPeriod)
	} else {
		core.Log.Info(r, "Created route", "name", params.Name, "faceid", faceID, "origin", origin,
			"cost", cost, "flags", strconv.FormatUint(flags, 16))
	}
	responseParams := &mgmt.ControlArgs{
		Name:   params.Name,
		FaceId: enc.Some(faceID),
		Origin: enc.Some(origin),
		Cost:   enc.Some(cost),
		Flags:  enc.Some(flags),
	}
	if expirationPeriod != nil {
		responseParams.ExpirationPeriod = enc.Some(uint64(expirationPeriod.Milliseconds()))
	}
	r.manager.sendCtrlResp(interest, 200, "OK", responseParams)
}

func (r *RIBModule) unregister(interest *Interest) {
	if len(interest.Name()) < len(LOCAL_PREFIX)+3 {
		r.manager.sendCtrlResp(interest, 400, "ControlParameters is incorrect", nil)
		return
	}

	params := decodeControlParameters(r, interest)
	if params == nil {
		r.manager.sendCtrlResp(interest, 400, "ControlParameters is incorrect", nil)
		return
	}

	if params.Name == nil {
		r.manager.sendCtrlResp(interest, 400, "ControlParameters is incorrect (missing Name)", nil)
		return
	}

	faceID := interest.inFace.Unwrap()
	if fid, ok := params.FaceId.Get(); ok && fid != 0 {
		faceID = fid
	}

	origin := params.Origin.GetOr(uint64(mgmt.RouteOriginApp))
	table.Rib.RemoveRouteEnc(params.Name, faceID, origin)

	r.manager.sendCtrlResp(interest, 200, "OK", &mgmt.ControlArgs{
		Name:   params.Name,
		FaceId: enc.Some(faceID),
		Origin: enc.Some(origin),
	})

	core.Log.Info(r, "Removed route", "name", params.Name, "faceid", faceID, "origin", origin)
}

func (r *RIBModule) announce(interest *Interest) {
	if len(interest.Name()) != len(LOCAL_PREFIX)+3 || interest.Name()[len(LOCAL_PREFIX)+2].Typ != enc.TypeParametersSha256DigestComponent {
		r.manager.sendCtrlResp(interest, 400, "Name is incorrect", nil)
		return
	}

	// Get PrefixAnnouncement
	appParam := interest.AppParam()
	if appParam.Length() == 0 {
		r.manager.sendCtrlResp(interest, 400, "PrefixAnnouncement is missing", nil)
		return
	}

	data, _, err := spec.Spec{}.ReadData(enc.NewFastReader(appParam))
	if err != nil {
		r.manager.sendCtrlResp(interest, 400, "PrefixAnnouncement is invalid", nil)
		return
	}
	if data != nil {
	}

	r.manager.sendCtrlResp(interest, 501, "PrefixAnnouncement not implemented yet", nil)
}

func (r *RIBModule) list(interest *Interest) {
	if len(interest.Name()) > len(LOCAL_PREFIX)+2 {
		// Ignore because contains version and/or segment components
		return
	}

	// Generate new dataset
	entries := table.Rib.GetAllEntries()
	dataset := &mgmt.RibStatus{}
	for _, entry := range entries {
		ribEntry := &mgmt.RibEntry{
			Name:   entry.Name,
			Routes: make([]*mgmt.Route, len(entry.GetRoutes())),
		}
		for i, route := range entry.GetRoutes() {
			ribEntry.Routes[i] = &mgmt.Route{}
			ribEntry.Routes[i].FaceId = route.FaceID
			ribEntry.Routes[i].Origin = route.Origin
			ribEntry.Routes[i].Cost = route.Cost
			ribEntry.Routes[i].Flags = route.Flags
			if route.ExpirationPeriod != nil {
				ribEntry.Routes[i].ExpirationPeriod = enc.Some(uint64(*route.ExpirationPeriod / time.Millisecond))
			}
		}

		dataset.Entries = append(dataset.Entries, ribEntry)
	}

	name := interest.Name()[:len(LOCAL_PREFIX)].
		Append(enc.NewGenericComponent("rib")).
		Append(enc.NewGenericComponent("list"))
	r.manager.sendStatusDataset(interest, name, dataset.Encode())
}
