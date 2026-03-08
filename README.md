# gstring

> Readable, self-documenting string formatting for Go.

---

## Why I Built This

Go's `fmt` package is powerful. It can do everything. But after writing enough Go, you start noticing a pattern — formatting strings feels more like writing assembly than writing code.

```go
// this is valid, correct Go
fmt.Printf("%[3]s (%[1]d) -> balance: %[2].2f\n", id, balance, name)
```

You have to mentally map `%[1]`, `%[2]`, `%[3]` back to the argument list. Change the order of args? Update every index. Add a column to a table? Recount every width. Six months later, nobody — including you — knows what `%-12s | %8.2f | %-5t` means at a glance.

Other ecosystems solved this years ago. Python has f-strings. Rust has named captures. Go still has C-style positional verbs.

`gstring` isn't a replacement for `fmt`. It's a layer on top for the cases where readability and maintainability matter more than raw terseness. Named placeholders, a fluent row/table API, and a handful of string utilities that `strings` and `fmt` left out.

---

## Install

```bash
go get github.com/Nurysso/gstring
```

---

## Features

### Named Interpolation

Replace positional `%[1]d` with named `{key}` placeholders. Supports full fmt verb specs after `:`.

```go
gstring.Println(
    "{name} (id:{id:03d}) -> balance: {balance:.2f}",
    gstring.With("id", 1, "name", "Alice", "balance", 1234.5),
)
// Alice (id:001) -> balance: 1234.50
```

| Placeholder      | Equivalent  |
|------------------|-------------|
| `{name}`         | `%v`        |
| `{balance:.2f}`  | `%.2f`      |
| `{id:05d}`       | `%05d`      |

`Sprintf` returns a string. `Println` prints it. `Format` is an alias for `Sprintf`.

---

### Row Builder

Fluent, chainable column alignment for single lines.

```go
gstring.NewRow().
    Left(u.ID, 5).Sep("|").
    Left(u.Name, 12).Sep("|").
    Right(u.Balance, 12, 2).Sep("|").
    Left(u.Active, 6).
    Print()
```

```
1     | Alice        |      1234.50 | true
```

---

### Table Builder

Declare columns once, call `.Row()` for each entry. Optional ANSI colors on headers, columns, and alternating rows.

```go
t := gstring.NewTable().
    HeaderColor(gstring.ColorCyan).
    AltRowColor(gstring.ColorGray)

t.Col("ID", 5, gstring.AlignRight).
    Col("Name", 12, gstring.AlignLeft).
    Col("Balance", 12, gstring.AlignRight, 2).ColColor(gstring.ColorGreen).
    Col("Active", 6, gstring.AlignLeft)

t.Header()
for _, u := range users {
    t.Row(u.ID, u.Name, u.Balance, u.Active)
}
```

Use `t.String(rows)` to get the full table as a string instead of printing it.

---

### String Utilities

Functions `fmt` and `strings` left out:

```go
gstring.Truncate("Hello, World!", 8)   // "Hello..."
gstring.Pad("hi", 10, '-')             // "hi--------"
gstring.Pad("hi", -10, '-')            // "--------hi"
gstring.Center("hi", 10, '-')          // "----hi----"
gstring.Wrap("long text here", 20)     // word-wrapped string
gstring.Repeat("─", 30)                // "──────────────────────────────"
gstring.Strip("  hello  ")             // "hello"
gstring.Title("hello world")           // "Hello World"
gstring.Snake("HelloWorld")            // "hello_world"
gstring.Camel("hello_world")           // "helloWorld"
```

---

## Pros

**Readable format strings.** `{name:.2f}` is immediately obvious. `%[2].2f` is not.

**Maintainable tables.** Column definitions live in one place. Adding or reordering columns doesn't require touching a format string full of magic numbers.

**Named args are refactor-safe.** Reorder your `With(...)` call however you like — the template doesn't care about position.

**Zero dependencies.** Pure stdlib under the hood. No transitive bloat.

**String utilities that should exist.** `Truncate`, `Center`, `Wrap`, `Snake`, `Camel` — things you end up reimplementing in every project anyway.

**`Table.String()` for testability.** Get the full rendered table as a string instead of printing to stdout — useful for logging, testing, or embedding in other output.

**ANSI color support baked in.** Header colors, per-column colors, alternating row colors — without pulling in a dependency.

---

## Cons

**Not a drop-in for `fmt`.** You can't just swap `fmt.Printf` for `gstring.Println` — the template syntax is different and you need to wrap args in `With(...)`.

**Performance overhead on hot paths.** Named interpolation uses regex under the hood. For tight loops formatting millions of strings, `fmt.Sprintf` with positional verbs will be faster. `gstring` optimizes for readability, not throughput.

**No compile-time safety.** A typo in `{nme}` silently passes through as `{nme}` at runtime. Go's `fmt` at least panics loudly on wrong arg counts.

**`With(...)` is variadic pairs, not typed.** It's stringly-typed — flexible, but at the cost of some safety. A future version could explore a typed builder.

**Table API is print-first.** `.Row()` prints immediately. For complex buffering scenarios you'd use `.String(rows)`, but that requires collecting rows upfront.

---

## Comparison

```go
// stdlib
fmt.Printf("%-5d | %-12s | %12.2f | %-5t\n", u.ID, u.Name, u.Balance, u.Active)

// gstring — row builder
gstring.NewRow().Left(u.ID, 5).Sep("|").Left(u.Name, 12).Right(u.Balance, 12, 2).Left(u.Active, 5).Print()

// gstring — named interpolation
gstring.Println("{id:<5} | {name:<12} | {balance:12.2f} | {active:<5}",
    gstring.With("id", u.ID, "name", u.Name, "balance", u.Balance, "active", u.Active))
```

---

## License

MIT
