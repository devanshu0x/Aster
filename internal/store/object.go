package store

import "time"

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
	LRUClock  uint32
}

func touch(o *Obj) {
	if o == nil {
		return
	}
	LRUClock++
	o.LRUClock = LRUClock
}

func NewObject(value interface{}, durationMs int64, objType ObjType) *Obj {
	expiresAt := int64(-1)
	if durationMs > 0 {
		expiresAt = time.Now().UnixMilli() + durationMs
	}

	return &Obj{
		Value:     value,
		ExpiresAt: expiresAt,
		Type:      objType,
	}
}