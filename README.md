# Mario Kart 8 (Wii U) server
Replacement NEX server for Mario Kart 8 on the Wii U

## About
This is a NEX game server for the Wii U version of Mario Kart 8, built on top
of the [Pretendo Network](https://github.com/PretendoNetwork) NEX libraries.
It handles authentication, Ranking (time trial leaderboards), DataStore (ghost
upload/download), Worldwide/friend matchmaking and NAT traversal.

| | |
|---|---|
| Game server ID | `1010EB00` |
| Access key | `25dbf96a` |
| NEX version | 3.5.4 |

### CTGP-Cafe support
The server also understands CTGP-Cafe time trials:
- custom-track leaderboards live in persistence slots `128+` (category = slot + 16),
  described by [`ctgp/wire_catalog.json`](ctgp/wire_catalog.json);
- uploaded ghosts are checked for integrity (size, CTG0 header and CRC-32) and
  dropped if they fail;
- when a player opens a ghost from a custom-track leaderboard, the client asks for
  the original track's slot; the server answers from the custom board the player
  was just viewing.

CTGP players can be kept apart from vanilla players by giving the CTGP build its
own game server ID on the account server (for example `1010EB01`) and routing it
to the same instance.

## Requirements
- Docker + Docker Compose
- An account server that issues NEX tokens for the game server ID
  (e.g. [Pretendo's account server](https://github.com/PretendoNetwork/account)),
  or local mode for LAN testing
- Optional: an S3-compatible storage (MinIO works) for ghost uploads

PostgreSQL is bundled in `compose.yaml`.

## Running
```
cp .env.example .env
cp settings.example.json settings.json
docker compose up -d --build
```

The server listens on UDP `29500` (authentication) and `29501` (secure).

## Configuration
All settings are environment variables, see [.env.example](.env.example).

| Variable | Description |
|---|---|
| `PN_MK8_AUTH_PORT` / `PN_MK8_SECURE_PORT` | UDP ports |
| `PN_MK8_SECURE_HOST` | Address given to consoles for the secure server. Keep it short (the game truncates it at ~15 characters) |
| `PN_MK8_KERBEROS_PASSWORD` | Kerberos password for the server accounts |
| `PN_MK8_LOCAL_MODE` | `1` = accounts from `settings.json`, tokens not checked (LAN/testing only). `0` = validate tokens from your account server |
| `PN_MK8_NEX_TOKEN_AES_KEY` | 64 hex chars, must match your account server (when `LOCAL_MODE=0`) |
| `PN_MK8_NEX_PASSWORD_SECRET` | 32+ bytes hex, must match your account server (when `LOCAL_MODE=0`) |
| `PN_MK8_POSTGRES_URI` | PostgreSQL connection string |
| `PN_MK8_LEGACY_SIG` / `PN_MK8_STRUCT_HEADER` | PRUDP/stream flags, defaults (`0` / `1`) are the ones Mario Kart 8 needs |
| `PN_S3_*` | S3 endpoint/bucket/credentials for ghost uploads |

Change every default password in `.env` and `compose.yaml` before exposing
the server.

## Known issues
- Online races can occasionally end with a communication error after 1-2 minutes.

## License
AGPL-3.0, see [LICENSE](LICENSE). The patched copy of
`nex-protocols-common-go` in `internal/` keeps its own license.
