#!/usr/bin/env bash
set -euo pipefail

DATA_DIR="${DATA_DIR:-./data}"
ADDR="${ADDR:-127.0.0.1:7379}"

cleanup() {
  if [[ -n "${SERVER_PID:-}" ]]; then
    kill "${SERVER_PID}" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

rm -rf "${DATA_DIR}"

go run ./cmd/kv-server --addr "${ADDR}" --data_dir "${DATA_DIR}" &
SERVER_PID=$!
sleep 0.3

go run ./cmd/kv-cli --addr "${ADDR}" SET demo hello
go run ./cmd/kv-cli --addr "${ADDR}" GET demo
go run ./cmd/kv-cli --addr "${ADDR}" SETEX tmp 1 goodbye
go run ./cmd/kv-cli --addr "${ADDR}" GET tmp
sleep 1.2
go run ./cmd/kv-cli --addr "${ADDR}" GET tmp

kill "${SERVER_PID}"
wait "${SERVER_PID}" 2>/dev/null || true

go run ./cmd/kv-server --addr "${ADDR}" --data_dir "${DATA_DIR}" &
SERVER_PID=$!
sleep 0.3
go run ./cmd/kv-cli --addr "${ADDR}" GET demo
