package main

import (
	"sort"
	"testing"
)

func TestGetEmptyDb(t *testing.T) {
	db := &lockUrnDb{}

	actual, ok := db.get("")
	if ok != false {
		t.Errorf("Expected: %v; Actual: %v\n", false, ok)
	}
	if len(actual) != 0 {
		t.Errorf("Expected: %v; Actual: %v\n", 0, len(actual))
	}
}

func TestAppendOnMissingKey(t *testing.T) {
	db := &lockUrnDb{}

	db.append("key", "value")

	actual, ok := db.get("this key is not there")
	if ok != false {
		t.Errorf("Expected: %v; Actual: %v\n", false, ok)
	}
	if len(actual) != 0 {
		t.Errorf("Expected: %v; Actual: %v\n", 0, len(actual))
	}

	actual, ok = db.get("key")
	if ok != true {
		t.Errorf("Expected: %v; Actual: %v\n", true, ok)
	}
	if len(actual) != 1 {
		t.Errorf("Expected: %v; Actual: %v\n", 1, len(actual))
	}
	if actual[0] != "value" {
		t.Errorf("Expected: %v; Actual: %v\n", "value", actual[0])
	}
}

func TestAppendOnExistingKey(t *testing.T) {
	db := &lockUrnDb{}

	db.append("key", "value1")
	db.append("key", "value2")

	actual, ok := db.get("key")
	if ok != true {
		t.Errorf("Expected: %v; Actual: %v\n", true, ok)
	}
	if len(actual) != 2 {
		t.Errorf("Expected: %v; Actual: %v\n", 2, len(actual))
	}
	if actual[0] != "value1" {
		t.Errorf("Expected: %v; Actual: %v\n", "value", actual[0])
	}
	if actual[1] != "value2" {
		t.Errorf("Expected: %v; Actual: %v\n", "value", actual[1])
	}
}

func TestKeysEmpty(t *testing.T) {
	db := &lockUrnDb{}

	actual := db.keys()
	if len(actual) != 0 {
		t.Errorf("Expected: %v; Actual: %v\n", 0, len(actual))
	}
}

func TestKeysSingleItem(t *testing.T) {
	db := &lockUrnDb{}

	db.append("key1", "value1")
	actual := db.keys()
	if len(actual) != 1 {
		t.Errorf("Expected: %v; Actual: %v\n", 1, len(actual))
	}
	if actual[0] != "key1" {
		t.Errorf("Expected: %v; Actual: %v\n", "key1", actual[0])
	}

	// single key with multiple values should also return only one key
	db.append("key1", "value2")
	actual = db.keys()
	if len(actual) != 1 {
		t.Errorf("Expected: %v; Actual: %v\n", 1, len(actual))
	}
	if actual[0] != "key1" {
		t.Errorf("Expected: %v; Actual: %v\n", "key1", actual[0])
	}
}

func TestKeysMultipleItems(t *testing.T) {
	db := &lockUrnDb{}

	db.append("key1", "value1")
	db.append("key1", "value2")
	db.append("key2", "value1")
	db.append("key2", "value2")

	actual := db.keys()
	if len(actual) != 2 {
		t.Errorf("Expected: %v; Actual: %v\n", 2, len(actual))
	}
	sort.Strings(actual)
	if actual[0] != "key1" {
		t.Errorf("Expected: %v; Actual: %v\n", "key1", actual[0])
	}
	if actual[1] != "key2" {
		t.Errorf("Expected: %v; Actual: %v\n", "key2", actual[0])
	}
}
