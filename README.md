# fastretro-cli

Terminal client for [fastRetro](https://github.com/helmedeiros/fastRetro) retrospective sessions.

Join a running retro session from the terminal and participate using a text-based interface.

## Features

- **Join rooms** by room code or URL
- **Pick identity** from existing participants
- **Brainstorm** — add cards to columns
- **Vote** — cast votes within budget
- **Discuss** — follow discussion carousel with notes
- **Review** — see action items and board overview
- **Close** — view retro summary and stats
- **Real-time sync** — all changes sync via WebSocket

## Install

```bash
go install github.com/helmedeiros/fastretro-cli/cmd/fastretro@latest
```

Or build from source:

```bash
make build
./bin/fastretro join ABC-123-DEF
```

## Usage

```bash
# Join by room code
fastretro join ABC-123-DEF

# Join by URL
fastretro join "http://localhost:5173/#room=ABC-123-DEF"

# Custom server
fastretro join ABC-123-DEF --server https://retro.example.com
```

## Development

```bash
make test       # Run tests
make cover      # Coverage report
make lint       # Go vet
make build      # Build binary
```

## Architecture

```
cmd/fastretro/     CLI entry point (cobra)
internal/
  protocol/        WebSocket message types (shared contract)
  client/          WebSocket connection manager
  tui/             Bubble Tea views (join, brainstorm, vote, discuss, review, close)
  styles/          Lip Gloss terminal styling
```

## Requirements

- Go 1.21+
- A running fastRetro web app instance
