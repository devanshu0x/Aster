package store

import (
	"time"

	"github.com/devanshu0x/Aster/internal/config"
)

type ObjType string

const (
	StringObject ObjType = "string"
	HashObject   ObjType = "hash"
	ListObject   ObjType = "list"
)

type Obj struct {
	Value     interface{}
	ExpiresAt int64
	Type      ObjType
	EvictData  uint32
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

func NewObject(value interface{}, durationMs int64, objType ObjType) *Obj {
	expiresAt := int64(-1)
	if durationMs > 0 {
		expiresAt = time.Now().UnixMilli() + durationMs
	}

	obj:=&Obj{
		Value:     value,
		ExpiresAt: expiresAt,
		Type:      objType,
	}

	if config.EVICTION_POLICY == config.LFU {
	setLFUCounter(obj, config.LFU_INIT_VAL)
	setLFULastDecay(obj, currentMinute())
	}

	return obj
}