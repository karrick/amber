package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"flag"
	"fmt"
	"hash"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

////////////////////////////////////////
// types
////////////////////////////////////////

// for the server, someplace to store our hostname, bound port; for
// the client, someplace to store the same information for the remote
// server

type remote struct {
	hostname string
	port     int
}

// metadata for a resource stored in amber

type metadata struct {
	cHash     string // hash of cipher text (name of resource)
	eName     string // name of encryption algorithm
	hName     string // name of hash algorithm
	mpathname string // pathname of resource meta file
	bpathname string // pathname of resource blob file
	size      string // string representation of resource size
	uName     string // user that owns resource, or "-"

	pHash string // hash of plain text
}

////////////////////////////////////////
// global variables
////////////////////////////////////////

const (
	CommunityUName    = "-"
	DefaultEncryption = "rc4"
	DefaultHash       = "sha256"
	MaxUrlLength      = 2083 // IE 9 limitation
	crlf              = "\r\n"
	nis               = "x-amber"
)

// these variables need to be accessible from within functions that
// are called by http server library, so making them...global

var (
	rem        remote
	ErrNoRepos = errors.New("no repository found")
)

////////////////////////////////////////
// common
////////////////////////////////////////

// Search for and return the repository root
func repositoryRoot() (repos string, err error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// start here, work way up until find .amber in pwd
	repos = pwd
	for {
		pathname := fmt.Sprintf("%s%c.amber", repos, filepath.Separator)
		fi, err := os.Stat(pathname)
		if err == nil {
			if fi.IsDir() {
				repos = pathname
				break // found it
			}
			return "", ErrNoRepos
		}
		if !os.IsNotExist(err) {
			return "", err // some other error
		}
		// have not found it yet
		parent := filepath.Dir(repos)
		if parent == repos {
			return "", ErrNoRepos
		}
		repos = parent
	}
	return repos, nil
}

func urlFromRemoteAndResource(rem *remote, cHash string) (url string) {
	return fmt.Sprintf("http://%s:%d/resource/%s", rem.hostname, rem.port, cHash)
}

func isRuneInvalidForHash(r rune) bool {
	switch {
	case '0' <= r && r <= '9':
		return false
	case 'a' <= r && r <= 'f':
		return false
	}
	return true
}

func isHashInvalid(s string) bool {
	switch {
	case len(s) == 0:
		return true
	case strings.IndexFunc(s, isRuneInvalidForHash) != -1:
		return true
	}
	return false
}

func mustLookupHeader(h http.Header, name string) (string, error) {
	values := h[name]
	if len(values) != 1 {
		return "", fmt.Errorf("cannot resolve header: %v", name)
	}
	return values[0], nil
}

// construct metadata object from URI and Headers
func request2metadata(r *http.Request) (meta metadata, err error) {
	// "/resource/contentHash" => []string{ "", "resource", "contentHash", }
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		err = fmt.Errorf("invalid url: %s", r.URL.Path)
		return
	}

	switch {
	case parts[1] == "resource":
		// no-op
	// case parts[1] == "account":
	// 	// no-op
	default:
		err = fmt.Errorf("invalid url: %s", r.URL.Path)
		return
	}

	meta.cHash = parts[2]
	if isHashInvalid(meta.cHash) {
		err = fmt.Errorf("invalid url: %s", r.URL.Path)
		return
	}
	if meta.uName, err = mustLookupHeader(r.Header, "X-Amber-User"); err != nil {
		meta.uName = "-"
		err = nil
	}
	if meta.uName != "-" && isHashInvalid(meta.uName) {
		err = fmt.Errorf("invalid: %s", meta.uName)
		return
	}
	if meta.eName, err = mustLookupHeader(r.Header, "X-Amber-Encryption"); err != nil {
		meta.eName = "-"
		err = nil
	}
	if meta.hName, err = mustLookupHeader(r.Header, "X-Amber-Hash"); err != nil {
		meta.hName = "-"
		err = nil
	}
	meta.size = fmt.Sprintf("%d", r.ContentLength)

	meta.bpathname = fmt.Sprintf("%s/users/%s", r.URL.Path[1:], meta.uName)
	meta.mpathname = fmt.Sprintf("%s/meta", r.URL.Path[1:])
	return
}

func checkHash(hName string, blob []byte, expectedHash string) (valid bool, err error) {
	actualHash, err := computeHash(hName, blob)
	if err != nil {
		return
	}
	if actualHash != expectedHash {
		err = fmt.Errorf("expected hash: %v, actual: %v", expectedHash, actualHash)
		return
	}
	return true, nil
}

func computeHash(hName string, blob []byte) (string, error) {
	var h hash.Hash
	switch {
	case hName == "sha1":
		h = sha1.New()
	case hName == "sha256":
		h = sha256.New()
	case hName == "sha512":
		h = sha512.New()
	default:
		return "", fmt.Errorf("unknown hash: %s", hName)
	}
	h.Write(blob)
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func writeFile(pathname string, blob []byte) (err error) {
	dirname := filepath.Dir(pathname)
	if err = os.MkdirAll(dirname, 0700); err != nil {
		return
	}
	if err = ioutil.WriteFile(pathname, blob, 0600); err != nil {
		return
	}
	return
}

////////////////////////////////////////
// main
////////////////////////////////////////

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %v [--hostname localhost] [--port 49154] [ server reposDir | download urn pathname pHash | upload pathname ]\n", filepath.Base(os.Args[0]))
}

func main() {
	flag.StringVar(&rem.hostname, "hostname", "localhost", "server hostname")
	flag.IntVar(&rem.port, "port", 49154, "server port")
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
		os.Exit(2)
	}
	cmds := map[string]int{
		"server":   2,
		"download": 4,
		"upload":   2,
	}
	cmd := strings.ToLower(flag.Arg(0))
	count, ok := cmds[cmd]
	if !ok || flag.NArg() != count {
		usage()
		os.Exit(2)
	}

	// TODO: should be loaded from config
	defaults := &metadata{
		hName: DefaultHash,
		eName: DefaultEncryption,
		uName: "-",
	}

	// TODO: will want to hold onto longer when doing more than
	// single upload
	client := &http.Client{}

	switch {
	case cmd == "help":
		usage()
	case cmd == "server":
		server(rem.port, flag.Arg(1))
	case cmd == "download":
		doDownload(rem, flag.Arg(1), flag.Arg(2), flag.Arg(3))
	case cmd == "upload":
		doUpload(flag.Arg(1), defaults, client, &rem)
	default:
		fmt.Println("pardon?")
		usage()
		os.Exit(1)
	}
}
