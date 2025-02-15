// Code generated by ndn tlv codegen DO NOT EDIT.
package svs_ps

import (
	"io"

	enc "github.com/named-data/ndnd/std/encoding"
)

type HistorySnapEncoder struct {
	length uint

	wirePlan []uint

	Entries_subencoder []struct {
		Entries_encoder HistorySnapEntryEncoder
	}
}

type HistorySnapParsingContext struct {
	Entries_context HistorySnapEntryParsingContext
}

func (encoder *HistorySnapEncoder) Init(value *HistorySnap) {
	{
		Entries_l := len(value.Entries)
		encoder.Entries_subencoder = make([]struct {
			Entries_encoder HistorySnapEntryEncoder
		}, Entries_l)
		for i := 0; i < Entries_l; i++ {
			pseudoEncoder := &encoder.Entries_subencoder[i]
			pseudoValue := struct {
				Entries *HistorySnapEntry
			}{
				Entries: value.Entries[i],
			}
			{
				encoder := pseudoEncoder
				value := &pseudoValue
				if value.Entries != nil {
					encoder.Entries_encoder.Init(value.Entries)
				}
				_ = encoder
				_ = value
			}
		}
	}

	l := uint(0)
	if value.Entries != nil {
		for seq_i, seq_v := range value.Entries {
			pseudoEncoder := &encoder.Entries_subencoder[seq_i]
			pseudoValue := struct {
				Entries *HistorySnapEntry
			}{
				Entries: seq_v,
			}
			{
				encoder := pseudoEncoder
				value := &pseudoValue
				if value.Entries != nil {
					l += 1
					l += uint(enc.TLNum(encoder.Entries_encoder.length).EncodingLength())
					l += encoder.Entries_encoder.length
				}
				_ = encoder
				_ = value
			}
		}
	}
	encoder.length = l

	wirePlan := make([]uint, 0, 8)
	l = uint(0)
	if value.Entries != nil {
		for seq_i, seq_v := range value.Entries {
			pseudoEncoder := &encoder.Entries_subencoder[seq_i]
			pseudoValue := struct {
				Entries *HistorySnapEntry
			}{
				Entries: seq_v,
			}
			{
				encoder := pseudoEncoder
				value := &pseudoValue
				if value.Entries != nil {
					l += 1
					l += uint(enc.TLNum(encoder.Entries_encoder.length).EncodingLength())
					if encoder.Entries_encoder.length > 0 {
						l += encoder.Entries_encoder.wirePlan[0]
						for i := 1; i < len(encoder.Entries_encoder.wirePlan); i++ {
							wirePlan = append(wirePlan, l)
							l = 0
							l = encoder.Entries_encoder.wirePlan[i]
						}
						if l == 0 {
							wirePlan = append(wirePlan, l)
							l = 0
						}
					}
				}
				_ = encoder
				_ = value
			}
		}
	}
	if l > 0 {
		wirePlan = append(wirePlan, l)
	}
	encoder.wirePlan = wirePlan
}

func (context *HistorySnapParsingContext) Init() {
	context.Entries_context.Init()
}

func (encoder *HistorySnapEncoder) EncodeInto(value *HistorySnap, wire enc.Wire) {

	wireIdx := 0
	buf := wire[wireIdx]

	pos := uint(0)

	if value.Entries != nil {
		for seq_i, seq_v := range value.Entries {
			pseudoEncoder := &encoder.Entries_subencoder[seq_i]
			pseudoValue := struct {
				Entries *HistorySnapEntry
			}{
				Entries: seq_v,
			}
			{
				encoder := pseudoEncoder
				value := &pseudoValue
				if value.Entries != nil {
					buf[pos] = byte(130)
					pos += 1
					pos += uint(enc.TLNum(encoder.Entries_encoder.length).EncodeInto(buf[pos:]))
					if encoder.Entries_encoder.length > 0 {
						{
							subWire := make(enc.Wire, len(encoder.Entries_encoder.wirePlan))
							subWire[0] = buf[pos:]
							for i := 1; i < len(subWire); i++ {
								subWire[i] = wire[wireIdx+i]
							}
							encoder.Entries_encoder.EncodeInto(value.Entries, subWire)
							for i := 1; i < len(subWire); i++ {
								wire[wireIdx+i] = subWire[i]
							}
							if lastL := encoder.Entries_encoder.wirePlan[len(subWire)-1]; lastL > 0 {
								wireIdx += len(subWire) - 1
								if len(subWire) > 1 {
									pos = lastL
								} else {
									pos += lastL
								}
							} else {
								wireIdx += len(subWire)
								pos = 0
							}
							if wireIdx < len(wire) {
								buf = wire[wireIdx]
							} else {
								buf = nil
							}
						}
					}
				}
				_ = encoder
				_ = value
			}
		}
	}
}

