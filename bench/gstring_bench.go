package gstring_test

// ─── gstring benchmark suite ──────────────────────────────────────────────────
//
// Run all benchmarks:
//   go test ./gstring/... -bench=. -benchmem -count=3
//
// Run a specific group:
//   go test ./gstring/... -bench=BenchmarkSprintf -benchmem -count=3
//
// CPU profile:
//   go test ./gstring/... -bench=BenchmarkSprintf_Cached -benchmem -cpuprofile=cpu.prof
//   go tool pprof cpu.prof
//
// Memory profile:
//   go test ./gstring/... -bench=. -benchmem -memprofile=mem.prof
//   go tool pprof mem.prof

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nurysso/gstrings"
)

// ─── sink vars — prevent the compiler from optimising away results ────────────

var (
	sinkStr   string
	sinkBytes []byte
)

// ─── shared test data ─────────────────────────────────────────────────────────

type benchUser struct {
	ID      int
	Name    string
	Balance float64
	Active  bool
}

var (
	smallUsers = []benchUser{
		{1, "Alice", 1234.5, true},
		{2, "Bob", 98.12, false},
		{3, "Charlotte", 100000.99, true},
	}

	largeUsers = func() []benchUser {
		names := []string{"Alice", "Bob", "Charlotte", "Dave", "Eve", "Frank", "Grace", "Heidi"}
		out := make([]benchUser, 100)
		for i := range out {
			out[i] = benchUser{i + 1, names[i%len(names)], float64(i)*123.45 + 0.99, i%2 == 0}
		}
		return out
	}()
)

// ═════════════════════════════════════════════════════════════════════════════
// GROUP 1 — Sprintf: cold vs warm (cache miss vs cache hit)
// ═════════════════════════════════════════════════════════════════════════════

// BenchmarkSprintf_Cold measures the first-call parse+cache cost.
// Each iteration uses a unique template so the cache never hits.
func BenchmarkSprintf_Cold(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// unique template every iteration → always a cache miss
		tpl := fmt.Sprintf("hello {name} iteration-%d end", i)
		sinkStr = gstring.Sprintf(tpl, gstring.With("name", "Alice"))
	}
}

// BenchmarkSprintf_Warm measures steady-state throughput after the cache is hot.
func BenchmarkSprintf_Warm(b *testing.B) {
	vars := gstring.With("name", "Alice", "age", 30)
	// warm the cache
	gstring.Sprintf("Hello {name}, age {age}", vars)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Sprintf("Hello {name}, age {age}", vars)
	}
}

// BenchmarkSprintf_Warm_vs_Fmt is a direct apples-to-apples warm comparison.
func BenchmarkSprintf_Warm_vs_Fmt(b *testing.B) {
	vars := gstring.With("name", "Alice", "age", 30)
	gstring.Sprintf("Hello {name}, age {age}", vars)

	b.Run("gstring_warm", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sinkStr = gstring.Sprintf("Hello {name}, age {age}", vars)
		}
	})
	b.Run("fmt_sprintf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sinkStr = fmt.Sprintf("Hello %s, age %d", "Alice", 30)
		}
	})
}

// ═════════════════════════════════════════════════════════════════════════════
// GROUP 2 — Sprintf: template complexity
// ═════════════════════════════════════════════════════════════════════════════

// BenchmarkSprintf_NoPlaceholders — template with no {} at all (fast-path).
func BenchmarkSprintf_NoPlaceholders(b *testing.B) {
	vars := gstring.With()
	gstring.Sprintf("no placeholders in this string at all", vars)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Sprintf("no placeholders in this string at all", vars)
	}
}

// BenchmarkSprintf_OnePlaceholder — single {key} substitution.
func BenchmarkSprintf_OnePlaceholder(b *testing.B) {
	vars := gstring.With("name", "Alice")
	gstring.Sprintf("Hello {name}!", vars)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Sprintf("Hello {name}!", vars)
	}
}

// BenchmarkSprintf_ThreePlaceholders — typical real-world usage.
func BenchmarkSprintf_ThreePlaceholders(b *testing.B) {
	vars := gstring.With("name", "Alice", "id", 1, "balance", 1234.5)
	gstring.Sprintf("{name} (id:{id:03d}) -> balance: {balance:.2f}", vars)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Sprintf("{name} (id:{id:03d}) -> balance: {balance:.2f}", vars)
	}
}

// BenchmarkSprintf_SixPlaceholders — stress linear-scan Vars at its limit.
func BenchmarkSprintf_SixPlaceholders(b *testing.B) {
	vars := gstring.With("a", "Alice", "b", 42, "c", 3.14159, "d", true, "e", "extra", "f", 99)
	gstring.Sprintf("{a} {b} {c:.2f} {d} {e} {f}", vars)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Sprintf("{a} {b} {c:.2f} {d} {e} {f}", vars)
	}
}

