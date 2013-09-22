// client
package main

// TODO: timeout on network requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	MAX_DIR_NAMES   = 1000
	REPOSITORY_ROOT = ".amber"
)

////////////////////////////////////////
// repository root
////////////////////////////////////////

func repositoryRoot(rootName string) (repos string, err error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	suffix := fmt.Sprintf("%c%s", filepath.Separator, rootName)
	// start here, work way up until find .amber in pwd
	repos = pwd
	for {
		pathname := fmt.Sprintf("%s%s", repos, suffix)
		fi, err := os.Stat(pathname)
		if err == nil {
			if fi.IsDir() {
				// found it
				repos = pathname
				break
			}
			// something other than a directory
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

////////////////////////////////////////
// createCommit
//
// directory loaded into cache (pcache, then encrypted into ecache)
// new tip created
////////////////////////////////////////

// TODO: revise this struct; do I need all metadata?
type commit struct {
	name   string // some urn?
	meta   metadata
	parent *commit
	// when merging, a commit has two parents, primary is parent,
	// while secondary is merge
	merge *commit
}

func createCommit(pathname string) (c commit, err error) {
	root, err := repositoryRoot(REPOSITORY_ROOT)
	if err != nil {
		return
	}
	meta := new(metadata)
	// TODO: should be loaded from config
	meta.hName = DefaultHash
	meta.eName = DefaultEncryption
	meta.uName = "-"
	err = commitPathname(root, pathname, meta)
	if err != nil {
		return
	}
	c = commit{name: pathname, meta: *meta}
	return
}

func commitPathname(repositoryRoot, pathname string, meta *metadata) (err error) {
	fi, err := os.Stat(pathname)
	if err != nil {
		return
	}
	mode := fi.Mode()
	meta.Mode = fmt.Sprintf("%o", mode)
	meta.Name = fi.Name()
	switch {
	case mode&os.ModeDir != 0:
		err = commitDirectory(repositoryRoot, pathname, meta)
	case mode&os.ModeSymlink != 0:
		err = fmt.Errorf("TODO: implement commitSymlink")
	default:
		err = commitFile(repositoryRoot, pathname, meta)
	}
	return
}

func commitDirectory(repositoryRoot, pathname string, meta *metadata) (err error) {
	if debug {
		log.Println("COMMIT DIRECTORY:", pathname)
	}
	meta.Type = "directory"
	fh, err := os.Open(pathname)
	if err != nil {
		return
	}
	defer fh.Close()

	meta.Children = make([]metadata, 0, MAX_DIR_NAMES)
	for {
		var names []string
		names, err = fh.Readdirnames(MAX_DIR_NAMES)
		if err != nil {
			if err == io.EOF {
				// err = nil // ignored by overwrite
				break // done
			}
			return
		}
		for _, name := range names {
			switch {
			case name == "." || name == "..":
			case name == ".amber" || name == ".git":
			default:
				childMeta := new(metadata)
				childMeta.hName = meta.hName
				childMeta.eName = meta.eName
				childMeta.uName = meta.uName
				childName := fmt.Sprintf("%s/%s", pathname, name)
				if err = commitPathname(repositoryRoot, childName, childMeta); err != nil {
					return
				}
				meta.Children = append(meta.Children, *childMeta) // ??? how efficient with large directories ???
			}
		}
	}
	blob, err := json.Marshal(meta.Children)
	if err != nil {
		return
	}
	if err = commitBytes(repositoryRoot, blob, meta); err != nil {
		return
	}
	// once directory committed, do not want to propagate Children up
	meta.Children = nil
	return
}

func commitFile(repositoryRoot, pathname string, meta *metadata) (err error) {
	if debug {
		log.Println("COMMIT FILE:", pathname)
	}
	meta.Type = "file"
	plainBytes, err := ioutil.ReadFile(pathname)
	if err != nil {
		return
	}
	return commitBytes(repositoryRoot, plainBytes, meta)
}

func commitBytes(repositoryRoot string, blob []byte, meta *metadata) (err error) {
	meta.size = fmt.Sprintf("%d", len(blob))
	meta.Phash, err = computeHash(meta.hName, blob)
	if err != nil {
		return
	}
	fname := fmt.Sprintf("%s/pcache/resource/%s", repositoryRoot, meta.Phash)
	if _, err = os.Stat(fname); os.IsNotExist(err) {
		if err = writeFile(fname, blob); err != nil {
			return
		}
	}
	iv, err := selectIV(meta.eName, meta.hName, blob)
	if err != nil {
		return
	}
	cipherBytes, err := encrypt(blob, meta.eName, meta.Phash, iv)
	if err != nil {
		return
	}

	meta.Chash, err = computeHash(meta.hName, cipherBytes)
	if err != nil {
		return
	}
	fname = fmt.Sprintf("%s/ecache/resource/%s", repositoryRoot, meta.Chash)
	if _, err = os.Stat(fname); os.IsNotExist(err) {
		if err = writeFile(fname, cipherBytes); err != nil {
			return
		}
	}
	return
}

////////////////////////////////////////
// push
//
// all resources not on remote is copied to remote
////////////////////////////////////////

////////////////////////////////////////
// pull
//
// all remote resources not on localhost is copied to localhost
////////////////////////////////////////

////////////////////////////////////////
// update
//
// tip of cached data copied to directory (brute overwrite of directory data)
////////////////////////////////////////

////////////////////////////////////////
// upload / download
////////////////////////////////////////

func doUpload(pathname string, meta *metadata, client *http.Client, rem *remote) {
	err := upload(pathname, meta, client, rem)
	if err != nil {
		log.Fatal(err)
	}
}

func upload(pathname string, meta *metadata, client *http.Client, rem *remote) (err error) {
	fi, err := os.Stat(pathname)
	if err != nil {
		return
	}
	plainBytes, err := ioutil.ReadFile(pathname)
	if err != nil {
		return
	}
	meta.Phash, err = computeHash(meta.hName, plainBytes)
	if err != nil {
		return
	}
	iv, err := selectIV(meta.eName, meta.hName, plainBytes)
	if err != nil {
		return
	}
	cipherBytes, err := encrypt(plainBytes, meta.eName, meta.Phash, iv)
	if err != nil {
		return
	}
	meta.Chash, err = computeHash(meta.hName, cipherBytes)
	if err != nil {
		return
	}
	url := urlFromRemoteAndResource(rem, meta.Chash)
	if debug {
		log.Print("PUT: " + url)
	}
	reader := bytes.NewReader(cipherBytes)
	req, err := http.NewRequest("PUT", url, reader)
	if err != nil {
		return
	}
	req.Header = http.Header{
		"X-Amber-Hash":       {meta.hName},
		"X-Amber-Encryption": {meta.eName},
	}
	req.ContentLength = fi.Size()
	// PUT
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	out, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return
	}
	if debug {
		log.Printf("pHash: %s\n", meta.Phash)
		log.Printf("upload response:\n%v", string(out))
	}
	return
}

func doDownload(rem remote, urn, pathname, pHash string) (err error) {
	i := strings.LastIndex(urn, ":")
	if i == -1 {
		err = fmt.Errorf("cannot find colon: %v", urn)
		return
	}
	resource := urn[i+1:]

	urls, err := resolveUrls(urn, rem)
	if err != nil {
		return
	}

	meta, cipherBytes, err := downloadResourceFromUrls(urls, resource)
	if err != nil {
		return
	}

	iv, err := selectIV(meta.eName, meta.hName, cipherBytes)
	if err != nil {
		return
	}
	plainBytes, err := decrypt(cipherBytes, meta.eName, pHash, iv)
	if err != nil {
		return
	}
	if _, err = checkHash(meta.hName, plainBytes, pHash); err != nil {
		return
	}

	return writeFile(pathname, plainBytes)
}

func downloadResourceFromUrls(urls []string, Chash string) (meta metadata, blob []byte, err error) {
	var last_err error
	for _, url := range urls {
		meta, blob, err = downloadResource(url, Chash)
		if err == nil {
			return
		}
		log.Println(err)
		last_err = err
	}
	err = last_err
	return
}

func downloadResource(url, Chash string) (meta metadata, blob []byte, err error) {
	if debug {
		log.Printf("downloadResource: %s", url)
	}
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	meta.Chash = Chash
	meta.hName, err = mustLookupHeader(resp.Header, "X-Amber-Hash")
	if err != nil {
		return
	}
	meta.eName, err = mustLookupHeader(resp.Header, "X-Amber-Encryption")
	if err != nil {
		meta.eName = "-"
		err = nil
	}
	blob, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if _, err = checkHash(meta.hName, blob, meta.Chash); err != nil {
		if err := sendBadHashNotice(url, Chash); err != nil {
			log.Printf(err.Error())
		}
		return
	}
	return
}

func sendBadHashNotice(url, Chash string) (err error) {
	// optionally send bad hash message to server. this message
	// signed by client and sig verified by server to prevent DOS.
	// SERVER may remove resource if hash invalid, which must be
	// protected against DOS attack.
	return
}

func resolveUrls(urn string, rem remote) (urls []string, err error) {
	query := fmt.Sprintf("http://%s:%d/N2Ls?%s", rem.hostname, rem.port, urn)
	log.Printf("resolveUrls: %s", query)
	resp, err := http.Get(query)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("%s", resp.Status)
		return
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return parseUriList(string(bytes)), nil
}

func parseUriList(blob string) (urls []string) {
	lines := strings.Split(blob, crlf)
	for _, line := range lines {
		if len(line) > 0 {
			i := strings.IndexRune(line, '#')
			switch {
			case i == 0: // comment
				continue
			case i == -1: // not found; append entire line
				urls = append(urls, line)
			default: // skip comment portion of line
				urls = append(urls, line[:i])
			}
		}
	}
	return
}

func resolveMeta(urn string, rem remote) (meta metadata, err error) {
	query := fmt.Sprintf("http://%s:%d/N2C?%s", rem.hostname, rem.port, urn)
	log.Printf("resolveMeta: %s", query)
	resp, err := http.Get(query)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("%s", resp.Status)
		return
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	meta, err = parseUrc(bytes)
	if err != nil {
		return
	}
	i := strings.LastIndex(urn, ":")
	if i == -1 {
		err = fmt.Errorf("cannot find colon: %s", urn)
		return
	}
	meta.Chash = urn[i+1:]
	return
}

func parseUrc(blob []byte) (meta metadata, err error) {
	lines := strings.Split(string(blob), crlf)
	for _, line := range lines {
		if line == "" {
			break
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			err = fmt.Errorf("invalid line format: %s", line)
			return
		}
		switch {
		case fields[0] == "X-Amber-Hash:":
			meta.hName = fields[1]
		case fields[0] == "X-Amber-Encryption:":
			meta.eName = fields[1]
		case fields[0] == "Content-Length:":
			meta.size = fields[1]
		}
	}
	return
}
