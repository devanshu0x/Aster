package store

import (
	"math/rand"

	"github.com/devanshu0x/Aster/internal/config"
)

/*
Approximate LRU eviction (similar to Redis).

Maintaining a perfect LRU list would require moving an entry on every
read/write, adding extra memory overhead and pointer updates.

Instead, when memory needs to be reclaimed:
  1. Randomly sample a small number of keys from the hash table.
  2. Compare their LRU timestamps.
  3. Evict the key that was accessed the longest time ago.

This is only an approximation of true LRU, but with a small sample size
(e.g. 5-10 keys) it closely matches true LRU in practice while keeping
normal GET and SET operations O(1).
*/
func evictLRUSample(){
	sampleSize:=config.SAMPLE_SIZE
	var oldest *Entry
	for sampleSize!=0{
		idx:=rand.Intn(len(store.ht[0].buckets))
		if store.ht[0].buckets[idx]==nil{
			continue
		}
		curr:=store.ht[0].buckets[idx]
		bucketSize:=0
		for curr!=nil{
			bucketSize++
			curr=curr.Next
		}

		bucketVal:=rand.Intn(bucketSize)
		curr=store.ht[0].buckets[idx]
		for range bucketVal{
			curr=curr.Next
		}
		sampleSize--
		if oldest==nil ||  curr.Obj.LRUClock<oldest.Obj.LRUClock{
			oldest=curr
		}
	}
	if oldest==nil{
		return
	}
	Del(oldest.Key)

}