// BenchmarkSprintf_TwelvePlaceholders — beyond optimal range for linear scan.
func BenchmarkSprintf_TwelvePlaceholders(b *testing.B) {
	vars := gstring.With(
		"k1", "v1", "k2", 2, "k3", 3.0, "k4", true,
		"k5", "v5", "k6", 6, "k7", 7.0, "k8", false,
		"k9", "v9", "k10", 10, "k11", 11.0, "k12", "v12",
	)
	tpl := "{k1} {k2} {k3:.1f} {k4} {k5} {k6} {k7:.1f} {k8} {k9} {k10} {k11:.1f} {k12}"
	gstring.Sprintf(tpl, vars)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Sprintf(tpl, vars)
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// GROUP 3 — Sprintf: value types (strconv fast paths)
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkSprintf_TypeString(b *testing.B) {
	vars := gstring.With("v", "hello world")
	gstring.Sprintf("{v}", vars)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Sprintf("{v}", vars)
	}
}

func BenchmarkSprintf_TypeInt(b *testing.B) {
	vars := gstring.With("v", 123456)
	gstring.Sprintf("{v}", vars)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Sprintf("{v}", vars)
	}
}

func BenchmarkSprintf_TypeFloat64(b *testing.B) {
	vars := gstring.With("v", 99999.9999)
	gstring.Sprintf("{v}", vars)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Sprintf("{v}", vars)
	}
}

func BenchmarkSprintf_TypeFloat64WithSpec(b *testing.B) {
	vars := gstring.With("v", 99999.9999)
	gstring.Sprintf("{v:.2f}", vars)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Sprintf("{v:.2f}", vars)
	}
}

func BenchmarkSprintf_TypeBool(b *testing.B) {
	vars := gstring.With("v", true)
	gstring.Sprintf("{v}", vars)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Sprintf("{v}", vars)
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// GROUP 4 — WithStruct
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkWithStruct_Build(b *testing.B) {
	u := smallUsers[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gstring.WithStruct(u)
	}
}

func BenchmarkWithStruct_BuildAndSprintf(b *testing.B) {
	u := smallUsers[0]
	gstring.Sprintf("{Name} {ID:03d} {Balance:.2f}", gstring.WithStruct(u))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Sprintf("{Name} {ID:03d} {Balance:.2f}", gstring.WithStruct(u))
	}
}

func BenchmarkWith_BuildAndSprintf(b *testing.B) {
	u := smallUsers[0]
	gstring.Sprintf("{name} {id:03d} {balance:.2f}", gstring.With("name", u.Name, "id", u.ID, "balance", u.Balance))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Sprintf(
			"{name} {id:03d} {balance:.2f}",
			gstring.With("name", u.Name, "id", u.ID, "balance", u.Balance),
		)
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// GROUP 5 — Row builder
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkRow_TwoColumns(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.NewRow().Left("Alice", 12).Sep("|").Right(1234.5, 10, 2).String()
	}
}

func BenchmarkRow_FourColumns(b *testing.B) {
	u := smallUsers[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.NewRow().
			Left(u.ID, 5).Sep("|").
			Left(u.Name, 12).Sep("|").
			Right(u.Balance, 12, 2).Sep("|").
			Left(u.Active, 6).
			String()
	}
}

func BenchmarkRow_FourColumns_vs_Fmt(b *testing.B) {
	u := smallUsers[0]

	b.Run("gstring_row", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sinkStr = gstring.NewRow().
				Left(u.ID, 5).Sep("|").
				Left(u.Name, 12).Sep("|").
				Right(u.Balance, 12, 2).Sep("|").
				Left(u.Active, 6).
				String()
		}
	})
	b.Run("fmt_sprintf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sinkStr = fmt.Sprintf("%-5d | %-12s | %12.2f | %-6t", u.ID, u.Name, u.Balance, u.Active)
		}
	})
}

