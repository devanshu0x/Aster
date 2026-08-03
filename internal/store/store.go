package store

import (
	"time"
	"github.com/devanshu0x/Aster/internal/config"
)

var store *Dict
var LRUClock uint32

func init() {
	store = &Dict{
		ht: [2]*HashTable{
			NewHashTable(config.HASH_TABLE_SIZE),
			nil,
		},
		rehashIdx: -1,
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
