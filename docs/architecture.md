# gstring — Architecture & Design Decisions

This document explains the internal structure of gstring, why each component was written the way it was, and the tradeoffs made at each decision point.

---

## Overview

gstring is organized into four independent layers:

```
┌─────────────────────────────────────────────────────┐
│                  Public API                         │
│  Sprintf · Println · Format · With · WithStruct     │
│  NewRow · NewTable · Truncate · Snake · Camel ...   │
├─────────────────────────────────────────────────────┤
│              Template Engine                        │
│  parse() · token · templateCache (sync.Map)         │
├─────────────────────────────────────────────────────┤
│              Value Serialization                    │
│  writeValue() · valueToString()                     │
├─────────────────────────────────────────────────────┤
│              Layout Engine                          │
│  Row (strings.Builder) · Table (column · rowBuf)    │
└─────────────────────────────────────────────────────┘
```

Each layer is independent. The template engine doesn't know about tables. The layout engine doesn't know about interpolation. This keeps each part testable and replaceable in isolation.

---

## Vars: flat slice instead of map

```go
type Vars struct {
    pairs []any // [k0, v0, k1, v1, ...]
}
```

The first decision anyone would question: why not `map[string]any`?

A map lookup in Go costs a hash computation, a bucket probe, and a potential cache miss. For the typical format string with 2-6 named keys, a linear scan over a `[]any` slice is faster because the entire slice fits in one or two cache lines and the CPU's prefetcher handles it trivially.

The crossover point where a map starts winning is around 10-12 keys. Format strings almost never have that many. At 6 keys, linear scan is measurably faster. At 12, it's roughly even. Beyond 12, use a map.

`With(pairs ...any)` takes the variadic directly as the pairs slice — no copy, no allocation beyond what the caller already made for the variadic. This is intentional: `gstring.With("a", 1, "b", 2)` compiles to a single stack-allocated array in many cases.

---

## Template cache: sync.Map over regex

```go
var templateCache sync.Map

func parse(template string) []token {
    if v, ok := templateCache.Load(template); ok {
        return v.([]token)
    }
    // ... parse once, store, return
}
```

The original implementation used `regexp.MustCompile` and `ReplaceAllStringFunc`. The problem: even with a pre-compiled regex, `ReplaceAllStringFunc` allocates a closure, walks the entire string with the DFA, and calls the closure once per match. For a template called in a loop, you pay that cost every time.

The cache trades a small amount of memory (one `[]token` per unique template string seen) for eliminating all parsing work on the hot path. `sync.Map` is used instead of a `RWMutex`-protected `map` because `sync.Map` is optimized for the "write once, read many" pattern — which is exactly what a template cache is.

The `Load` path in `sync.Map` is lock-free on the hot path (after the key is stored). Concurrent readers never block each other, which matters for the parallel benchmark results.

### Token structure

```go
type token struct {
    kind tokenKind // tokLiteral or tokPlaceholder
    text string    // literal text, or the key name
    spec string    // fmt spec, e.g. ".2f" (empty → use writeValue)
}
```

The parse result is a flat `[]token` rather than a tree. Format strings don't need trees — they're linear. A flat slice is faster to allocate, faster to iterate, and simpler to reason about.

### No-placeholder fast path

```go
if strings.IndexByte(template, '{') == -1 {
    toks := []token{{kind: tokLiteral, text: template}}
    templateCache.Store(template, toks)
    return toks
}
```

`strings.IndexByte` is implemented as a SIMD scan in the Go runtime — it checks 16 or 32 bytes at a time. For a template with no `{`, it returns -1 in a few nanoseconds without any allocation. The resulting single-token result gets a `Sprintf` fast path:

```go
if len(toks) == 1 && toks[0].kind == tokLiteral {
    return toks[0].text
}
```

Zero allocations, zero work. The benchmark shows 13.6 ns/op, 0 B/op for this path.

---

## writeValue: strconv over fmt

