package main

import (
	"testing"
)

func TestString(t *testing.T) {
	var (
		got      string
		expected string
	)

	got = Pointer{13, 3}.String()
	expected = "S13E03"
	if got != expected {
		t.Errorf("Expected '%s', got '%s'", expected, got)
	}

	got = Pointer{2, 123}.String()
	expected = "S02E123"
	if got != expected {
		t.Errorf("Expected '%s', got '%s'", expected, got)
	}
}

func TestPointerFromString(t *testing.T) {
	var (
		got      Pointer
		expected Pointer
	)

	got, _ = PointerFromString("S04E221")
	expected = Pointer{4, 221}
	if got != expected {
		t.Errorf("Expected '%v', got '%v'", expected, got)
	}

	got, _ = PointerFromString("S8E1")
	expected = Pointer{8, 1}
	if got != expected {
		t.Errorf("Expected '%v', got '%v'", expected, got)
	}
}
