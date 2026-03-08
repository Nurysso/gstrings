# gstring — Usage Guide

## Table of Contents

1. [Installation](https://www.google.com/search?q=%23installation)
2. [Named Interpolation](https://www.google.com/search?q=%23named-interpolation)
3. [WithStruct](https://www.google.com/search?q=%23withstruct)
4. [Row Builder](https://www.google.com/search?q=%23row-builder)
5. [Table Builder](https://www.google.com/search?q=%23table-builder)
6. [String Utilities](https://www.google.com/search?q=%23string-utilities)
7. [Colors](https://www.google.com/search?q=%23colors)
8. [Recipes](https://www.google.com/search?q=%23recipes)

---

## Installation

```bash
go get github.com/Nurysso/gstring
```

### Import Path

Because the library is now structured with the source code at the root of the repository, the import path is clean:

```go
import "github.com/Nurysso/gstring"
```

**Local development**:

```go
import "gstring"
```

### Project Structure

Run tests and examples from the root:

```bash
go test .                # Run root package tests
go run examples/main.go  # Run the example code

```

---

## Named Interpolation

Replace positional `%[1]d` fmt verbs with readable `{key}` placeholders. Templates are parsed once and cached — repeat calls on the same template do zero parsing work.

### Sprintf

Returns the formatted string.

```go
result := gstring.Sprintf(
    "{name} (id:{id:03d}) -> balance: {balance:.2f}",
    gstring.With("id", 1, "name", "Alice", "balance", 1234.5),
)
// => "Alice (id:001) -> balance: 1234.50"

```

### Placeholder syntax

| Placeholder     | Equivalent | Output         |
| --------------- | ---------- | -------------- |
| `{name}`        | `%v`       | `Alice`        |
| `{id:03d}`      | `%03d`     | `001`          |
| `{balance:.2f}` | `%.2f`     | `1234.50`      |
| `{val:>12s}`    | `%12s`     | `       hello` |

---

## WithStruct

Extract exported fields from a struct directly into a `Vars`. Field names become the placeholder keys.

```go
type User struct {
    ID      int
    Name    string
    Balance float64
}

u := User{1, "Alice", 1234.5}

gstring.Println(
    "{Name} (id:{ID:03d}) -> balance: {Balance:.2f}",
    gstring.WithStruct(u),
)

```

---

## Row Builder

Build a single formatted line with fluent chained column calls.

```go
gstring.NewRow().
    Left(u.ID, 5).Sep("|").
    Left(u.Name, 12).Sep("|").
    Right(u.Balance, 12, 2).
    Print()
// => "1     | Alice        |      1234.50"

```

---

## Table Builder

Declare columns once with `Col()`, then call `Row()` for each data entry. Column widths and format verbs are pre-built at declaration time for maximum performance.

### Basic table

```go
t := gstring.NewTable()
t.Col("ID", 5, gstring.AlignRight).
    Col("Name", 12, gstring.AlignLeft).
    Col("Balance", 12, gstring.AlignRight, 2)

t.Header()
t.Row(1, "Alice", 1234.5)
t.Row(2, "Bob", 98.12)

```

### AutoWidth

Let the table compute column widths from your data automatically.

```go
rows := [][]any{{1, "Alice", 1234.5}, {2, "Bob", 98.12}}
t.AutoWidth(rows)  // Scans data to find the widest content per column
t.Header()

```

---

## String Utilities

High-performance utilities with full UTF-8/Rune support.

| Method       | Usage                                | Result             |
| ------------ | ------------------------------------ | ------------------ |
| **Truncate** | `gstring.Truncate("Hello World", 8)` | `"Hello..."`       |
| **Snake**    | `gstring.Snake("CamelCase")`         | `"camel_case"`     |
| **Camel**    | `gstring.Camel("snake_case")`        | `"snakeCase"`      |
| **Wrap**     | `gstring.Wrap(longText, 20)`         | `Line-broken text` |

---

## Colors

ANSI color constants for use with `Table`, `Row`, and `colorize`.

```go
t := gstring.NewTable().
    HeaderColor(gstring.ColorCyan).
    AltRowColor(gstring.ColorGray)

t.Col("Name", 12, gstring.AlignLeft).
    Col("Balance", 10, gstring.AlignRight).ColColor(gstring.ColorGreen)

```

---

## Recipes

### CLI Report Generation

```go
func printInventory(products []Product) {
    t := gstring.NewTable().HeaderColor(gstring.ColorBold)
    t.Col("SKU", 10, gstring.AlignLeft).
      Col("Name", 20, gstring.AlignLeft).
      Col("Price", 10, gstring.AlignRight, 2).ColColor(gstring.ColorGreen)

    t.Header()
    for _, p := range products {
        t.Row(p.SKU, p.Name, p.Price)
    }
}

```

### Slug generation pipeline

```go
func toSlug(s string) string {
    // Strip -> Snake -> Replace underscores with hyphens
    return strings.ReplaceAll(gstring.Snake(gstring.Strip(s)), "_", "-")
}

```
