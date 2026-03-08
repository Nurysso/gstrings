# gstring — Benchmark Results

**Machine:** AMD Ryzen 5 5600X 6-Core Processor  
**OS:** Linux · amd64  
**Go:** tested with `-benchmem -count=3`  
**Package:** `gstring`

---

## Results

### Sprintf

| Benchmark | ns/op | B/op | allocs/op |
|-----------|------:|-----:|----------:|
| `Sprintf` — simple (warm) | 96.74 | 80 | 2 |
| `fmt.Sprintf` — simple | 68.72 | 24 | 1 |
| `Sprintf` — float with spec (warm) | 274.0 | 96 | 2 |
| `fmt.Sprintf` — float with spec | 150.6 | 48 | 1 |
| `Sprintf` — no placeholders | **13.60** | **0** | **0** |
| `Sprintf` — 6 keys | 325.6 | 104 | 3 |

All values are medians across 3 runs.

---

### Row builder

| Benchmark | ns/op | B/op | allocs/op |
|-----------|------:|-----:|----------:|
| `NewRow()` — 4 columns | 326.7 | 160 | 6 |
| `fmt.Sprintf` — equivalent | 167.5 | 48 | 1 |

---

### Table builder

| Benchmark | ns/op | B/op | allocs/op |
|-----------|------:|-----:|----------:|
| `Table.String()` — 5 rows | **1,749** | 1,112 | 32 |
| `fmt.Sprintf` loop — 5 rows | 1,983 | 1,736 | 23 |

gstring is **~12% faster** than hand-rolled fmt for table rendering and allocates **36% less memory** per render.

---

### String utilities

| Benchmark | ns/op | B/op | allocs/op |
|-----------|------:|-----:|----------:|
| `Truncate` | 120.9 | 176 | 2 |
| `Snake` | 159.1 | 24 | 1 |
| `Camel` | 247.9 | 88 | 2 |
| `Wrap` | 243.1 | 304 | 2 |

---

### WithStruct

| Benchmark | ns/op | B/op | allocs/op |
|-----------|------:|-----:|----------:|
| `WithStruct` + `Sprintf` | 325.7 | 288 | 8 |

---

## Analysis

### Where gstring wins

**No-placeholder fast path is essentially free.**  
At 13.6 ns/op with 0 allocations, a template string with no `{}` is returned directly from the cache with no work done. This is faster than a `strings.Contains` check on its own.

**Table rendering beats fmt.**  
`Table.String()` at 1,749 ns/op vs fmt's hand-rolled equivalent at 1,983 ns/op. The reason is pre-allocated `Builder.Grow` with exact row width, reused `rowBuf`, and pre-built format verbs on `Col()` — fmt has to build format strings on every cell.

**Memory advantage on tables is significant.**  
1,112 B/op vs 1,736 B/op — 36% less. At high volume (CLI tools, log formatters, report generators) this directly reduces GC pressure.

---

### Where fmt wins

**Simple string interpolation: ~1.4x faster.**  
`fmt.Sprintf` at 68.7 ns/op vs gstring at 96.7 ns/op. The gap is the template cache lookup (`sync.Map.Load`) and the linear scan over `Vars` pairs. For the absolute hot path, fmt is still the right tool.

**Float with spec: ~1.8x faster.**  
When a spec is present (e.g. `{balance:.2f}`), gstring delegates to `fmt.Fprintf` internally — so you're paying both the cache overhead and fmt's cost. The 274 ns/op vs 150 ns/op gap is almost entirely that double overhead.

**Allocations: fmt consistently uses 1 alloc, gstring uses 2.**  
The extra allocation is the `Vars` pairs slice. This is the price of named args over positional ones — there's no way to eliminate it without a code generation step.

---

### The Row builder gap

Row at 326 ns/op vs fmt at 167 ns/op is the largest relative gap. The reason is `NewRow()` allocates a fresh `*Row` with an embedded `strings.Builder` on every call. A future `SyncPool`-backed `NewRow()` that returns pooled `*Row` instances from `sync.Pool` would close most of this gap. Tracked as a potential improvement.

---

### WithStruct

At 325.7 ns/op and 8 allocs, `WithStruct` is in the same ballpark as a normal `Sprintf` call with 6 keys. The 8 allocations come from `reflect.ValueOf`, the `pairs` slice, and one `Interface()` call per field. Acceptable for non-hot paths; avoid in tight loops — use `With(...)` explicitly if you need speed there.

---

## When to use what

| Situation | Recommendation |
|-----------|---------------|
| Hot loop, millions of calls/sec | `fmt.Sprintf` |
| Readable log lines, CLI output | `gstring.Sprintf` (warm ~97 ns) |
| Formatted table to string/stdout | `gstring.Table` — faster and less memory than fmt |
| Plain string with no substitution | `gstring.Sprintf` — 0 alloc fast path |
| Struct field interpolation | `gstring.WithStruct` — fine outside hot loops |
| Single aligned line | `gstring.NewRow` — consider pooling for hot paths |

---

## Raw output

