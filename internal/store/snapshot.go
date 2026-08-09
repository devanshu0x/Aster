package store

import "time"

type SnapshotEntry struct {
	Key       string
	Type      ObjType
	Value     interface{}
	ExpiresAt int64
}

type Snapshot struct {
	Entries []SnapshotEntry
}

func cloneValue(obj *Obj) interface{} {
    switch obj.Type {
    case StringObject:
        return obj.Value
    default:
        return nil
    }
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
                    Type:      obj.Type,
                    Value:     cloneValue(obj),
                    ExpiresAt: obj.ExpiresAt,
                })
            }
        }
    }

    return snapshot
}