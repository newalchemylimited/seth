package seth

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"encoding/json"

	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Address) DecodeMsg(dc *msgp.Reader) (err error) {
	err = dc.ReadExactBytes((z)[:])
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Address) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteBytes((z)[:])
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Address) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendBytes(o, (z)[:])
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Address) UnmarshalMsg(bts []byte) (o []byte, err error) {
	bts, err = msgp.ReadExactBytes(bts, (z)[:])
	if err != nil {
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Address) Msgsize() (s int) {
	s = msgp.ArrayHeaderSize + (20 * (msgp.ByteSize))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Block) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Number":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.Number = nil
			} else {
				if z.Number == nil {
					z.Number = new(Int)
				}
				err = z.Number.DecodeMsg(dc)
				if err != nil {
					return
				}
			}
		case "Hash":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.Hash = nil
			} else {
				if z.Hash == nil {
					z.Hash = new(Hash)
				}
				err = dc.ReadExactBytes((*z.Hash)[:])
				if err != nil {
					return
				}
			}
		case "Parent":
			err = dc.ReadExactBytes((z.Parent)[:])
			if err != nil {
				return
			}
		case "Nonce":
			err = z.Nonce.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "UncleHash":
			err = dc.ReadExactBytes((z.UncleHash)[:])
			if err != nil {
				return
			}
		case "Bloom":
			{
				var zb0002 []byte
				zb0002, err = dc.ReadBytes([]byte(z.Bloom))
				if err != nil {
					return
				}
				z.Bloom = Data(zb0002)
			}
		case "TxRoot":
			err = dc.ReadExactBytes((z.TxRoot)[:])
			if err != nil {
				return
			}
		case "StateRoot":
			err = dc.ReadExactBytes((z.StateRoot)[:])
			if err != nil {
				return
			}
		case "ReceiptsRoot":
			err = dc.ReadExactBytes((z.ReceiptsRoot)[:])
			if err != nil {
				return
			}
		case "Miner":
			err = dc.ReadExactBytes((z.Miner)[:])
			if err != nil {
				return
			}
		case "GasLimit":
			err = z.GasLimit.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "GasUsed":
			err = z.GasUsed.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "Transactions":
			var zb0003 uint32
			zb0003, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Transactions) >= int(zb0003) {
				z.Transactions = (z.Transactions)[:zb0003]
			} else {
				z.Transactions = make([]json.RawMessage, zb0003)
			}
			for za0008 := range z.Transactions {
				{
					var zb0004 string
					zb0004, err = dc.ReadString()
					if err != nil {
						return
					}
					z.Transactions[za0008] = json.RawMessage(zb0004)
				}
			}
		case "Uncles":
			var zb0005 uint32
			zb0005, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Uncles) >= int(zb0005) {
				z.Uncles = (z.Uncles)[:zb0005]
			} else {
				z.Uncles = make([]Hash, zb0005)
			}
			for za0009 := range z.Uncles {
				err = dc.ReadExactBytes((z.Uncles[za0009])[:])
				if err != nil {
					return
				}
			}
		case "Difficulty":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.Difficulty = nil
			} else {
				if z.Difficulty == nil {
					z.Difficulty = new(Int)
				}
				err = z.Difficulty.DecodeMsg(dc)
				if err != nil {
					return
				}
			}
		case "TotalDifficulty":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.TotalDifficulty = nil
			} else {
				if z.TotalDifficulty == nil {
					z.TotalDifficulty = new(Int)
				}
				err = z.TotalDifficulty.DecodeMsg(dc)
				if err != nil {
					return
				}
			}
		case "Timestamp":
			err = z.Timestamp.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "Extra":
			{
				var zb0006 []byte
				zb0006, err = dc.ReadBytes([]byte(z.Extra))
				if err != nil {
					return
				}
				z.Extra = Data(zb0006)
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Block) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 18
	// write "Number"
	err = en.Append(0xde, 0x0, 0x12, 0xa6, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72)
	if err != nil {
		return err
	}
	if z.Number == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = z.Number.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	// write "Hash"
	err = en.Append(0xa4, 0x48, 0x61, 0x73, 0x68)
	if err != nil {
		return err
	}
	if z.Hash == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = en.WriteBytes((*z.Hash)[:])
		if err != nil {
			return
		}
	}
	// write "Parent"
	err = en.Append(0xa6, 0x50, 0x61, 0x72, 0x65, 0x6e, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteBytes((z.Parent)[:])
	if err != nil {
		return
	}
	// write "Nonce"
	err = en.Append(0xa5, 0x4e, 0x6f, 0x6e, 0x63, 0x65)
	if err != nil {
		return err
	}
	err = z.Nonce.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "UncleHash"
	err = en.Append(0xa9, 0x55, 0x6e, 0x63, 0x6c, 0x65, 0x48, 0x61, 0x73, 0x68)
	if err != nil {
		return err
	}
	err = en.WriteBytes((z.UncleHash)[:])
	if err != nil {
		return
	}
	// write "Bloom"
	err = en.Append(0xa5, 0x42, 0x6c, 0x6f, 0x6f, 0x6d)
	if err != nil {
		return err
	}
	err = en.WriteBytes([]byte(z.Bloom))
	if err != nil {
		return
	}
	// write "TxRoot"
	err = en.Append(0xa6, 0x54, 0x78, 0x52, 0x6f, 0x6f, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteBytes((z.TxRoot)[:])
	if err != nil {
		return
	}
	// write "StateRoot"
	err = en.Append(0xa9, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x6f, 0x6f, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteBytes((z.StateRoot)[:])
	if err != nil {
		return
	}
	// write "ReceiptsRoot"
	err = en.Append(0xac, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x52, 0x6f, 0x6f, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteBytes((z.ReceiptsRoot)[:])
	if err != nil {
		return
	}
	// write "Miner"
	err = en.Append(0xa5, 0x4d, 0x69, 0x6e, 0x65, 0x72)
	if err != nil {
		return err
	}
	err = en.WriteBytes((z.Miner)[:])
	if err != nil {
		return
	}
	// write "GasLimit"
	err = en.Append(0xa8, 0x47, 0x61, 0x73, 0x4c, 0x69, 0x6d, 0x69, 0x74)
	if err != nil {
		return err
	}
	err = z.GasLimit.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "GasUsed"
	err = en.Append(0xa7, 0x47, 0x61, 0x73, 0x55, 0x73, 0x65, 0x64)
	if err != nil {
		return err
	}
	err = z.GasUsed.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "Transactions"
	err = en.Append(0xac, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.Transactions)))
	if err != nil {
		return
	}
	for za0008 := range z.Transactions {
		err = en.WriteString(string(z.Transactions[za0008]))
		if err != nil {
			return
		}
	}
	// write "Uncles"
	err = en.Append(0xa6, 0x55, 0x6e, 0x63, 0x6c, 0x65, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.Uncles)))
	if err != nil {
		return
	}
	for za0009 := range z.Uncles {
		err = en.WriteBytes((z.Uncles[za0009])[:])
		if err != nil {
			return
		}
	}
	// write "Difficulty"
	err = en.Append(0xaa, 0x44, 0x69, 0x66, 0x66, 0x69, 0x63, 0x75, 0x6c, 0x74, 0x79)
	if err != nil {
		return err
	}
	if z.Difficulty == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = z.Difficulty.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	// write "TotalDifficulty"
	err = en.Append(0xaf, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x44, 0x69, 0x66, 0x66, 0x69, 0x63, 0x75, 0x6c, 0x74, 0x79)
	if err != nil {
		return err
	}
	if z.TotalDifficulty == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = z.TotalDifficulty.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	// write "Timestamp"
	err = en.Append(0xa9, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70)
	if err != nil {
		return err
	}
	err = z.Timestamp.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "Extra"
	err = en.Append(0xa5, 0x45, 0x78, 0x74, 0x72, 0x61)
	if err != nil {
		return err
	}
	err = en.WriteBytes([]byte(z.Extra))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Block) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 18
	// string "Number"
	o = append(o, 0xde, 0x0, 0x12, 0xa6, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72)
	if z.Number == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.Number.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "Hash"
	o = append(o, 0xa4, 0x48, 0x61, 0x73, 0x68)
	if z.Hash == nil {
		o = msgp.AppendNil(o)
	} else {
		o = msgp.AppendBytes(o, (*z.Hash)[:])
	}
	// string "Parent"
	o = append(o, 0xa6, 0x50, 0x61, 0x72, 0x65, 0x6e, 0x74)
	o = msgp.AppendBytes(o, (z.Parent)[:])
	// string "Nonce"
	o = append(o, 0xa5, 0x4e, 0x6f, 0x6e, 0x63, 0x65)
	o, err = z.Nonce.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "UncleHash"
	o = append(o, 0xa9, 0x55, 0x6e, 0x63, 0x6c, 0x65, 0x48, 0x61, 0x73, 0x68)
	o = msgp.AppendBytes(o, (z.UncleHash)[:])
	// string "Bloom"
	o = append(o, 0xa5, 0x42, 0x6c, 0x6f, 0x6f, 0x6d)
	o = msgp.AppendBytes(o, []byte(z.Bloom))
	// string "TxRoot"
	o = append(o, 0xa6, 0x54, 0x78, 0x52, 0x6f, 0x6f, 0x74)
	o = msgp.AppendBytes(o, (z.TxRoot)[:])
	// string "StateRoot"
	o = append(o, 0xa9, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x6f, 0x6f, 0x74)
	o = msgp.AppendBytes(o, (z.StateRoot)[:])
	// string "ReceiptsRoot"
	o = append(o, 0xac, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x52, 0x6f, 0x6f, 0x74)
	o = msgp.AppendBytes(o, (z.ReceiptsRoot)[:])
	// string "Miner"
	o = append(o, 0xa5, 0x4d, 0x69, 0x6e, 0x65, 0x72)
	o = msgp.AppendBytes(o, (z.Miner)[:])
	// string "GasLimit"
	o = append(o, 0xa8, 0x47, 0x61, 0x73, 0x4c, 0x69, 0x6d, 0x69, 0x74)
	o, err = z.GasLimit.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "GasUsed"
	o = append(o, 0xa7, 0x47, 0x61, 0x73, 0x55, 0x73, 0x65, 0x64)
	o, err = z.GasUsed.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Transactions"
	o = append(o, 0xac, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Transactions)))
	for za0008 := range z.Transactions {
		o = msgp.AppendString(o, string(z.Transactions[za0008]))
	}
	// string "Uncles"
	o = append(o, 0xa6, 0x55, 0x6e, 0x63, 0x6c, 0x65, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Uncles)))
	for za0009 := range z.Uncles {
		o = msgp.AppendBytes(o, (z.Uncles[za0009])[:])
	}
	// string "Difficulty"
	o = append(o, 0xaa, 0x44, 0x69, 0x66, 0x66, 0x69, 0x63, 0x75, 0x6c, 0x74, 0x79)
	if z.Difficulty == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.Difficulty.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "TotalDifficulty"
	o = append(o, 0xaf, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x44, 0x69, 0x66, 0x66, 0x69, 0x63, 0x75, 0x6c, 0x74, 0x79)
	if z.TotalDifficulty == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.TotalDifficulty.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "Timestamp"
	o = append(o, 0xa9, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70)
	o, err = z.Timestamp.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Extra"
	o = append(o, 0xa5, 0x45, 0x78, 0x74, 0x72, 0x61)
	o = msgp.AppendBytes(o, []byte(z.Extra))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Block) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Number":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.Number = nil
			} else {
				if z.Number == nil {
					z.Number = new(Int)
				}
				bts, err = z.Number.UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		case "Hash":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.Hash = nil
			} else {
				if z.Hash == nil {
					z.Hash = new(Hash)
				}
				bts, err = msgp.ReadExactBytes(bts, (*z.Hash)[:])
				if err != nil {
					return
				}
			}
		case "Parent":
			bts, err = msgp.ReadExactBytes(bts, (z.Parent)[:])
			if err != nil {
				return
			}
		case "Nonce":
			bts, err = z.Nonce.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "UncleHash":
			bts, err = msgp.ReadExactBytes(bts, (z.UncleHash)[:])
			if err != nil {
				return
			}
		case "Bloom":
			{
				var zb0002 []byte
				zb0002, bts, err = msgp.ReadBytesBytes(bts, []byte(z.Bloom))
				if err != nil {
					return
				}
				z.Bloom = Data(zb0002)
			}
		case "TxRoot":
			bts, err = msgp.ReadExactBytes(bts, (z.TxRoot)[:])
			if err != nil {
				return
			}
		case "StateRoot":
			bts, err = msgp.ReadExactBytes(bts, (z.StateRoot)[:])
			if err != nil {
				return
			}
		case "ReceiptsRoot":
			bts, err = msgp.ReadExactBytes(bts, (z.ReceiptsRoot)[:])
			if err != nil {
				return
			}
		case "Miner":
			bts, err = msgp.ReadExactBytes(bts, (z.Miner)[:])
			if err != nil {
				return
			}
		case "GasLimit":
			bts, err = z.GasLimit.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "GasUsed":
			bts, err = z.GasUsed.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Transactions":
			var zb0003 uint32
			zb0003, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Transactions) >= int(zb0003) {
				z.Transactions = (z.Transactions)[:zb0003]
			} else {
				z.Transactions = make([]json.RawMessage, zb0003)
			}
			for za0008 := range z.Transactions {
				{
					var zb0004 string
					zb0004, bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
					z.Transactions[za0008] = json.RawMessage(zb0004)
				}
			}
		case "Uncles":
			var zb0005 uint32
			zb0005, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Uncles) >= int(zb0005) {
				z.Uncles = (z.Uncles)[:zb0005]
			} else {
				z.Uncles = make([]Hash, zb0005)
			}
			for za0009 := range z.Uncles {
				bts, err = msgp.ReadExactBytes(bts, (z.Uncles[za0009])[:])
				if err != nil {
					return
				}
			}
		case "Difficulty":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.Difficulty = nil
			} else {
				if z.Difficulty == nil {
					z.Difficulty = new(Int)
				}
				bts, err = z.Difficulty.UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		case "TotalDifficulty":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.TotalDifficulty = nil
			} else {
				if z.TotalDifficulty == nil {
					z.TotalDifficulty = new(Int)
				}
				bts, err = z.TotalDifficulty.UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		case "Timestamp":
			bts, err = z.Timestamp.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Extra":
			{
				var zb0006 []byte
				zb0006, bts, err = msgp.ReadBytesBytes(bts, []byte(z.Extra))
				if err != nil {
					return
				}
				z.Extra = Data(zb0006)
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Block) Msgsize() (s int) {
	s = 3 + 7
	if z.Number == nil {
		s += msgp.NilSize
	} else {
		s += z.Number.Msgsize()
	}
	s += 5
	if z.Hash == nil {
		s += msgp.NilSize
	} else {
		s += msgp.ArrayHeaderSize + (32 * (msgp.ByteSize))
	}
	s += 7 + msgp.ArrayHeaderSize + (32 * (msgp.ByteSize)) + 6 + z.Nonce.Msgsize() + 10 + msgp.ArrayHeaderSize + (32 * (msgp.ByteSize)) + 6 + msgp.BytesPrefixSize + len([]byte(z.Bloom)) + 7 + msgp.ArrayHeaderSize + (32 * (msgp.ByteSize)) + 10 + msgp.ArrayHeaderSize + (32 * (msgp.ByteSize)) + 13 + msgp.ArrayHeaderSize + (32 * (msgp.ByteSize)) + 6 + msgp.ArrayHeaderSize + (20 * (msgp.ByteSize)) + 9 + z.GasLimit.Msgsize() + 8 + z.GasUsed.Msgsize() + 13 + msgp.ArrayHeaderSize
	for za0008 := range z.Transactions {
		s += msgp.StringPrefixSize + len(string(z.Transactions[za0008]))
	}
	s += 7 + msgp.ArrayHeaderSize + (len(z.Uncles) * (32 * (msgp.ByteSize))) + 11
	if z.Difficulty == nil {
		s += msgp.NilSize
	} else {
		s += z.Difficulty.Msgsize()
	}
	s += 16
	if z.TotalDifficulty == nil {
		s += msgp.NilSize
	} else {
		s += z.TotalDifficulty.Msgsize()
	}
	s += 10 + z.Timestamp.Msgsize() + 6 + msgp.BytesPrefixSize + len([]byte(z.Extra))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *BlockIterator) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z BlockIterator) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 0
	err = en.Append(0x80)
	if err != nil {
		return err
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z BlockIterator) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 0
	o = append(o, 0x80)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *BlockIterator) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z BlockIterator) Msgsize() (s int) {
	s = 1
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Client) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z Client) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 0
	err = en.Append(0x80)
	if err != nil {
		return err
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Client) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 0
	o = append(o, 0x80)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Client) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z Client) Msgsize() (s int) {
	s = 1
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Data) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zb0001 []byte
		zb0001, err = dc.ReadBytes([]byte((*z)))
		if err != nil {
			return
		}
		(*z) = Data(zb0001)
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z Data) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteBytes([]byte(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Data) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendBytes(o, []byte(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Data) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zb0001 []byte
		zb0001, bts, err = msgp.ReadBytesBytes(bts, []byte((*z)))
		if err != nil {
			return
		}
		(*z) = Data(zb0001)
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z Data) Msgsize() (s int) {
	s = msgp.BytesPrefixSize + len([]byte(z))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Hash) DecodeMsg(dc *msgp.Reader) (err error) {
	err = dc.ReadExactBytes((z)[:])
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Hash) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteBytes((z)[:])
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Hash) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendBytes(o, (z)[:])
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Hash) UnmarshalMsg(bts []byte) (o []byte, err error) {
	bts, err = msgp.ReadExactBytes(bts, (z)[:])
	if err != nil {
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Hash) Msgsize() (s int) {
	s = msgp.ArrayHeaderSize + (32 * (msgp.ByteSize))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Log) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Removed":
			z.Removed, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "LogIndex":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.LogIndex = nil
			} else {
				if z.LogIndex == nil {
					z.LogIndex = new(Int)
				}
				err = z.LogIndex.DecodeMsg(dc)
				if err != nil {
					return
				}
			}
		case "TxIndex":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.TxIndex = nil
			} else {
				if z.TxIndex == nil {
					z.TxIndex = new(Int)
				}
				err = z.TxIndex.DecodeMsg(dc)
				if err != nil {
					return
				}
			}
		case "TxHash":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.TxHash = nil
			} else {
				if z.TxHash == nil {
					z.TxHash = new(Hash)
				}
				err = dc.ReadExactBytes((*z.TxHash)[:])
				if err != nil {
					return
				}
			}
		case "BlockHash":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.BlockHash = nil
			} else {
				if z.BlockHash == nil {
					z.BlockHash = new(Hash)
				}
				err = dc.ReadExactBytes((*z.BlockHash)[:])
				if err != nil {
					return
				}
			}
		case "BlockNumber":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.BlockNumber = nil
			} else {
				if z.BlockNumber == nil {
					z.BlockNumber = new(Int)
				}
				err = z.BlockNumber.DecodeMsg(dc)
				if err != nil {
					return
				}
			}
		case "Address":
			err = dc.ReadExactBytes((z.Address)[:])
			if err != nil {
				return
			}
		case "Data":
			{
				var zb0002 []byte
				zb0002, err = dc.ReadBytes([]byte(z.Data))
				if err != nil {
					return
				}
				z.Data = Data(zb0002)
			}
		case "Topics":
			var zb0003 uint32
			zb0003, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Topics) >= int(zb0003) {
				z.Topics = (z.Topics)[:zb0003]
			} else {
				z.Topics = make([]Data, zb0003)
			}
			for za0004 := range z.Topics {
				{
					var zb0004 []byte
					zb0004, err = dc.ReadBytes([]byte(z.Topics[za0004]))
					if err != nil {
						return
					}
					z.Topics[za0004] = Data(zb0004)
				}
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Log) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 9
	// write "Removed"
	err = en.Append(0x89, 0xa7, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.Removed)
	if err != nil {
		return
	}
	// write "LogIndex"
	err = en.Append(0xa8, 0x4c, 0x6f, 0x67, 0x49, 0x6e, 0x64, 0x65, 0x78)
	if err != nil {
		return err
	}
	if z.LogIndex == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = z.LogIndex.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	// write "TxIndex"
	err = en.Append(0xa7, 0x54, 0x78, 0x49, 0x6e, 0x64, 0x65, 0x78)
	if err != nil {
		return err
	}
	if z.TxIndex == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = z.TxIndex.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	// write "TxHash"
	err = en.Append(0xa6, 0x54, 0x78, 0x48, 0x61, 0x73, 0x68)
	if err != nil {
		return err
	}
	if z.TxHash == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = en.WriteBytes((*z.TxHash)[:])
		if err != nil {
			return
		}
	}
	// write "BlockHash"
	err = en.Append(0xa9, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x61, 0x73, 0x68)
	if err != nil {
		return err
	}
	if z.BlockHash == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = en.WriteBytes((*z.BlockHash)[:])
		if err != nil {
			return
		}
	}
	// write "BlockNumber"
	err = en.Append(0xab, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72)
	if err != nil {
		return err
	}
	if z.BlockNumber == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = z.BlockNumber.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	// write "Address"
	err = en.Append(0xa7, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteBytes((z.Address)[:])
	if err != nil {
		return
	}
	// write "Data"
	err = en.Append(0xa4, 0x44, 0x61, 0x74, 0x61)
	if err != nil {
		return err
	}
	err = en.WriteBytes([]byte(z.Data))
	if err != nil {
		return
	}
	// write "Topics"
	err = en.Append(0xa6, 0x54, 0x6f, 0x70, 0x69, 0x63, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.Topics)))
	if err != nil {
		return
	}
	for za0004 := range z.Topics {
		err = en.WriteBytes([]byte(z.Topics[za0004]))
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Log) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 9
	// string "Removed"
	o = append(o, 0x89, 0xa7, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x64)
	o = msgp.AppendBool(o, z.Removed)
	// string "LogIndex"
	o = append(o, 0xa8, 0x4c, 0x6f, 0x67, 0x49, 0x6e, 0x64, 0x65, 0x78)
	if z.LogIndex == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.LogIndex.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "TxIndex"
	o = append(o, 0xa7, 0x54, 0x78, 0x49, 0x6e, 0x64, 0x65, 0x78)
	if z.TxIndex == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.TxIndex.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "TxHash"
	o = append(o, 0xa6, 0x54, 0x78, 0x48, 0x61, 0x73, 0x68)
	if z.TxHash == nil {
		o = msgp.AppendNil(o)
	} else {
		o = msgp.AppendBytes(o, (*z.TxHash)[:])
	}
	// string "BlockHash"
	o = append(o, 0xa9, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x61, 0x73, 0x68)
	if z.BlockHash == nil {
		o = msgp.AppendNil(o)
	} else {
		o = msgp.AppendBytes(o, (*z.BlockHash)[:])
	}
	// string "BlockNumber"
	o = append(o, 0xab, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72)
	if z.BlockNumber == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.BlockNumber.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "Address"
	o = append(o, 0xa7, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73)
	o = msgp.AppendBytes(o, (z.Address)[:])
	// string "Data"
	o = append(o, 0xa4, 0x44, 0x61, 0x74, 0x61)
	o = msgp.AppendBytes(o, []byte(z.Data))
	// string "Topics"
	o = append(o, 0xa6, 0x54, 0x6f, 0x70, 0x69, 0x63, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Topics)))
	for za0004 := range z.Topics {
		o = msgp.AppendBytes(o, []byte(z.Topics[za0004]))
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Log) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Removed":
			z.Removed, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "LogIndex":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.LogIndex = nil
			} else {
				if z.LogIndex == nil {
					z.LogIndex = new(Int)
				}
				bts, err = z.LogIndex.UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		case "TxIndex":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.TxIndex = nil
			} else {
				if z.TxIndex == nil {
					z.TxIndex = new(Int)
				}
				bts, err = z.TxIndex.UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		case "TxHash":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.TxHash = nil
			} else {
				if z.TxHash == nil {
					z.TxHash = new(Hash)
				}
				bts, err = msgp.ReadExactBytes(bts, (*z.TxHash)[:])
				if err != nil {
					return
				}
			}
		case "BlockHash":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.BlockHash = nil
			} else {
				if z.BlockHash == nil {
					z.BlockHash = new(Hash)
				}
				bts, err = msgp.ReadExactBytes(bts, (*z.BlockHash)[:])
				if err != nil {
					return
				}
			}
		case "BlockNumber":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.BlockNumber = nil
			} else {
				if z.BlockNumber == nil {
					z.BlockNumber = new(Int)
				}
				bts, err = z.BlockNumber.UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		case "Address":
			bts, err = msgp.ReadExactBytes(bts, (z.Address)[:])
			if err != nil {
				return
			}
		case "Data":
			{
				var zb0002 []byte
				zb0002, bts, err = msgp.ReadBytesBytes(bts, []byte(z.Data))
				if err != nil {
					return
				}
				z.Data = Data(zb0002)
			}
		case "Topics":
			var zb0003 uint32
			zb0003, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Topics) >= int(zb0003) {
				z.Topics = (z.Topics)[:zb0003]
			} else {
				z.Topics = make([]Data, zb0003)
			}
			for za0004 := range z.Topics {
				{
					var zb0004 []byte
					zb0004, bts, err = msgp.ReadBytesBytes(bts, []byte(z.Topics[za0004]))
					if err != nil {
						return
					}
					z.Topics[za0004] = Data(zb0004)
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Log) Msgsize() (s int) {
	s = 1 + 8 + msgp.BoolSize + 9
	if z.LogIndex == nil {
		s += msgp.NilSize
	} else {
		s += z.LogIndex.Msgsize()
	}
	s += 8
	if z.TxIndex == nil {
		s += msgp.NilSize
	} else {
		s += z.TxIndex.Msgsize()
	}
	s += 7
	if z.TxHash == nil {
		s += msgp.NilSize
	} else {
		s += msgp.ArrayHeaderSize + (32 * (msgp.ByteSize))
	}
	s += 10
	if z.BlockHash == nil {
		s += msgp.NilSize
	} else {
		s += msgp.ArrayHeaderSize + (32 * (msgp.ByteSize))
	}
	s += 12
	if z.BlockNumber == nil {
		s += msgp.NilSize
	} else {
		s += z.BlockNumber.Msgsize()
	}
	s += 8 + msgp.ArrayHeaderSize + (20 * (msgp.ByteSize)) + 5 + msgp.BytesPrefixSize + len([]byte(z.Data)) + 7 + msgp.ArrayHeaderSize
	for za0004 := range z.Topics {
		s += msgp.BytesPrefixSize + len([]byte(z.Topics[za0004]))
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *RPCError) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Code":
			z.Code, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "Message":
			z.Message, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Data":
			{
				var zb0002 string
				zb0002, err = dc.ReadString()
				if err != nil {
					return
				}
				z.Data = json.RawMessage(zb0002)
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z RPCError) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "Code"
	err = en.Append(0x83, 0xa4, 0x43, 0x6f, 0x64, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.Code)
	if err != nil {
		return
	}
	// write "Message"
	err = en.Append(0xa7, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Message)
	if err != nil {
		return
	}
	// write "Data"
	err = en.Append(0xa4, 0x44, 0x61, 0x74, 0x61)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.Data))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z RPCError) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "Code"
	o = append(o, 0x83, 0xa4, 0x43, 0x6f, 0x64, 0x65)
	o = msgp.AppendInt(o, z.Code)
	// string "Message"
	o = append(o, 0xa7, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65)
	o = msgp.AppendString(o, z.Message)
	// string "Data"
	o = append(o, 0xa4, 0x44, 0x61, 0x74, 0x61)
	o = msgp.AppendString(o, string(z.Data))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RPCError) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Code":
			z.Code, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "Message":
			z.Message, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Data":
			{
				var zb0002 string
				zb0002, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				z.Data = json.RawMessage(zb0002)
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z RPCError) Msgsize() (s int) {
	s = 1 + 5 + msgp.IntSize + 8 + msgp.StringPrefixSize + len(z.Message) + 5 + msgp.StringPrefixSize + len(string(z.Data))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Receipt) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Hash":
			err = dc.ReadExactBytes((z.Hash)[:])
			if err != nil {
				return
			}
		case "Index":
			err = z.Index.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "BlockHash":
			err = dc.ReadExactBytes((z.BlockHash)[:])
			if err != nil {
				return
			}
		case "BlockNumber":
			err = z.BlockNumber.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "GasUsed":
			err = z.GasUsed.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "Cumulative":
			err = z.Cumulative.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "Address":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.Address = nil
			} else {
				if z.Address == nil {
					z.Address = new(Address)
				}
				err = dc.ReadExactBytes((*z.Address)[:])
				if err != nil {
					return
				}
			}
		case "Logs":
			var zb0002 uint32
			zb0002, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Logs) >= int(zb0002) {
				z.Logs = (z.Logs)[:zb0002]
			} else {
				z.Logs = make([]Log, zb0002)
			}
			for za0004 := range z.Logs {
				err = z.Logs[za0004].DecodeMsg(dc)
				if err != nil {
					return
				}
			}
		case "LikelyThrew":
			z.LikelyThrew, err = dc.ReadBool()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Receipt) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 9
	// write "Hash"
	err = en.Append(0x89, 0xa4, 0x48, 0x61, 0x73, 0x68)
	if err != nil {
		return err
	}
	err = en.WriteBytes((z.Hash)[:])
	if err != nil {
		return
	}
	// write "Index"
	err = en.Append(0xa5, 0x49, 0x6e, 0x64, 0x65, 0x78)
	if err != nil {
		return err
	}
	err = z.Index.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "BlockHash"
	err = en.Append(0xa9, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x61, 0x73, 0x68)
	if err != nil {
		return err
	}
	err = en.WriteBytes((z.BlockHash)[:])
	if err != nil {
		return
	}
	// write "BlockNumber"
	err = en.Append(0xab, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72)
	if err != nil {
		return err
	}
	err = z.BlockNumber.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "GasUsed"
	err = en.Append(0xa7, 0x47, 0x61, 0x73, 0x55, 0x73, 0x65, 0x64)
	if err != nil {
		return err
	}
	err = z.GasUsed.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "Cumulative"
	err = en.Append(0xaa, 0x43, 0x75, 0x6d, 0x75, 0x6c, 0x61, 0x74, 0x69, 0x76, 0x65)
	if err != nil {
		return err
	}
	err = z.Cumulative.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "Address"
	err = en.Append(0xa7, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73)
	if err != nil {
		return err
	}
	if z.Address == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = en.WriteBytes((*z.Address)[:])
		if err != nil {
			return
		}
	}
	// write "Logs"
	err = en.Append(0xa4, 0x4c, 0x6f, 0x67, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.Logs)))
	if err != nil {
		return
	}
	for za0004 := range z.Logs {
		err = z.Logs[za0004].EncodeMsg(en)
		if err != nil {
			return
		}
	}
	// write "LikelyThrew"
	err = en.Append(0xab, 0x4c, 0x69, 0x6b, 0x65, 0x6c, 0x79, 0x54, 0x68, 0x72, 0x65, 0x77)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.LikelyThrew)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Receipt) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 9
	// string "Hash"
	o = append(o, 0x89, 0xa4, 0x48, 0x61, 0x73, 0x68)
	o = msgp.AppendBytes(o, (z.Hash)[:])
	// string "Index"
	o = append(o, 0xa5, 0x49, 0x6e, 0x64, 0x65, 0x78)
	o, err = z.Index.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "BlockHash"
	o = append(o, 0xa9, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x61, 0x73, 0x68)
	o = msgp.AppendBytes(o, (z.BlockHash)[:])
	// string "BlockNumber"
	o = append(o, 0xab, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72)
	o, err = z.BlockNumber.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "GasUsed"
	o = append(o, 0xa7, 0x47, 0x61, 0x73, 0x55, 0x73, 0x65, 0x64)
	o, err = z.GasUsed.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Cumulative"
	o = append(o, 0xaa, 0x43, 0x75, 0x6d, 0x75, 0x6c, 0x61, 0x74, 0x69, 0x76, 0x65)
	o, err = z.Cumulative.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Address"
	o = append(o, 0xa7, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73)
	if z.Address == nil {
		o = msgp.AppendNil(o)
	} else {
		o = msgp.AppendBytes(o, (*z.Address)[:])
	}
	// string "Logs"
	o = append(o, 0xa4, 0x4c, 0x6f, 0x67, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Logs)))
	for za0004 := range z.Logs {
		o, err = z.Logs[za0004].MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "LikelyThrew"
	o = append(o, 0xab, 0x4c, 0x69, 0x6b, 0x65, 0x6c, 0x79, 0x54, 0x68, 0x72, 0x65, 0x77)
	o = msgp.AppendBool(o, z.LikelyThrew)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Receipt) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Hash":
			bts, err = msgp.ReadExactBytes(bts, (z.Hash)[:])
			if err != nil {
				return
			}
		case "Index":
			bts, err = z.Index.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "BlockHash":
			bts, err = msgp.ReadExactBytes(bts, (z.BlockHash)[:])
			if err != nil {
				return
			}
		case "BlockNumber":
			bts, err = z.BlockNumber.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "GasUsed":
			bts, err = z.GasUsed.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Cumulative":
			bts, err = z.Cumulative.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Address":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.Address = nil
			} else {
				if z.Address == nil {
					z.Address = new(Address)
				}
				bts, err = msgp.ReadExactBytes(bts, (*z.Address)[:])
				if err != nil {
					return
				}
			}
		case "Logs":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Logs) >= int(zb0002) {
				z.Logs = (z.Logs)[:zb0002]
			} else {
				z.Logs = make([]Log, zb0002)
			}
			for za0004 := range z.Logs {
				bts, err = z.Logs[za0004].UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		case "LikelyThrew":
			z.LikelyThrew, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Receipt) Msgsize() (s int) {
	s = 1 + 5 + msgp.ArrayHeaderSize + (32 * (msgp.ByteSize)) + 6 + z.Index.Msgsize() + 10 + msgp.ArrayHeaderSize + (32 * (msgp.ByteSize)) + 12 + z.BlockNumber.Msgsize() + 8 + z.GasUsed.Msgsize() + 11 + z.Cumulative.Msgsize() + 8
	if z.Address == nil {
		s += msgp.NilSize
	} else {
		s += msgp.ArrayHeaderSize + (20 * (msgp.ByteSize))
	}
	s += 5 + msgp.ArrayHeaderSize
	for za0004 := range z.Logs {
		s += z.Logs[za0004].Msgsize()
	}
	s += 12 + msgp.BoolSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *TokenTransfer) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Block":
			z.Block, err = dc.ReadInt64()
			if err != nil {
				return
			}
		case "TxHeight":
			z.TxHeight, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "Token":
			err = dc.ReadExactBytes((z.Token)[:])
			if err != nil {
				return
			}
		case "From":
			err = dc.ReadExactBytes((z.From)[:])
			if err != nil {
				return
			}
		case "To":
			err = dc.ReadExactBytes((z.To)[:])
			if err != nil {
				return
			}
		case "Amount":
			err = z.Amount.DecodeMsg(dc)
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *TokenTransfer) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 6
	// write "Block"
	err = en.Append(0x86, 0xa5, 0x42, 0x6c, 0x6f, 0x63, 0x6b)
	if err != nil {
		return err
	}
	err = en.WriteInt64(z.Block)
	if err != nil {
		return
	}
	// write "TxHeight"
	err = en.Append(0xa8, 0x54, 0x78, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.TxHeight)
	if err != nil {
		return
	}
	// write "Token"
	err = en.Append(0xa5, 0x54, 0x6f, 0x6b, 0x65, 0x6e)
	if err != nil {
		return err
	}
	err = en.WriteBytes((z.Token)[:])
	if err != nil {
		return
	}
	// write "From"
	err = en.Append(0xa4, 0x46, 0x72, 0x6f, 0x6d)
	if err != nil {
		return err
	}
	err = en.WriteBytes((z.From)[:])
	if err != nil {
		return
	}
	// write "To"
	err = en.Append(0xa2, 0x54, 0x6f)
	if err != nil {
		return err
	}
	err = en.WriteBytes((z.To)[:])
	if err != nil {
		return
	}
	// write "Amount"
	err = en.Append(0xa6, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74)
	if err != nil {
		return err
	}
	err = z.Amount.EncodeMsg(en)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *TokenTransfer) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 6
	// string "Block"
	o = append(o, 0x86, 0xa5, 0x42, 0x6c, 0x6f, 0x63, 0x6b)
	o = msgp.AppendInt64(o, z.Block)
	// string "TxHeight"
	o = append(o, 0xa8, 0x54, 0x78, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74)
	o = msgp.AppendInt(o, z.TxHeight)
	// string "Token"
	o = append(o, 0xa5, 0x54, 0x6f, 0x6b, 0x65, 0x6e)
	o = msgp.AppendBytes(o, (z.Token)[:])
	// string "From"
	o = append(o, 0xa4, 0x46, 0x72, 0x6f, 0x6d)
	o = msgp.AppendBytes(o, (z.From)[:])
	// string "To"
	o = append(o, 0xa2, 0x54, 0x6f)
	o = msgp.AppendBytes(o, (z.To)[:])
	// string "Amount"
	o = append(o, 0xa6, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74)
	o, err = z.Amount.MarshalMsg(o)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *TokenTransfer) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Block":
			z.Block, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return
			}
		case "TxHeight":
			z.TxHeight, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "Token":
			bts, err = msgp.ReadExactBytes(bts, (z.Token)[:])
			if err != nil {
				return
			}
		case "From":
			bts, err = msgp.ReadExactBytes(bts, (z.From)[:])
			if err != nil {
				return
			}
		case "To":
			bts, err = msgp.ReadExactBytes(bts, (z.To)[:])
			if err != nil {
				return
			}
		case "Amount":
			bts, err = z.Amount.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *TokenTransfer) Msgsize() (s int) {
	s = 1 + 6 + msgp.Int64Size + 9 + msgp.IntSize + 6 + msgp.ArrayHeaderSize + (20 * (msgp.ByteSize)) + 5 + msgp.ArrayHeaderSize + (20 * (msgp.ByteSize)) + 3 + msgp.ArrayHeaderSize + (20 * (msgp.ByteSize)) + 7 + z.Amount.Msgsize()
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Transaction) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Hash":
			err = dc.ReadExactBytes((z.Hash)[:])
			if err != nil {
				return
			}
		case "Nonce":
			{
				var zb0002 []byte
				zb0002, err = dc.ReadBytes([]byte(z.Nonce))
				if err != nil {
					return
				}
				z.Nonce = Data(zb0002)
			}
		case "Block":
			err = dc.ReadExactBytes((z.Block)[:])
			if err != nil {
				return
			}
		case "BlockNumber":
			err = z.BlockNumber.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "To":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.To = nil
			} else {
				if z.To == nil {
					z.To = new(Address)
				}
				err = dc.ReadExactBytes((*z.To)[:])
				if err != nil {
					return
				}
			}
		case "TxIndex":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.TxIndex = nil
			} else {
				if z.TxIndex == nil {
					z.TxIndex = new(Int)
				}
				err = z.TxIndex.DecodeMsg(dc)
				if err != nil {
					return
				}
			}
		case "From":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					return
				}
				z.From = nil
			} else {
				if z.From == nil {
					z.From = new(Address)
				}
				err = dc.ReadExactBytes((*z.From)[:])
				if err != nil {
					return
				}
			}
		case "Value":
			err = z.Value.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "GasPrice":
			err = z.GasPrice.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "Gas":
			err = z.Gas.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "Input":
			{
				var zb0003 []byte
				zb0003, err = dc.ReadBytes([]byte(z.Input))
				if err != nil {
					return
				}
				z.Input = Data(zb0003)
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Transaction) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 11
	// write "Hash"
	err = en.Append(0x8b, 0xa4, 0x48, 0x61, 0x73, 0x68)
	if err != nil {
		return err
	}
	err = en.WriteBytes((z.Hash)[:])
	if err != nil {
		return
	}
	// write "Nonce"
	err = en.Append(0xa5, 0x4e, 0x6f, 0x6e, 0x63, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteBytes([]byte(z.Nonce))
	if err != nil {
		return
	}
	// write "Block"
	err = en.Append(0xa5, 0x42, 0x6c, 0x6f, 0x63, 0x6b)
	if err != nil {
		return err
	}
	err = en.WriteBytes((z.Block)[:])
	if err != nil {
		return
	}
	// write "BlockNumber"
	err = en.Append(0xab, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72)
	if err != nil {
		return err
	}
	err = z.BlockNumber.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "To"
	err = en.Append(0xa2, 0x54, 0x6f)
	if err != nil {
		return err
	}
	if z.To == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = en.WriteBytes((*z.To)[:])
		if err != nil {
			return
		}
	}
	// write "TxIndex"
	err = en.Append(0xa7, 0x54, 0x78, 0x49, 0x6e, 0x64, 0x65, 0x78)
	if err != nil {
		return err
	}
	if z.TxIndex == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = z.TxIndex.EncodeMsg(en)
		if err != nil {
			return
		}
	}
	// write "From"
	err = en.Append(0xa4, 0x46, 0x72, 0x6f, 0x6d)
	if err != nil {
		return err
	}
	if z.From == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = en.WriteBytes((*z.From)[:])
		if err != nil {
			return
		}
	}
	// write "Value"
	err = en.Append(0xa5, 0x56, 0x61, 0x6c, 0x75, 0x65)
	if err != nil {
		return err
	}
	err = z.Value.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "GasPrice"
	err = en.Append(0xa8, 0x47, 0x61, 0x73, 0x50, 0x72, 0x69, 0x63, 0x65)
	if err != nil {
		return err
	}
	err = z.GasPrice.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "Gas"
	err = en.Append(0xa3, 0x47, 0x61, 0x73)
	if err != nil {
		return err
	}
	err = z.Gas.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "Input"
	err = en.Append(0xa5, 0x49, 0x6e, 0x70, 0x75, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteBytes([]byte(z.Input))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Transaction) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 11
	// string "Hash"
	o = append(o, 0x8b, 0xa4, 0x48, 0x61, 0x73, 0x68)
	o = msgp.AppendBytes(o, (z.Hash)[:])
	// string "Nonce"
	o = append(o, 0xa5, 0x4e, 0x6f, 0x6e, 0x63, 0x65)
	o = msgp.AppendBytes(o, []byte(z.Nonce))
	// string "Block"
	o = append(o, 0xa5, 0x42, 0x6c, 0x6f, 0x63, 0x6b)
	o = msgp.AppendBytes(o, (z.Block)[:])
	// string "BlockNumber"
	o = append(o, 0xab, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72)
	o, err = z.BlockNumber.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "To"
	o = append(o, 0xa2, 0x54, 0x6f)
	if z.To == nil {
		o = msgp.AppendNil(o)
	} else {
		o = msgp.AppendBytes(o, (*z.To)[:])
	}
	// string "TxIndex"
	o = append(o, 0xa7, 0x54, 0x78, 0x49, 0x6e, 0x64, 0x65, 0x78)
	if z.TxIndex == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.TxIndex.MarshalMsg(o)
		if err != nil {
			return
		}
	}
	// string "From"
	o = append(o, 0xa4, 0x46, 0x72, 0x6f, 0x6d)
	if z.From == nil {
		o = msgp.AppendNil(o)
	} else {
		o = msgp.AppendBytes(o, (*z.From)[:])
	}
	// string "Value"
	o = append(o, 0xa5, 0x56, 0x61, 0x6c, 0x75, 0x65)
	o, err = z.Value.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "GasPrice"
	o = append(o, 0xa8, 0x47, 0x61, 0x73, 0x50, 0x72, 0x69, 0x63, 0x65)
	o, err = z.GasPrice.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Gas"
	o = append(o, 0xa3, 0x47, 0x61, 0x73)
	o, err = z.Gas.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Input"
	o = append(o, 0xa5, 0x49, 0x6e, 0x70, 0x75, 0x74)
	o = msgp.AppendBytes(o, []byte(z.Input))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Transaction) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Hash":
			bts, err = msgp.ReadExactBytes(bts, (z.Hash)[:])
			if err != nil {
				return
			}
		case "Nonce":
			{
				var zb0002 []byte
				zb0002, bts, err = msgp.ReadBytesBytes(bts, []byte(z.Nonce))
				if err != nil {
					return
				}
				z.Nonce = Data(zb0002)
			}
		case "Block":
			bts, err = msgp.ReadExactBytes(bts, (z.Block)[:])
			if err != nil {
				return
			}
		case "BlockNumber":
			bts, err = z.BlockNumber.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "To":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.To = nil
			} else {
				if z.To == nil {
					z.To = new(Address)
				}
				bts, err = msgp.ReadExactBytes(bts, (*z.To)[:])
				if err != nil {
					return
				}
			}
		case "TxIndex":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.TxIndex = nil
			} else {
				if z.TxIndex == nil {
					z.TxIndex = new(Int)
				}
				bts, err = z.TxIndex.UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		case "From":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.From = nil
			} else {
				if z.From == nil {
					z.From = new(Address)
				}
				bts, err = msgp.ReadExactBytes(bts, (*z.From)[:])
				if err != nil {
					return
				}
			}
		case "Value":
			bts, err = z.Value.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "GasPrice":
			bts, err = z.GasPrice.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Gas":
			bts, err = z.Gas.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Input":
			{
				var zb0003 []byte
				zb0003, bts, err = msgp.ReadBytesBytes(bts, []byte(z.Input))
				if err != nil {
					return
				}
				z.Input = Data(zb0003)
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Transaction) Msgsize() (s int) {
	s = 1 + 5 + msgp.ArrayHeaderSize + (32 * (msgp.ByteSize)) + 6 + msgp.BytesPrefixSize + len([]byte(z.Nonce)) + 6 + msgp.ArrayHeaderSize + (32 * (msgp.ByteSize)) + 12 + z.BlockNumber.Msgsize() + 3
	if z.To == nil {
		s += msgp.NilSize
	} else {
		s += msgp.ArrayHeaderSize + (20 * (msgp.ByteSize))
	}
	s += 8
	if z.TxIndex == nil {
		s += msgp.NilSize
	} else {
		s += z.TxIndex.Msgsize()
	}
	s += 5
	if z.From == nil {
		s += msgp.NilSize
	} else {
		s += msgp.ArrayHeaderSize + (20 * (msgp.ByteSize))
	}
	s += 6 + z.Value.Msgsize() + 9 + z.GasPrice.Msgsize() + 4 + z.Gas.Msgsize() + 6 + msgp.BytesPrefixSize + len([]byte(z.Input))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Uint64) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zb0001 uint64
		zb0001, err = dc.ReadUint64()
		if err != nil {
			return
		}
		(*z) = Uint64(zb0001)
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z Uint64) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteUint64(uint64(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Uint64) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendUint64(o, uint64(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Uint64) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zb0001 uint64
		zb0001, bts, err = msgp.ReadUint64Bytes(bts)
		if err != nil {
			return
		}
		(*z) = Uint64(zb0001)
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z Uint64) Msgsize() (s int) {
	s = msgp.Uint64Size
	return
}
