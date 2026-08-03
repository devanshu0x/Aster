# Aster

> A Redis-inspired, RESP2-compatible in-memory database written from scratch in Go.

Aster is a learning project for exploring the building blocks behind Redis-like systems: TCP sockets, non-blocking I/O, Linux `epoll`, the RESP wire protocol, command dispatch, expiration, and hash-table maintenance. It currently implements a focused subset of Redis commands rather than the complete Redis command surface.

## What is implemented

- RESP2 encoder and streaming decoder, including pipelined commands
- Linux `epoll`-based, non-blocking TCP event loop
- In-memory string key-value store
- Incremental hash-table rehashing as the store grows
- Lazy expiration on reads plus periodic active expiration
- RESP-compatible replies for the supported commands

## Supported commands

| Command | Notes |
| --- | --- |
| `PING [message]` | Returns `PONG`, or echoes a message. |
| `SET key value [EX seconds]` | Stores a string; `EX` assigns a TTL in seconds. |
| `GET key` | Returns the string value or a null bulk string. |
| `DEL key [key ...]` | Deletes one or more keys and returns the number removed. |
| `EXPIRE key seconds` | Sets a key's TTL and returns whether it was set. |
| `TTL key` | Returns remaining TTL in seconds, `-1` without expiry, or `-2` when missing/expired. |

## Requirements

- Go 1.26 or newer
- Linux (the server uses `epoll` and Linux system calls)

## Getting started

```bash
git clone https://github.com/devanshu0x/Aster.git
cd Aster
go run .
```

The server listens on `0.0.0.0:6969` by default. Configure it with flags:

```bash
go run . -host 127.0.0.1 -port 6969
```

Try it with a RESP-compatible client:

```bash
redis-cli -p 6969 PING
redis-cli -p 6969 SET greeting hello EX 60
redis-cli -p 6969 GET greeting
redis-cli -p 6969 TTL greeting
```

## Project structure

```
internal/
    command/        # Command parsing and implementations
    config/         # Server and store configuration
    resp/           # RESP2 encoder and streaming decoder
    server/         # TCP server and epoll event loop
    store/          # Dictionary, expiry, rehashing, and eviction hooks
```

## Roadmap

The next items are ordered to build a dependable core before adding distributed features.

### 1. Harden the server and protocol

- [ ] Add unit and integration tests for every command, expiry boundary, rehashing, pipelining, and partial writes.
- [ ] Add fuzz tests and request-size/depth limits to the RESP decoder.
- [ ] Handle `EPOLLOUT` so buffered responses cannot stall after an `EAGAIN` write.
- [ ] Standardize Redis-style error messages and command arity/type validation.
- [ ] Add graceful shutdown, connection limits, structured logging, and basic metrics.

### 2. Complete the string and keyspace foundation

- [ ] `EXISTS`, `TYPE`, `DBSIZE`, `FLUSHDB`, `MGET`, `MSET`, `INCR`/`DECR`, and `APPEND`.
- [ ] Redis-compatible `SET` options: `NX`, `XX`, `PX`, `KEEPTTL`, and `GET`.
- [ ] `PTTL`, `PEXPIRE`, `PERSIST`, `EXPIREAT`, and `PEXPIREAT`.
- [ ] Cursor-based `SCAN` with `MATCH` and `COUNT` rather than a blocking `KEYS` implementation.

### 3. Add data types and their core commands

- [ ] Hashes: `HSET`, `HGET`, `HDEL`, `HGETALL`.
- [ ] Lists: `LPUSH`, `RPUSH`, `LPOP`, `RPOP`, `LRANGE`.
- [ ] Sets: `SADD`, `SREM`, `SMEMBERS`, `SISMEMBER`.
- [ ] Sorted sets: `ZADD`, `ZRANGE`, `ZRANK`.

### 4. Add server semantics

- [ ] Logical databases and `SELECT`.
- [ ] Transactions: `MULTI`, `EXEC`, `DISCARD`, `WATCH`.
- [ ] Pub/Sub with per-client subscription state and back-pressure handling.
- [ ] Authentication and ACLs before exposing the server beyond a trusted network.

### 5. Make data durable and scalable

- [ ] Configurable memory limits and an actual eviction policy (LRU/LFU/TTL).
- [ ] Append-only file (AOF), rewrite/compaction, and crash recovery.
- [ ] RDB-style snapshots and background save.
- [ ] Leader/replica replication, replication offsets, and resynchronization.

## Development

```bash
go test ./...
```

## Scope

Aster is currently intended for learning and local experimentation. It should not be used as a drop-in production Redis replacement: persistence, authentication, replication, comprehensive command compatibility, and production hardening are still roadmap work.
