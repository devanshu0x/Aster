package store


type Encoding uint8

const (
	RawEncoding Encoding = iota
	IntEncoding
	HashTableEncoding
	LinkedListEncoding
)

func setEncoding(o *Obj, encoding Encoding) {
	o.EvictData &^= encodingMask
	o.EvictData |= uint32(encoding) << 24
}

func getEncoding(o *Obj) Encoding {
	return Encoding(
		(o.EvictData & encodingMask) >> 24,
	)
}

func TryEncodeInt(o *Obj) bool{
	if getEncoding(o)==IntEncoding{
		return true
	}
	if _,ok:=o.Value.(int);ok{
		setEncoding(o,IntEncoding)
		return true
	}
	return false
}
