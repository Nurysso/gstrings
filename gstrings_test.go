package gstring_test

import (
	"fmt"
	"testing"
	"gstring"
)

// ─── Sprintf benchmarks ───────────────────────────────────────────────────────

var sink string // prevent compiler optimising away results

func BenchmarkSprintfSimple_gstring(b *testing.B) {
	vars := gstring.With("name", "Alice", "age", 30)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = gstring.Sprintf("Hello {name}, age {age}", vars)
	}
}

func BenchmarkSprintfSimple_fmt(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = fmt.Sprintf("Hello %s, age %d", "Alice", 30)
	}
}

func BenchmarkSprintfFloat_gstring(b *testing.B) {
	vars := gstring.With("name", "Alice", "id", 1, "balance", 1234.5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = gstring.Sprintf("{name} (id:{id:03d}) -> balance: {balance:.2f}", vars)
	}
}

func BenchmarkSprintfFloat_fmt(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = fmt.Sprintf("%s (id:%03d) -> balance: %.2f", "Alice", 1, 1234.5)
	}
}

func BenchmarkSprintfNoPlaceholders_gstring(b *testing.B) {
	vars := gstring.With()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = gstring.Sprintf("no placeholders here at all", vars)
	}
}

func BenchmarkSprintfManyKeys_gstring(b *testing.B) {
	vars := gstring.With(
		"a", "Alice", "b", 42, "c", 3.14,
		"d", true, "e", "extra", "f", 99,
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = gstring.Sprintf("{a} {b} {c:.2f} {d} {e} {f}", vars)
	}
}

// ─── Row benchmarks ───────────────────────────────────────────────────────────

func BenchmarkRow_gstring(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = gstring.NewRow().
			Left("Alice", 12).Sep("|").
			Right(1234.5, 10, 2).Sep("|").
			Left(true, 6).
			String()
	}
}

func BenchmarkRow_fmt(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = fmt.Sprintf("%-12s | %10.2f | %-6v", "Alice", 1234.5, true)
	}
}

// ─── Table benchmarks ─────────────────────────────────────────────────────────

type benchUser struct {
	ID      int
	Name    string
	Balance float64
	Active  bool
}

var benchUsers = []benchUser{
	{1, "Alice", 1234.5, true},
	{2, "Bob", 98.12, false},
	{3, "Charlotte", 100000.99, true},
	{4, "Dave", 0.01, false},
	{5, "Eve", 55555.55, true},
}

func BenchmarkTableString_gstring(b *testing.B) {
	t := gstring.NewTable()
	t.Col("ID", 5, gstring.AlignRight).
		Col("Name", 12, gstring.AlignLeft).
		Col("Balance", 12, gstring.AlignRight, 2).
		Col("Active", 6, gstring.AlignLeft)

	rows := make([][]any, len(benchUsers))
	for i, u := range benchUsers {
		rows[i] = []any{u.ID, u.Name, u.Balance, u.Active}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = t.String(rows)
	}
}

func BenchmarkTableString_fmt(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s string
		s += fmt.Sprintf("%-5s | %-12s | %12s | %-6s\n", "ID", "Name", "Balance", "Active")
		s += fmt.Sprintf("%-5s | %-12s | %12s | %-6s\n", "-----", "------------", "------------", "------")
		for _, u := range benchUsers {
			s += fmt.Sprintf("%-5d | %-12s | %12.2f | %-6t\n", u.ID, u.Name, u.Balance, u.Active)
		}
		sink = s
	}
}

// ─── String utility benchmarks ────────────────────────────────────────────────

func BenchmarkTruncate(b *testing.B) {
	s := "Hello, World! This is a longer string."
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = gstring.Truncate(s, 10)
	}
}

func BenchmarkSnake(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = gstring.Snake("HelloWorldFooBar")
	}
}

func BenchmarkCamel(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = gstring.Camel("hello_world_foo_bar")
	}
}

func BenchmarkWrap(b *testing.B) {
	s := "The quick brown fox jumps over the lazy dog and then does it again"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = gstring.Wrap(s, 20)
	}
}

// ─── WithStruct benchmark ─────────────────────────────────────────────────────

func BenchmarkWithStruct(b *testing.B) {
	u := benchUsers[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := gstring.WithStruct(u)
		sink = gstring.Sprintf("{Name} {ID}", v)
	}
}
