package main

import (
	"os"
	"testing"
)

type parseConfigFileCase struct {
	name     string
	pathname string
	expected configuration
	err      string
}

func TestParseConfigFileMissingFile(t *testing.T) {
	expected := "open no-such-file: no such file or directory"
	config, err := parseConfigFile("no-such-file")
	if err.Error() != expected {
		t.Errorf("Expected error: %v; Actual error: %#v\n", expected, err.Error())
	}
	if config != nil {
		t.Errorf("Expected: nil; Actual: %#v\n", config)
	}
}

func TestParseConfigFile(t *testing.T) {
	pathname := "test/config"
	contents := " ; Comment\n\nAlpha=One\nBravo=Two\n[Section1]\nCharlie=Three\n" +
		"[Section2]\nDelta=Four\n"
	err := writeFile(pathname, []byte(contents))
	if err != nil {
		t.Errorf("cannot write fixture file: %s\n", pathname)
	}
	defer func() { 
		if err := os.RemoveAll("test"); err != nil {
			t.Errorf("failed to remove test directory: %s\n", err)
		}
	}()

	expected := map[string]map[string]string{
		"General": {
			"Alpha": "One",
			"Bravo": "Two",
		},
		"Section1": {
			"Charlie": "Three",
		},
		"Section2": {
			"Delta": "Four",
		},
	}

	config, err := parseConfigFile(pathname)

	if err != nil {
		t.Errorf("Did not expect error: %v\n", err.Error())
	}
	if len(config) != len(expected) {
		t.Errorf("Expected: %#v; Actual: %#v\n", expected, config)
		return
	}
	if config["General"]["Alpha"] != "One" {
		t.Errorf("Expected: %#v; Actual: %#v\n", expected, config)
		return
	}
	if config["General"]["Bravo"] != "Two" {
		t.Errorf("Expected: %#v; Actual: %#v\n", expected, config)
		return
	}
	if config["Section1"]["Charlie"] != "Three" {
		t.Errorf("Expected: %#v; Actual: %#v\n", expected, config)
		return
	}
	if config["Section2"]["Delta"] != "Four" {
		t.Errorf("Expected: %#v; Actual: %#v\n", expected, config)
		return
	}
}