func (encoder *HistorySnapEncoder) Encode(value *HistorySnap) enc.Wire {
	total := uint(0)
	for _, l := range encoder.wirePlan {
		total += l
	}
	content := make([]byte, total)

	wire := make(enc.Wire, len(encoder.wirePlan))
	for i, l := range encoder.wirePlan {
		if l > 0 {
			wire[i] = content[:l]
			content = content[l:]
		}
	}
	encoder.EncodeInto(value, wire)

	return wire
}

func (context *HistorySnapParsingContext) Parse(reader enc.WireView, ignoreCritical bool) (*HistorySnap, error) {

	var handled_Entries bool = false

	progress := -1
	_ = progress

	value := &HistorySnap{}
	var err error
	var startPos int
	for {
		startPos = reader.Pos()
		if startPos >= reader.Length() {
			break
		}
		typ := enc.TLNum(0)
		l := enc.TLNum(0)
		typ, err = reader.ReadTLNum()
		if err != nil {
			return nil, enc.ErrFailToParse{TypeNum: 0, Err: err}
		}
		l, err = reader.ReadTLNum()
		if err != nil {
			return nil, enc.ErrFailToParse{TypeNum: 0, Err: err}
		}

		err = nil
		if handled := false; true {
			switch typ {
			case 130:
				if true {
					handled = true
					handled_Entries = true
					if value.Entries == nil {
						value.Entries = make([]*HistorySnapEntry, 0)
					}
					{
						pseudoValue := struct {
							Entries *HistorySnapEntry
						}{}
						{
							value := &pseudoValue
							value.Entries, err = context.Entries_context.Parse(reader.Delegate(int(l)), ignoreCritical)
							_ = value
						}
						value.Entries = append(value.Entries, pseudoValue.Entries)
					}
					progress--
				}
			default:
				if !ignoreCritical && ((typ <= 31) || ((typ & 1) == 1)) {
					return nil, enc.ErrUnrecognizedField{TypeNum: typ}
				}
				handled = true
				err = reader.Skip(int(l))
			}
			if err == nil && !handled {
			}
			if err != nil {
				return nil, enc.ErrFailToParse{TypeNum: typ, Err: err}
			}
		}
	}

	startPos = reader.Pos()
	err = nil

	if !handled_Entries && err == nil {
		// sequence - skip
	}

	if err != nil {
		return nil, err
	}

	return value, nil
}

func (value *HistorySnap) Encode() enc.Wire {
	encoder := HistorySnapEncoder{}
	encoder.Init(value)
	return encoder.Encode(value)
}

func (value *HistorySnap) Bytes() []byte {
	return value.Encode().Join()
}

func ParseHistorySnap(reader enc.WireView, ignoreCritical bool) (*HistorySnap, error) {
	context := HistorySnapParsingContext{}
	context.Init()
	return context.Parse(reader, ignoreCritical)
}

type HistorySnapEntryEncoder struct {
	length uint

	wirePlan []uint

	Content_length uint
}

type HistorySnapEntryParsingContext struct {
}

func (encoder *HistorySnapEntryEncoder) Init(value *HistorySnapEntry) {

	if value.Content != nil {
		encoder.Content_length = 0
		for _, c := range value.Content {
			encoder.Content_length += uint(len(c))
		}
	}

	l := uint(0)
	l += 1
	l += uint(1 + enc.Nat(value.SeqNo).EncodingLength())
	if value.Content != nil {
		l += 1
		l += uint(enc.TLNum(encoder.Content_length).EncodingLength())
		l += encoder.Content_length
	}
	encoder.length = l

	wirePlan := make([]uint, 0, 8)
	l = uint(0)
	l += 1
	l += uint(1 + enc.Nat(value.SeqNo).EncodingLength())
	if value.Content != nil {
		l += 1
		l += uint(enc.TLNum(encoder.Content_length).EncodingLength())
		wirePlan = append(wirePlan, l)
		l = 0
		for range value.Content {
			wirePlan = append(wirePlan, l)
			l = 0
		}
	}
	if l > 0 {
		wirePlan = append(wirePlan, l)
	}
	encoder.wirePlan = wirePlan
}

func (context *HistorySnapEntryParsingContext) Init() {

}

func (encoder *HistorySnapEntryEncoder) EncodeInto(value *HistorySnapEntry, wire enc.Wire) {

	wireIdx := 0
	buf := wire[wireIdx]

	pos := uint(0)

	buf[pos] = byte(214)
	pos += 1

	buf[pos] = byte(enc.Nat(value.SeqNo).EncodeInto(buf[pos+1:]))
	pos += uint(1 + buf[pos])
	if value.Content != nil {
		buf[pos] = byte(131)
		pos += 1
		pos += uint(enc.TLNum(encoder.Content_length).EncodeInto(buf[pos:]))
		wireIdx++
		pos = 0
		if wireIdx < len(wire) {
			buf = wire[wireIdx]
		} else {
			buf = nil
		}
		for _, w := range value.Content {
			wire[wireIdx] = w
			wireIdx++
			pos = 0
			if wireIdx < len(wire) {
				buf = wire[wireIdx]
			} else {
				buf = nil
			}
		}
	}
}

