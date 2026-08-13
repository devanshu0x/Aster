# Benchmarking Aster against Valkey

This is a small, local apples-to-apples comparison between Aster and Valkey.
The runner sends the same `valkey-benchmark` workload to both, starts Aster
for you, and leaves the raw output behind so you can look at the details later.

## Prerequisites

- Linux and Go 1.26+ (needed only when the script starts Aster)
- `valkey-benchmark` and `valkey-cli` available on `PATH`
- A local Valkey server listening on `127.0.0.1:6379`

For a throwaway Valkey instance with persistence turned off:

```bash
valkey-server --port 6379 --save '' --appendonly no
```

Run that in another terminal. The benchmark sends `SET` commands, so avoid
pointing it at a Valkey instance with data you care about.

## Run the default benchmark

From the repository root:

```bash
./benchmark/run.sh
```

By default it sends 100,000 requests each of RESP `PING`, inline `PING`,
`SET`, and `GET`, using 50 clients, no pipelining, and 16-byte values. Aster
runs on `127.0.0.1:6969` with RDB loading and AOF disabled, then the runner
shuts it down when it is done.

## Tune a run

You can tweak the workload with environment variables:

| Variable | Default | Meaning |
| --- | --- | --- |
| `HOST` | `127.0.0.1` | Address for both servers. |
| `ASTER_PORT` | `6969` | Aster port. |
| `VALKEY_PORT` | `6379` | Valkey port. |
| `REQUESTS` | `100000` | Requests per selected command. |
| `CLIENTS` | `50` | Concurrent benchmark clients. |
| `PIPELINE` | `1` | Requests sent per pipeline. |
| `DATA_SIZE` | `16` | `SET` value size in bytes. |
| `TESTS` | `ping,ping_inline,set,get` | Comma-separated `valkey-benchmark` tests. |
| `START_ASTER` | `1` | Set to `0` to benchmark an already-running Aster instance. |
| `RESULTS_DIR` | `benchmark/results` | Parent directory for run output. |

For example, this runs a larger pipelined workload:

```bash
REQUESTS=1000000 CLIENTS=100 PIPELINE=16 DATA_SIZE=128 ./benchmark/run.sh
```

To use an already-running Aster instance on a different port:

```bash
START_ASTER=0 ASTER_PORT=7000 ./benchmark/run.sh
```

## Results

Each run is written to `benchmark/results/<timestamp>/`:

```text
metadata.txt        workload settings
aster.txt           valkey-benchmark output for Aster
valkey.txt          valkey-benchmark output for Valkey
aster-server.log    Aster startup and server logs (when started by the runner)
```

Start with requests/sec, then look at the latency percentiles for the same
command in `aster.txt` and `valkey.txt`. Git ignores these results, so you can
run the benchmark as often as you like without cluttering the working tree.

## Observed result: 2026-08-13 (clean rerun)

This clean run used the defaults: 100,000 requests per command, 50 clients,
no pipelining, and 16-byte values. Aster started successfully on port 6969
with RDB loading and AOF disabled. The raw results live in
`benchmark/results/20260813-171704/`.

Treat this as a snapshot, not a promise. A few runs on a quiet machine will
give you a much more trustworthy picture than any single run.

| Command | Aster requests/sec | Valkey requests/sec | Aster p95 | Valkey p95 |
| --- | ---: | ---: | ---: | ---: |
| `PING_INLINE` | 92,166 | 94,518 | 0.439 ms | 0.407 ms |
| RESP `PING` | 84,459 | 96,899 | 0.575 ms | 0.383 ms |
| `SET` | 93,023 | 99,701 | 0.447 ms | 0.351 ms |
| `GET` | 94,340 | 100,604 | 0.423 ms | 0.351 ms |

Aster lands within 2.5–12.8% of Valkey's throughput here. `GET` and `SET` are
only about 6–7% behind; RESP `PING` is the widest gap at roughly 13%. Aster's
p95 latency stays below 0.6 ms with 50 clients.

That is a very encouraging result for a small learning project. Valkey has had
years of focused performance work behind it, so being this close on a local
benchmark is a good sign—not a reason to panic about the remaining gap.

## Making comparisons useful

The script keeps both sides on the same host, client count, pipeline depth,
payload size, and command list. For useful comparisons, keep the machine as
quiet as you can, run the test several times, and compare the middle result
rather than the best one. Aster intentionally implements a smaller command set
than Valkey, so stick to commands it supports.
