package store

import (
	"hash/fnv"
	"time"
)

type Entry struct {
	Key  string
	Obj  *Obj
	Next *Entry
}

type HashTable struct {
	buckets []*Entry
	used    int
}

func NewHashTable(s int) *HashTable {
	return &HashTable{
		buckets: make([]*Entry, s),
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