package main

import (
	"testing"
)

const (
	TestDefault = "application/json"
)

func TestParseAcceptStringEmptyStringReturnsDefault(t *testing.T) {
	actual := parseAcceptString("", TestDefault)
	expected := TestDefault
	if actual != expected {
		t.Errorf("Expected: %#v; Actual: %#v\n", expected, actual)
	}
}

// func TestParseAcceptStringUnknownReturnsError(t *testing.T) {
// 	actual := parseAcceptString("unknown", TestDefault)
// 	expected := TestDefault
// 	if actual != expected {
// 		t.Errorf("Expected: %#v; Actual: %#v\n", expected, actual)
// 	}
// }

func TestParseAcceptStringValid(t *testing.T) {
	actual := parseAcceptString("application/json", TestDefault)
	expected := "application/json"
	if actual != expected {
		t.Errorf("Expected: %#v; Actual: %#v\n", expected, actual)
	}
}
