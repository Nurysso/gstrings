// comparison.go
//
// Side-by-side comparison of Go string formatting approaches:
//   - stdlib fmt
//   - github.com/slongfield/pyfmt   (Python .format() style, positional/named via struct/map)
//   - github.com/ZiadMansourM/fstr  (Python f-string style, named map)
//   - github.com/Wissance/stringFormatter (indexed {0},{1} + named map)
//   - gstring                        (this lib)
//
// Run the gstring examples directly:
//
//	go run examples/comparison.go
package main

import (
	"fmt"
	"strings"
	"gstring"
)

// ─── Shared data ──────────────────────────────────────────────────────────────

type User struct {
	ID      int
	Name    string
	Balance float64
	Active  bool
}

var users = []User{
	{1, "Alice", 1234.5, true},
	{2, "Bob", 98.12, false},
	{3, "Charlotte", 100000.99, true},
}

func divider(label string) {
	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println(" " + label)
	fmt.Println(strings.Repeat("─", 60))
}

// ─── 1. stdlib fmt ───────────────────────────────────────────────────────────
//
// The baseline. Positional verbs, powerful but opaque.
//
// Pros:
//   - zero deps, built-in, compile-time vet checks
//   - fastest — no extra allocations
//   - full verb support (%v, %d, %f, %T, %p, etc.)
//
// Cons:
//   - positional args break silently when reordered
//   - %[n] indexed notation is hard to read at a glance
//   - no named args — duplicate values must be passed twice
//   - table formatting is a wall of magic numbers with no structure

func exampleFmt() {
	divider("1. stdlib fmt")

	// Basic
	fmt.Printf("Hello %s, your balance is %.2f\n", "Alice", 1234.5)

	// Indexed — correct but unreadable
	fmt.Printf("%[3]s (%[1]d) -> balance: %[2].2f\n", 1, 1234.5, "Alice")

	// Table — cryptic width numbers, fragile to change
	fmt.Printf("%-5s | %-12s | %12s | %-6s\n", "ID", "Name", "Balance", "Active")
	fmt.Printf("%-5s | %-12s | %12s | %-6s\n",
		strings.Repeat("-", 5), strings.Repeat("-", 12),
		strings.Repeat("-", 12), strings.Repeat("-", 6))
	for _, u := range users {
		fmt.Printf("%-5d | %-12s | %12.2f | %-6t\n", u.ID, u.Name, u.Balance, u.Active)
	}
}

// ─── 2. pyfmt ────────────────────────────────────────────────────────────────
//
// Mimics Python's str.format(). Uses {} positional or named via struct/map.
// Implements PEP3101 style formatting in Go.
//
// Pros:
//   - named access via struct fields or map keys
//   - supports nested field access ({user.Name})
//   - familiar to Python developers
//   - full fmt verb compatibility under the hood
//
// Cons:
//   - panics on undefined key with Must(), error-only with Fmt()
//   - no table/row API — just string interpolation
//   - slower than fmt due to reflection
//   - no alignment/padding utilities beyond what fmt provides
//   - last commit years ago, low maintenance activity
//
// Example (requires: go get github.com/slongfield/pyfmt):
//
//	import "github.com/slongfield/pyfmt"
//
//	// positional
//	pyfmt.Must("{} (id:{}) -> balance: {:.2f}", "Alice", 1, 1234.5)
//	// => "Alice (id:1) -> balance: 1234.50"
//
//	// named via map
//	pyfmt.Must("{name} -> {balance:.2f}", map[string]any{"name": "Alice", "balance": 1234.5})
//	// => "Alice -> 1234.50"
//
//	// named via struct field
//	pyfmt.Must("{Name} -> {Balance:.2f}", users[0])
//	// => "Alice -> 1234.50"

func examplePyfmt() {
	divider("2. pyfmt — documented example (requires: go get github.com/slongfield/pyfmt)")
	fmt.Println(`
  // positional
  pyfmt.Must("{} (id:{}) -> balance: {:.2f}", "Alice", 1, 1234.5)
  // => "Alice (id:1) -> balance: 1234.50"

  // named via map
  pyfmt.Must("{name} -> {balance:.2f}", map[string]any{"name": "Alice", "balance": 1234.5})
  // => "Alice -> 1234.50"

  // named via struct fields (uses reflection)
  pyfmt.Must("{Name} -> {Balance:.2f}", users[0])
  // => "Alice -> 1234.50"
	`)
}

