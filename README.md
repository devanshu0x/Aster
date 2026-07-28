# Aster

> A Redis-compatible in-memory database written from scratch in Go.

Aster is a learning project built to understand how databases like Redis work internally. Rather than relying on Go's networking abstractions, the goal is to progressively implement the core building blocks of a high-performance server—from the networking layer to the command execution engine.

## Features

- RESP (Redis Serialization Protocol) parser
- Redis-compatible TCP server
- Event-driven architecture using Linux `epoll`
- Single-threaded I/O multiplexing
- In-memory key-value store
- Compatible with existing Redis clients (`redis-cli`, `valkey-cli`, etc.)

## Supported Commands

- `PING`
- `SET`
- `GET`

> More Redis commands will be added over time.

## Why?

The goal of this project is to learn how systems like Redis are built under the hood by implementing them from scratch.

Some of the topics explored include:

- TCP sockets
- Linux system calls
- Non-blocking I/O
- I/O multiplexing (`epoll`)
- RESP parsing
- Event loops
- Database internals
- Memory management

## Getting Started

```bash
git clone https://github.com/devanshu0x/Aster.git
cd Aster

go run ./cmd/aster
```

Then connect using any Redis client:

```bash
redis-cli -p 6969
```

or

```bash
valkey-cli -p 6969
```

## Project Structure

```
internal/
    command/        # Redis command implementations
    resp/           # RESP encoder/decoder
    server/         # TCP server and event loop
    store/          # In-memory database
```

## Roadmap

- [x] RESP parser
- [x] TCP server
- [x] Redis client compatibility
- [x] Event-driven server using epoll
- [x] PING
- [x] SET / GET
- [ ] Expiry (TTL)
- [ ] Multiple data types
- [ ] Persistence (RDB/AOF)
- [ ] Transactions
- [ ] Replication
- [ ] Pub/Sub
