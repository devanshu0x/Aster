package store

import (
	"log"
	"github.com/devanshu0x/Aster/internal/config"
)

type Dict struct {
	ht        [2]*HashTable
	rehashIdx int
}

func (d *Dict) totalKeys() int{
	total:=0
	total+=d.ht[0].used
	if d.isRehashing(){
		total+=d.ht[1].used
	}
	return total
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