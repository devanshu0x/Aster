package store

import (
	"log"
	"math/rand"
	"time"

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
func evictSample() {
	if config.EVICTION_POLICY == config.NO_EVICTION {
		return
	}

	sampleSize := config.SAMPLE_SIZE
	var oldest *Entry
	for sampleSize != 0 {
		idx := rand.Intn(len(store.ht[0].buckets))
		if store.ht[0].buckets[idx] == nil {
			continue
		}
		curr := store.ht[0].buckets[idx]
		bucketSize := 0
		for curr != nil {
			bucketSize++
			curr = curr.Next
		}

		bucketVal := rand.Intn(bucketSize)
		curr = store.ht[0].buckets[idx]
		for range bucketVal {
			curr = curr.Next
		}
		sampleSize--
		switch config.EVICTION_POLICY {
		case config.LRU:
			if oldest == nil || curr.Obj.EvictData < oldest.Obj.EvictData {
				oldest = curr
			}
		case config.LFU:
			if oldest == nil {
				oldest = curr
				continue
			}
			currFreq := LFUDecay(getLFUCounter(curr.Obj),currentMinute()-getLFULastDecay(curr.Obj))
			oldFreq := LFUDecay(getLFUCounter(oldest.Obj),currentMinute()-getLFULastDecay(oldest.Obj))

			if currFreq < oldFreq {
				oldest = curr
			}
		}
	}
	if oldest == nil {
		log.Printf("Eviction skipped: no candidate found for %s policy", evictionPolicyName())
		return
	}

	if Del(oldest.Key) {
		log.Printf("Evicted key %q using %s policy (%d/%d keys remaining)", oldest.Key, evictionPolicyName(), store.totalKeys(), config.MAX_OBJECTS)
	}

}

func evictionPolicyName() string {
	switch config.EVICTION_POLICY {
	case config.LRU:
		return "LRU"
	case config.LFU:
		return "LFU"
	default:
		return "noeviction"
	}
}

// IN LFU, EvictData-> 8 bit unused+ 16 bit last decay min+ 8 bit counter in

func getLFUCounter(o *Obj) uint8 {
	return uint8(o.EvictData & 0xFF)
}

func setLFUCounter(o *Obj, counter uint8) {
	o.EvictData &= ^uint32(0xFF)
	o.EvictData |= uint32(counter)
}

func getLFULastDecay(o *Obj) uint16 {
	return uint16((o.EvictData >> 8) & 0xFFFF)
}

func setLFULastDecay(o *Obj, minute uint16) {
	o.EvictData &= ^uint32(0xFFFF << 8)
	o.EvictData |= uint32(minute) << 8
}

func currentMinute() uint16 {
	return uint16(time.Now().Unix() / 60)
}

func LFUIncrement(counter uint8) uint8 {
	if counter == 255 {
		return counter
	}
	base := float64(0)

	if counter > config.LFU_INIT_VAL {
		base = float64(counter - config.LFU_INIT_VAL)
	}

	p := 1.0 / (base*float64(config.LFU_LOG_FACTOR) + 1)

	if rand.Float64() < p {
		counter++
	}
	return counter
}

func LFUDecay(counter uint8, elapsed uint16) uint8 {

	decays := elapsed / uint16(config.DECAY_TIME)

	if decays == 0 || counter <= config.LFU_INIT_VAL {
		return counter
	}

	if decays >= uint16(counter-config.LFU_INIT_VAL) {
		return config.LFU_INIT_VAL
	}

	return counter - uint8(decays)
}

