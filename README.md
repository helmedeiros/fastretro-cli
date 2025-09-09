# fastretro-cli

A terminal client for [fastRetro](https://github.com/helmedeiros/fastRetro) sprint retrospectives. Join a session from the command line, participate in every stage, and stay in sync with the web app — all without leaving your terminal.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

## Quick start

```bash
go install github.com/helmedeiros/fastretro-cli/cmd/fastretro@latest

fastretro join "http://localhost:5173/#room=ABC-123-DEF"
```

Or build from source:

```bash
git clone https://github.com/helmedeiros/fastretro-cli.git
cd fastretro-cli
make build
./bin/fastretro join ABC-123-DEF
```

## Stages

Every retro follows the same flow. The stage bar at the top always shows where you are:

```
ICEBREAKER  BRAINSTORM  GROUP  VOTE  DISCUSS  REVIEW  CLOSE
```

---

### Join

Pick your identity from the participant list or add a new name.

![Join screen](docs/screenshots/01-join.png)

| Key       | Action                |
|-----------|-----------------------|
| `j` / `k` | Navigate participants |
| `Enter`   | Select identity       |
| `n`       | Add new name          |
| `q`       | Quit                  |

---

### Icebreaker

Watch the icebreaker question and see who's answering. Controlled from the web app.

![Icebreaker](docs/screenshots/02-icebreaker.png)

---

### Brainstorm

Add cards to columns. Navigate between columns with `Tab`. Each column shows its template description so everyone knows what to write about.

![Brainstorm](docs/screenshots/03-brainstorm.png)

| Key              | Action            |
|------------------|-------------------|
| `j` / `k`        | Navigate cards    |
| `Tab` / `l` / `h` | Switch column    |
| `a`              | Add card          |
| `d`              | Delete your card  |
| `Enter`          | Submit card       |
| `Esc`            | Cancel input      |

---

### Group

Merge related cards into clusters. Select a card, press `m`, navigate to the target, press `m` again. Rename groups with `r`, ungroup with `u`.

![Group](docs/screenshots/04-group.png)

| Key              | Action             |
|------------------|--------------------|
| `j` / `k`        | Navigate items     |
| `Tab` / `l` / `h` | Switch column     |
| `m`              | Merge (two-step)   |
| `r`              | Rename group       |
| `u`              | Ungroup card       |
| `Esc`            | Cancel merge       |

---

### Vote

Cast votes on cards and groups within your budget. Vote counts and your own votes are shown inline. Groups expand to show their cards for context.

![Vote](docs/screenshots/05-vote.png)

| Key              | Action            |
|------------------|-------------------|
| `j` / `k`        | Navigate items    |
| `Tab` / `l` / `h` | Switch column    |
| `Enter` / `Space` | Cast vote        |
| `u`              | Remove your vote  |

---

### Discuss

Walk through items in the discussion carousel. View context and action notes side by side. Navigate between items with `p`/`n`, switch lanes with `Tab`, and add notes with `a`.

![Discuss](docs/screenshots/06-discuss.png)

| Key       | Action                    |
|-----------|---------------------------|
| `j` / `k` | Navigate notes            |
| `Tab`     | Switch context/actions     |
| `l` / `h` | Switch context/actions    |
| `p`       | Previous item              |
| `n`       | Next item                  |
| `a`       | Add note to active lane    |

---

### Review

Browse action items and assign owners. The board overview shows all columns side by side.

| Key       | Action          |
|-----------|-----------------|
| `j` / `k` | Navigate items  |
| `o`       | Assign owner    |

---

### Close

View the retro summary: stats, action items with owners, and a full board overview.

---

## Supported templates

The CLI supports all 6 facilitation templates, each with proper column titles and descriptions:

| Template         | Columns                                    |
|------------------|--------------------------------------------|
| Start / Stop     | Stop, Start                                |
| Anchors & Engines | Anchors, Engines                          |
| Mad Sad Glad     | Mad, Sad, Glad                             |
| Four Ls          | Liked, Learned, Lacked, Longed for         |
| KALM             | Keep, Add, Less, More                      |
| Starfish         | Start, More of, Continue, Less of, Stop    |

## How it works

The CLI connects to the same WebSocket server as the web app. Any change made in the CLI appears in the browser and vice versa.

```
Browser (host)  <--- WebSocket --->  Vite dev server  <--- WebSocket --->  CLI (participant)
```

1. Host creates a retro and starts a sync session in the web app
2. CLI joins with `fastretro join <room-code-or-url>`
3. State syncs in real time — cards, votes, notes, stage navigation

## Usage

```bash
# Join by room code
fastretro join ABC-123-DEF

# Join by URL (copies directly from browser)
fastretro join "http://localhost:5173/#room=ABC-123-DEF"

# Custom server
fastretro join ABC-123-DEF --server https://retro.example.com
```

## Development

```bash
make build      # Build binary to ./bin/fastretro
make test       # Run all tests
make cover      # Coverage report (93%+)
make lint       # Go vet
make cover-html # Open coverage in browser
```

### Project structure

```
cmd/fastretro/        CLI entry point (cobra)
internal/
  protocol/           WebSocket message types + facilitation templates
  client/             WebSocket connection manager
  tui/                Bubble Tea views per stage
  styles/             Lip Gloss dark theme
```

### Test pyramid

| Layer      | Focus                                  | Coverage |
|------------|----------------------------------------|----------|
| Protocol   | JSON serialization, templates          | 100%     |
| Client     | Room codes, WebSocket integration      | 88%      |
| TUI        | Views, key handlers, state mutations   | 94%      |
| **Total**  |                                        | **93%**  |

## Requirements

- Go 1.21+
- A running [fastRetro](https://github.com/helmedeiros/fastRetro) instance

## License

MIT
