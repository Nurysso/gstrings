# gstring

> Readable, self-documenting string formatting for Go.

---

## Why I Built This

Go's `fmt` package is powerful. It can do everything. But after writing enough Go, you start noticing a pattern — formatting strings feels more like writing assembly than writing code.

```go
// this is valid, correct Go — but what does it even mean?
fmt.Printf("%[3]s (%[1]d) -> balance: %[2].2f\n", id, balance, name)
```

You have to mentally map `%[1]`, `%[2]`, `%[3]` back to the argument list every single time you read it. Change the order of args? Update every index by hand. Add a column to a table? Recount every magic width number. Come back six months later? Nobody — including you — knows what `%-12s | %8.2f | %-5t` means at a glance.

Other ecosystems solved this years ago. Python has f-strings. Rust has named format args. Go still ships C-style positional verbs.

`gstring` isn't a replacement for `fmt`. It's a thin layer on top for the times when readability matters more than terseness — named placeholders, a fluent table/row API, and a handful of string utilities that `strings` and `fmt` just never got around to including.

> And if you are wondering why I named it gstring, then these are the reasonse g-stings as in GO strings, python had f strings and i wanted this to be similare in some ways so picked the next letter of the english alphabet and there is clearly no other reason I swear.

---

## Install

```bash
go get github.com/nurysso/gstring
```

```go
import "github.com/nurysso/gstring"
```

---

## Features

### Named Interpolation

No more `%[2]d`. Just write `{name}` and pass your values by name.

```go
gstring.Println(
    "{name} (id:{id:03d}) -> balance: {balance:.2f}",
    gstring.With("id", 1, "name", "Alice", "balance", 1234.5),
)
// Alice (id:001) -> balance: 1234.50
```

Full Go `fmt` verb specs work after the colon — anything you'd put after `%` in `fmt.Sprintf` works here too.

| Placeholder     | Same as |
| --------------- | ------- |
| `{name}`        | `%v`    |
| `{balance:.2f}` | `%.2f`  |
| `{id:05d}`      | `%05d`  |

`Sprintf` returns a string. `Println` prints it. `Format` is just an alias for `Sprintf`.

Missing keys pass through unchanged — `{typo}` stays `{typo}`, no panic.

---

### WithStruct

Don't want to manually write out all your field names? Pass your struct directly.

```go
type User struct {
    ID      int
    Name    string
    Balance float64
}

gstring.Println(
    "{Name} (id:{ID:03d}) -> {Balance:.2f}",
    gstring.WithStruct(user),
)
// Alice (id:001) -> 1234.50
```

Exported fields only. Pointer structs work too.

---

### Row Builder

Fluent, chainable API for building a single aligned line. No format strings, no magic numbers to remember.

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

`.String()` gives you the result without printing if you need it.

---

### Table Builder

Declare your columns once, render as many rows as you want. ANSI colors on headers, specific columns, and alternating rows are all built in.

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

```
ID    | Name         |      Balance | Active
----- | ------------ | ------------ | ------
    1 | Alice        |      1234.50 | true
    2 | Bob          |        98.12 | false
    3 | Charlotte    |    100000.99 | true
```

Hate counting column widths manually? `AutoWidth` scans your data and figures it out.

```go
t.Col("ID", 0, gstring.AlignRight).
    Col("Name", 0, gstring.AlignLeft)

t.AutoWidth(rows) // done, no manual counting
t.Header()
```

`t.String(rows)` returns the whole table as a string — useful for tests or logging.

---

### String Utilities

Functions `fmt` and `strings` never bothered to include:

```go
gstring.Truncate("Hello, World!", 8)   // "Hello..."
gstring.Pad("hi", 10, '-')             // "hi--------"
gstring.Pad("hi", -10, '-')            // "--------hi"
gstring.Center("hi", 10, '-')          // "----hi----"
gstring.Wrap("long sentence here", 10) // word-wrapped at 10 chars
gstring.Repeat("─", 30)                // "──────────────────────────────"
gstring.Strip("  hello  ")             // "hello"
gstring.Title("hello world")           // "Hello World"
gstring.Snake("HelloWorld")            // "hello_world"
gstring.Camel("hello_world")           // "helloWorld"
```

