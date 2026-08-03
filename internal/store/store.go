package store

import (
	"hash/fnv"
	"log"
	"time"

	"github.com/devanshu0x/Aster/internal/config"
)

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

type HashTable struct {
	buckets []*Entry
	used    int
}

type Dict struct {
	ht        [2]*HashTable
	rehashIdx int
}

var store *Dict
var LRUClock uint32

func NewHashTable(s int) *HashTable {
	return &HashTable{
		buckets: make([]*Entry, s),
	}
}

func touch(o *Obj) {
	if o == nil {
		return
	}
	LRUClock++
	o.LRUClock = LRUClock
}

func init() {
	store = &Dict{
		ht: [2]*HashTable{
			NewHashTable(config.HASH_TABLE_SIZE),
			nil,
		},
		rehashIdx: -1,
	}
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

// handles both insertion and updating
func (ht *HashTable) insertEntry(entry *Entry) {
	idx := ht.bucketIndex(entry.Key)
	curr := ht.buckets[idx]
	for curr != nil && curr.Key != entry.Key {
		curr = curr.Next
	}

	if curr != nil {
		curr.Obj = entry.Obj
	} else {
		ht.used++
		entry.Next = ht.buckets[idx]
		ht.buckets[idx] = entry
	}
}

func (ht *HashTable) findObj(k string) *Obj {
	idx := ht.bucketIndex(k)
	curr := ht.buckets[idx]
	for curr != nil && curr.Key != k {
		curr = curr.Next
	}

	if curr == nil {
		return nil
	}

	return curr.Obj
}

func (ht *HashTable) retrieveObj(k string) *Obj {
	obj := ht.findObj(k)
	if obj == nil {
		return nil
	}
	if obj.ExpiresAt != -1 && obj.ExpiresAt <= time.Now().UnixMilli() {
		ht.deleteEntry(k)
		return nil
	}
	return obj
}

func (ht *HashTable) deleteEntry(k string) bool {
	idx := ht.bucketIndex(k)
	curr := ht.buckets[idx]
	var prev *Entry
	for curr != nil && curr.Key != k {
		prev = curr
		curr = curr.Next
	}

	if curr == nil {
		return false
	}

	obj := curr.Obj

	if prev == nil {
		ht.buckets[idx] = curr.Next
	} else {
		prev.Next = curr.Next
		curr.Next = nil
	}
	ht.used--

	if obj.ExpiresAt != -1 && obj.ExpiresAt <= time.Now().UnixMilli() {
		return false
	}

	return true
}

func (d *Dict) loadFactor() float64 {
	return float64(d.ht[0].used) / float64(len(d.ht[0].buckets))
}

func (d *Dict) shouldShrink() bool {
	if d.loadFactor() <= 0.25 && len(d.ht[0].buckets)/2 >= config.HASH_TABLE_SIZE {
		return true
	}
	return false
}

func (d *Dict) shouldExpand() bool {
	if d.loadFactor() > 1 {
		return true
	}
	return false
}

func (d *Dict) startRehash(newSize int) {
	if d.rehashIdx != -1 {
		return
	}
	log.Println("Started Rehashing to size: ",newSize)
	d.rehashIdx = 0
	d.ht[1] = NewHashTable(newSize)
}

func (d *Dict) isRehashing() bool {
	return d.rehashIdx != -1
}

func (d *Dict) rehashStep() {
	if !d.isRehashing() {
		return
	}

	curr := d.ht[0].buckets[d.rehashIdx]

	for curr != nil {
		next := curr.Next
		d.ht[1].insertEntry(curr)
		d.ht[0].used--
		curr = next
	}
	d.ht[0].buckets[d.rehashIdx] = nil
	d.rehashIdx++
	if d.rehashIdx == len(d.ht[0].buckets) {
		d.rehashIdx = -1
		d.ht[0] = d.ht[1]
		d.ht[1] = nil
	}
}

func Put(k string, obj *Obj) {
	if store.shouldExpand() {
		store.startRehash(len(store.ht[0].buckets) * 2)
	}

	touch(obj)
	if store.isRehashing() {
		store.rehashStep()
		// rehash step might finish hashing so it may give nil pointer
		//derefrence errror when using store.ht[1]
		if store.isRehashing(){
			store.ht[0].deleteEntry(k)
			store.ht[1].insertEntry(&Entry{
			Key: k,
			Obj: obj,
		})
		}else{
			store.ht[0].insertEntry(&Entry{
			Key: k,
			Obj: obj,
		})
		}

	} else {
		store.ht[0].insertEntry(&Entry{
			Key: k,
			Obj: obj,
		})
	}

}

func Get(k string) *Obj {

	if store.isRehashing() {
		store.rehashStep()
		if obj := store.ht[0].retrieveObj(k); obj != nil {
			touch(obj)
			return obj
		}
		obj := store.ht[1].retrieveObj(k)
		touch(obj)
		return obj
	} else {
		obj := store.ht[0].retrieveObj(k)
		touch(obj)
		return obj
	}
}

func Del(k string) bool {
	if store.shouldShrink() {
		store.startRehash(len(store.ht[0].buckets) / 2)
	}

	if store.isRehashing() {
		store.rehashStep()
		return (store.ht[0].deleteEntry(k) || store.ht[1].deleteEntry(k))
	} else {
		return store.ht[0].deleteEntry(k)
	}
}

func Expire(k string, expInMilli int64) bool {
	now := time.Now().UnixMilli()
	expTime := now + expInMilli
	if store.isRehashing() {
		store.rehashStep()
		obj := store.ht[0].findObj(k)
		obj2 := store.ht[1].findObj(k)
		if obj == nil && obj2 == nil {
			return false
		}
		if obj != nil && obj.ExpiresAt != -1 && obj.ExpiresAt <= now {
			Del(k)
			return false
		}
		if obj2 != nil && obj2.ExpiresAt != -1 && obj2.ExpiresAt <= now {
			Del(k)
			return false
		}

		if obj != nil {
			obj.ExpiresAt = expTime
		}

		if obj2 != nil {
			obj2.ExpiresAt = expTime
		}

		return true

	} else {
		obj := store.ht[0].findObj(k)
		if obj == nil {
			return false
		}

		obj.ExpiresAt = expTime

		return true
	}
}
