package seth

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

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
			err = z.Hash.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "Nonce":
			err = z.Nonce.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "Block":
			err = z.Block.DecodeMsg(dc)
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
				err = z.To.DecodeMsg(dc)
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
					z.TxIndex = new(Uint64)
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
				err = z.From.DecodeMsg(dc)
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
			err = z.Input.DecodeMsg(dc)
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
func (z *Transaction) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 11
	// write "Hash"
	err = en.Append(0x8b, 0xa4, 0x48, 0x61, 0x73, 0x68)
	if err != nil {
		return err
	}
	err = z.Hash.EncodeMsg(en)
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
	// write "Block"
	err = en.Append(0xa5, 0x42, 0x6c, 0x6f, 0x63, 0x6b)
	if err != nil {
		return err
	}
	err = z.Block.EncodeMsg(en)
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
		err = z.To.EncodeMsg(en)
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
		err = z.From.EncodeMsg(en)
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
	err = z.Input.EncodeMsg(en)
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
	o, err = z.Hash.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Nonce"
	o = append(o, 0xa5, 0x4e, 0x6f, 0x6e, 0x63, 0x65)
	o, err = z.Nonce.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Block"
	o = append(o, 0xa5, 0x42, 0x6c, 0x6f, 0x63, 0x6b)
	o, err = z.Block.MarshalMsg(o)
	if err != nil {
		return
	}
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
		o, err = z.To.MarshalMsg(o)
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
	// string "From"
	o = append(o, 0xa4, 0x46, 0x72, 0x6f, 0x6d)
	if z.From == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.From.MarshalMsg(o)
		if err != nil {
			return
		}
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
	o, err = z.Input.MarshalMsg(o)
	if err != nil {
		return
	}
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
			bts, err = z.Hash.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Nonce":
			bts, err = z.Nonce.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Block":
			bts, err = z.Block.UnmarshalMsg(bts)
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
				bts, err = z.To.UnmarshalMsg(bts)
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
					z.TxIndex = new(Uint64)
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
				bts, err = z.From.UnmarshalMsg(bts)
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
			bts, err = z.Input.UnmarshalMsg(bts)
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
func (z *Transaction) Msgsize() (s int) {
	s = 1 + 5 + z.Hash.Msgsize() + 6 + z.Nonce.Msgsize() + 6 + z.Block.Msgsize() + 12 + z.BlockNumber.Msgsize() + 3
	if z.To == nil {
		s += msgp.NilSize
	} else {
		s += z.To.Msgsize()
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
		s += z.From.Msgsize()
	}
	s += 6 + z.Value.Msgsize() + 9 + z.GasPrice.Msgsize() + 4 + z.Gas.Msgsize() + 6 + z.Input.Msgsize()
	return
}
