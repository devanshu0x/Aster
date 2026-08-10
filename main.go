package main

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/devanshu0x/Aster/internal/config"
	"github.com/devanshu0x/Aster/internal/persistence"
	"github.com/devanshu0x/Aster/internal/server"
	"github.com/devanshu0x/Aster/internal/store"
)


func setupFlags(){
	flag.StringVar(&config.HOST,"host","0.0.0.0","host for the aster server")
	flag.IntVar(&config.PORT,"port",6969,"port for the aster server")
	flag.Parse()
}

func loadPersistence() {
	snapshot, err := persistence.LoadRDB(config.RDB_PATH)

	if err != nil {
		// No RDB on first startup is normal.
		if errors.Is(err, os.ErrNotExist) {
			log.Println("No RDB file found, starting with empty database")
			return
		}

		log.Println("Failed to load RDB: ", err)
	}

	store.RestoreSnapshot(snapshot)

	log.Printf("Loaded %d keys from RDB\n", len(snapshot.Entries))
}

func main(){
	setupFlags()
	log.Println("Aster is starting...")
	if config.LOAD_RDB_ON_START{
		log.Println("Attempting to load RDB snapshot")
		loadPersistence()
	}
	var aof *persistence.AOF
	if config.USE_AOF{
		var err error
		aof,err=persistence.OpenAOF(config.AOF_PATH)
		if err!=nil{
			log.Fatalln(err)
		}

		defer aof.Close()
	}
	err:=server.RunAsyncTCPServer(aof)
	log.Fatalln("Error closing server: ",err)
}