```go
func writeValue(b *strings.Builder, val any) {
    switch v := val.(type) {
    case string:  b.WriteString(v)
    case int:     b.WriteString(strconv.Itoa(v))
    case float64: b.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
    case bool:    if v { b.WriteString("true") } else { b.WriteString("false") }
    // ...
    default:      fmt.Fprintf(b, "%v", v)
    }
}
```

`fmt.Sprintf("%v", val)` goes through reflection, format string parsing, and a general-purpose value printer on every call. For common types — `string`, `int`, `float64`, `bool` — we can skip all of that.

`strconv.Itoa` and `strconv.FormatInt` write directly to a byte buffer with no intermediate allocation. `strings.Builder.WriteString` does a single `copy` into the builder's buffer. The entire path for an `int` is: type switch hit → `strconv.Itoa` → `WriteString` → done.

The `default` case falls back to `fmt.Fprintf` for unusual types (`net.IP`, custom structs, etc.). This is correct and safe — the fast path handles 95%+ of real usage and the fallback handles the rest.

The same logic applies to `valueToString` used by the Row and Table builders — identical type switch, returns a string instead of writing to a builder.

---

## Row: embedded strings.Builder

```go
type Row struct {
    b strings.Builder
}

func NewRow() *Row { return &Row{} }

func (r *Row) Left(val any, width int) *Row {
    s := valueToString(val)
    r.b.WriteString(s)
    pad := width - utf8.RuneCountInString(s)
    for ; pad > 0; pad-- {
        r.b.WriteByte(' ')
    }
    return r
}
```

The `strings.Builder` is embedded directly in the `Row` struct rather than held as a pointer. This means `NewRow()` allocates one object — the `Row` — and the builder's internal buffer grows inside it. No separate heap allocation for the builder.

`Left` writes the value string and then writes pad bytes directly to the builder in a loop. This avoids the double allocation of the old approach (`fmt.Sprintf("%-Nv", val)` which allocates a format string, then allocates the result string, then copies it into the builder). Now it's: `valueToString` → `WriteString` → loop `WriteByte`. One allocation if `valueToString` allocates (for non-string types), zero otherwise.

**Why `NewRow()` still allocates on every call:**
A `*Row` is returned so callers can chain methods. The alternative — a stack-allocated `Row` — doesn't work with method chaining in Go because the receiver would escape to heap anyway. A `sync.Pool` of `*Row` objects is the right fix but adds API complexity (callers would need to `Put` the row back). Left as a future improvement.

---

## Table: pre-built verbs and reused rowBuf

```go
type column struct {
    verbLeft  string  // pre-built, e.g. "%-12.2f"
    verbRight string  // pre-built, e.g. "%12.2f"
    // ...
}
```

The old `formatCell` built the format verb string (`"%-12.2f"`) on every single cell of every single row. For a 100-row table with 4 columns, that's 400 `fmt.Sprintf` calls just to construct verb strings, before any actual formatting happens.

The fix: build the verbs once when `Col()` is called, store them in the `column` struct, reuse forever. `Col()` is called once at setup time. `Row()` is called N times. Moving work from `Row()` to `Col()` is almost always the right trade.

```go
type Table struct {
    rowBuf []string  // reused across Row() calls
    rowWidth int     // pre-calculated exact width
}
```

`rowBuf` is allocated once in `rebuildMetrics()` when columns are added. Every `Row()` call writes into the same slice — no `make([]string, n)` per row. For a 100-row table this eliminates 100 slice allocations.

`rowWidth` is the exact byte count of a rendered row (sum of column widths + separators). `joinRow` and `String` use it to `Grow` the builder to the right size before writing anything. A correctly-sized `Grow` means zero reallocations during the write phase.

### Why String() allocates a separate buf instead of using rowBuf

```go
func (t *Table) String(rows [][]any) string {
    buf := make([]string, len(t.cols))  // local, not t.rowBuf
    // ...
}
```

`String()` is safe to call concurrently with other `String()` calls on the same table. If it used `t.rowBuf`, concurrent calls would race on the shared slice. The local `buf` is stack-allocated in many cases (small column counts) and always goroutine-local. `Row()` uses `t.rowBuf` because `Row()` is inherently a sequential print operation.

