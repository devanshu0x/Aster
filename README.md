# Aster

> A Redis-inspired, RESP2-compatible in-memory database written from scratch in Go.

Aster is a learning project for exploring the building blocks behind Redis-like systems: TCP sockets, non-blocking I/O, Linux `epoll`, the RESP wire protocol, command dispatch, expiration, and hash-table maintenance. It currently implements a focused subset of Redis commands rather than the complete Redis command surface.

## What is implemented

- RESP2 encoder and streaming decoder, including pipelined commands
- Linux `epoll`-based, non-blocking TCP event loop
- In-memory string key-value store
- Incremental hash-table rehashing as the store grows
- Lazy expiration on reads plus periodic active expiration
- Configurable approximate LRU and LFU eviction
- RDB-style snapshots and an append-only write log
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
| `INCR key` | Increments an integer string, creating it with value `1` when absent. |
| `SAVE` | Writes an RDB snapshot to the configured path. |

## Requirements

- Go 1.26 or newer
- Linux (the server uses `epoll` and Linux system calls)

## Getting started

```bash
git clone https://github.com/devanshu0x/Aster.git
cd Aster
go run .
```

The server listens on `0.0.0.0:6969` by default. For a local-only server:

```bash
go run . -host 127.0.0.1 -port 6969
```

## Configuration

All runtime configuration is available as command-line flags. Run `go run . -help` for the generated help text.

| Flag | Default | Purpose |
| --- | --- | --- |
| `-host` | `0.0.0.0` | IPv4 bind address. |
| `-port` | `6969` | TCP listen port. |
| `-max-objects` | `10` | Key limit at which an eviction is attempted. Use `lru` or `lfu` for this to enforce a limit. |
| `-eviction-policy` | `noeviction` | `noeviction`, `lru`, or `lfu`. |
| `-sample-size` | `2` | Candidates sampled for an approximate eviction decision. |
| `-hash-table-size` | `2` | Initial bucket count for the in-memory hash table. |
| `-lfu-init-val` | `5` | Initial LFU frequency counter. |
| `-lfu-log-factor` | `10` | LFU logarithmic increment factor. |
| `-decay-time` | `1` | LFU counter decay interval, in minutes. |
| `-rdb-path` | `./data/dump.rdb` | Snapshot file used by `SAVE` and startup loading. |
| `-load-rdb-on-start` | `true` | Load the RDB snapshot during startup. |
| `-aof-path` | `./data/appendonly.aof` | Append-only log file path. |
| `-use-aof` | `true` | Append successful mutating commands to the AOF. |

For example, to run with a 1,000-key approximate LRU cache and custom persistence paths:

```bash
go run . \
  -max-objects 1000 \
  -eviction-policy lru \
  -rdb-path ./data/aster.rdb \
  -aof-path ./data/aster.aof
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

## Current limitations

- Aster is a learning project, not a drop-in production Redis replacement.
- The AOF records successful writes, but it is not replayed during startup yet; use `SAVE` for restorable persistence.
- The server currently accepts numeric IPv4 addresses, not hostnames.
- Back-pressure after a non-blocking write reaches `EAGAIN`, graceful shutdown, authentication, replication, and comprehensive command compatibility are still future work.

## Development

```bash
go test ./...
```
