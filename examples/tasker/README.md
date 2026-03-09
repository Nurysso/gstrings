# tasker

A Dummy CLI task manager written in Go to show case the `gstring` library for all output — tables, named interpolation, string utilities, colors, and the row builder.

## Project structure

```
tasker/
├── go.mod
├── main.go       — CLI entry point, command routing
├── store.go      — Task model + JSON persistence
└── ui.go         — All rendering
```

## Build & run

```bash
go build -o tasker .
```

### Commands

| Command   | Args                               | Description                         |
| --------- | ---------------------------------- | ----------------------------------- |
| `list`    |                                    | List all tasks                      |
| `list`    | `--status=todo\|in-progress\|done` | Filter by status                    |
| `list`    | `--project=<name>`                 | Filter by project                   |
| `add`     | `[title]`                          | Add a task (interactive prompts)    |
| `show`    | `<id>`                             | Full task detail view               |
| `done`    | `<id>`                             | Mark task done                      |
| `start`   | `<id>`                             | Mark task in-progress               |
| `delete`  | `<id>`                             | Delete a task                       |
| `summary` |                                    | Progress stats + project breakdown  |
| `demo`    |                                    | Load demo data and render all views |
| `help`    |                                    | Show help                           |

### Quick demo

```bash
./tasker demo
```

## How gstring is used

### Named interpolation (`Sprintf` / `Println` / `With`)

Every message uses named placeholders instead of positional `%v`:

```go
// In ui.go — success/error/info messages
gstring.Println(
    "Task #{id} created: {title}",
    gstring.With("id", t.ID, "title", t.Title),
)

// Format specs work too
gstring.Println(
    "  Progress  [{bar}] {pct}% complete",
    gstring.With("bar", bar, "pct", pct),
)
```

### Table builder

The task list uses `NewTable()` with `AutoWidth()` to size columns to content:

```go
tbl := gstring.NewTable().
    Col("ID",       4,  gstring.AlignRight).
    Col("S P",      7,  gstring.AlignCenter).
    Col("TITLE",    32, gstring.AlignLeft).
    Col("PROJECT",  12, gstring.AlignLeft).
    Col("PRIORITY", 8,  gstring.AlignLeft).
    Col("DUE",      10, gstring.AlignLeft).
    HeaderColor(gstring.ColorCyan).
    Separator("  ")

tbl.AutoWidth(rows)
tbl.Header()
```

### Row builder

Stats in the summary view are built with the fluent `Row` API:

```go
gstring.NewRow().
    Left("In Progress", 10).
    Sep("·").
    Left(fmt.Sprintf("%d", inProg), 12).
    Print()
```

### String utilities

- `gstring.Truncate(title, 32)` — caps task titles in the list view
- `gstring.Pad(label, 12, ' ')` — aligns detail view labels
- `gstring.Center(title, width, '─')` — section headers
- `gstring.Repeat("█", filled)` — progress bar
- `gstring.Repeat("░", remaining)` — progress bar empty portion
- `gstring.Wrap(t.Notes, 44)` — wraps task notes in the detail view
- `gstring.Strip(input)` — trims user input from prompts

### ANSI colors

Each priority and status has a color mapping used throughout:

```go
func priorityColor(p Priority) gstring.Color {
    switch p {
    case PriorityHigh:   return gstring.ColorRed
    case PriorityMedium: return gstring.ColorYellow
    case PriorityLow:    return gstring.ColorGreen
    }
    return gstring.ColorNone
}
```
