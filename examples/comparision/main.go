package main

import (
	"fmt"
	"strings"
	"time"

	fstr "github.com/ZiadMansourM/fstr"
	"github.com/nurysso/gstrings"
	"github.com/slongfield/pyfmt"
	sf "github.com/wissance/stringFormatter"
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

const benchIterations = 100_000

// ─── Helpers ──────────────────────────────────────────────────────────────────

func divider(label string) {
	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf(" %s\n", label)
	fmt.Println(strings.Repeat("─", 60))
}

func benchmarkLabel(label string, ns int64) {
	fmt.Printf("  ⏱  %-40s %6d ns/op\n", label, ns)
}

// bench runs fn N times and returns nanoseconds per operation.
func bench(n int, fn func()) int64 {
	start := time.Now()
	for i := 0; i < n; i++ {
		fn()
	}
	return time.Since(start).Nanoseconds() / int64(n)
}

// ─── 1. stdlib fmt ───────────────────────────────────────────────────────────

func exampleFmt() {
	divider("1. stdlib fmt")

	// Basic interpolation
	result := fmt.Sprintf("Hello %s, your balance is %.2f", "Alice", 1234.5)
	fmt.Println(" Basic:   ", result)

	// Indexed — reordered args, hard to follow at a glance
	// args: [1]=name, [2]=id, [3]=balance
	result = fmt.Sprintf("%[1]s (%[2]d) -> balance: %.2f", "Alice", 1, 1234.5)
	fmt.Println(" Indexed: ", result)

	// Table
	fmt.Println()
	fmt.Printf(" %-5s | %-12s | %12s | %-6s\n", "ID", "Name", "Balance", "Active")
	fmt.Printf(" %-5s | %-12s | %12s | %-6s\n",
		strings.Repeat("-", 5), strings.Repeat("-", 12),
		strings.Repeat("-", 12), strings.Repeat("-", 6))
	for _, u := range users {
		fmt.Printf(" %-5d | %-12s | %12.2f | %-6t\n", u.ID, u.Name, u.Balance, u.Active)
	}

	// Benchmarks
	fmt.Println()
	nsBasic := bench(benchIterations, func() {
		_ = fmt.Sprintf("Hello %s, your balance is %.2f", "Alice", 1234.5)
	})
	nsFloat := bench(benchIterations, func() {
		_ = fmt.Sprintf("%[3]s (%[1]d) -> balance: %[2].2f", 1, 1234.5, "Alice")
	})
	benchmarkLabel("fmt.Sprintf basic", nsBasic)
	benchmarkLabel("fmt.Sprintf indexed float", nsFloat)
}

// ─── 2. pyfmt ────────────────────────────────────────────────────────────────

func examplePyfmt() {
	divider("2. pyfmt (github.com/slongfield/pyfmt)")

	// Positional
	result, _ := pyfmt.Fmt("{} (id:{}) -> balance: {:.2f}", "Alice", 1, 1234.5)
	fmt.Println(" Positional:", result)

	// Named via map
	result, _ = pyfmt.Fmt("{name} -> {balance:.2f}",
		map[string]any{"name": "Alice", "balance": 1234.5})
	fmt.Println(" Named map: ", result)

	// Named via struct field (reflection)
	result, _ = pyfmt.Fmt("{Name} (id:{ID}) -> {Balance:.2f}", users[0])
	fmt.Println(" Struct:    ", result)

	// Benchmarks
	fmt.Println()
	nsPositional := bench(benchIterations, func() {
		_, _ = pyfmt.Fmt("{} (id:{}) -> balance: {:.2f}", "Alice", 1, 1234.5)
	})
	nsNamed := bench(benchIterations, func() {
		_, _ = pyfmt.Fmt("{name} -> {balance:.2f}",
			map[string]any{"name": "Alice", "balance": 1234.5})
	})
	nsStruct := bench(benchIterations, func() {
		_, _ = pyfmt.Fmt("{Name} (id:{ID}) -> {Balance:.2f}", users[0])
	})
	benchmarkLabel("pyfmt positional", nsPositional)
	benchmarkLabel("pyfmt named map", nsNamed)
	benchmarkLabel("pyfmt struct (reflection)", nsStruct)
}

// ─── 3. fstr ─────────────────────────────────────────────────────────────────

func exampleFstr() {
	divider("3. fstr (github.com/ZiadMansourM/fstr)")

	// Named map
	result := fstr.Eval("{name} (id:{id}) -> balance: {balance:.2f}",
		map[string]interface{}{"id": 1, "name": "Alice", "balance": 1234.5})
	fmt.Println(" Named:     ", result)

	// Debug syntax — unique to fstr
	result = fstr.Eval("{name=} {balance=:.2f}",
		map[string]interface{}{"name": "Alice", "balance": 1234.5})
	fmt.Println(" Debug:     ", result)

	// Thousands separator — unique to fstr
	result = fstr.Eval("balance: {balance:,.2f}",
		map[string]interface{}{"balance": 123456789.64})
	fmt.Println(" Thousands: ", result)

	// Benchmarks
	fmt.Println()
	nsNamed := bench(benchIterations, func() {
		_ = fstr.Eval("{name} (id:{id}) -> balance: {balance:.2f}",
			map[string]interface{}{"id": 1, "name": "Alice", "balance": 1234.5})
	})
	nsDebug := bench(benchIterations, func() {
		_ = fstr.Eval("{name=} {balance=:.2f}",
			map[string]interface{}{"name": "Alice", "balance": 1234.5})
	})
	benchmarkLabel("fstr named map", nsNamed)
	benchmarkLabel("fstr debug syntax", nsDebug)
}

// ─── 4. stringFormatter ──────────────────────────────────────────────────────

func exampleStringFormatter() {
	divider("4. stringFormatter (github.com/Wissance/stringFormatter)")

	// Indexed
	result := sf.Format("Hello {0}, balance: {1}", "Alice", 1234.5)
	fmt.Println(" Indexed: ", result)

	// Named
	result = sf.FormatComplex("{name} -> balance: {balance}",
		map[string]any{"name": "Alice", "balance": 1234.5})
	fmt.Println(" Named:   ", result)

	// Slice helper
	slice := []any{1, "two", 3.0}
	sep := ", "
	fmt.Println(" Slice:   ", sf.SliceToString(&slice, &sep))

	// Benchmarks
	fmt.Println()
	nsIndexed := bench(benchIterations, func() {
		_ = sf.Format("Hello {0}, balance: {1}", "Alice", 1234.5)
	})
	nsNamed := bench(benchIterations, func() {
		_ = sf.FormatComplex("{name} -> balance: {balance}",
			map[string]any{"name": "Alice", "balance": 1234.5})
	})
	benchmarkLabel("stringFormatter indexed", nsIndexed)
	benchmarkLabel("stringFormatter named map", nsNamed)
}

// ─── 5. gstring ─────────────────────────────────────────────────────────────

func examplegstring() {
	divider("5. gstring (github.com/nurysso/gstring)")

	// Named With()
	fmt.Println(" Named interpolation (With):")
	for _, u := range users {
		gstring.Println(
			"  {name} (id:{id:03d}) -> balance: {balance:.2f} active:{active}",
			gstring.With("id", u.ID, "name", u.Name, "balance", u.Balance, "active", u.Active),
		)
	}

	// WithStruct
	fmt.Println()
	fmt.Println(" Named interpolation (WithStruct):")
	for _, u := range users {
		gstring.Println(
			"  {Name} (id:{ID:03d}) -> balance: {Balance:.2f}",
			gstring.WithStruct(u),
		)
	}

	// Row builder
	fmt.Println()
	fmt.Println(" Row builder:")
	for _, u := range users {
		fmt.Print("  ")
		gstring.NewRow().
			Left(u.ID, 5).Sep("|").
			Left(u.Name, 12).Sep("|").
			Right(u.Balance, 12, 2).Sep("|").
			Left(u.Active, 6).
			Print()
	}

	// Table builder
	fmt.Println()
	fmt.Println(" Table builder:")
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

	// AutoWidth
	fmt.Println()
	fmt.Println(" Table builder (AutoWidth):")
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

	// String utilities
	fmt.Println()
	fmt.Println(" String utilities:")
	fmt.Printf("  Truncate:  %s\n", gstring.Truncate("Hello, World!", 8))
	fmt.Printf("  Pad:       [%s]\n", gstring.Pad("hi", 10, '·'))
	fmt.Printf("  Center:    [%s]\n", gstring.Center("hi", 10, '─'))
	fmt.Printf("  Title:     %s\n", gstring.Title("hello world"))
	fmt.Printf("  Snake:     %s\n", gstring.Snake("HelloWorld"))
	fmt.Printf("  Camel:     %s\n", gstring.Camel("hello_world"))

	// Benchmarks
	fmt.Println()

	// Cold (first call, no cache)
	nsCold := bench(1, func() {
		_ = gstring.Sprintf(
			"{name} (id:{id:03d}) -> balance: {balance:.2f}",
			gstring.With("id", 1, "name", "Alice", "balance", 1234.5),
		)
	})

	// Warm (cached template)
	template := "{name} (id:{id:03d}) -> balance: {balance:.2f}"
	args := gstring.With("id", 1, "name", "Alice", "balance", 1234.5)
	_ = gstring.Sprintf(template, args) // prime cache
	nsWarm := bench(benchIterations, func() {
		_ = gstring.Sprintf(template, args)
	})

	nsNoPlaceholder := bench(benchIterations, func() {
		_ = gstring.Sprintf("no placeholders here at all", gstring.With())
	})

	nsStruct := bench(benchIterations, func() {
		_ = gstring.Sprintf("{Name} (id:{ID:03d}) -> {Balance:.2f}", gstring.WithStruct(users[0]))
	})

	benchRows := [][]any{}
	for _, u := range users {
		benchRows = append(benchRows, []any{u.ID, u.Name, u.Balance, u.Active})
	}
	nsTable := bench(benchIterations, func() {
		tBench := gstring.NewTable()
		tBench.Col("ID", 5, gstring.AlignRight).
			Col("Name", 12, gstring.AlignLeft).
			Col("Balance", 12, gstring.AlignRight, 2).
			Col("Active", 6, gstring.AlignLeft)
		_ = tBench.String(benchRows)
	})

	benchmarkLabel("gstring cold (first call)", nsCold)
	benchmarkLabel("gstring warm (cached template)", nsWarm)
	benchmarkLabel("gstring no-placeholder fast path", nsNoPlaceholder)
	benchmarkLabel("gstring WithStruct", nsStruct)
	benchmarkLabel("gstring table (5 rows)", nsTable)
}

// ─── Summary table ────────────────────────────────────────────────────────────

func summaryTable(results map[string]int64) {
	divider("Benchmark Summary (ns/op, lower is better)")

	baseline := results["fmt_basic"]

	rows := []struct {
		label string
		key   string
	}{
		{"fmt.Sprintf basic", "fmt_basic"},
		{"fmt.Sprintf indexed", "fmt_indexed"},
		{"pyfmt positional", "pyfmt_positional"},
		{"pyfmt named map", "pyfmt_named"},
		{"pyfmt struct (reflection)", "pyfmt_struct"},
		{"fstr named map", "fstr_named"},
		{"fstr debug syntax", "fstr_debug"},
		{"stringFormatter indexed", "sf_indexed"},
		{"stringFormatter named", "sf_named"},
		{"gstring cold", "gs_cold"},
		{"gstring warm (cached)", "gs_warm"},
		{"gstring no-placeholder", "gs_noop"},
		{"gstring WithStruct", "gs_struct"},
		{"gstring table 5 rows", "gs_table"},
	}

	fmt.Printf("\n  %-38s %10s %8s\n", "Operation", "ns/op", "vs fmt")
	fmt.Println(" ", strings.Repeat("─", 60))
	for _, r := range rows {
		ns, ok := results[r.key]
		if !ok {
			continue
		}
		ratio := float64(ns) / float64(baseline)
		fmt.Printf("  %-38s %10d %7.1fx\n", r.label, ns, ratio)
	}
}

// ─── Feature matrix ───────────────────────────────────────────────────────────

func featureMatrix() {
	divider("Feature Matrix")

	t := gstring.NewTable().HeaderColor(gstring.ColorBold)
	t.Col("Feature", 32, gstring.AlignLeft).
		Col("fmt", 6, gstring.AlignCenter).
		Col("pyfmt", 6, gstring.AlignCenter).
		Col("fstr", 6, gstring.AlignCenter).
		Col("strFmt", 7, gstring.AlignCenter).
		Col("gstring", 9, gstring.AlignCenter)
	t.Header()

	matrix := [][]any{
		{"Named {key} placeholders", "✗", "✓", "✓", "✓", "✓"},
		{"Indexed {0} placeholders", "✓", "✓", "✗", "✓", "✗"},
		{"Struct field access", "✓", "✓", "✗", "✗", "✓"},
		{"Full Go fmt verb specs", "✓", "✓", "partial", "custom", "✓"},
		{"{name=} debug syntax", "✗", "✗", "✓", "✗", "✗"},
		{"Thousands sep {val:,.2f}", "✗", "✗", "✓", "✗", "✗"},
		{"Table / Row alignment API", "✗", "✗", "✗", "✗", "✓"},
		{"AutoWidth from data", "✗", "✗", "✗", "✗", "✓"},
		{"ANSI color support", "✗", "✗", "✗", "✗", "✓"},
		{"String utilities", "✗", "✗", "✗", "✗", "✓"},
		{"Template caching", "✗", "✗", "✗", "✗", "✓"},
		{"Zero dependencies", "✓", "✓", "✓", "✓", "✓"},
		{"Missing key behavior", "panic", "error", "error", "error", "passthru"},
	}

	for _, r := range matrix {
		t.Row(r...)
	}
}

// ─── Main ─────────────────────────────────────────────────────────────────────

func main() {
	results := make(map[string]int64)

	// Run examples and collect benchmark results inline
	exampleFmt()
	results["fmt_basic"] = bench(benchIterations, func() {
		_ = fmt.Sprintf("Hello %s, your balance is %.2f", "Alice", 1234.5)
	})
	results["fmt_indexed"] = bench(benchIterations, func() {
		_ = fmt.Sprintf("%[1]s (%[2]d) -> balance: %.2f", "Alice", 1, 1234.5)
	})

	examplePyfmt()
	results["pyfmt_positional"] = bench(benchIterations, func() {
		_, _ = pyfmt.Fmt("{} (id:{}) -> balance: {:.2f}", "Alice", 1, 1234.5)
	})
	results["pyfmt_named"] = bench(benchIterations, func() {
		_, _ = pyfmt.Fmt("{name} -> {balance:.2f}",
			map[string]any{"name": "Alice", "balance": 1234.5})
	})
	results["pyfmt_struct"] = bench(benchIterations, func() {
		_, _ = pyfmt.Fmt("{Name} (id:{ID}) -> {Balance:.2f}", users[0])
	})

	exampleFstr()
	results["fstr_named"] = bench(benchIterations, func() {
		_ = fstr.Eval("{name} (id:{id}) -> balance: {balance:.2f}",
			map[string]interface{}{"id": 1, "name": "Alice", "balance": 1234.5})
	})
	results["fstr_debug"] = bench(benchIterations, func() {
		_ = fstr.Eval("{name=} {balance=:.2f}",
			map[string]interface{}{"name": "Alice", "balance": 1234.5})
	})

	exampleStringFormatter()
	results["sf_indexed"] = bench(benchIterations, func() {
		_ = sf.Format("Hello {0}, balance: {1}", "Alice", 1234.5)
	})
	results["sf_named"] = bench(benchIterations, func() {
		_ = sf.FormatComplex("{name} -> balance: {balance}",
			map[string]any{"name": "Alice", "balance": 1234.5})
	})

	examplegstring()
	results["gs_cold"] = bench(1, func() {
		_ = gstring.Sprintf("cold:{name} {balance:.2f} {id}",
			gstring.With("id", 1, "name", "Alice", "balance", 1234.5))
	})
	template := "{name} (id:{id:03d}) -> balance: {balance:.2f}"
	args := gstring.With("id", 1, "name", "Alice", "balance", 1234.5)
	_ = gstring.Sprintf(template, args) // prime cache
	results["gs_warm"] = bench(benchIterations, func() {
		_ = gstring.Sprintf(template, args)
	})
	results["gs_noop"] = bench(benchIterations, func() {
		_ = gstring.Sprintf("no placeholders here", gstring.With())
	})
	results["gs_struct"] = bench(benchIterations, func() {
		_ = gstring.Sprintf("{Name} (id:{ID:03d}) -> {Balance:.2f}", gstring.WithStruct(users[0]))
	})
	tableRows := [][]any{}
	for _, u := range users {
		tableRows = append(tableRows, []any{u.ID, u.Name, u.Balance, u.Active})
	}
	results["gs_table"] = bench(benchIterations, func() {
		tb := gstring.NewTable()
		tb.Col("ID", 5, gstring.AlignRight).
			Col("Name", 12, gstring.AlignLeft).
			Col("Balance", 12, gstring.AlignRight, 2).
			Col("Active", 6, gstring.AlignLeft)
		_ = tb.String(tableRows)
	})

	summaryTable(results)
	featureMatrix()

	fmt.Println()
	fmt.Println(gstring.Repeat("─", 60))
	fmt.Println(" When to use each")
	fmt.Println(gstring.Repeat("─", 60))
	fmt.Println(`
  fmt          → hot loops, compile-time vet, full verb coverage
  pyfmt        → Python-style with struct field access
  fstr         → debug output {name=}, thousands separators
  strFormatter → indexed + named in one lib, slice/map helpers
  gstring     → CLI tables, aligned output, struct interpolation,
                 AutoWidth, ANSI colors, string utilities
	`)
}
