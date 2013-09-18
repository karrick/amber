package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func directoryContents(dirname string) (contents []string, err error) {
	fileInfos, err := ioutil.ReadDir(dirname)
	if err != nil {
		return
	}
	contents = make([]string, len(fileInfos))
	for i, fi := range fileInfos {
		contents[i] = fi.Name()
	}
	return
}
func directoryIncludes(dirname string, child string) (bool, error) {
	pathname := filepath.Join(dirname, child)
	_, err := os.Stat(pathname)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
func includesString(xs []string, item string) bool {
	for _, x := range xs {
		if item == x {
			return true
		}
	}
	return false
}
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, value := range a {
		if value != b[i] {
			return false
		}
	}
	return true
}
func test1(dirname string) error {
	fmt.Printf("#### Readdirnames\n")

	dir, err := os.Open(dirname)
	if err != nil {
		return err
	}
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	if err != nil {
		panic(err)
	}

	for _, name := range names {
		fmt.Printf("NAME: [%v]\n", name)
	}
	return nil
}
func test2(dirname string) error {
	fmt.Printf("#### Readdir\n")

	dir, err := os.Open(dirname)
	if err != nil {
		panic(err)
	}
	defer dir.Close()

	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		panic(err)
	}

	for _, fi := range fileInfos {
		fmt.Printf("FI: [%v]\n", fi.Name())
	}
	return nil
}
func readdir_vs_readdirnames() {
	if err := test1("."); err != nil {
		log.Fatal(err)
	}
	if err := test2("foo"); err != nil {
		log.Fatal(err)
	}
}

// Convert global address of file to pathname on disk
func AddressToPathname(address string) string {
	prefix := address[:3]
	suffix := address[3:]
	return filepath.Join("resource", prefix, suffix)
}

////////////////////////////////////////

func TestdirectoryContents(t *testing.T) {
	// setup
	save_pwd, _ := os.Getwd()
	if err := os.MkdirAll("test/artifacts/.amber", 0700); err != nil {
		t.Error("Error creating artifacts: ", err)
	}
	defer os.RemoveAll("test/artifacts")
	defer os.Chdir(save_pwd)

	// test
	actual, err := directoryContents("test/artifacts")
	if err != nil {
		t.Error(err)
	}

	// verify pwd same
	if pwd, _ := os.Getwd(); save_pwd != pwd {
		t.Errorf("Expected: %v, actual: %v", save_pwd, pwd)
	}

	// verify
	expected := []string{".amber"}
	if !stringSlicesEqual(expected, actual) {
		t.Errorf("Expected: %v, actual: %v", expected, actual)
	}
}
func TestIncludesItem(t *testing.T) {
	expected := true
	if actual := includesString([]string{"a", "b", "c"}, "a"); expected != actual {
		t.Errorf("Expected: %v, Actual: %v", expected, actual)
	}

	expected = true
	if actual := includesString([]string{"a", "b", "c"}, "b"); expected != actual {
		t.Errorf("Expected: %v, Actual: %v", expected, actual)
	}

	expected = false
	if actual := includesString([]string{"a", "b", "c"}, "d"); expected != actual {
		t.Errorf("Expected: %v, Actual: %v", expected, actual)
	}
}
func TestAddressToPathname(t *testing.T) {
	components := []string{"resource", "01e", "06a68df2f0598042449c4088842bb4e92ca75"}
	expected := filepath.Join(components...)

	actual := AddressToPathname("01e06a68df2f0598042449c4088842bb4e92ca75")

	if actual != expected {
		t.Errorf("Data mismatch:\n   actual: [%s]\n expected: [%s]\n", actual, expected)
	}
}
