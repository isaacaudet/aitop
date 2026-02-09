<div align="center">

# aitop

### `htop` for your AI coding spend

**See exactly how much Claude, Cursor, Gemini, and Codex cost you.**<br>
**In a beautiful terminal dashboard. Right now.**

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=for-the-badge)](http://makeapullrequest.com)

<br>

```
  aitop — AI Usage Dashboard
  ◈ Claude Code  ✦ Gemini  ⊡ Codex  ⌘ Cursor
  Max $200/mo — $153.28 (77%)

  ╭────────────────╮╭────────────────╮╭────────────────╮╭────────────────╮
  │ Today          ││ This Week      ││ This Month     ││ All Time       │
  │ $2.55          ││ $36.22         ││ $153.28        ││ $4,152.50      │
  │ 101.8K tokens  ││ 1.4M tokens    ││ 6.2M tokens    ││ 4.9B tokens    │
  │ ▁▂▃▄▅▆▇█▅▃▂▁   ││ ▂▄▆█▇▅▃▁       ││ ▁▂▃▄▅▆▇██▆▃▁   ││ ▁▂▃▅▇█▇▅▃▂▁    │
  │ 5,071 msgs     ││ 73,983 msgs    ││ 173,821 msgs   ││ 173,821 msgs   │
  ╰────────────────╯╰────────────────╯╰────────────────╯╰────────────────╯

  Daily Token Usage (Last 30 Days)
   45M █
       █     █                       █
       █   █ █ █               █     █ █
       █ █ █ █ █   █   █     █ █   █ █ █ █   █
       █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █
     0 ████████████████████████████████████████████████
       05 08 11 14 17 20 23 26 29 01 04 07

  Model Cost Breakdown
   opus-4-5          ████████████████████████████████░░  $3,395.66
  opus-4-6          █████░░░░░░░░░░░░░░░░░░░░░░░░░░░░  $614.10
  sonnet-4-5        █░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  $137.98
```

</div>

---

## Why

You're burning tokens across Claude Code, Cursor, Gemini, and Codex every day. Do you know what they cost?

- **Claude Code** burns through tokens fast — cache writes alone can cost hundreds
- **Cursor** tracks generations but hides the bill
- **Gemini CLI** and **Codex** have zero cost visibility
- You probably have a $200/mo plan and no idea if you're over

**aitop reads data already on your machine** — no API keys, no cloud, no telemetry. Just `go install` and see the truth.

## Install

```bash
go install github.com/isaacaudet/aitop@latest
```

Or build from source:

```bash
git clone https://github.com/isaacaudet/aitop.git
cd aitop && make install
```

> Requires `CGO_ENABLED=1` for Cursor's SQLite database. If you don't use Cursor, `CGO_ENABLED=0` works fine.

Then just run:

```bash
aitop
```

## What It Reads

| Provider | Data Source | What You See |
|----------|-----------|--------------|
| **Claude Code** | `~/.claude/stats-cache.json` + `~/.claude/projects/*.jsonl` | Full token breakdown, cost per model, per-session detail |
| **Cursor** | `~/.cursor/ai-tracking/ai-code-tracking.db` | Code generations by file type, conversation history |
| **Gemini CLI** | `~/.gemini/tmp/*/chats/session-*.json` | Input/output/cached tokens, cost per session |
| **Codex** | `~/.codex/sessions/YYYY/MM/DD/rollout-*.jsonl` | Token counts, reasoning tokens, rate limit usage |

**Nothing leaves your machine. No API calls. No telemetry. Read-only.**

## Views

### 1. Dashboard

The home screen. Four stat boxes with sparklines, daily usage chart, model cost breakdown with aligned bars, burn rate projections, and hourly activity patterns.

```
  tab: views | 1-5: jump | j/k: scroll | s: sort | t: period | q: quit
```

### 2. Sessions

Every session across all providers. Sort by cost to find the expensive ones.

| Key | Action |
|-----|--------|
| `s` | Cycle sort: date -> cost -> tokens -> duration |
| `enter` | Expand session detail |
| `esc` | Collapse |

Includes project grouping with cost subtotals.

### 3. Providers

Per-provider breakdown with model tables, cost distribution bars, and generation counts. Filter by time period to see this month vs all time.

| Key | Action |
|-----|--------|
| `t` | Cycle: All Time -> This Month -> This Week -> Today |

### 4. Heatmap

GitHub-style activity heatmap — 16 weeks of history with separate views for cost and message volume. Uses log-scale coloring so you can see gradients, not just spikes.

### 5. Live

Real-time activity timeline. Each row is a provider, each column is a time bucket. Intensity shows token usage. Auto-refreshes every 5 seconds.

| Key | Action |
|-----|--------|
| `h` | Scroll back in time |
| `l` | Scroll forward / return to now |

## Navigation

| Key | Action |
|-----|--------|
| `1`-`5` | Jump to view |
| `tab` | Cycle views |
| `j`/`k` | Scroll |
| `ctrl+u`/`ctrl+d` | Half-page scroll |
| `r` | Refresh data |
| `?` | Help |
| `q` | Quit |

## Configuration

Optional. Create `~/.config/aitop/config.toml`:

```toml
# Custom data paths (defaults shown)
stats_cache_path = "~/.claude/stats-cache.json"
projects_dir = "~/.claude/projects"

# Your subscription plan (for the usage banner)
[plan]
provider = "claude"
name = "Max"
monthly_cost = 200
```

The plan banner shows: `Max $200/mo — $153.28 (77%)` — green under 70%, yellow 70-90%, red above 90%.

## Non-Interactive Mode

Pipe-friendly summary for scripts and dashboards:

```bash
$ aitop summary

aitop — AI Usage Dashboard
═════════════════════════════════════════════════

  ◈ Claude Code  $4,152.50  2,007 sessions  4.9B tokens
  ⌘ Cursor  6,021 generations
  ✦ Gemini  $2.33  4 sessions  3.2M tokens
  ⊡ Codex  31 sessions  759.4M tokens

  Grand Total: $4,154.83 across 4 providers
```

## Pricing

aitop uses current API pricing to calculate costs:

| Model | Input | Output |
|-------|-------|--------|
| Claude Opus 4.5/4.6 | $5/MTok | $25/MTok |
| Claude Sonnet 4.5 | $3/MTok | $15/MTok |
| Claude Haiku 4.5 | $0.80/MTok | $4/MTok |
| GPT-4o | $2.50/MTok | $10/MTok |
| o3 | $10/MTok | $40/MTok |
| Gemini 2.5 Pro | $1.25/MTok | $10/MTok |
| Gemini 2.5 Flash | $0.30/MTok | $2.50/MTok |

Cache read/write pricing included for models that support it.

## Built With

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — terminal UI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — styling
- [Bubbles](https://github.com/charmbracelet/bubbles) — components
- [Catppuccin Mocha](https://github.com/catppuccin/catppuccin) — color palette

## License

MIT