func (encoder *HistorySnapEntryEncoder) Encode(value *HistorySnapEntry) enc.Wire {
	total := uint(0)
	for _, l := range encoder.wirePlan {
		total += l
	}
	content := make([]byte, total)

	wire := make(enc.Wire, len(encoder.wirePlan))
	for i, l := range encoder.wirePlan {
		if l > 0 {
			wire[i] = content[:l]
			content = content[l:]
		}
	}
	encoder.EncodeInto(value, wire)

	return wire
}

func (context *HistorySnapEntryParsingContext) Parse(reader enc.WireView, ignoreCritical bool) (*HistorySnapEntry, error) {

	var handled_SeqNo bool = false
	var handled_Content bool = false

	progress := -1
	_ = progress

	value := &HistorySnapEntry{}
	var err error
	var startPos int
	for {
		startPos = reader.Pos()
		if startPos >= reader.Length() {
			break
		}
		typ := enc.TLNum(0)
		l := enc.TLNum(0)
		typ, err = reader.ReadTLNum()
		if err != nil {
			return nil, enc.ErrFailToParse{TypeNum: 0, Err: err}
		}
		l, err = reader.ReadTLNum()
		if err != nil {
			return nil, enc.ErrFailToParse{TypeNum: 0, Err: err}
		}

		err = nil
		if handled := false; true {
			switch typ {
			case 214:
				if true {
					handled = true
					handled_SeqNo = true
					value.SeqNo = uint64(0)
					{
						for i := 0; i < int(l); i++ {
							x := byte(0)
							x, err = reader.ReadByte()
							if err != nil {
								if err == io.EOF {
									err = io.ErrUnexpectedEOF
								}
								break
							}
							value.SeqNo = uint64(value.SeqNo<<8) | uint64(x)
						}
					}
				}
			case 131:
				if true {
					handled = true
					handled_Content = true
					value.Content, err = reader.ReadWire(int(l))
				}
			default:
				if !ignoreCritical && ((typ <= 31) || ((typ & 1) == 1)) {
					return nil, enc.ErrUnrecognizedField{TypeNum: typ}
				}
				handled = true
				err = reader.Skip(int(l))
			}
			if err == nil && !handled {
			}
			if err != nil {
				return nil, enc.ErrFailToParse{TypeNum: typ, Err: err}
			}
		}
	}

	startPos = reader.Pos()
	err = nil

	if !handled_SeqNo && err == nil {
		err = enc.ErrSkipRequired{Name: "SeqNo", TypeNum: 214}
	}
	if !handled_Content && err == nil {
		value.Content = nil
	}

	if err != nil {
		return nil, err
	}

	return value, nil
}

func (value *HistorySnapEntry) Encode() enc.Wire {
	encoder := HistorySnapEntryEncoder{}
	encoder.Init(value)
	return encoder.Encode(value)
}

func (value *HistorySnapEntry) Bytes() []byte {
	return value.Encode().Join()
}

func ParseHistorySnapEntry(reader enc.WireView, ignoreCritical bool) (*HistorySnapEntry, error) {
	context := HistorySnapEntryParsingContext{}
	context.Init()
	return context.Parse(reader, ignoreCritical)
}

type HistoryIndexEncoder struct {
	length uint

	wirePlan []uint

	SeqNos_subencoder []struct {
	}
}

type HistoryIndexParsingContext struct {
}

func (encoder *HistoryIndexEncoder) Init(value *HistoryIndex) {
	{
		SeqNos_l := len(value.SeqNos)
		encoder.SeqNos_subencoder = make([]struct {
		}, SeqNos_l)
		for i := 0; i < SeqNos_l; i++ {
			pseudoEncoder := &encoder.SeqNos_subencoder[i]
			pseudoValue := struct {
				SeqNos uint64
			}{
				SeqNos: value.SeqNos[i],
			}
			{
				encoder := pseudoEncoder
				value := &pseudoValue

				_ = encoder
				_ = value
			}
		}
	}

	l := uint(0)
	if value.SeqNos != nil {
		for seq_i, seq_v := range value.SeqNos {
			pseudoEncoder := &encoder.SeqNos_subencoder[seq_i]
			pseudoValue := struct {
				SeqNos uint64
			}{
				SeqNos: seq_v,
			}
			{
				encoder := pseudoEncoder
				value := &pseudoValue
				l += 1
				l += uint(1 + enc.Nat(value.SeqNos).EncodingLength())
				_ = encoder
				_ = value
			}
		}
	}
	encoder.length = l

	wirePlan := make([]uint, 0, 8)
	l = uint(0)
	if value.SeqNos != nil {
		for seq_i, seq_v := range value.SeqNos {
			pseudoEncoder := &encoder.SeqNos_subencoder[seq_i]
			pseudoValue := struct {
				SeqNos uint64
			}{
				SeqNos: seq_v,
			}
			{
				encoder := pseudoEncoder
				value := &pseudoValue
				l += 1
				l += uint(1 + enc.Nat(value.SeqNos).EncodingLength())
				_ = encoder
				_ = value
			}
		}
	}
	if l > 0 {
		wirePlan = append(wirePlan, l)
	}
	encoder.wirePlan = wirePlan
}

