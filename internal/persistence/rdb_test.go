package persistence

import (
	"os"
	"testing"

	"github.com/devanshu0x/Aster/internal/store"
)

func TestRDB(t *testing.T) {
	snapshot := store.Snapshot{
		Entries: []store.SnapshotEntry{
			{
				Key:       "foo",
				Value:     "bar",
				ExpiresAt: -1,
			},
			{
				Key:       "hello",
				Value:     "world",
				ExpiresAt: 123456789,
			},
		},
	}

	path := "test.rdb"

	err := SaveRDB(snapshot, path)
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(path)

	loaded, err := LoadRDB(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(loaded.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(loaded.Entries))
	}

	t.Logf("%+v", loaded)
}