// Package gstring provides readable, self-documenting string formatting for Go.
// Optimized for maximum performance via template caching, zero-alloc paths,
// strconv-based number formatting, and linear-scan Vars for small key sets.
package gstring

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

// ─── ANSI Colors ──────────────────────────────────────────────────────────────

const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorGray   = "\033[90m"
)

type Color string

const (
	ColorRed    Color = colorRed
	ColorGreen  Color = colorGreen
	ColorYellow Color = colorYellow
	ColorBlue   Color = colorBlue
	ColorCyan   Color = colorCyan
	ColorWhite  Color = colorWhite
	ColorGray   Color = colorGray
	ColorBold   Color = colorBold
	ColorNone   Color = ""
)

func colorize(s string, c Color) string {
	if c == ColorNone {
		return s
	}
	return string(c) + s + colorReset
}

// ─── Vars: linear-scan pairs (faster than map for <~8 keys) ──────────────────

// Vars holds named values for interpolation.
// Internally a flat []any slice of alternating key, value pairs.
// Linear scan beats map lookup for the small key counts typical in format strings.
type Vars struct {
	pairs []any // [k0, v0, k1, v1, ...]
}

// With constructs a Vars from alternating key/value pairs.
//
//	gstring.With("name", "Alice", "age", 30)
func With(pairs ...any) Vars {
	return Vars{pairs: pairs}
}

// get does a linear scan for key. Returns (value, true) or (nil, false).
func (v Vars) get(key string) (any, bool) {
	for i := 0; i+1 < len(v.pairs); i += 2 {
		if k, ok := v.pairs[i].(string); ok && k == key {
			return v.pairs[i+1], true
		}
	}
	return nil, false
}

// WithStruct extracts exported fields from a struct into a Vars.
func WithStruct(s any) Vars {
	val := reflect.Indirect(reflect.ValueOf(s))
	typ := val.Type()
	if typ.Kind() != reflect.Struct {
		return Vars{}
	}
	n := typ.NumField()
	pairs := make([]any, 0, n*2)
	for i := 0; i < n; i++ {
		f := typ.Field(i)
		if f.PkgPath == "" { // exported only
			pairs = append(pairs, f.Name, val.Field(i).Interface())
		}
	}
	return Vars{pairs: pairs}
}

// ─── Template cache ───────────────────────────────────────────────────────────
//
// Parse each template string once into a []token (literals + placeholders).
// On repeat calls we skip regex entirely and just walk the token slice.

type tokenKind uint8

const (
	tokLiteral     tokenKind = iota // raw text segment
	tokPlaceholder                  // {key} or {key:spec}
)

type token struct {
	kind tokenKind
	text string // literal text, or key name
	spec string // fmt spec (empty → %v)
}

// templateCache maps template string → []token
var templateCache sync.Map

// parse converts a template string into a token slice.
// Runs once per unique template string; result is cached.
func parse(template string) []token {
	if v, ok := templateCache.Load(template); ok {
		return v.([]token)
	}

	// Fast path: no placeholders at all
	if strings.IndexByte(template, '{') == -1 {
		toks := []token{{kind: tokLiteral, text: template}}
		templateCache.Store(template, toks)
		return toks
	}

	toks := make([]token, 0, 8)
	s := template
	for len(s) > 0 {
		open := strings.IndexByte(s, '{')
		if open == -1 {
			toks = append(toks, token{kind: tokLiteral, text: s})
			break
		}
		if open > 0 {
			toks = append(toks, token{kind: tokLiteral, text: s[:open]})
		}
		close := strings.IndexByte(s[open:], '}')
		if close == -1 {
			// unclosed brace — treat rest as literal
			toks = append(toks, token{kind: tokLiteral, text: s[open:]})
			break
		}
		inner := s[open+1 : open+close]
		if colon := strings.IndexByte(inner, ':'); colon != -1 {
			toks = append(toks, token{kind: tokPlaceholder, text: inner[:colon], spec: inner[colon+1:]})
		} else {
			toks = append(toks, token{kind: tokPlaceholder, text: inner})
		}
		s = s[open+close+1:]
	}

	templateCache.Store(template, toks)
	return toks
}

// ─── Named Interpolation ──────────────────────────────────────────────────────

