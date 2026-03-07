// Package gstring provides readable, simple string formatting for Go.
// It replaces cryptic fmt verbs with named placeholders and a fluent table/column API.

package gstring

import (
	"fmt"
	"regexp"
	"strings"
)

// ─── Named Interpolation ────────────────────────────────────────────────────

var placeholderRe = regexp.MustCompile(`\{(\w+)(?::([^}]*))?\}`)

// Vars holds named values for interpolation.
type Vars map[string]any

// With is a convenience constructor for Vars.

//	gstring.With("name", "Alice", "age", 30)
func With(pairs ...any) Vars {
	v := Vars{}
	for i := 0; i+1 < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			continue
		}
		v[key] = pairs[i+1]
	}
	return v
}

// Format replaces named placeholders in the template with values from vars.
// Supports optional format specs after a colon:
//
//	{name}        → plain string conversion
//	{balance:.2f} → fmt verb applied to value
//	{id:05d}      → padded integer
//
// Example:
//
//	gstring.Format("{name} ({id}) -> balance: {balance:.2f}", gstring.With("id", 1, "name", "Alice", "balance", 1234.5))
//	// => "Alice (1) -> balance: 1234.50"
func Format(template string, vars Vars) string {
	return placeholderRe.ReplaceAllStringFunc(template, func(match string) string {
		sub := placeholderRe.FindStringSubmatch(match)
		if sub == nil {
			return match
		}
		key, spec := sub[1], sub[2]
		val, ok := vars[key]
		if !ok {
			return "{" + key + "}"
		}
		if spec == "" {
			return fmt.Sprintf("%v", val)
		}
		return fmt.Sprintf("%"+spec, val)
	})
}

// Println formats and prints with a newline.
func Println(template string, vars Vars) {
	fmt.Println(Format(template, vars))
}

// ─── Row Builder ────────────────────────────────────────────────────────────

// Row builds a single formatted line with aligned columns.
//
//	gstring.NewRow().Left(u.ID, 5).Sep("|").Left(u.Name, 12).Right(u.Balance, 12, 2).Print()
type Row struct {
	parts []string
}

// NewRow creates a new Row builder.
func NewRow() *Row {
	return &Row{}
}

// Left appends a left-aligned column of given width.
func (r *Row) Left(val any, width int) *Row {
	r.parts = append(r.parts, fmt.Sprintf(fmt.Sprintf("%%-%dv", width), val))
	return r
}

// Right appends a right-aligned column of given width.
// For floats, pass precision as an optional third argument.
func (r *Row) Right(val any, width int, precision ...int) *Row {
	if len(precision) > 0 {
		r.parts = append(r.parts, fmt.Sprintf(fmt.Sprintf("%%%d.%df", width, precision[0]), val))
	} else {
		r.parts = append(r.parts, fmt.Sprintf(fmt.Sprintf("%%%dv", width), val))
	}
	return r
}

// Sep appends a literal separator string (e.g. " | ").
func (r *Row) Sep(s string) *Row {
	r.parts = append(r.parts, " "+s+" ")
	return r
}

// String returns the built row as a string.
func (r *Row) String() string {
	return strings.Join(r.parts, "")
}

// Print prints the row followed by a newline.
func (r *Row) Print() {
	fmt.Println(r.String())
}

// ─── Table Builder ───────────────────────────────────────────────────────────

// Align controls column alignment.
type Align int

const (
	AlignLeft  Align = iota
	AlignRight Align = iota
)

type column struct {
	header    string
	width     int
	align     Align
	precision int // for floats; -1 means not set
}

// Table builds aligned, readable tables from structured data.
//
//	t := gstring.NewTable()
//	t.Col("ID", 5, gstring.AlignRight).Col("Name", 12, gstring.AlignLeft).Col("Balance", 12, gstring.AlignRight, 2)
//	t.Header()
//	t.Row(1, "Alice", 1234.5)
//	t.Row(2, "Bob", 98.12)
type Table struct {
	cols      []column
	separator string
}

// NewTable creates a new Table with default " | " separator.
func NewTable() *Table {
	return &Table{separator: " | "}
}

// Separator sets a custom column separator.
func (t *Table) Separator(s string) *Table {
	t.separator = s
	return t
}

// Col adds a column definition. Optional precision for float columns.
func (t *Table) Col(header string, width int, align Align, precision ...int) *Table {
	p := -1
	if len(precision) > 0 {
		p = precision[0]
	}
	t.cols = append(t.cols, column{header, width, align, p})
	return t
}

// Header prints the header row and a divider.
func (t *Table) Header() {
	parts := make([]string, len(t.cols))
	divparts := make([]string, len(t.cols))
	for i, c := range t.cols {
		parts[i] = fmt.Sprintf(fmt.Sprintf("%%-%ds", c.width), c.header)
		divparts[i] = strings.Repeat("-", c.width)
	}
	fmt.Println(strings.Join(parts, t.separator))
	fmt.Println(strings.Join(divparts, t.separator))
}

// Row prints a data row. Values must match column order.
func (t *Table) Row(vals ...any) {
	parts := make([]string, len(t.cols))
	for i, c := range t.cols {
		if i >= len(vals) {
			parts[i] = strings.Repeat(" ", c.width)
			continue
		}
		val := vals[i]
		var verb string
		if c.precision >= 0 {
			if c.align == AlignLeft {
				verb = fmt.Sprintf("%%-%d.%df", c.width, c.precision)
			} else {
				verb = fmt.Sprintf("%%%d.%df", c.width, c.precision)
			}
		} else {
			if c.align == AlignLeft {
				verb = fmt.Sprintf("%%-%dv", c.width)
			} else {
				verb = fmt.Sprintf("%%%dv", c.width)
			}
		}
		parts[i] = fmt.Sprintf(verb, val)
	}
	fmt.Println(strings.Join(parts, t.separator))
}