---

## AutoWidth: single scan, then rebuild

```go
func (t *Table) AutoWidth(rows [][]any) {
    for colIdx := range t.cols {
        max := utf8.RuneCountInString(t.cols[colIdx].header)
        for _, row := range rows {
            if colIdx < len(row) {
                s := valueToString(row[colIdx])
                if l := utf8.RuneCountInString(s); l > max {
                    max = l
                }
            }
        }
        t.cols[colIdx].width = max
    }
    // rebuild verbs and metrics after all widths are set
    for i := range t.cols {
        c := &t.cols[i]
        if c.precision >= 0 {
            c.verbLeft  = fmt.Sprintf("%%-%d.%df", c.width, c.precision)
            c.verbRight = fmt.Sprintf("%%%d.%df", c.width, c.precision)
        }
    }
    t.rebuildMetrics()
}
```

The scan is O(rows × cols) — unavoidable since you have to look at every cell to know the max width. The verb rebuild and `rebuildMetrics` happen once after all widths are determined, not once per column. This keeps the constant factors small.

`valueToString` is used instead of `fmt.Sprintf("%v", ...)` for the same reason as everywhere else — the strconv fast path avoids allocations for common types during the scan.

---

## WithStruct: reflect once, scan after

```go
func WithStruct(s any) Vars {
    val := reflect.Indirect(reflect.ValueOf(s))
    typ := val.Type()
    // ...
    for i := 0; i < n; i++ {
        f := typ.Field(i)
        if f.PkgPath == "" {
            pairs = append(pairs, f.Name, val.Field(i).Interface())
        }
    }
    return Vars{pairs: pairs}
}
```

`reflect.Indirect` dereferences a pointer transparently — both `User{}` and `&User{}` work. `f.PkgPath == ""` is the canonical Go check for exported fields: unexported fields have a non-empty package path.

The result is a `Vars` with the same flat-slice layout as `With(...)`, so `Sprintf` doesn't know or care whether the vars came from a manual `With` or from `WithStruct`. The lookup path is identical.

The reflection cost is paid once per `WithStruct` call. If you're calling `WithStruct` in a loop over many structs of the same type, the `typ.Field(i)` calls hit the reflect type cache after the first call so subsequent calls are cheaper. Still not free — use `With(...)` if you're in a tight loop and know your field names at compile time.

---

## Unicode correctness: utf8.RuneCountInString

```go
pad := width - utf8.RuneCountInString(s)
```

Column width is measured in runes, not bytes. A Chinese character is 3 bytes but 1 rune and occupies 1 terminal column. `len(s)` would give the wrong answer for any non-ASCII string.

`utf8.RuneCountInString` counts runes in O(n) time — it's a single pass over the string. There's no cheaper correct option in Go without unsafe code.

**Known limitation:** terminal column width for CJK characters (wide characters that occupy 2 columns) is not handled. Correct wide-character support would require `golang.org/x/text/width` which would add a dependency. For now, ASCII and single-width Unicode are correct; CJK alignment will be off by approximately the number of wide characters in the cell.

---

## What was deliberately not added

**Indexed placeholders `{0}`, `{1}`**
Named args are the entire value proposition of gstring. Adding indexed placeholders would make templates ambiguous (is `{1}` a key named "1" or the second positional arg?) and recreate the readability problems of `fmt`.

**Error returns on Sprintf**
Missing keys return the original placeholder text. Returning `(string, error)` would make every call site require error handling for what is almost always a programmer mistake caught in development. The pass-through behavior makes templates self-documenting in their failure mode.

**Compile-time template validation**
This would require either code generation or a custom vet pass. Both are significant scope increases. The template cache means a typo surfaces immediately on first run, which is good enough for the intended use case.

**sync.Pool for Row**
Correct but adds API surface — callers would need to return rows to the pool. Left as a future optimization with a clear path: `AcquireRow() / ReleaseRow()` pair, similar to `fasthttp`'s approach.
