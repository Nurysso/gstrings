package gstring

import (
	"strings"
	"testing"
)

// ─── Test data ────────────────────────────────────────────────────────────────

type User struct {
	ID      int
	Name    string
	Balance float64
	Active  bool
}

var testUser = User{1, "Alice", 1234.5, true}

// ─── Vars / With ─────────────────────────────────────────────────────────────

func TestWith_GetExistingKey(t *testing.T) {
	v := With("name", "Alice", "age", 30)
	val, ok := v.get("name")
	if !ok || val != "Alice" {
		t.Errorf("expected Alice, got %v (ok=%v)", val, ok)
	}
}

func TestWith_GetMissingKey(t *testing.T) {
	v := With("name", "Alice")
	_, ok := v.get("missing")
	if ok {
		t.Error("expected false for missing key")
	}
}

func TestWith_Empty(t *testing.T) {
	v := With()
	_, ok := v.get("anything")
	if ok {
		t.Error("expected false on empty Vars")
	}
}

func TestWith_DuplicateKeys_ReturnsFirst(t *testing.T) {
	v := With("name", "Alice", "name", "Bob")
	val, ok := v.get("name")
	if !ok || val != "Alice" {
		t.Errorf("expected first value Alice, got %v", val)
	}
}

// ─── WithStruct ───────────────────────────────────────────────────────────────

func TestWithStruct_ExportedFields(t *testing.T) {
	v := WithStruct(testUser)
	name, ok := v.get("Name")
	if !ok || name != "Alice" {
		t.Errorf("expected Alice, got %v", name)
	}
	id, ok := v.get("ID")
	if !ok || id != 1 {
		t.Errorf("expected 1, got %v", id)
	}
}

func TestWithStruct_Pointer(t *testing.T) {
	v := WithStruct(&testUser)
	name, ok := v.get("Name")
	if !ok || name != "Alice" {
		t.Errorf("pointer struct: expected Alice, got %v", name)
	}
}

func TestWithStruct_NonStruct(t *testing.T) {
	v := WithStruct("not a struct")
	if len(v.pairs) != 0 {
		t.Error("expected empty Vars for non-struct input")
	}
}

// ─── Sprintf ─────────────────────────────────────────────────────────────────