// ─── 3. fstr ─────────────────────────────────────────────────────────────────
//
// Closest to Python f-strings. Named map-based interpolation with {key} and
// {key:.2f} syntax. Also supports {name=} debug output.
//
// Pros:
//   - clean {key} syntax, very readable
//   - supports {balance:,.2f} including thousands separator
//   - {name=} debug syntax outputs "name=value"
//   - Interpolate() returns (string, error) — safe
//   - Eval() panics on error — quick scripts
//
// Cons:
//   - map[string]interface{} only — no struct field access
//   - no table/row API
//   - no string utilities (Truncate, Wrap, case conversions, etc.)
//   - ~4x slower than fmt.Sprintf for simple strings
//   - thousands separator deviates from standard Go fmt verbs
//
// Benchmark from their repo:
//
//	BenchmarkSimpleString/fmt.Sprintf    4376912   310 ns/op    32 B/op   2 allocs
//	BenchmarkSimpleString/fstr.Sprintf    854262  1344 ns/op   184 B/op  11 allocs
//
// Example (requires: go get github.com/ZiadMansourM/fstr):
//
//	import "github.com/ZiadMansourM/fstr"
//
//	fstr.Println(
//	    "{name} (id:{id}) -> balance: {balance:.2f}",
//	    map[string]interface{}{"id": 1, "name": "Alice", "balance": 1234.5},
//	)
//	// => "Alice (id:1) -> balance: 1234.50"
//
//	// debug syntax
//	fstr.Eval("{name=} {balance=:.2f}", map[string]interface{}{"name": "Alice", "balance": 1234.5})
//	// => "name=Alice balance=1234.50"
//
//	// thousands separator
//	fstr.Eval("balance: {balance:,.2f}", map[string]interface{}{"balance": 123456789.64})
//	// => "balance: 123,456,789.64"

func exampleFstr() {
	divider("3. fstr — documented example (requires: go get github.com/ZiadMansourM/fstr)")
	fmt.Println(`
  fstr.Println(
      "{name} (id:{id}) -> balance: {balance:.2f}",
      map[string]interface{}{"id": 1, "name": "Alice", "balance": 1234.5},
  )
  // => "Alice (id:1) -> balance: 1234.50"

  // debug syntax — unique to fstr
  fstr.Eval("{name=} {balance=:.2f}", map[string]interface{}{"name": "Alice", "balance": 1234.5})
  // => "name=Alice balance=1234.50"

  // thousands separator — unique to fstr
  fstr.Eval("balance: {balance:,.2f}", map[string]interface{}{"balance": 123456789.64})
  // => "balance: 123,456,789.64"
	`)
}

// ─── 4. stringFormatter ──────────────────────────────────────────────────────
//
// Indexed {0},{1} AND named map formatting. Has MapToString, SliceToString.
//
// Pros:
//   - both indexed {0} and named {key} in the same lib
//   - FormatComplex() for named args — map[string]any
//   - MapToString / SliceToString helpers
//   - claims performance edge over fmt for slice printing
//   - active development / recent commits
//
// Cons:
//   - no table/row alignment API
//   - no string utilities (Truncate, Wrap, Center, etc.)
//   - {0:B}/{0:X}/{0:F} are custom codes, NOT Go fmt verbs
//   - cannot mix indexed and named in a single call
//
// Example (requires: go get github.com/Wissance/stringFormatter):
//
//	import sf "github.com/Wissance/stringFormatter"
//
//	// indexed
//	sf.Format("Hello {0}, balance: {1}", "Alice", 1234.5)
//	// => "Hello Alice, balance: 1234.5"
//
//	// named
//	sf.FormatComplex("{name} -> balance: {balance}", map[string]any{"name": "Alice", "balance": 1234.5})
//	// => "Alice -> balance: 1234.5"
//
//	// custom format codes (NOT Go fmt verbs)
//	sf.FormatComplex("bin:{val:B} hex:{val:X4}", map[string]any{"val": 255})
//	// => "bin:11111111 hex:00ff"