func BenchmarkRow_EightColumns(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.NewRow().
			Left("Alice", 10).Sep("|").
			Right(1, 5).Sep("|").
			Right(1234.5, 10, 2).Sep("|").
			Left(true, 6).Sep("|").
			Left("admin", 8).Sep("|").
			Left("active", 8).Sep("|").
			Right(42, 4).Sep("|").
			Left("US", 4).
			String()
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// GROUP 6 — Table builder: Row() throughput
// ═════════════════════════════════════════════════════════════════════════════

// newBenchTable builds a standard 4-col table for reuse across benchmarks.
func newBenchTable() *gstring.Table {
	t := gstring.NewTable()
	t.Col("ID", 5, gstring.AlignRight).
		Col("Name", 12, gstring.AlignLeft).
		Col("Balance", 12, gstring.AlignRight, 2).
		Col("Active", 6, gstring.AlignLeft)
	return t
}

// BenchmarkTable_RowPrint — measures single Row() print call.
func BenchmarkTable_RowPrint(b *testing.B) {
	t := newBenchTable()
	u := smallUsers[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Row(u.ID, u.Name, u.Balance, u.Active)
	}
}

// BenchmarkTable_String_SmallDataset — 3 rows rendered to string.
func BenchmarkTable_String_Small(b *testing.B) {
	t := newBenchTable()
	rows := make([][]any, len(smallUsers))
	for i, u := range smallUsers {
		rows[i] = []any{u.ID, u.Name, u.Balance, u.Active}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = t.String(rows)
	}
}

// BenchmarkTable_String_LargeDataset — 100 rows rendered to string.
func BenchmarkTable_String_Large(b *testing.B) {
	t := newBenchTable()
	rows := make([][]any, len(largeUsers))
	for i, u := range largeUsers {
		rows[i] = []any{u.ID, u.Name, u.Balance, u.Active}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = t.String(rows)
	}
}

// BenchmarkTable_String_Large_vs_Fmt — gstring vs hand-rolled fmt for 100 rows.
func BenchmarkTable_String_Large_vs_Fmt(b *testing.B) {
	t := newBenchTable()
	rows := make([][]any, len(largeUsers))
	for i, u := range largeUsers {
		rows[i] = []any{u.ID, u.Name, u.Balance, u.Active}
	}

	b.Run("gstring_table", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sinkStr = t.String(rows)
		}
	})
	b.Run("fmt_manual", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var sb strings.Builder
			sb.Grow(4096)
			sb.WriteString(fmt.Sprintf("%-5s | %-12s | %12s | %-6s\n", "ID", "Name", "Balance", "Active"))
			sb.WriteString(fmt.Sprintf("%-5s | %-12s | %12s | %-6s\n", "-----", "------------", "------------", "------"))
			for _, u := range largeUsers {
				sb.WriteString(fmt.Sprintf("%-5d | %-12s | %12.2f | %-6t\n", u.ID, u.Name, u.Balance, u.Active))
			}
			sinkStr = sb.String()
		}
	})
}

// BenchmarkTable_AutoWidth — cost of auto-sizing columns from data.
func BenchmarkTable_AutoWidth(b *testing.B) {
	rows := make([][]any, len(largeUsers))
	for i, u := range largeUsers {
		rows[i] = []any{u.ID, u.Name, u.Balance, u.Active}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t := gstring.NewTable()
		t.Col("ID", 0, gstring.AlignRight).
			Col("Name", 0, gstring.AlignLeft).
			Col("Balance", 0, gstring.AlignRight, 2).
			Col("Active", 0, gstring.AlignLeft)
		t.AutoWidth(rows)
		sinkStr = t.String(rows)
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// GROUP 7 — String utilities
// ═════════════════════════════════════════════════════════════════════════════

var longString = strings.Repeat("Hello, World! ", 20)

func BenchmarkTruncate_Short(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Truncate("hi", 10)
	}
}

func BenchmarkTruncate_Long(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Truncate(longString, 20)
	}
}

func BenchmarkPad_Left(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Pad("hello", 20, '-')
	}
}

func BenchmarkPad_Right(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Pad("hello", -20, '-')
	}
}

func BenchmarkCenter(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Center("hello", 30, '─')
	}
}

func BenchmarkWrap_Short(b *testing.B) {
	s := "The quick brown fox jumps over the lazy dog"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Wrap(s, 20)
	}
}

func BenchmarkWrap_Long(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Wrap(longString, 40)
	}
}

func BenchmarkStrip(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Strip("    lots of whitespace around this string    ")
	}
}

func BenchmarkRepeat(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Repeat("─", 80)
	}
}

func BenchmarkTitle(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Title("the quick brown fox jumps over the lazy dog")
	}
}

func BenchmarkSnake_CamelInput(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Snake("TheQuickBrownFoxJumpsOverTheLazyDog")
	}
}

func BenchmarkSnake_SpaceInput(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Snake("The Quick Brown Fox Jumps Over The Lazy Dog")
	}
}

func BenchmarkCamel_SnakeInput(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Camel("the_quick_brown_fox_jumps_over_the_lazy_dog")
	}
}

func BenchmarkCamel_SpaceInput(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Camel("The Quick Brown Fox Jumps Over The Lazy Dog")
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// GROUP 8 — Parallel throughput (simulates concurrent CLI/server usage)
// ═════════════════════════════════════════════════════════════════════════════

func BenchmarkSprintf_Parallel(b *testing.B) {
	vars := gstring.With("name", "Alice", "id", 1, "balance", 1234.5)
	gstring.Sprintf("{name} (id:{id:03d}) -> {balance:.2f}", vars)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sinkStr = gstring.Sprintf("{name} (id:{id:03d}) -> {balance:.2f}", vars)
		}
	})
}

