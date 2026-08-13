package config

var HOST string
var PORT int
var MAX_OBJECTS = 100000

var SAMPLE_SIZE = 5
var HASH_TABLE_SIZE = 1024
var LFU_INIT_VAL uint8 = 5
var LFU_LOG_FACTOR uint8 = 10
var DECAY_TIME = 1 // time in min

type EvictonPolicy int

const (
	NO_EVICTION EvictonPolicy = 0
	LRU         EvictonPolicy = 1
	LFU         EvictonPolicy = 2
)

var RDB_PATH = "./data/dump.rdb"
var AOF_PATH = "./data/appendonly.aof"
var LOAD_RDB_ON_START = true
var USE_AOF = true
var EVICTION_POLICY = NO_EVICTION
