package store

import (
	"time"

	"github.com/devanshu0x/Aster/internal/config"
)

type SnapshotEntry struct {
	Key       string
	Value     interface{}
	ExpiresAt int64
}

type Snapshot struct {
	Entries []SnapshotEntry
}



func SnapshotStore() Snapshot {
    now := time.Now().UnixMilli()

    snapshot := Snapshot{
        Entries: make([]SnapshotEntry, 0, store.ht[0].used),
    }

    seen := make(map[string]struct{})

    for _, ht := range store.ht {
        if ht == nil {
            continue
        }

        for _, bucket := range ht.buckets {
            for entry := bucket; entry != nil; entry = entry.Next {

                // Shouldn't normally happen, but protects us
                // if an entry somehow exists in both tables.
                if _, exists := seen[entry.Key]; exists {
                    continue
                }

                seen[entry.Key] = struct{}{}

                obj := entry.Obj

                // Don't persist already-expired objects.
                if obj.ExpiresAt != -1 && obj.ExpiresAt <= now {
                    continue
                }

                snapshot.Entries = append(snapshot.Entries, SnapshotEntry{
                    Key:       entry.Key,
                    Value:     obj.Value,
                    ExpiresAt: obj.ExpiresAt,
                })
            }
        }
    }

    return snapshot
}

func RestoreSnapshot(snapshot Snapshot) {
	now := time.Now().UnixMilli()

	for _, entry := range snapshot.Entries {
	
		if entry.ExpiresAt != -1 && entry.ExpiresAt <= now {
			continue
		}

		obj := &Obj{
			Value:     entry.Value,
			ExpiresAt: entry.ExpiresAt,
		}

		// Reinitialize eviction metadata.
		if config.EVICTION_POLICY == config.LFU {
			setLFUCounter(obj, config.LFU_INIT_VAL)
			setLFULastDecay(obj, currentMinute())
		}

		Put(entry.Key, obj)
	}
}