func (context *HistoryIndexParsingContext) Init() {

}

func (encoder *HistoryIndexEncoder) EncodeInto(value *HistoryIndex, wire enc.Wire) {

	wireIdx := 0
	buf := wire[wireIdx]

	pos := uint(0)

	if value.SeqNos != nil {
		for seq_i, seq_v := range value.SeqNos {
			pseudoEncoder := &encoder.SeqNos_subencoder[seq_i]
			pseudoValue := struct {
				SeqNos uint64
			}{
				SeqNos: seq_v,
			}
			{
				encoder := pseudoEncoder
				value := &pseudoValue
				buf[pos] = byte(132)
				pos += 1

				buf[pos] = byte(enc.Nat(value.SeqNos).EncodeInto(buf[pos+1:]))
				pos += uint(1 + buf[pos])
				_ = encoder
				_ = value
			}
		}
	}
}

func (encoder *HistoryIndexEncoder) Encode(value *HistoryIndex) enc.Wire {
	total := uint(0)
	for _, l := range encoder.wirePlan {
		total += l
	}
	content := make([]byte, total)

	wire := make(enc.Wire, len(encoder.wirePlan))
	for i, l := range encoder.wirePlan {
		if l > 0 {
			wire[i] = content[:l]
			content = content[l:]
		}
	}
	encoder.EncodeInto(value, wire)

	return wire
}

func (context *HistoryIndexParsingContext) Parse(reader enc.WireView, ignoreCritical bool) (*HistoryIndex, error) {

	var handled_SeqNos bool = false

	progress := -1
	_ = progress

	value := &HistoryIndex{}
	var err error
	var startPos int
	for {
		startPos = reader.Pos()
		if startPos >= reader.Length() {
			break
		}
		typ := enc.TLNum(0)
		l := enc.TLNum(0)
		typ, err = reader.ReadTLNum()
		if err != nil {
			return nil, enc.ErrFailToParse{TypeNum: 0, Err: err}
		}
		l, err = reader.ReadTLNum()
		if err != nil {
			return nil, enc.ErrFailToParse{TypeNum: 0, Err: err}
		}

		err = nil
		if handled := false; true {
			switch typ {
			case 132:
				if true {
					handled = true
					handled_SeqNos = true
					if value.SeqNos == nil {
						value.SeqNos = make([]uint64, 0)
					}
					{
						pseudoValue := struct {
							SeqNos uint64
						}{}
						{
							value := &pseudoValue
							value.SeqNos = uint64(0)
							{
								for i := 0; i < int(l); i++ {
									x := byte(0)
									x, err = reader.ReadByte()
									if err != nil {
										if err == io.EOF {
											err = io.ErrUnexpectedEOF
										}
										break
									}
									value.SeqNos = uint64(value.SeqNos<<8) | uint64(x)
								}
							}
							_ = value
						}
						value.SeqNos = append(value.SeqNos, pseudoValue.SeqNos)
					}
					progress--
				}
			default:
				if !ignoreCritical && ((typ <= 31) || ((typ & 1) == 1)) {
					return nil, enc.ErrUnrecognizedField{TypeNum: typ}
				}
				handled = true
				err = reader.Skip(int(l))
			}
			if err == nil && !handled {
			}
			if err != nil {
				return nil, enc.ErrFailToParse{TypeNum: typ, Err: err}
			}
		}
	}

	startPos = reader.Pos()
	err = nil

	if !handled_SeqNos && err == nil {
		// sequence - skip
	}

	if err != nil {
		return nil, err
	}

	return value, nil
}

func (value *HistoryIndex) Encode() enc.Wire {
	encoder := HistoryIndexEncoder{}
	encoder.Init(value)
	return encoder.Encode(value)
}

func (value *HistoryIndex) Bytes() []byte {
	return value.Encode().Join()
}

func ParseHistoryIndex(reader enc.WireView, ignoreCritical bool) (*HistoryIndex, error) {
	context := HistoryIndexParsingContext{}
	context.Init()
	return context.Parse(reader, ignoreCritical)
}
