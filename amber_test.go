package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

////////////////////////////////////////

func TestUrlFromRemoteAndResource(t *testing.T) {
	rem := remote{hostname: "localhost", port: 8080}
	cHash := "def"
	actual := urlFromRemoteAndResource(&rem, cHash)
	expected := "http://localhost:8080/resource/def"
	if actual != expected {
		t.Errorf("expected: %v, actual: %v", expected, actual)
	}
}

////////////////////////////////////////

func TestCheckHashReturnsTrueWhenHashMatch(t *testing.T) {
	bytes := []byte("just some blob of data")
	actual, err := checkHash("sha256", bytes, "0f60742ed4cc07265128fda3343cd4932bdecb1eeceea73653334259d6a02af0")
	expected := true
	if actual != expected {
		t.Errorf("expected: %v, actual: %v", expected, actual)
	}
	if err != nil {
		t.Errorf("expected: %v, actual: %v", nil, err)
	}
}

func TestCheckHashReturnsFalseWhenHashMismatch(t *testing.T) {
	bytes := []byte("just some different blob of data")
	actual, err := checkHash("sha256", bytes, "0f60742ed4cc07265128fda3343cd4932bdecb1eeceea73653334259d6a02af0")
	expected := false
	if actual != expected {
		t.Errorf("expected: %v, actual: %v", expected, actual)
	}
	if err == nil {
		t.Errorf("expected: %v, actual: %v", errors.New("hash mismatch"), err)
	}
}

////////////////////////////////////////

func TestComputeHashReturnsErrorWhenUnknownHash(t *testing.T) {
	actual, err := computeHash("non-existant-hash", make([]byte, 10))
	expected := ""
	if actual != expected {
		t.Errorf("expected: %v, actual: %v", expected, actual)
	}
	if err == nil {
		t.Errorf("expected: %v, actual: %v", errors.New("hash mismatch"), err)
	}
}

func TestComputeHashReturnsHashStringOfBytes(t *testing.T) {
	bytes := []byte("just some blob of data")
	actual, err := computeHash("sha256", bytes)
	expected := "0f60742ed4cc07265128fda3343cd4932bdecb1eeceea73653334259d6a02af0"
	if actual != expected {
		t.Errorf("expected: %v, actual: %v", expected, actual)
	}
	if err != nil {
		t.Errorf("expected: %v, actual: %v", nil, err)
	}
}

////////////////////////////////////////

func TestInvalidHashFormat(t *testing.T) {
	var cases = map[string]bool{
		"":        true,
		"flubber": true,
		"..":      true,
		".foo":    true,
		"foo.":    true,
		"foo.bar": true,
		"abc":     false,
		"ABC":     true,
		"123":     false,
	}

	for hash, expected := range cases {
		actual := isHashInvalid(hash)
		if actual != expected {
			t.Errorf("Data mismatch:\n   actual: [%v]\n expected: [%v]\n", actual, expected)
		}
	}
}

////////////////////////////////////////

func TestResourceRequest2metadataRejectsInvalidURL(t *testing.T) {
	var cases = map[string]string{
		"":                 "invalid url: ",
		"/":                "invalid url: /",
		"flubber":          "invalid url: flubber",
		"resource/../foo":  "invalid url: resource/../foo",
		"/resource/../foo": "invalid url: /resource/../foo",
		"/resource/.foo":   "invalid url: /resource/.foo",
		"/resource/":       "invalid url: /resource/",
		"/resource/ABC":    "invalid url: /resource/ABC",
	}

	for path, expected := range cases {
		r := &http.Request{URL: &url.URL{Path: path}}
		_, actual := resourceRequest2metadata(r)
		if actual.Error() != expected {
			t.Errorf("Data mismatch:\n   actual: [%s]\n expected: [%s]\n", actual, expected)
		}
	}
}

func TestResourceRequest2metadataRejectsInvalidUser(t *testing.T) {
	var cases = map[string]string{
		"":       "invalid: ",
		"../abc": "invalid: ../abc",
		"ABC":    "invalid: ABC",
	}
	for user, expected := range cases {
		path := "/resource/abc"
		headers := map[string][]string{
			"X-Amber-User": {user},
		}
		r := &http.Request{URL: &url.URL{Path: path}, Header: headers}
		meta, err := resourceRequest2metadata(r)
		if err.Error() != expected {
			t.Errorf("Data mismatch:\n   actual: [%s]\n expected: [%s]\n", err.Error(), expected)
		}
		if meta.bpathname != "" {
			t.Errorf("Data mismatch:\n   actual: [%s]\n expected: [%s]\n", meta.bpathname, "")
		}
	}
}

func TestResourceRequest2metadata(t *testing.T) {
	var cases = map[string]map[string]string{
		"resource/abc123/users/0000abcdef": {
			"cHash": "abc123",
			"uName": "0000abcdef",
		},
		"resource/abc123/users/-": {
			"cHash": "abc123",
			"uName": "-",
		},
	}

	for expected, req := range cases {
		path := fmt.Sprintf("/resource/%s", req["cHash"])
		headers := map[string][]string{
			"X-Amber-User": {req["uName"]},
		}
		r := &http.Request{URL: &url.URL{Path: path}, Header: headers}
		actual, err := resourceRequest2metadata(r)
		if actual.bpathname != expected {
			t.Errorf("Data mismatch:\n   actual: [%s]\n expected: [%s]\n", actual, expected)
		}
		if err != nil {
			t.Errorf("Data mismatch:\n   actual: [%s]\n expected: [%s]\n", err.Error(), nil)
		}
	}
}