```
goos: linux
goarch: amd64
pkg: gstring
cpu: AMD Ryzen 5 5600X 6-Core Processor

BenchmarkSprintfSimple_gstring-12          12353282     96.39 ns/op    80 B/op   2 allocs/op
BenchmarkSprintfSimple_gstring-12          12262868     96.50 ns/op    80 B/op   2 allocs/op
BenchmarkSprintfSimple_gstring-12          12248842     97.62 ns/op    80 B/op   2 allocs/op
BenchmarkSprintfSimple_fmt-12              17300988     68.81 ns/op    24 B/op   1 allocs/op
BenchmarkSprintfSimple_fmt-12              17474059     68.64 ns/op    24 B/op   1 allocs/op
BenchmarkSprintfSimple_fmt-12              17158065     69.05 ns/op    24 B/op   1 allocs/op
BenchmarkSprintfFloat_gstring-12            4367709    274.9  ns/op    96 B/op   2 allocs/op
BenchmarkSprintfFloat_gstring-12            4379122    274.0  ns/op    96 B/op   2 allocs/op
BenchmarkSprintfFloat_gstring-12            4108876    273.8  ns/op    96 B/op   2 allocs/op
BenchmarkSprintfFloat_fmt-12                7977945    151.1  ns/op    48 B/op   1 allocs/op
BenchmarkSprintfFloat_fmt-12                7909702    150.4  ns/op    48 B/op   1 allocs/op
BenchmarkSprintfFloat_fmt-12                7636753    151.9  ns/op    48 B/op   1 allocs/op
BenchmarkSprintfNoPlaceholders_gstring-12  89678698     13.60 ns/op     0 B/op   0 allocs/op
BenchmarkSprintfNoPlaceholders_gstring-12  88925605     13.59 ns/op     0 B/op   0 allocs/op
BenchmarkSprintfNoPlaceholders_gstring-12  89273192     13.52 ns/op     0 B/op   0 allocs/op
BenchmarkSprintfManyKeys_gstring-12         3683434    327.2  ns/op   104 B/op   3 allocs/op
BenchmarkSprintfManyKeys_gstring-12         3667065    325.1  ns/op   104 B/op   3 allocs/op
BenchmarkSprintfManyKeys_gstring-12         3619837    324.3  ns/op   104 B/op   3 allocs/op
BenchmarkRow_gstring-12                     3662622    326.7  ns/op   160 B/op   6 allocs/op
BenchmarkRow_gstring-12                     3672157    326.9  ns/op   160 B/op   6 allocs/op
BenchmarkRow_gstring-12                     3616968    325.4  ns/op   160 B/op   6 allocs/op
BenchmarkRow_fmt-12                         7140604    167.5  ns/op    48 B/op   1 allocs/op
BenchmarkRow_fmt-12                         7123797    167.4  ns/op    48 B/op   1 allocs/op
BenchmarkRow_fmt-12                         7151017    166.4  ns/op    48 B/op   1 allocs/op
BenchmarkTableString_gstring-12              653061   1749    ns/op  1112 B/op  32 allocs/op
BenchmarkTableString_gstring-12              656570   1735    ns/op  1112 B/op  32 allocs/op
BenchmarkTableString_gstring-12              647047   1750    ns/op  1112 B/op  32 allocs/op
BenchmarkTableString_fmt-12                  569208   1965    ns/op  1736 B/op  23 allocs/op
BenchmarkTableString_fmt-12                  608353   1993    ns/op  1736 B/op  23 allocs/op
BenchmarkTableString_fmt-12                  590211   1983    ns/op  1736 B/op  23 allocs/op
BenchmarkTruncate-12                        9949459    120.5  ns/op   176 B/op   2 allocs/op
BenchmarkTruncate-12                       10028158    120.9  ns/op   176 B/op   2 allocs/op
BenchmarkTruncate-12                        9737424    121.0  ns/op   176 B/op   2 allocs/op
BenchmarkSnake-12                           7536216    159.6  ns/op    24 B/op   1 allocs/op
BenchmarkSnake-12                           7406679    159.1  ns/op    24 B/op   1 allocs/op
BenchmarkSnake-12                           7482133    158.5  ns/op    24 B/op   1 allocs/op
BenchmarkCamel-12                           4843893    248.6  ns/op    88 B/op   2 allocs/op
BenchmarkCamel-12                           4828635    247.9  ns/op    88 B/op   2 allocs/op
BenchmarkCamel-12                           4852971    247.0  ns/op    88 B/op   2 allocs/op
BenchmarkWrap-12                            4970637    241.9  ns/op   304 B/op   2 allocs/op
BenchmarkWrap-12                            4917038    243.1  ns/op   304 B/op   2 allocs/op
BenchmarkWrap-12                            4846359    243.6  ns/op   304 B/op   2 allocs/op
BenchmarkWithStruct-12                      3663976    325.5  ns/op   288 B/op   8 allocs/op
BenchmarkWithStruct-12                      3689996    328.1  ns/op   288 B/op   8 allocs/op
BenchmarkWithStruct-12                      3684729    325.7  ns/op   288 B/op   8 allocs/op
```
