package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/devanshu0x/Aster/internal/config"
	"github.com/devanshu0x/Aster/internal/persistence"
	"github.com/devanshu0x/Aster/internal/server"
	"github.com/devanshu0x/Aster/internal/store"
)

func setupFlags() error {
	lfuInitVal := uint(config.LFU_INIT_VAL)
	lfuLogFactor := uint(config.LFU_LOG_FACTOR)
	policy := "noeviction"

	flag.StringVar(&config.HOST, "host", "0.0.0.0", "IPv4 address for the Aster server")
	flag.IntVar(&config.PORT, "port", 6969, "TCP port for the Aster server")
	flag.IntVar(&config.MAX_OBJECTS, "max-objects", config.MAX_OBJECTS, "maximum number of keys before eviction")
	flag.IntVar(&config.HASH_TABLE_SIZE, "hash-table-size", config.HASH_TABLE_SIZE, "initial hash-table bucket count")
	flag.IntVar(&config.SAMPLE_SIZE, "sample-size", config.SAMPLE_SIZE, "keys sampled when choosing an eviction candidate")
	flag.UintVar(&lfuInitVal, "lfu-init-val", lfuInitVal, "initial LFU counter value (0-255)")
	flag.UintVar(&lfuLogFactor, "lfu-log-factor", lfuLogFactor, "LFU counter logarithmic factor (0-255)")
	flag.IntVar(&config.DECAY_TIME, "decay-time", config.DECAY_TIME, "LFU counter decay interval in minutes")
	flag.StringVar(&policy, "eviction-policy", policy, "eviction policy: noeviction, lru, or lfu")
	flag.StringVar(&config.RDB_PATH, "rdb-path", config.RDB_PATH, "RDB snapshot file path")
	flag.StringVar(&config.AOF_PATH, "aof-path", config.AOF_PATH, "append-only file path")
	flag.BoolVar(&config.LOAD_RDB_ON_START, "load-rdb-on-start", config.LOAD_RDB_ON_START, "load the RDB snapshot on startup")
	flag.BoolVar(&config.USE_AOF, "use-aof", config.USE_AOF, "append mutating commands to the AOF")
	flag.Parse()

	if config.MAX_OBJECTS < 1 || config.HASH_TABLE_SIZE < 1 || config.SAMPLE_SIZE < 1 || config.DECAY_TIME < 1 {
		return errors.New("max-objects, hash-table-size, sample-size, and decay-time must be positive")
	}
	if lfuInitVal > 255 || lfuLogFactor > 255 {
		return errors.New("lfu-init-val and lfu-log-factor must be between 0 and 255")
	}
	config.LFU_INIT_VAL = uint8(lfuInitVal)
	config.LFU_LOG_FACTOR = uint8(lfuLogFactor)

	switch strings.ToLower(policy) {
	case "noeviction":
		config.EVICTION_POLICY = config.NO_EVICTION
	case "lru":
		config.EVICTION_POLICY = config.LRU
	case "lfu":
		config.EVICTION_POLICY = config.LFU
	default:
		return fmt.Errorf("invalid eviction-policy %q (want noeviction, lru, or lfu)", policy)
	}

	return nil
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
		return
	}

	store.RestoreSnapshot(snapshot)

	log.Printf("Loaded %d keys from RDB\n", len(snapshot.Entries))
}

func main() {
	if err := setupFlags(); err != nil {
		log.Fatal(err)
	}
	store.Initialize()
	log.Printf("Aster is starting (max-objects=%d, eviction-policy=%s)", config.MAX_OBJECTS, policyName())
	if config.LOAD_RDB_ON_START {
		log.Println("Attempting to load RDB snapshot")
		loadPersistence()
	}
	var aof *persistence.AOF
	if config.USE_AOF {
		var err error
		aof, err = persistence.OpenAOF(config.AOF_PATH)
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("Append-only file enabled at %q", config.AOF_PATH)

		defer aof.Close()
	}
	err := server.RunAsyncTCPServer(aof)
	log.Fatalln("Error closing server: ", err)
}

func policyName() string {
	switch config.EVICTION_POLICY {
	case config.LRU:
		return "lru"
	case config.LFU:
		return "lfu"
	default:
		return "noeviction"
	}
}
