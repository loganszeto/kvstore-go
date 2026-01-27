#!/usr/bin/env bash
set -euo pipefail

ADDR="${ADDR:-127.0.0.1:7379}"

CVE_JSON='{"cve":"CVE-2024-1234","severity":"HIGH","summary":"Example vuln for demo","affected":["cpe:microsoft:office:2019"]}'
CPE_JSON='["CVE-2024-1234","CVE-2023-9999"]'

go run ./cmd/kv-cli --addr "${ADDR}" SET "cve:CVE-2024-1234" "${CVE_JSON}"
go run ./cmd/kv-cli --addr "${ADDR}" SET "cpe:microsoft:office:2019" "${CPE_JSON}"
go run ./cmd/kv-cli --addr "${ADDR}" GET "cpe:microsoft:office:2019"
