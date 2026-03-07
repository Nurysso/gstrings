package main

import (
    "gstring/gstring"
)

type User struct {
	ID      int
	Name    string
	Balance float64
	Active  bool
	Subscription bool
}

func main() {
	users := []User{
		{1, "Alice", 1234.5, true, false},
		{2, "Bob", 98.12, false, true},
		{3, "Charlotte", 100000.99, true, true},
	}

	// ── Named interpolation ──────────────────────────────────────────────────
	println("=== Named Interpolation ===")
	for _, u := range users {
		gstring.Println(
			"{name} (id:{id:03d}) -> balance: {balance:.2f} active:{active}",
			gstring.With("id", u.ID, "name", u.Name, "balance", u.Balance, "active", u.Active),
		)
	}

	println()

	// ── Row builder ──────────────────────────────────────────────────────────
	println("=== Row Builder ===")
	for _, u := range users {
		gstring.NewRow().
			Left(u.ID, 5).Sep("|").
			Left(u.Name, 12).Sep("|").
			Right(u.Balance, 12, 2).Sep("|").
			Left(u.Active, 6).Sep("|").
			Right(u.Subscription, 6).
			Print()
	}

	println()

	// ── Table builder ────────────────────────────────────────────────────────
	println("=== Table Builder ===")
	t := gstring.NewTable()
	t.Col("ID", 5, gstring.AlignRight).
		Col("Name", 12, gstring.AlignLeft).
		Col("Balance", 12, gstring.AlignRight, 2).
		Col("Active", 6, gstring.AlignLeft).
		Col("Subscription", 6, gstring.AlignLeft)
	t.Header()
	for _, u := range users {
		t.Row(u.ID, u.Name, u.Balance, u.Active, u.Subscription)
	}
}