func exampleStringFormatter() {
	divider("4. stringFormatter — documented example (requires: go get github.com/Wissance/stringFormatter)")
	fmt.Println(`
  // indexed
  sf.Format("Hello {0}, balance: {1}", "Alice", 1234.5)
  // => "Hello Alice, balance: 1234.5"

  // named
  sf.FormatComplex(
      "{name} -> balance: {balance}",
      map[string]any{"name": "Alice", "balance": 1234.5},
  )
  // => "Alice -> balance: 1234.5"

  // slice helper
  slice := []any{1, "two", 3.0}
  sep := ", "
  sf.SliceToString(&slice, &sep)
  // => "1, two, 3"

  // custom format codes (different from Go fmt verbs!)
  sf.FormatComplex("bin:{val:B} hex:{val:X4}", map[string]any{"val": 255})
  // => "bin:11111111 hex:00ff"
	`)
}

// ─── 5. gstring ──────────────────────────────────────────────────────────────
//
// Named {key} interpolation + WithStruct + fluent Row/Table + AutoWidth + utilities.
// Pure stdlib, zero deps, template-cached for performance.
//
// Pros:
//   - named placeholders with full Go fmt verb specs {balance:.2f}
//   - WithStruct() — interpolate directly from exported struct fields
//   - fluent Row builder for single-line column alignment
//   - Table builder — declare once, render many rows with ANSI colors
//   - AutoWidth() — auto-size columns from data, no manual width counting
//   - String utilities: Truncate, Pad, Center, Wrap, Repeat, Strip, Title, Snake, Camel
//   - Table.String() returns table as string (useful for testing/logging)
//   - template cache — regex runs once per unique template, zero regex on repeat calls
//   - zero dependencies, pure stdlib under the hood
//
// Cons:
//   - no {name=} debug syntax (unlike fstr)
//   - no thousands separator (unlike fstr's {val:,.2f})
//   - no indexed {0},{1} positional syntax (unlike pyfmt / stringFormatter)
//   - first call on a new template has parse overhead (cached after that)

func exampleGstring() {
	divider("5. gstring (this lib)")

	// Named interpolation via With()
	fmt.Println("── Named interpolation (With)")
	for _, u := range users {
		gstring.Println(
			"{name} (id:{id:03d}) -> balance: {balance:.2f} active:{active}",
			gstring.With("id", u.ID, "name", u.Name, "balance", u.Balance, "active", u.Active),
		)
	}

	fmt.Println()

	// Named interpolation via WithStruct() — new
	fmt.Println("── Named interpolation (WithStruct)")
	for _, u := range users {
		gstring.Println(
			"{Name} (id:{ID:03d}) -> balance: {Balance:.2f}",
			gstring.WithStruct(u),
		)
	}

	fmt.Println()

	// Row builder
	fmt.Println("── Row builder")
	for _, u := range users {
		gstring.NewRow().
			Left(u.ID, 5).Sep("|").
			Left(u.Name, 12).Sep("|").
			Right(u.Balance, 12, 2).Sep("|").
			Left(u.Active, 6).
			Print()
	}

	fmt.Println()

	// Table builder with explicit widths + colors
	fmt.Println("── Table builder (explicit widths)")
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

	fmt.Println()

	// AutoWidth — widths computed from data, no manual counting
	fmt.Println("── Table builder (AutoWidth — widths computed from data)")
	t2 := gstring.NewTable().HeaderColor(gstring.ColorYellow)
	t2.Col("ID", 0, gstring.AlignRight).
		Col("Name", 0, gstring.AlignLeft).
		Col("Balance", 0, gstring.AlignRight, 2).
		Col("Active", 0, gstring.AlignLeft)
	rows := [][]any{}
	for _, u := range users {
		rows = append(rows, []any{u.ID, u.Name, u.Balance, u.Active})
	}
	t2.AutoWidth(rows)
	t2.Header()
	for _, u := range users {
		t2.Row(u.ID, u.Name, u.Balance, u.Active)
	}

	fmt.Println()

	// String utilities — none of the other libs have these
	fmt.Println("── String utilities (unique to gstring)")
	fmt.Printf("Truncate:  %s\n", gstring.Truncate("Hello, World!", 8))
	fmt.Printf("Pad:       [%s]\n", gstring.Pad("hi", 10, '·'))
	fmt.Printf("Center:    [%s]\n", gstring.Center("hi", 10, '─'))
	fmt.Printf("Wrap:\n%s\n", gstring.Wrap("The quick brown fox jumps over the lazy dog", 20))
	fmt.Printf("Title:     %s\n", gstring.Title("hello world"))
	fmt.Printf("Snake:     %s\n", gstring.Snake("HelloWorld"))
	fmt.Printf("Camel:     %s\n", gstring.Camel("hello_world"))
	fmt.Printf("Repeat:    %s\n", gstring.Repeat("─", 40))
}