func BenchmarkTable_String_Parallel(b *testing.B) {
	t := newBenchTable()
	rows := make([][]any, len(largeUsers))
	for i, u := range largeUsers {
		rows[i] = []any{u.ID, u.Name, u.Balance, u.Active}
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sinkStr = t.String(rows)
		}
	})
}

func BenchmarkRow_Parallel(b *testing.B) {
	u := smallUsers[0]
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sinkStr = gstring.NewRow().
				Left(u.ID, 5).Sep("|").
				Left(u.Name, 12).Sep("|").
				Right(u.Balance, 12, 2).Sep("|").
				Left(u.Active, 6).
				String()
		}
	})
}

// ═════════════════════════════════════════════════════════════════════════════
// GROUP 9 — Allocation pressure: zero-alloc verification
// ═════════════════════════════════════════════════════════════════════════════
//
// These use b.ReportAllocs() explicitly — if the "allocs/op" column
// in the output shows 0, the function is allocation-free on that path.

func BenchmarkAllocs_Sprintf_Warm(b *testing.B) {
	vars := gstring.With("name", "Alice", "age", 30)
	gstring.Sprintf("Hello {name}, age {age}", vars)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Sprintf("Hello {name}, age {age}", vars)
	}
}

func BenchmarkAllocs_NoPlaceholders(b *testing.B) {
	vars := gstring.With()
	gstring.Sprintf("plain string no placeholders", vars)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.Sprintf("plain string no placeholders", vars)
	}
}

func BenchmarkAllocs_Row(b *testing.B) {
	u := smallUsers[0]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = gstring.NewRow().
			Left(u.ID, 5).Sep("|").
			Left(u.Name, 12).
			String()
	}
}

func BenchmarkAllocs_TableString_Large(b *testing.B) {
	t := newBenchTable()
	rows := make([][]any, len(largeUsers))
	for i, u := range largeUsers {
		rows[i] = []any{u.ID, u.Name, u.Balance, u.Active}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = t.String(rows)
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// GROUP 10 — Stress: sustained high-volume throughput
// ═════════════════════════════════════════════════════════════════════════════

// BenchmarkStress_SprintfLoop simulates a logger formatting 1000 lines.
func BenchmarkStress_SprintfLoop(b *testing.B) {
	tpl := "[{level}] {ts} {service}: {msg} (req:{req_id})"
	vars := gstring.With(
		"level", "INFO",
		"ts", "2025-03-07T12:00:00Z",
		"service", "api-gateway",
		"msg", "request received",
		"req_id", "abc123xyz",
	)
	gstring.Sprintf(tpl, vars)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var sb strings.Builder
		sb.Grow(120 * 1000)
		for j := 0; j < 1000; j++ {
			sb.WriteString(gstring.Sprintf(tpl, vars))
			sb.WriteByte('\n')
		}
		sinkStr = sb.String()
	}
}

// BenchmarkStress_TableLargeLoop renders a 100-row table 100 times.
func BenchmarkStress_TableLargeLoop(b *testing.B) {
	t := newBenchTable()
	rows := make([][]any, len(largeUsers))
	for i, u := range largeUsers {
		rows[i] = []any{u.ID, u.Name, u.Balance, u.Active}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			sinkStr = t.String(rows)
		}
	}
}

// BenchmarkStress_MixedWorkload simulates a CLI tool mixing all APIs.
func BenchmarkStress_MixedWorkload(b *testing.B) {
	headerTpl := "=== Report: {title} ==="
	rowTpl := "{name} (id:{id:04d}) balance:{balance:.2f}"
	t := newBenchTable()
	rows := make([][]any, len(largeUsers))
	for i, u := range largeUsers {
		rows[i] = []any{u.ID, u.Name, u.Balance, u.Active}
	}
	// warm all templates
	gstring.Sprintf(headerTpl, gstring.With("title", "test"))
	gstring.Sprintf(rowTpl, gstring.With("name", "Alice", "id", 1, "balance", 1.0))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// header
		sinkStr = gstring.Sprintf(headerTpl, gstring.With("title", "User Summary"))
		// table
		sinkStr = t.String(rows)
		// per-row interpolation
		for _, u := range largeUsers {
			sinkStr = gstring.Sprintf(rowTpl, gstring.With("name", u.Name, "id", u.ID, "balance", u.Balance))
		}
		// utilities
		sinkStr = gstring.Snake(sinkStr)
		sinkStr = gstring.Truncate(sinkStr, 30)
		sinkStr = gstring.Center(sinkStr, 40, '─')
	}
}
