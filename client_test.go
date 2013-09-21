package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

////////////////////////////////////////

func TestRespositoryRootReturnsRootWhenThere(t *testing.T) {
	// setup
	pwd, _ := os.Getwd()
	if err := os.MkdirAll("test/artifacts/.amber", 0700); err != nil {
		t.Errorf("Error creating artifacts: ", err)
	}
	if err := os.Chdir("test/artifacts"); err != nil {
		t.Errorf("Error changing directory: ", err)
	}
	defer os.RemoveAll("test/artifacts")
	defer os.Chdir(pwd)

	// test
	actual, err := repositoryRoot(".amber")
	if err != nil {
		t.Error(err)
	}

	// verify
	var expected = filepath.Join(pwd, "test", "artifacts", ".amber")
	if actual != expected {
		t.Errorf("Data mismatch:\n   actual: [%s]\n expected: [%s]\n", actual, expected)
	}
}

func TestRespositoryRootReturnsRootWhenBelow(t *testing.T) {
	// setup
	pwd, _ := os.Getwd()
	if err := os.MkdirAll("test/artifacts/.amber", 0700); err != nil {
		t.Errorf("Error creating artifacts: ", err)
	}
	if err := os.MkdirAll("test/artifacts/foo/bar", 0700); err != nil {
		t.Errorf("Error creating artifacts: ", err)
	}
	if err := os.Chdir("test/artifacts/foo/bar"); err != nil {
		t.Errorf("Error changing directory: ", err)
	}
	defer os.RemoveAll("test/artifacts")
	defer os.Chdir(pwd)

	// test
	actual, err := repositoryRoot(".amber")
	if err != nil {
		t.Error(err)
	}

	// verify
	var expected = filepath.Join(pwd, "test", "artifacts", ".amber")
	if actual != expected {
		t.Errorf("Data mismatch:\n   actual: [%s]\n expected: [%s]\n", actual, expected)
	}
}

func TestRespositoryRootReturnsErrorWhenDotAmberIsFile(t *testing.T) {
	// setup
	pwd, _ := os.Getwd()
	if err := os.MkdirAll("test/artifacts/foo/bar", 0700); err != nil {
		t.Errorf("Error creating artifacts: ", err)
	}
	if err := ioutil.WriteFile("test/artifacts/.amber", []byte{}, 0600); err != nil {
		t.Error(err)
	}
	if err := os.Chdir("test/artifacts/foo/bar"); err != nil {
		t.Errorf("Error changing directory: ", err)
	}
	defer os.RemoveAll("test/artifacts")
	defer os.Chdir(pwd)

	// test
	_, err := repositoryRoot(".amber")

	// verify
	if err == nil {
		t.Errorf("expected error %v", ErrNoRepos)
	}
}

func TestRespositoryRootReturnsErrorWhenNotFound(t *testing.T) {
	// setup
	pwd, _ := os.Getwd()
	if err := os.MkdirAll("test/artifacts/foo/bar", 0700); err != nil {
		t.Errorf("Error creating artifacts: ", err)
	}
	if err := os.Chdir("test/artifacts/foo/bar"); err != nil {
		t.Errorf("Error changing directory: ", err)
	}
	defer os.RemoveAll("test/artifacts")
	defer os.Chdir(pwd)

	// test
	repos, err := repositoryRoot(".amber-should-never-see")

	// verify
	if repos != "" {
		t.Errorf("expected: %v, actual: %v", "", repos)
	}
	if err == nil {
		t.Errorf("expected error %v", ErrNoRepos)
	}
}

////////////////////////////////////////

////////////////////////////////////////

func TestParseUriList(t *testing.T) {
	uriList := "# this is a comment\r\nhttp://example.com/1\r\nhttp://example.com/2"
	expected := []string{
		"http://example.com/1",
		"http://example.com/2",
	}
	actual := parseUriList(uriList)
	if len(expected) != len(actual) {
		t.Errorf("expected: %v, actual: %v", len(expected), len(actual))
	}
	for i, _ := range expected {
		if expected[i] != actual[i] {
			t.Errorf("expected: %v, actual: %v", expected[i], actual[i])
		}
	}
}

type parseUrcCase struct {
	name		string
	blob		[]byte
	output		metadata
	err			error
}

func TestParseUrcCatchesErrors(t *testing.T) {
	cases := []parseUrcCase{
		{
			name:	"three fields",
			blob:	[]byte("one two three"),
			output:	metadata{},
			err:	fmt.Errorf("invalid line format: one two three"),
		},
		{
			name:	"splits on crlf",
			blob:	[]byte("X-Amber-Hash: foo\nX-Amber-Encryption: bar\n"),
			output:	metadata{},
			err:	errors.New("invalid line format: X-Amber-Hash: foo\nX-Amber-Encryption: bar\n"),
		},
	}
	for _, item := range cases {
		output, err := parseUrc(item.blob)
		if item.err.Error() != err.Error() {
			t.Errorf("Case: %v; Expected error: %v; Acutal error: %v\n", item.name, item.err.Error(), err.Error())
		}
		if fmt.Sprintf("%#v", item.output) != fmt.Sprintf("%#v", output) {
			t.Errorf("Case: %v; Expected: %#v; Acutal: %#v\n", item.name, item.output, output)
		}
	}
}

func TestParseUrcExpectedResults(t *testing.T) {
	cases := []parseUrcCase{
		{
			name:	"gets hash name",
			blob:	[]byte("X-Amber-Hash: foo"),
			output:	metadata{hName: "foo"},
			err:	nil,
		},
		{
			name:	"gets encryption name",
			blob:	[]byte("X-Amber-Encryption: bar"),
			output:	metadata{eName: "bar"},
			err:	nil,
		},
		{
			name:	"gets hash and encryption name",
			blob:	[]byte("X-Amber-Hash: foo\r\nX-Amber-Encryption: bar\r\n"),
			output:	metadata{hName: "foo", eName: "bar"},
			err:	nil,
		},
		{
			name:	"stops at empty line",
			blob:	[]byte("X-Amber-Hash: foo\r\n\r\nX-Amber-Encryption: bar\r\n"),
			output:	metadata{hName: "foo"},
			err:	nil,
		},
	}
	for _, item := range cases {
		output, err := parseUrc(item.blob)
		if err != nil {
			t.Errorf("Case: %v; Didn't expect error: %v\n", item.name, err.Error())
		}
		if fmt.Sprintf("%#v", item.output) != fmt.Sprintf("%#v", output) {
			t.Errorf("Case: %v; Expected: %#v; Acutal: %#v\n", item.name, item.output, output)
		}
	}
}