// Sprintf replaces named placeholders and returns the result.
// Templates are parsed once and cached; subsequent calls do zero regex work.
//
//	{name}        → %v
//	{balance:.2f} → %.2f
//	{id:05d}      → %05d
func Sprintf(template string, vars Vars) string {
	toks := parse(template)

	// Fast path: single literal token (no placeholders)
	if len(toks) == 1 && toks[0].kind == tokLiteral {
		return toks[0].text
	}

	// Pre-calculate capacity: sum of literal lengths + rough estimate for values
	cap := 0
	for i := range toks {
		cap += len(toks[i].text)
	}
	var b strings.Builder
	b.Grow(cap * 2)

	for _, tok := range toks {
		if tok.kind == tokLiteral {
			b.WriteString(tok.text)
			continue
		}
		val, ok := vars.get(tok.text)
		if !ok {
			b.WriteByte('{')
			b.WriteString(tok.text)
			b.WriteByte('}')
			continue
		}
		if tok.spec == "" {
			writeValue(&b, val)
		} else {
			fmt.Fprintf(&b, "%"+tok.spec, val)
		}
	}
	return b.String()
}

// writeValue writes val to b without allocating a format string.
// Handles common types directly via strconv; falls back to fmt for the rest.
func writeValue(b *strings.Builder, val any) {
	switch v := val.(type) {
	case string:
		b.WriteString(v)
	case int:
		b.WriteString(strconv.Itoa(v))
	case int64:
		b.WriteString(strconv.FormatInt(v, 10))
	case float64:
		b.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
	case float32:
		b.WriteString(strconv.FormatFloat(float64(v), 'f', -1, 32))
	case bool:
		if v {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case int32:
		b.WriteString(strconv.FormatInt(int64(v), 10))
	case uint:
		b.WriteString(strconv.FormatUint(uint64(v), 10))
	case uint64:
		b.WriteString(strconv.FormatUint(v, 10))
	default:
		fmt.Fprintf(b, "%v", v)
	}
}

// Format is an alias for Sprintf.
func Format(template string, vars Vars) string { return Sprintf(template, vars) }

// Println formats and prints with a newline.
func Println(template string, vars Vars) { fmt.Println(Sprintf(template, vars)) }

// ─── String Utilities ─────────────────────────────────────────────────────────

// Truncate shortens s to at most n runes. Appends "..." if truncated.
func Truncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	if n <= 3 {
		return "..."[:n]
	}
	runes := []rune(s)
	return string(runes[:n-3]) + "..."
}

// Pad pads s to width using padChar.
// Positive width = left-align (pad right). Negative = right-align (pad left).
func Pad(s string, width int, padChar rune) string {
	right := width < 0
	if right {
		width = -width
	}
	count := utf8.RuneCountInString(s)
	if count >= width {
		return s
	}
	padding := strings.Repeat(string(padChar), width-count)
	if right {
		return padding + s
	}
	return s + padding
}

// Center centers s within a field of given width using padChar.
func Center(s string, width int, padChar rune) string {
	count := utf8.RuneCountInString(s)
	if count >= width {
		return s
	}
	total := width - count
	left := total / 2
	right := total - left
	pad := string(padChar)
	return strings.Repeat(pad, left) + s + strings.Repeat(pad, right)
}

// Wrap word-wraps s at the given column width.
func Wrap(s string, width int) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	lineLen := 0
	for i, w := range words {
		wlen := utf8.RuneCountInString(w)
		if i == 0 {
			b.WriteString(w)
			lineLen = wlen
			continue
		}
		if lineLen+1+wlen > width {
			b.WriteByte('\n')
			b.WriteString(w)
			lineLen = wlen
		} else {
			b.WriteByte(' ')
			b.WriteString(w)
			lineLen += 1 + wlen
		}
	}
	return b.String()
}

// Repeat repeats s n times.
func Repeat(s string, n int) string { return strings.Repeat(s, n) }

// Strip trims all leading and trailing whitespace.
func Strip(s string) string { return strings.TrimSpace(s) }

