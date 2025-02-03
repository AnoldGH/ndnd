//go:generate gondn_tlv_gen
package defn

import (
	"time"

	enc "github.com/named-data/ndnd/std/encoding"
)

// +tlv-model:nocopy
type FwPacket struct {
	//+field:struct:FwInterest:nocopy
	Interest *FwInterest `tlv:"0x05"`
	//+field:struct:FwData:nocopy
	Data *FwData `tlv:"0x06"`
	//+field:struct:FwLpPacket:nocopy
	LpPacket *FwLpPacket `tlv:"0x64"`
}

// +tlv-model:nocopy
type FwInterest struct {
	//+field:name
	NameV enc.Name `tlv:"0x07"`
	//+field:bool
	CanBePrefixV bool `tlv:"0x21"`
	//+field:bool
	MustBeFreshV bool `tlv:"0x12"`
	//+field:struct:FwLinks
	ForwardingHintV *FwLinks `tlv:"0x1e"`
	//+field:fixedUint:uint32:optional
	NonceV enc.Optional[uint32] `tlv:"0x0a"`
	//+field:time:optional
	InterestLifetimeV enc.Optional[time.Duration] `tlv:"0x0c"`
	//+field:byte
	HopLimitV *byte `tlv:"0x22"`

	//+field:bool
	ApplicationParameters bool `tlv:"0x24"`
	//+field:bool
	SignatureInfo bool `tlv:"0x2c"`
	//+field:bool
	SignatureValue bool `tlv:"0x2e"`
}

type FwLinks struct {
	//+field:sequence:enc.Name:name
	Names []enc.Name `tlv:"0x07"`
}

// +tlv-model:nocopy
type FwData struct {
	//+field:name
	NameV enc.Name `tlv:"0x07"`
	//+field:struct:FwMetaInfo
	MetaInfo *FwMetaInfo `tlv:"0x14"`

	//+field:bool
	ContentV bool `tlv:"0x15"`
	//+field:bool
	SignatureInfo bool `tlv:"0x16"`
	//+field:bool
	SignatureValue bool `tlv:"0x17"`
}

type FwMetaInfo struct {
	//+field:natural:optional
	ContentType enc.Optional[uint64] `tlv:"0x18"`
	//+field:time:optional
	FreshnessPeriod enc.Optional[time.Duration] `tlv:"0x19"`

	//+field:bool
	FinalBlockID bool `tlv:"0x1a"`
}

// +tlv-model:nocopy
type FwLpPacket struct {
	//+field:fixedUint:uint64:optional
	Sequence enc.Optional[uint64] `tlv:"0x51"`
	//+field:natural:optional
	FragIndex enc.Optional[uint64] `tlv:"0x52"`
	//+field:natural:optional
	FragCount enc.Optional[uint64] `tlv:"0x53"`
	//+field:binary
	PitToken []byte `tlv:"0x62"`
	//+field:struct:FwNetworkNack
	Nack *FwNetworkNack `tlv:"0x0320"`
	//+field:natural:optional
	IncomingFaceId enc.Optional[uint64] `tlv:"0x032C"`
	//+field:natural:optional
	NextHopFaceId enc.Optional[uint64] `tlv:"0x0330"`
	//+field:struct:FwCachePolicy
	CachePolicy *FwCachePolicy `tlv:"0x0334"`
	//+field:natural:optional
	CongestionMark enc.Optional[uint64] `tlv:"0x0340"`

	//+field:wire
	Fragment enc.Wire `tlv:"0x50"`
}

type FwNetworkNack struct {
	//+field:natural
	Reason uint64 `tlv:"0x0321"`
}

type FwCachePolicy struct {
	//+field:natural
	CachePolicyType uint64 `tlv:"0x0335"`
}

func (p *FwInterest) Name() enc.Name {
	return p.NameV
}

func (p *FwInterest) Lifetime() enc.Optional[time.Duration] {
	return p.InterestLifetimeV
}
