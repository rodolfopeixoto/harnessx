# Install

## Requirements

- Go 1.23+
- Git
- macOS or Linux (x86_64 or arm64)

Optional but recommended:

- `ripgrep` (`rg`) — exact search (Phase 5)
- `sqlite3` CLI — for inspecting the local database
- Node 20+ and `npm` — for the dashboard (Phase 0 scaffold)

## Build from source

```bash
git clone https://github.com/ropeixoto/harnessx
cd harnessx
make build
./bin/harness version
```

## Verify

```bash
./bin/harness doctor
```

Doctor exits 0 when every required tool is present, 1 otherwise.
