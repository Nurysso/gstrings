package gstring_test

import (
	"testing"
	"gstring/gstring"
	// "github.com/Nurysso/gstring/gstring"
)

func TestFormat_basic(t *testing.T) {
	got := gstring.Format("Hello {name}!", gstring.With("name", "Alice"))
	want := "Hello Alice!"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestFormat_multipleVars(t *testing.T) {
	got := gstring.Format("{name} is {age} years old", gstring.With("name", "Bob", "age", 30))
	want := "Bob is 30 years old"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestFormat_withSpec(t *testing.T) {
	got := gstring.Format("balance: {bal:.2f}", gstring.With("bal", 1234.5))
	want := "balance: 1234.50"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestFormat_missingKey(t *testing.T) {
	got := gstring.Format("hello {missing}", gstring.Vars{})
	want := "hello {missing}"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestFormat_paddedInt(t *testing.T) {
	got := gstring.Format("id: {id:05d}", gstring.With("id", 7))
	want := "id: 00007"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestRow_string(t *testing.T) {
	got := gstring.NewRow().Left("Alice", 10).Sep("|").Right(1234.5, 10, 2).String()
	want := "Alice      |       1234.50"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
