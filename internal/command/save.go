package command

import (
	"log"

	"github.com/devanshu0x/Aster/internal/config"
	"github.com/devanshu0x/Aster/internal/persistence"
	"github.com/devanshu0x/Aster/internal/resp"
	"github.com/devanshu0x/Aster/internal/store"
)

func cmdSAVE() *resp.RESPValue {
	snapshot := store.SnapshotStore()

	if err := persistence.SaveRDB(snapshot, config.RDB_PATH); err != nil {
		log.Println(err)
		return RESPError("Err failed to save snapshot")
	}
	return &resp.RESPValue{
		Type:  resp.RESPSimpleString,
		Value: "OK",
	}
}
