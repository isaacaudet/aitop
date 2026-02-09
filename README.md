<div align="center">
<br>

# clawdtop

**`htop` for your AI coding spend.**

You're burning tokens across Claude Code, Cursor, Gemini, and Codex every day.<br>
clawdtop reads data already on your machine and shows you what it costs — beautifully.

<br>

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](http://makeapullrequest.com)

[Install](#install) · [Features](#features) · [Configuration](#configuration) · [How It Works](#how-it-works)

</div>

<br>

> **No API keys. No cloud. No telemetry. Just `go install` and see the truth.**

<br>

<!-- Replace with actual screenshot: -->
<!-- ![clawdtop dashboard](assets/dashboard.png) -->

## Install

**Homebrew:**

```bash
brew install isaacaudet/tap/clawdtop
```

**Go:**

```bash
go install github.com/isaacaudet/clawdtop@latest
```

**From source:**

```bash
git clone https://github.com/isaacaudet/clawdtop.git
cd clawdtop && make install
```

> Requires `CGO_ENABLED=1` for Cursor's SQLite database. If you don't use Cursor, `CGO_ENABLED=0` works fine.

Then just run:

```bash
clawdtop
```

## Features

### Dashboard

The home view. Four stat cards with sparklines for Today / This Week / This Month / All Time. Daily usage bar chart, model cost breakdown with aligned bars, burn rate projections, and hourly activity heatmap.

### Sessions

Every AI coding session across all providers in one table. Sort by cost to find the expensive ones. Expand any session for full token/model detail. Grouped by project with cost subtotals.

### Providers

Per-provider deep dive. Model tables, cost distribution bars, generation counts. Toggle between time periods to compare this month vs. all time.

### Heatmap

GitHub-style contribution heatmap — 16 weeks of history with separate views for cost and message volume. Log-scale coloring so you see gradients, not just spikes.

### Live

Real-time activity timeline. Each row is a provider, each column is a time bucket. Intensity shows token usage. Auto-refreshes every 5 seconds. Scroll back in time with `h`/`l`.

### Non-Interactive Mode

Pipe-friendly for scripts and dashboards:

```bash
$ clawdtop summary

clawdtop — AI Usage Summary
═══════════════════════════════════════════

  ◈ Claude Code   $4,152.50   2,007 sessions   4.9B tokens
  ⌘ Cursor         6,021 generations
  ✦ Gemini         $2.33       4 sessions       3.2M tokens
  ⊡ Codex          31 sessions                  759.4M tokens

  Grand Total: $4,154.83 across 4 providers
```

## How It Works

clawdtop reads data from files that already exist on your machine. Nothing leaves your system.

| Provider | Source | What You See |
|----------|--------|--------------|
| **Claude Code** | `~/.claude/stats-cache.json` + `~/.claude/projects/*.jsonl` | Token breakdown by model, cost per session, cache hit rates |
| **Cursor** | `~/.cursor/ai-tracking/ai-code-tracking.db` | Code generations by file type, conversation history |
| **Gemini CLI** | `~/.gemini/tmp/*/chats/session-*.json` | Input/output/cached tokens, cost per session |
| **Codex** | `~/.codex/sessions/YYYY/MM/DD/rollout-*.jsonl` | Token counts, reasoning tokens, rate limit usage |

Cost estimates use current API pricing:

| Model | Input | Output |
|-------|-------|--------|
| Claude Opus 4.5 / 4.6 | $5/MTok | $25/MTok |
| Claude Sonnet 4.5 | $3/MTok | $15/MTok |
| Claude Haiku 4.5 | $0.80/MTok | $4/MTok |
| GPT-4o | $2.50/MTok | $10/MTok |
| o3 | $10/MTok | $40/MTok |
| Gemini 2.5 Pro | $1.25/MTok | $10/MTok |
| Gemini 2.5 Flash | $0.30/MTok | $2.50/MTok |

Cache read/write pricing included where applicable.

## Keybindings

| Key | Action |
|-----|--------|
| `1`–`5` | Jump to view |
| `tab` | Cycle views |
| `j` / `k` | Scroll |
| `ctrl+u` / `ctrl+d` | Half-page scroll |
| `s` | Cycle sort order |
| `t` | Cycle time period |
| `r` | Refresh data |
| `?` | Help |
| `q` | Quit |

## Configuration

Optional. Create `~/.config/clawdtop/config.toml`:

```toml
# Your subscription plan (shown in the usage banner)
[plan]
provider = "claude"
name = "Max"
monthly_cost = 200

# Custom data paths (defaults shown)
stats_cache_path = "~/.claude/stats-cache.json"
projects_dir = "~/.claude/projects"
```

The plan banner shows `Max $200/mo — $153.28 (77%)` — green under 70%, yellow 70–90%, red above 90%.

## Built With

[Bubble Tea](https://github.com/charmbracelet/bubbletea) ·
[Lip Gloss](https://github.com/charmbracelet/lipgloss) ·
[Bubbles](https://github.com/charmbracelet/bubbles) ·
[Catppuccin Mocha](https://github.com/catppuccin/catppuccin)

## License

[MIT](LICENSE)