// ─── Feature matrix ───────────────────────────────────────────────────────────

func featureMatrix() {
	divider("Feature Matrix")

	t := gstring.NewTable().HeaderColor(gstring.ColorBold)
	t.Col("Feature", 28, gstring.AlignLeft).
		Col("fmt", 7, gstring.AlignCenter).
		Col("pyfmt", 7, gstring.AlignCenter).
		Col("fstr", 7, gstring.AlignCenter).
		Col("strFmt", 7, gstring.AlignCenter).
		Col("gstring", 8, gstring.AlignCenter)
	t.Header()

	rows := [][]any{
		{"Named {key} placeholders", "✗", "✓", "✓", "✓", "✓"},
		{"Indexed {0} placeholders", "✓", "✓", "✗", "✓", "✗"},
		{"Struct field access", "✓", "✓", "✗", "✗", "✓"},
		{"Full Go fmt verb specs", "✓", "✓", "partial", "custom", "✓"},
		{"{name=} debug syntax", "✗", "✗", "✓", "✗", "✗"},
		{"Thousands sep {val:,.2f}", "✗", "✗", "✓", "✗", "✗"},
		{"Table/Row alignment API", "✗", "✗", "✗", "✗", "✓"},
		{"AutoWidth from data", "✗", "✗", "✗", "✗", "✓"},
		{"ANSI color support", "✗", "✗", "✗", "✗", "✓"},
		{"String utilities", "✗", "✗", "✗", "✗", "✓"},
		{"Table.String() return", "✗", "✗", "✗", "✗", "✓"},
		{"Template caching", "✗", "✗", "✗", "✗", "✓"},
		{"Zero dependencies", "✓", "✓", "✓", "✓", "✓"},
		{"Error on missing key", "panic", "error", "error", "error", "passthru"},
		{"Relative perf vs fmt", "1x", "~2x", "~4x", "~1.2x", "~1.5x"},
	}

	for _, r := range rows {
		t.Row(r...)
	}
}

func main() {
	exampleFmt()
	examplePyfmt()
	exampleFstr()
	exampleStringFormatter()
	exampleGstring()
	featureMatrix()

	fmt.Println()
	fmt.Println(gstring.Repeat("─", 60))
	fmt.Println(" Summary")
	fmt.Println(gstring.Repeat("─", 60))
	fmt.Println(`
  Use fmt when:
    - performance is critical (hot loops, high-throughput formatting)
    - you want compile-time vet checks on format strings
    - you need full verb coverage (%T, %p, %#v, etc.)

  Use pyfmt when:
    - you want Python .format() style with struct field access
    - your team comes from a Python background

  Use fstr when:
    - you want {name=} debug output
    - you need thousands separators ({val:,.2f})
    - you're doing one-off script-style formatting

  Use stringFormatter when:
    - you want both indexed AND named in one lib
    - you need SliceToString / MapToString helpers

  Use gstring when:
    - you're building CLI output, tables, or aligned reports
    - you want to interpolate struct fields directly with WithStruct()
    - you want AutoWidth so you never count column widths manually
    - readability and maintainability matter more than raw speed
    - you want a fluent table API with color support
    - you need string utilities (Truncate, Wrap, Snake, Camel, etc.)
	`)
}
