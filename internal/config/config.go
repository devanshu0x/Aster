package config

var HOST string
var PORT int
var MAX_KEYS int=5

var SAMPLE_SIZE int=2
var NUMBER_OF_SAMPLE int=2
var HASH_TABLE_SIZE=2
var LFU_INIT_VAL uint8 =5
var LFU_LOG_FACTOR uint8=10
var DECAY_TIME=1 // time in min

type EvictonPolicy int

const(
	NO_EVICTION EvictonPolicy=0
	LRU EvictonPolicy=1
	LFU EvictonPolicy=2
	RDB_PATH = "./data/dump.rdb"
    AOF_PATH = "./data/appendonly.aof"
	LOAD_RDB_ON_START=true
	USE_AOF=true
	MAX_OBJECTS= 100000
)

var EVICTION_POLICY=NO_EVICTION