---

## Benchmarks

Tested on AMD Ryzen 5 5600X · Linux(Arch BTW, I just have to lol) · amd64.

| What                        | gstring                | fmt                |
| --------------------------- | ---------------------- | ------------------ |
| Simple interpolation (warm) | 97 ns                  | 69 ns              |
| Float with spec (warm)      | 275 ns                 | 151 ns             |
| No placeholders             | **14 ns / 0 alloc**    | —                  |
| Table (5 rows, to string)   | **1,749 ns / 1,112 B** | 1,983 ns / 1,736 B |

The no-placeholder fast path is zero-allocation — if there's no `{` in the template, gstring returns the string immediately without doing anything.

Table rendering beats hand-rolled `fmt` loops because column widths and format verbs are computed once at setup, not rebuilt on every row.

Full benchmark results in [BENCHMARKS.md](BENCHMARKS.md).

---

## Pros

**Readable format strings.** `{name:.2f}` tells you exactly what's happening. `%[2].2f` tells you nothing without scrolling to the arg list.

**Maintainable tables.** Column definitions live in one place. Reordering or adding columns is a one-line change, not a find-and-replace across multiple format strings.

**Refactor-safe named args.** Reorder your `With(...)` call however you like — the template doesn't care about position.

**Zero dependencies.** Pure stdlib under the hood. No transitive bloat, no `go.sum` surprises.

**String utilities that should exist.** `Truncate`, `Center`, `Wrap`, `Snake`, `Camel` — things you end up copy-pasting into every project anyway.

**`Table.String()` for testing.** Capture table output as a string and assert on it in tests, or pipe it into a logger. No stdout hijacking needed.

**ANSI colors without a dep.** Header colors, per-column colors, alternating row colors — all built in, no extra package.

**Template caching.** Each unique format string is parsed once. Repeat calls skip all parsing and go straight to substitution.

---

## Cons

**Not a drop-in for `fmt`.** The template syntax is different and you need to wrap args in `With(...)`. There's no migration path, you adopt it for new code.

**Slower than fmt on raw interpolation.** For hot loops formatting millions of strings a second, `fmt.Sprintf` is still faster. gstring trades a bit of speed for a lot of readability.

**No compile-time safety.** A typo in `{nme}` silently passes through as `{nme}` at runtime. `fmt` at least panics loudly on arg count mismatches.

**`With(...)` is stringly-typed.** Flexible, but you won't get a compile error if you swap a key name. A typed builder API is something I might explore later.

---

## Comparison

```go
// stdlib fmt — works, but what does it mean?
fmt.Printf("%-5d | %-12s | %12.2f | %-5t\n", u.ID, u.Name, u.Balance, u.Active)

// gstring row builder — reads left to right, self-explanatory
gstring.NewRow().Left(u.ID, 5).Sep("|").Left(u.Name, 12).Right(u.Balance, 12, 2).Left(u.Active, 5).Print()

// gstring named interpolation — looks like every other modern language
gstring.Println(
    "{id:<5} | {name:<12} | {balance:12.2f} | {active:<5}",
    gstring.With("id", u.ID, "name", u.Name, "balance", u.Balance, "active", u.Active),
)
```

For a full comparison against `pyfmt`, `fstr`, and `stringFormatter` see [examples/comparison.go](examples/comparison.go).

---

## Docs

| File                               | What's in it                                 |
| ---------------------------------- | -------------------------------------------- |
| [USAGE.md](USAGE.md)               | Full API reference with examples and recipes |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Why things are implemented the way they are  |
| [BENCHMARKS.md](BENCHMARKS.md)     | Benchmark results with analysis              |

---

> [!NOTE]
> I've only been writing Go for about 6 months. Some decisions in here probably make experienced Go devs raise an eyebrow and that's fair.
> This started as a fun project to scratch my own itch and to learn more about go.
> The code isnt 100% me I used Claude for performance optimizations (the template cache, strconv fast paths) and gemini to help edit the docs. If you spot something that could be done better the Go way, open an issue, I'm genuinely trying to learn.

---

## [LICENSE](LICENSE)
