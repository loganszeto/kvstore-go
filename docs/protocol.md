# Protocol

vulnkv uses a simple line-oriented protocol over TCP.

## Requests

- `PING`
- `GET <key>`
- `DEL <key>`
- `EXISTS <key>`
- `EXPIRE <key> <ttlSeconds>`
- `KEYS <prefix>` (or `<prefix>*`)
- `SET <key> <len>\n<raw>\n`
- `SETEX <key> <ttlSeconds> <len>\n<raw>\n`
- `STATS`

Values for `SET` and `SETEX` are raw bytes; length is in bytes. All commands end with `\n`.

## Responses

- `OK`
- `ERR <msg>`
- `NOT_FOUND`
- `VALUE <len>\n<raw>\n`
- `INT <n>`
- `ARRAY <count>\n<item>\n...`

`ARRAY` items are returned as single-line strings.
