package main

import (
	"fmt"
	"github.com/nurysso/gstrings"
)

type User struct {
	ID      int
	Name    string
	Balance float64
	Active  bool
}

func main() {
	users := []User{
		{1, "Pickachu", 1234.5, true},
		{2, "doremon", 98.12, false},
		{3, "Charlotte", 100000.99, true},
	}

	fmt.Println(gstring.Repeat("─", 50))
	fmt.Println(" gstring — examples")
	fmt.Println(gstring.Repeat("─", 50))

	// ── Named interpolation ──────────────────────────────────────────────────
	fmt.Println("\n=== Named Interpolation (With) ===")
	for _, u := range users {
		gstring.Println(
			"{name} (id:{id:03d}) -> balance: {balance:.2f} active:{active}",
			gstring.With("id", u.ID, "name", u.Name, "balance", u.Balance, "active", u.Active),
		)
	}

	// ── WithStruct — new: interpolate directly from struct fields ────────────
	fmt.Println("\n=== Named Interpolation (WithStruct) ===")
	for _, u := range users {
		gstring.Println(
			"{Name} (id:{ID:03d}) -> balance: {Balance:.2f}",
			gstring.WithStruct(u),
		)
	}

	// ── Row builder ──────────────────────────────────────────────────────────
	fmt.Println("\n=== Row Builder ===")
	for _, u := range users {
		gstring.NewRow().
			Left(u.ID, 5).Sep("|").
			Left(u.Name, 12).Sep("|").
			Right(u.Balance, 12, 2).Sep("|").
			Left(u.Active, 6).
			Print()
	}

	// ── Table builder with colors ────────────────────────────────────────────
	fmt.Println("\n=== Table Builder (with colors) ===")
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

	// ── AutoWidth — new: let the table figure out column widths ─────────────
	fmt.Println("\n=== Table Builder (AutoWidth) ===")
	t3 := gstring.NewTable().HeaderColor(gstring.ColorYellow)
	t3.Col("ID", 0, gstring.AlignRight).
		Col("Name", 0, gstring.AlignLeft).
		Col("Balance", 0, gstring.AlignRight, 2).
		Col("Active", 0, gstring.AlignLeft)
	rows3 := [][]any{}
	for _, u := range users {
		rows3 = append(rows3, []any{u.ID, u.Name, u.Balance, u.Active})
	}
	t3.AutoWidth(rows3)
	t3.Header()
	for _, u := range users {
		t3.Row(u.ID, u.Name, u.Balance, u.Active)
	}

	// ── Table.String ─────────────────────────────────────────────────────────
	fmt.Println("\n=== Table.String (returns string, no print) ===")
	t2 := gstring.NewTable()
	t2.Col("ID", 4, gstring.AlignRight).
		Col("Name", 12, gstring.AlignLeft).
		Col("Balance", 10, gstring.AlignRight, 2)
	rows2 := [][]any{}
	for _, u := range users {
		rows2 = append(rows2, []any{u.ID, u.Name, u.Balance})
	}
	fmt.Print(t2.String(rows2))

	// ── String utilities ─────────────────────────────────────────────────────
	fmt.Println("\n=== Truncate ===")
	fmt.Println(gstring.Truncate("Hello, World!", 8)) // Hello...
	fmt.Println(gstring.Truncate("Short", 10))        // Short

	fmt.Println("\n=== Pad ===")
	fmt.Printf("[%s]\n", gstring.Pad("hi", 10, '-'))  // [hi--------]
	fmt.Printf("[%s]\n", gstring.Pad("hi", -10, '-')) // [--------hi]

	fmt.Println("\n=== Center ===")
	fmt.Printf("[%s]\n", gstring.Center("hi", 10, '-')) // [----hi----]

	fmt.Println("\n=== Wrap ===")
	fmt.Println(gstring.Wrap("The quick brown fox jumps over the lazy dog", 20))

	fmt.Println("\n=== Case Conversions ===")
	fmt.Println(gstring.Title("hello world")) // Hello World
	fmt.Println(gstring.Snake("HelloWorld"))  // hello_world
	fmt.Println(gstring.Snake("Hello World")) // hello_world
	fmt.Println(gstring.Camel("hello_world")) // helloWorld
	fmt.Println(gstring.Camel("Hello World")) // helloWorld

	fmt.Println("\n=== Strip & Repeat ===")
	fmt.Printf("[%s]\n", gstring.Strip("   trimmed   "))
	fmt.Println(gstring.Repeat("─", 50))
}
