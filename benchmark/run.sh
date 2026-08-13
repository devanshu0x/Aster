#!/usr/bin/env bash

# Compare Aster against a locally running Valkey instance using the same
# valkey-benchmark workload.  Aster is started with persistence disabled so
# disk I/O does not skew the in-memory comparison.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOST="${HOST:-127.0.0.1}"
ASTER_PORT="${ASTER_PORT:-6969}"
VALKEY_PORT="${VALKEY_PORT:-6379}"
REQUESTS="${REQUESTS:-100000}"
CLIENTS="${CLIENTS:-50}"
PIPELINE="${PIPELINE:-1}"
DATA_SIZE="${DATA_SIZE:-16}"
TESTS="${TESTS:-ping,ping_inline,set,get}"
START_ASTER="${START_ASTER:-1}"
RESULTS_DIR="${RESULTS_DIR:-$ROOT_DIR/benchmark/results}"
RUN_ID="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_DIR/$RUN_ID"
ASTER_PID=""

require_command() {
    command -v "$1" >/dev/null 2>&1 || {
        echo "error: $1 is required but was not found in PATH" >&2
        exit 1
    }
}

cleanup() {
    if [[ -n "$ASTER_PID" ]] && kill -0 "$ASTER_PID" 2>/dev/null; then
        kill "$ASTER_PID" 2>/dev/null || true
        wait "$ASTER_PID" 2>/dev/null || true
    fi
}

wait_for_server() {
    local port="$1"
    local name="$2"
    local attempt
    for attempt in {1..50}; do
        if valkey-cli -h "$HOST" -p "$port" ping 2>/dev/null | grep -qx 'PONG'; then
            return 0
        fi
        sleep 0.1
    done
    echo "error: $name did not become ready at $HOST:$port" >&2
    exit 1
}

require_command valkey-benchmark
require_command valkey-cli
if [[ "$START_ASTER" == "1" ]]; then
    require_command go
fi

if [[ "$ASTER_PORT" == "$VALKEY_PORT" ]]; then
    echo "error: ASTER_PORT and VALKEY_PORT must differ" >&2
    exit 1
fi

mkdir -p "$RUN_DIR"
trap cleanup EXIT INT TERM

if [[ "$START_ASTER" == "1" ]]; then
    if (exec 3<>"/dev/tcp/$HOST/$ASTER_PORT") 2>/dev/null; then
        echo "error: Aster port $HOST:$ASTER_PORT is already in use; stop the existing server or run with START_ASTER=0" >&2
        exit 1
    fi

    echo "Starting Aster on $HOST:$ASTER_PORT..."
    (
        cd "$ROOT_DIR"
        exec go run . \
            -host "$HOST" \
            -port "$ASTER_PORT" \
            -load-rdb-on-start=false \
            -use-aof=false
    ) >"$RUN_DIR/aster-server.log" 2>&1 &
    ASTER_PID=$!
fi

wait_for_server "$ASTER_PORT" "Aster"
wait_for_server "$VALKEY_PORT" "Valkey"

benchmark_args=(
    -h "$HOST"
    -n "$REQUESTS"
    -c "$CLIENTS"
    -P "$PIPELINE"
    -d "$DATA_SIZE"
    -t "$TESTS"
)

printf 'Aster vs Valkey benchmark\n\n' | tee "$RUN_DIR/metadata.txt"
{
    echo "timestamp=$RUN_ID"
    echo "host=$HOST"
    echo "requests=$REQUESTS"
    echo "clients=$CLIENTS"
    echo "pipeline=$PIPELINE"
    echo "data_size_bytes=$DATA_SIZE"
    echo "tests=$TESTS"
} >> "$RUN_DIR/metadata.txt"

echo "Benchmarking Aster..."
valkey-benchmark "${benchmark_args[@]}" -p "$ASTER_PORT" | tee "$RUN_DIR/aster.txt"

echo "Benchmarking Valkey..."
valkey-benchmark "${benchmark_args[@]}" -p "$VALKEY_PORT" | tee "$RUN_DIR/valkey.txt"

echo
echo "Results written to $RUN_DIR"