func TestSprintf_BasicSubstitution(t *testing.T) {
	got := Sprintf("Hello {name}!", With("name", "Alice"))
	want := "Hello Alice!"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSprintf_MultipleKeys(t *testing.T) {
	got := Sprintf("{name} is {age}", With("name", "Alice", "age", 30))
	want := "Alice is 30"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSprintf_WithFormatSpec(t *testing.T) {
	got := Sprintf("{balance:.2f}", With("balance", 1234.5))
	want := "1234.50"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSprintf_WithPaddedInt(t *testing.T) {
	got := Sprintf("{id:03d}", With("id", 7))
	want := "007"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSprintf_MissingKeyPassthrough(t *testing.T) {
	got := Sprintf("Hello {missing}!", With("name", "Alice"))
	want := "Hello {missing}!"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSprintf_NoPlaceholders(t *testing.T) {
	got := Sprintf("plain string", With())
	want := "plain string"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSprintf_EmptyTemplate(t *testing.T) {
	got := Sprintf("", With("name", "Alice"))
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestSprintf_UnclosedBrace(t *testing.T) {
	got := Sprintf("Hello {name", With("name", "Alice"))
	// unclosed brace treated as literal
	if !strings.Contains(got, "{name") {
		t.Errorf("expected unclosed brace to be preserved, got %q", got)
	}
}

func TestSprintf_WithStruct(t *testing.T) {
	got := Sprintf("{Name} ({ID})", WithStruct(testUser))
	want := "Alice (1)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSprintf_AllPrimitiveTypes(t *testing.T) {
	cases := []struct {
		template string
		vars     Vars
		want     string
	}{
		{"{v}", With("v", "hello"), "hello"},
		{"{v}", With("v", 42), "42"},
		{"{v}", With("v", int64(100)), "100"},
		{"{v}", With("v", float64(3.14)), "3.14"},
		{"{v}", With("v", float32(1.5)), "1.5"},
		{"{v}", With("v", true), "true"},
		{"{v}", With("v", false), "false"},
		{"{v}", With("v", int32(7)), "7"},
		{"{v}", With("v", uint(8)), "8"},
		{"{v}", With("v", uint64(9)), "9"},
	}
	for _, c := range cases {
		got := Sprintf(c.template, c.vars)
		if got != c.want {
			t.Errorf("Sprintf(%q) = %q, want %q", c.template, got, c.want)
		}
	}
}

// Template cache: same template called twice should return identical results
func TestSprintf_CacheConsistency(t *testing.T) {
	tmpl := "cached:{value:.2f}"
	vars := With("value", 99.9)
	first := Sprintf(tmpl, vars)
	second := Sprintf(tmpl, vars)
	if first != second {
		t.Errorf("cache inconsistency: %q != %q", first, second)
	}
}

// ─── String Utilities ─────────────────────────────────────────────────────────

func TestTruncate_ShortString(t *testing.T) {
	got := Truncate("Hi", 10)
	if got != "Hi" {
		t.Errorf("got %q, want %q", got, "Hi")
	}
}

func TestTruncate_ExactLength(t *testing.T) {
	got := Truncate("Hello", 5)
	if got != "Hello" {
		t.Errorf("got %q, want %q", got, "Hello")
	}
}

func TestTruncate_LongString(t *testing.T) {
	got := Truncate("Hello, World!", 8)
	if got != "Hello..." {
		t.Errorf("got %q, want %q", got, "Hello...")
	}
}

func TestTruncate_VeryShortLimit(t *testing.T) {
	got := Truncate("Hello", 2)
	if got != ".." {
		t.Errorf("got %q, want %q", got, "..")
	}
}

func TestTruncate_Unicode(t *testing.T) {
	got := Truncate("héllo wörld", 6)
	// 6 runes: h é l l o ' ' → truncated to 3 + "..."
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected ellipsis, got %q", got)
	}
}

func TestPad_RightPadding(t *testing.T) {
	got := Pad("hi", 6, '-')
	want := "hi----"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPad_LeftPadding(t *testing.T) {
	got := Pad("hi", -6, '-')
	want := "----hi"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPad_NoOpWhenWide(t *testing.T) {
	got := Pad("hello", 3, '-')
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestCenter_EvenPadding(t *testing.T) {
	got := Center("hi", 6, '-')
	want := "--hi--"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCenter_OddPadding(t *testing.T) {
	got := Center("hi", 7, '-')
	// odd: left gets floor, right gets ceil
	want := "--hi---"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCenter_NoOpWhenWide(t *testing.T) {
	got := Center("hello", 3, '-')
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestWrap_Basic(t *testing.T) {
	got := Wrap("the quick brown fox", 10)
	lines := strings.Split(got, "\n")
	for _, l := range lines {
		if len(l) > 10 {
			t.Errorf("line %q exceeds width 10", l)
		}
	}
}

func TestWrap_SingleWord(t *testing.T) {
	got := Wrap("hello", 3)
	if got != "hello" {
		t.Errorf("single word longer than width should not be split, got %q", got)
	}
}

func TestWrap_Empty(t *testing.T) {
	got := Wrap("", 10)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestRepeat(t *testing.T) {
	got := Repeat("ab", 3)
	want := "ababab"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStrip(t *testing.T) {
	got := Strip("  hello  ")
	want := "hello"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTitle(t *testing.T) {
	cases := [][2]string{
		{"hello world", "Hello World"},
		{"HELLO WORLD", "HELLO WORLD"},
		{"hello", "Hello"},
		{"", ""},
	}
	for _, c := range cases {
		got := Title(c[0])
		if got != c[1] {
			t.Errorf("Title(%q) = %q, want %q", c[0], got, c[1])
		}
	}
}

func TestSnake(t *testing.T) {
	cases := [][2]string{
		{"HelloWorld", "hello_world"},
		{"helloWorld", "hello_world"},
		{"hello world", "hello_world"},
		{"hello-world", "hello_world"},
		{"alreadysnake", "alreadysnake"},
	}
	for _, c := range cases {
		got := Snake(c[0])
		if got != c[1] {
			t.Errorf("Snake(%q) = %q, want %q", c[0], got, c[1])
		}
	}
}

func TestCamel(t *testing.T) {
	cases := [][2]string{
		{"hello_world", "helloWorld"},
		{"hello-world", "helloWorld"},
		{"hello world", "helloWorld"},
		{"already", "already"},
	}
	for _, c := range cases {
		got := Camel(c[0])
		if got != c[1] {
			t.Errorf("Camel(%q) = %q, want %q", c[0], got, c[1])
		}
	}
}

// ─── Row Builder ─────────────────────────────────────────────────────────────

func TestRow_LeftAlign(t *testing.T) {
	got := NewRow().Left("hi", 5).String()
	want := "hi   "
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRow_RightAlign(t *testing.T) {
	got := NewRow().Right("hi", 5).String()
	want := "   hi"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRow_RightAlignFloat(t *testing.T) {
	got := NewRow().Right(1234.5, 10, 2).String()
	want := "   1234.50"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRow_CenterCol(t *testing.T) {
	got := NewRow().CenterCol("hi", 6).String()
	want := "  hi  "
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRow_Sep(t *testing.T) {
	got := NewRow().Left("a", 1).Sep("|").Left("b", 1).String()
	want := "a | b"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRow_Chaining(t *testing.T) {
	got := NewRow().Left(1, 3).Sep("|").Left("Alice", 8).Sep("|").Right(99.5, 8, 2).String()
	if !strings.Contains(got, "Alice") || !strings.Contains(got, "99.50") {
		t.Errorf("unexpected row output: %q", got)
	}
}

// ─── Table Builder ────────────────────────────────────────────────────────────

func TestTable_StringOutput(t *testing.T) {
	tb := NewTable()
	tb.Col("Name", 8, AlignLeft).Col("Score", 6, AlignRight)
	rows := [][]any{
		{"Alice", 42},
		{"Bob", 7},
	}
	out := tb.String(rows)
	if !strings.Contains(out, "Alice") {
		t.Error("expected Alice in output")
	}
	if !strings.Contains(out, "Bob") {
		t.Error("expected Bob in output")
	}
	if !strings.Contains(out, "Name") {
		t.Error("expected header Name in output")
	}
}

func TestTable_AutoWidth(t *testing.T) {
	tb := NewTable()
	tb.Col("Name", 0, AlignLeft).Col("Score", 0, AlignRight)
	rows := [][]any{
		{"Alice", 42},
		{"A very long name indeed", 7},
	}
	tb.AutoWidth(rows)
	out := tb.String(rows)
	if !strings.Contains(out, "A very long name indeed") {
		t.Error("expected long name in output after AutoWidth")
	}
}

func TestTable_FloatPrecision(t *testing.T) {
	tb := NewTable()
	tb.Col("Balance", 10, AlignRight, 2)
	rows := [][]any{{1234.5}}
	out := tb.String(rows)
	if !strings.Contains(out, "1234.50") {
		t.Errorf("expected 1234.50 in output, got:\n%s", out)
	}
}

func TestTable_FewerValuesThanColumns(t *testing.T) {
	tb := NewTable()
	tb.Col("A", 4, AlignLeft).Col("B", 4, AlignLeft).Col("C", 4, AlignLeft)
	rows := [][]any{{"x"}} // only one value for three columns
	out := tb.String(rows)
	if !strings.Contains(out, "x") {
		t.Errorf("expected x in output, got:\n%s", out)
	}
}

func TestTable_CustomSeparator(t *testing.T) {
	tb := NewTable()
	tb.Separator(" , ")
	tb.Col("A", 3, AlignLeft).Col("B", 3, AlignLeft)
	rows := [][]any{{"x", "y"}}
	out := tb.String(rows)
	if !strings.Contains(out, " , ") {
		t.Errorf("expected custom separator in output, got:\n%s", out)
	}
}

// ─── Benchmarks ───────────────────────────────────────────────────────────────

func BenchmarkSprintf_Warm(b *testing.B) {
	tmpl := "{name} (id:{id:03d}) -> balance: {balance:.2f}"
	vars := With("id", 1, "name", "Alice", "balance", 1234.5)
	_ = Sprintf(tmpl, vars) // prime cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Sprintf(tmpl, vars)
	}
}

func BenchmarkSprintf_NoPlaceholder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Sprintf("no placeholders here", With())
	}
}

func BenchmarkSprintf_WithStruct(b *testing.B) {
	_ = Sprintf("{Name} ({ID})", WithStruct(testUser)) // prime cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Sprintf("{Name} ({ID})", WithStruct(testUser))
	}
}

func BenchmarkTable_String(b *testing.B) {
	tb := NewTable()
	tb.Col("ID", 5, AlignRight).
		Col("Name", 12, AlignLeft).
		Col("Balance", 12, AlignRight, 2).
		Col("Active", 6, AlignLeft)
	rows := [][]any{
		{1, "Alice", 1234.5, true},
		{2, "Bob", 98.12, false},
		{3, "Charlotte", 100000.99, true},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tb.String(rows)
	}
}

func BenchmarkRow_Build(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewRow().Left(1, 5).Sep("|").Left("Alice", 12).Sep("|").Right(1234.5, 12, 2).String()
	}
}
