package store

import (
	"time"

	"github.com/devanshu0x/Aster/internal/config"
)


const (
	objectTypeMask uint32 = 0xF0000000
	encodingMask   uint32 = 0x0F000000
)

const (
	StringObject uint32 = 1
	HashObject   uint32 = 2
	ListObject   uint32 = 3
)

/*
  Here I'm making some improvement, to reduce memory overhead
  EvictData is 4 bytes, its first byte represents object type and encoding
  type and next three bytes are used for eviction strategies.
  in first byte the first 4 bits will represent the object type and next 4 bits
  will represent encoding type.
  If eviction strategy is LRU than then then the next three bytes represent LRU
  Clock time and in case of LFU, first two byte represent last decay min and
  last 1 byte represent approximate frequency count using morris counter(not the
  orignal morris counter but a derived one which uses same probability logic)
*/

type Obj struct {
	Value     interface{}
	ExpiresAt int64
	EvictData  uint32
}

func setObjectType(o *Obj, t uint32) {
	o.EvictData &^= objectTypeMask
	o.EvictData |= t << 28
}

func getObjectType(o *Obj) uint32 {
	return (o.EvictData & objectTypeMask) >> 28
}

func touchLRU(o *Obj) {
	if o == nil {
		return
	}
	LRUClock++
	o.EvictData = LRUClock
}

func touchLFU(o *Obj) {
	if o == nil {
		return
	}
	
	now:=currentMinute()
	last:=getLFULastDecay(o)
	counter:=getLFUCounter(o)

	// decay the counter
	counter=LFUDecay(counter,now-last)

	// apply probabilistic increment
	counter=LFUIncrement(counter)

	setLFUCounter(o,counter)
	setLFULastDecay(o,now)

}

func touch(o *Obj){
	switch config.EVICTION_POLICY{
	case config.NO_EVICTION:
		return
	case config.LRU:
		touchLRU(o)
	case config.LFU:
		touchLFU(o)		
	}
}

func NewObject(value interface{}, durationMs int64, objType uint32) *Obj {
	expiresAt := int64(-1)
	if durationMs > 0 {
		expiresAt = time.Now().UnixMilli() + durationMs
	}

	obj:=&Obj{
		Value:     value,
		ExpiresAt: expiresAt,
	}

	// for now we only support string objects
	switch objType {
	case StringObject:
		setObjectType(obj, StringObject)

	case HashObject:
		setObjectType(obj, HashObject)

	case ListObject:
		setObjectType(obj, ListObject)
	}

	if config.EVICTION_POLICY == config.LFU {
	setLFUCounter(obj, config.LFU_INIT_VAL)
	setLFULastDecay(obj, currentMinute())
	}

	return obj
}