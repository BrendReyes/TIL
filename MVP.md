# TIL — MVP Specification

> **TIL** is a personal "Today I Learned" tracker: a CLI + TUI tool for capturing things you learn while studying, and reviewing them over time with a simple spaced repetition system.

---

## Overview

When you're deep in a study session and learn something important, you shouldn't break your flow. `til add` lets you capture it in seconds from the terminal. Later, `til review` opens an interactive TUI where you work through everything that's due for a revisit.

---

## Goals for MVP

1. Add a learning entry from the CLI, with optional tags
2. View all entries in a clean list
3. Delete a specific entry by ID
4. Edit an entry's text
5. Review due entries via a simple spaced repetition TUI

---
## Tech Stack

- GO
- Cobra
- Bubbletea, bubbles, lipgloss
- sqlite, sqlc
- Goose migration


---

## Spaced Repetition Algorithm (Simple)

- review things that haven't been reviewed or was long time last reviewed

---

## CLI Commands

### `til add`

```bash
til add "In Go, defer runs LIFO — last deferred call runs first"
til add "Postgres EXPLAIN ANALYZE shows actual row counts" --tag postgres --tag sql
til add "BFS uses a queue, DFS uses a stack" -t algorithms
```

Flags:
- `--tag` / `-t` (repeatable) — attach one or more tags

Output:
```
✓ Added entry #42  [go]
```

---

### `til list`

```bash
til list
```

Renders a table to stdout:

```
 ID   Body                                              Tags           Created
 ──   ────────────────────────────────────────────────  ─────────────  ────────────
 42   In Go, defer runs LIFO — last deferred...         go             2 hours ago
 41   Postgres EXPLAIN ANALYZE shows actual row...      postgres, sql  yesterday
 40   BFS uses a queue, DFS uses a stack                algorithms     3 days ago
```

Long bodies are truncated to ~60 chars in list view; full text shown in review.

---

### `til delete`

```bash
til delete 42
```

Output:
```
Delete entry #42: "In Go, defer runs LIFO — last deferred call runs first"? [y/N]
✓ Deleted.
```

Prompts for confirmation before deleting. No undo.

---

### `til edit`

```bash
til edit 42
```

Behavior:
1. Fetches the entry body
2. Opens it in TUI editor, then saves if user confirms


Output:
```
✓ Updated entry #42
```

---

### `til review`

```bash
til review
```

Launches the Bubbletea TUI. Shows how many entries are due:

```
  Review Session — 6 entries due
  ────────────────────────────────
  In Go, defer runs LIFO — last deferred
  call runs first.

  Tags: go               Created: 3 days ago

  [enter or space] Mark reviewed    [s] Skip    [q] Quit
```

Controls:
- `space or enter` — mark as reviewed (bumps interval), advance to next
- `s` — skip (leaves interval unchanged), advance to next
- `q` / `ctrl+c` — quit session (only reviewed entries are updated)

On completion:
```
  ✓ Session complete. Reviewed 4, skipped 2.
  Next due: tomorrow
```

If nothing is due:
```
  ✓ You're all caught up! Nothing due for review.
  Next entry due: in 3 days (entry #41)
```

---


## Out of Scope for MVP but nice for later

1. Search 
2. Reminders | "You have N entries overdue"
3. Export / import | JSON dump of `entries` + `tags` is straightforward |
4. Filter / sort / pagination
5. Tag selection menu (need this for consistent sorting and filtering)

---

# Build Plan (Milestones)
1. [x] Download and import the necessary tools and libraries for cli and tui
2. [x] Initialize SQLite
3. [x] Create the table schema using goose migrations
4. [x] Can add entry (basic)
5. [x] Can view all entries
6. [x] Can edit using tui bubbletea
7. [x] Can review each, bubbletea and cli
8. [x] view specific entry 
9. [] multiple delete flags / delete all
10.[] search by tags
10. [] Implement tui as a whole 