// Title converts s to Title Case.
func Title(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	upper := true
	for _, r := range s {
		if unicode.IsSpace(r) {
			upper = true
			b.WriteRune(r)
		} else if upper {
			b.WriteRune(unicode.ToUpper(r))
			upper = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Snake converts s to snake_case.
func Snake(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 4)
	runes := []rune(strings.TrimSpace(s))
	for i, r := range runes {
		if unicode.IsSpace(r) || r == '-' {
			b.WriteRune('_')
			continue
		}
		if unicode.IsUpper(r) && i > 0 && !unicode.IsUpper(runes[i-1]) && !unicode.IsSpace(runes[i-1]) {
			b.WriteRune('_')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// Camel converts s to camelCase.
func Camel(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || r == '_' || r == '-'
	})
	if len(words) == 0 {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i, w := range words {
		if w == "" {
			continue
		}
		if i == 0 {
			b.WriteString(strings.ToLower(w))
		} else {
			runes := []rune(w)
			b.WriteRune(unicode.ToUpper(runes[0]))
			b.WriteString(strings.ToLower(string(runes[1:])))
		}
	}
	return b.String()
}

// ─── Row Builder ──────────────────────────────────────────────────────────────

// Row builds a single formatted line with aligned columns.
// Uses a strings.Builder — writes directly, no intermediate allocations.
type Row struct {
	b strings.Builder
}

// NewRow creates a new Row builder.
func NewRow() *Row { return &Row{} }

// Left appends a left-aligned column of given width.
func (r *Row) Left(val any, width int) *Row {
	s := valueToString(val)
	r.b.WriteString(s)
	pad := width - utf8.RuneCountInString(s)
	for ; pad > 0; pad-- {
		r.b.WriteByte(' ')
	}
	return r
}

// Right appends a right-aligned column of given width.
// Pass optional precision for floats.
func (r *Row) Right(val any, width int, precision ...int) *Row {
	if len(precision) > 0 {
		// Use fmt only for float precision — no way around it cleanly
		fmt.Fprintf(&r.b, fmt.Sprintf("%%%d.%df", width, precision[0]), val)
	} else {
		s := valueToString(val)
		pad := width - utf8.RuneCountInString(s)
		for ; pad > 0; pad-- {
			r.b.WriteByte(' ')
		}
		r.b.WriteString(s)
	}
	return r
}

// CenterCol appends a center-aligned column of given width.
func (r *Row) CenterCol(val any, width int) *Row {
	r.b.WriteString(Center(valueToString(val), width, ' '))
	return r
}

// Sep appends a literal separator.
func (r *Row) Sep(s string) *Row {
	r.b.WriteByte(' ')
	r.b.WriteString(s)
	r.b.WriteByte(' ')
	return r
}

// String returns the built row as a string.
func (r *Row) String() string { return r.b.String() }

// Print prints the row followed by a newline.
func (r *Row) Print() { fmt.Println(r.b.String()) }

// valueToString converts common types without fmt overhead.
func valueToString(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ─── Table Builder ────────────────────────────────────────────────────────────

// Align controls column alignment.
type Align int

const (
	AlignLeft   Align = iota
	AlignRight  Align = iota
	AlignCenter Align = iota
)

type column struct {
	header    string
	width     int
	align     Align
	precision int // -1 = not set
	color     Color
	// pre-built format verbs to avoid rebuilding on every row
	verbLeft  string
	verbRight string
}

// Table builds aligned, readable tables from structured data.
type Table struct {
	cols        []column
	separator   string
	headerColor Color
	altColor    Color
	rowIndex    int
	// reusable row buffer — avoids allocating []string on every Row() call
	rowBuf []string
	// pre-calculated exact row width for Builder.Grow
	rowWidth int
}

// NewTable creates a new Table with default " | " separator.
func NewTable() *Table {
	return &Table{
		separator:   " | ",
		headerColor: ColorNone,
		altColor:    ColorNone,
	}
}

func (t *Table) Separator(s string) *Table  { t.separator = s; return t }
func (t *Table) HeaderColor(c Color) *Table { t.headerColor = c; return t }
func (t *Table) AltRowColor(c Color) *Table { t.altColor = c; return t }

// Col adds a column. Pre-builds format verbs immediately so Row() does no string building.
func (t *Table) Col(header string, width int, align Align, precision ...int) *Table {
	p := -1
	if len(precision) > 0 {
		p = precision[0]
	}
	c := column{
		header:    header,
		width:     width,
		align:     align,
		precision: p,
	}
	if p >= 0 {
		c.verbLeft = fmt.Sprintf("%%-%d.%df", width, p)
		c.verbRight = fmt.Sprintf("%%%d.%df", width, p)
	}
	t.cols = append(t.cols, c)
	t.rebuildMetrics()
	return t
}

// ColColor sets the color of the last added column.
func (t *Table) ColColor(c Color) *Table {
	if len(t.cols) > 0 {
		t.cols[len(t.cols)-1].color = c
	}
	return t
}

// rebuildMetrics recalculates rowWidth and resets rowBuf after column changes.
func (t *Table) rebuildMetrics() {
	n := len(t.cols)
	t.rowBuf = make([]string, n)
	sepLen := len(t.separator)
	total := 0
	for _, c := range t.cols {
		total += c.width
	}
	if n > 1 {
		total += sepLen * (n - 1)
	}
	t.rowWidth = total
}

// AutoWidth scans data to find the max content width for each column.
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
	// rebuild verbs after width change
	for i := range t.cols {
		c := &t.cols[i]
		if c.precision >= 0 {
			c.verbLeft = fmt.Sprintf("%%-%d.%df", c.width, c.precision)
			c.verbRight = fmt.Sprintf("%%%d.%df", c.width, c.precision)
		}
	}
	t.rebuildMetrics()
}

// formatCell formats a single cell into the reusable row buffer slot.
func (t *Table) formatCell(val any, c *column) string {
	if c.precision >= 0 {
		if c.align == AlignLeft {
			return fmt.Sprintf(c.verbLeft, val)
		}
		return fmt.Sprintf(c.verbRight, val)
	}
	s := valueToString(val)
	switch c.align {
	case AlignLeft:
		n := utf8.RuneCountInString(s)
		pad := c.width - n
		if pad <= 0 {
			return s
		}
		var b strings.Builder
		b.Grow(c.width)
		b.WriteString(s)
		for ; pad > 0; pad-- {
			b.WriteByte(' ')
		}
		return b.String()
	case AlignRight:
		n := utf8.RuneCountInString(s)
		pad := c.width - n
		if pad <= 0 {
			return s
		}
		var b strings.Builder
		b.Grow(c.width)
		for ; pad > 0; pad-- {
			b.WriteByte(' ')
		}
		b.WriteString(s)
		return b.String()
	default: // AlignCenter
		return Center(s, c.width, ' ')
	}
}

// joinRow assembles the row buffer into a single string with the separator.
// Uses a Builder pre-grown to the exact output size.
func (t *Table) joinRow(parts []string) string {
	var b strings.Builder
	b.Grow(t.rowWidth + 16) // +16 for ANSI codes if any
	for i, p := range parts {
		if i > 0 {
			b.WriteString(t.separator)
		}
		b.WriteString(p)
	}
	return b.String()
}

// Header prints the header row and a divider line.
func (t *Table) Header() {
	for i, c := range t.cols {
		h := Pad(c.header, c.width, ' ')
		if t.headerColor != ColorNone {
			h = colorize(h, t.headerColor)
		}
		t.rowBuf[i] = h
	}
	fmt.Println(t.joinRow(t.rowBuf))

	// divider
	for i, c := range t.cols {
		t.rowBuf[i] = strings.Repeat("-", c.width)
	}
	fmt.Println(t.joinRow(t.rowBuf))
	t.rowIndex = 0
}

// Row prints a data row. Values must match column order.
func (t *Table) Row(vals ...any) {
	for i := range t.cols {
		if i >= len(vals) {
			t.rowBuf[i] = strings.Repeat(" ", t.cols[i].width)
			continue
		}
		cell := t.formatCell(vals[i], &t.cols[i])
		if t.cols[i].color != ColorNone {
			cell = colorize(cell, t.cols[i].color)
		}
		t.rowBuf[i] = cell
	}
	line := t.joinRow(t.rowBuf)
	if t.altColor != ColorNone && t.rowIndex%2 == 1 {
		line = colorize(line, t.altColor)
	}
	fmt.Println(line)
	t.rowIndex++
}

// String renders the full table as a string without printing.
func (t *Table) String(rows [][]any) string {
	var b strings.Builder
	// exact capacity: (rowWidth + 1) * (numRows + 2 header lines)
	b.Grow((t.rowWidth + 1) * (len(rows) + 2))

	buf := make([]string, len(t.cols))

	// header
	for i, c := range t.cols {
		buf[i] = Pad(c.header, c.width, ' ')
	}
	b.WriteString(t.joinRow(buf))
	b.WriteByte('\n')

	for i, c := range t.cols {
		buf[i] = strings.Repeat("-", c.width)
	}
	b.WriteString(t.joinRow(buf))
	b.WriteByte('\n')

	// data rows
	for _, row := range rows {
		for i := range t.cols {
			if i >= len(row) {
				buf[i] = strings.Repeat(" ", t.cols[i].width)
				continue
			}
			buf[i] = t.formatCell(row[i], &t.cols[i])
		}
		b.WriteString(t.joinRow(buf))
		b.WriteByte('\n')
	}
	return b.String()
}
