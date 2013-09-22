package main

import (
	// "fmt"
	"io/ioutil"
	"testing"
)

type parseConfigFileCase struct {
	name string
	pathname string
	expected configuration
	err string
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

func TestParseConfigFileErrors(t *testing.T) {
	cases := []parseConfigFileCase{
		// {
		// 	name: "missing value",
		// },
	}
	for _, item := range cases {
		config, err := parseConfigFile(item.pathname)
		if err.Error() != item.err {
			t.Errorf("Case: %v; Expected error: %v; Actual error: %#v\n", item.name, item.err, err.Error())
		}
		if config != nil {
			t.Errorf("Case: %v; Expected: nil; Actual: %#v\n", item.name, config)
		}
	}
}

func TestParseConfigFile(t *testing.T) {
	pathname := "test/config"
	contents := "Alpha=One\nBravo=Two\n[Section1]\nCharlie=Three\n" +
		"[Section2]\nDelta=Four\n"
	err := ioutil.WriteFile(pathname, []byte(contents), 0600)
	if err != nil {
		t.Errorf("cannot write fixture file: %s\n", pathname)
	}

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

// can accept either \r\n, \r, or \n
