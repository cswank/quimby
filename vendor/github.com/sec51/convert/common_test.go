package convert

import "testing"

func TestRound(t *testing.T) {

	// TODO: test negative numbers, although not used in our case

	input := float64(3.7)
	expected := uint64(4)
	result := Round(input)
	if result != expected {
		t.Fatalf("Expected %d - got %d\n", expected, result)
	}

	input = float64(3.5)
	expected = uint64(4)
	result = Round(input)
	if result != expected {
		t.Fatalf("Expected %d - got %d\n", expected, result)
	}

	input = float64(3.499999999)
	expected = uint64(3)
	result = Round(input)
	if result != expected {
		t.Fatalf("Expected %d - got %d\n", expected, result)
	}

	input = float64(3.0)
	expected = uint64(3)
	result = Round(input)
	if result != expected {
		t.Fatalf("Expected %d - got %d\n", expected, result)
	}

	input = float64(3.9999)
	expected = uint64(4)
	result = Round(input)
	if result != expected {
		t.Fatalf("Expected %d - got %d\n", expected, result)
	}
}
