package command

import (
	"log"

	"github.com/devanshu0x/Aster/internal/config"
	"github.com/devanshu0x/Aster/internal/persistence"
	"github.com/devanshu0x/Aster/internal/resp"
	"github.com/devanshu0x/Aster/internal/store"
)

func cmdSAVE(argArr []*resp.RESPValue) *resp.RESPValue {
	if len(argArr)!=0{
		return RESPError("Err invalid number of argument for 'save' command")
	}
	snapshot := store.SnapshotStore()

	if err := persistence.SaveRDB(snapshot, config.RDB_PATH); err != nil {
		log.Printf("Failed to save RDB snapshot to %q: %v", config.RDB_PATH, err)
		return RESPError("Err failed to save snapshot")
	}
	log.Printf("Saved RDB snapshot with %d keys to %q", len(snapshot.Entries), config.RDB_PATH)
	return &resp.RESPValue{
		Type:  resp.RESPSimpleString,
		Value: "OK",
	}
}
