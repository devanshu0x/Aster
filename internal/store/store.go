package store

import (
	"hash/fnv"
	"time"

	"github.com/devanshu0x/Aster/internal/config"
)

type HashTable struct {
	buckets []*Entry
	used    int
}

var store *HashTable
var LRUClock uint32

type ObjType string

const (
	StringObject ObjType = "string"
	HashObject   ObjType = "hash"
	ListObject   ObjType = "list"
)

type Entry struct {
	Key  string
	Obj  *Obj
	Next *Entry
}

type Obj struct {
	Value     interface{}
	ExpiresAt int64
	Type      ObjType
	LRUClock  uint32
}

func NewHashTable(s int) *HashTable {
	return &HashTable{
		buckets: make([]*Entry, s),
	}
}

func touch(o *Obj) {
	LRUClock++
	o.LRUClock = LRUClock
}

func init() {
	store = NewHashTable(config.HASH_TABLE_SIZE)
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

func hash(key string) uint64 {
	fnv := fnv.New64a()
	fnv.Write([]byte(key))
	return fnv.Sum64()
}

func (ht *HashTable) bucketIndex(key string) int {
	hashSum := hash(key)
	return int(hashSum % uint64(len(ht.buckets)))
}

func Put(k string, obj *Obj) {
	idx := store.bucketIndex(k)
	curr := store.buckets[idx]
	for curr != nil && curr.Key != k {
		curr = curr.Next
	}
	
	touch(obj)

	if curr != nil {
		curr.Obj = obj
	} else {
		if store.used >= config.MAX_KEYS {
			// TODO: handle used decrement here only
		    AllKeysLRUEviction()
	    }
		store.used++
		curr=store.buckets[idx]
		store.buckets[idx]=&Entry{
			Key: k,
			Obj: obj,
			Next: curr,
		}
	}
	
}

func Get(k string) *Obj {
	idx := store.bucketIndex(k)
	var prev *Entry
	curr := store.buckets[idx]
	for curr != nil && curr.Key != k {
		prev = curr
		curr = curr.Next
	}
	if curr == nil {
		return nil
	}
	obj := curr.Obj
	if obj.ExpiresAt != -1 && obj.ExpiresAt <= time.Now().UnixMilli() {
		store.used--
		if prev == nil {
			store.buckets[idx] = curr.Next
		} else {
			prev.Next = curr.Next
		}
		curr.Next = nil
		return nil
	}

	touch(obj)

	return obj
}

func Del(k string) bool{
	idx:=store.bucketIndex(k)
	curr:=store.buckets[idx]
	var prev *Entry
	for curr!=nil && curr.Key!=k{
		prev=curr
		curr=curr.Next
	}
	if curr==nil{
		return false
	}
	obj:=curr.Obj
	store.used--
	if prev==nil{
		store.buckets[idx]=curr.Next
		curr.Next=nil
	}else{
		prev.Next=curr.Next
		curr.Next=nil
	}
	if obj.ExpiresAt!=-1 && obj.ExpiresAt<=time.Now().UnixMilli(){
		return false
	}

	return true
}

func Expire(k string, expInMilli int64) bool{
	idx:=store.bucketIndex(k)
	var prev *Entry
	curr:=store.buckets[idx]
	
	for curr!=nil && curr.Key!=k{
		prev=curr
		curr=curr.Next
	}

	if curr==nil{
		return false
	}
	obj:=curr.Obj
	now:=time.Now().UnixMilli()
	if obj.ExpiresAt!=-1 && obj.ExpiresAt<=now{
		store.used--
		if prev==nil{
			store.buckets[idx]=curr.Next
			curr.Next=nil
		}else{
			prev.Next=curr.Next
			curr.Next=nil
		}
		return false
	}

	expTime:=now+expInMilli

	obj.ExpiresAt=expTime

	return true
}
