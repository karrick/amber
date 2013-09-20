// client
package main

// TODO: for encryption, make sure use Message Authentication Code
// Alternatively, you can apply your own message authentication, as
// follows. First, encrypt the message using an appropriate
// symmetric-key encryption scheme (e.g., AES-CBC). Then, take the
// entire ciphertext (including any IVs, nonces, or other values
// needed for decryption), apply a message authentication code (e.g.,
// AES-CMAC, SHA1-HMAC, SHA256-HMAC), and append the resulting MAC
// digest to the ciphertext before transmission. On the receiving
// side, check that the MAC digest is valid before decrypting. This is
// known as the encrypt-then-authenticate construction. (See also: 1,
// 2.) This also works fine, but requires a little more care from you.

import (
	"bytes"
	"crypto/rc4"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	MAX_DIR_NAMES   = 1000
	RC4_TRASH_BYTES = 256
)

////////////////////////////////////////
// repository root
////////////////////////////////////////

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

////////////////////////////////////////
// encryption / decryption
////////////////////////////////////////

func selectIV(algorithm, hName string, blob []byte) (iv []byte, err error) {
	var ivSize int
	var hash string

	switch {
	case strings.HasSuffix(algorithm, "128"):
		ivSize = 16
	case strings.HasSuffix(algorithm, "192"):
		ivSize = 24
	case strings.HasSuffix(algorithm, "256"):
		ivSize = 32
	case algorithm == "rc4":
		return
	default:
		return nil, fmt.Errorf("unknown encryption algorithm: " + algorithm)
	}

	size := fmt.Sprintf("%d", len(blob))
	hash, err = computeHash(hName, []byte(size))
	if err != nil {
		return
	}
	iv = make([]byte, ivSize)
	copy(iv, []byte(hash))
	return
}

func encrypt(blob []byte, algorithm, key string, iv []byte) ([]byte, error) {
	switch {
	case algorithm == "-":
		return blob, nil
		// case strings.HasPrefix(algorithm, "aes"):
		//	// FIXME: key must be 16, 24, or 32 bytes for AES-128, -192, or -256
		//	c,err = aes.NewCipher([]byte(key))
		//	if err != nil {
		//		return nil, err
		//	}
		//	eblob = make([]byte, len(blob)) // padding?
		//	// TODO
		//	return eblob, nil
	case strings.HasPrefix(algorithm, "rc4"):
		return encryptRC4(blob, key)
	}
	return nil, fmt.Errorf("unknown encryption algorithm: %s", algorithm)
}

func encryptRC4(blob []byte, key string) ([]byte, error) {
	// TODO: assert key is between 1 and 256 bytes
	rc4Key := make([]byte, 256)
	copy(rc4Key, []byte(key))
	c, err := rc4.NewCipher(rc4Key)
	if err != nil {
		return nil, err
	}
	// throw away at least 256 bytes of data
	junk := make([]byte, RC4_TRASH_BYTES)
	// ??? how does decryption work with below reseeding
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < len(junk); i++ {
		junk[i] = byte(r.Intn(256))
	}
	c.XORKeyStream(junk, junk)
	// actual encryption
	c.XORKeyStream(blob, blob)
	c.Reset()
	return blob, nil
}

func decrypt(blob []byte, algorithm, key string, iv []byte) ([]byte, error) {
	switch {
	case algorithm == "-":
		return blob, nil
	case strings.HasPrefix(algorithm, "rc4"):
		return encryptRC4(blob, key)
	}
	return nil, fmt.Errorf("unknown encryption algorithm: " + algorithm)
}

////////////////////////////////////////
// commit
//
// directory loaded into cache (pcache, then encrypted into ecache)
// new tip created
////////////////////////////////////////

func commit(pathname string) (err error) {
	root, err := repositoryRoot()
	if err != nil {
		return
	}
	meta := new(metadata)
	// TODO: should be loaded from config
	meta.hName = DefaultHash
	meta.eName = DefaultEncryption
	meta.uName = "-"
	return commitPathname(root, pathname, meta)
}

func commitPathname(repositoryRoot, pathname string, meta *metadata) (err error) {
	// log.Println("COMMIT PATHNAME:", pathname)
	fi, err := os.Stat(pathname)
	if err != nil {
		return
	}
	mode := fi.Mode()
	meta.Mode = fmt.Sprintf("0%o", mode)
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
	log.Println("COMMIT DIRECTORY:", pathname)
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
	log.Println("COMMIT FILE:", pathname)
	meta.Type = "file"
	plainBytes, err := ioutil.ReadFile(pathname)
	if err != nil {
		return
	}
	if err = commitBytes(repositoryRoot, plainBytes, meta); err != nil {
		return
	}
	return
}

func commitBytes(repositoryRoot string, bytes []byte, meta *metadata) (err error) {
	meta.Phash, err = computeHash(meta.hName, bytes)
	if err != nil {
		return
	}
	fname := fmt.Sprintf("%s/pcache/resource/%s", repositoryRoot, meta.Phash)
	if _, err = os.Stat(fname); os.IsNotExist(err) {
		if err = writeFile(fname, bytes); err != nil {
			return
		}
	}
	iv, err := selectIV(meta.eName, meta.hName, bytes)
	if err != nil {
		return
	}
	cipherBytes, err := encrypt(bytes, meta.eName, meta.Phash, iv)
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
	meta.size = fmt.Sprintf("%d", len(bytes))
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
	log.Print("PUT: " + url)
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
	log.Printf("pHash: %s\n", meta.Phash)
	log.Printf("upload response:\n%v", string(out))
	return
}

func resourceFromUrl(url string) (Chash string) {
	i := strings.LastIndex(url, "/")
	if i == -1 {
		return ""
	}
	return url[i+1:]
}

// TODO: return proper error or refactor so failure to download a
// resource doesn't kill program
func doDownload(rem remote, urn, pathname, pHash string) {
	var meta metadata
	var err error

	i := strings.LastIndex(urn, ":")
	if i == -1 {
		err = fmt.Errorf("cannot find colon: %v", urn)
		return
	}
	resource := urn[i+1:]

	urls, err := resolveUrls(urn, rem)
	if err != nil {
		log.Fatal(err)
	}

	meta, cipherBytes, err := downloadResourceFromUrls(urls, resource)
	if err != nil {
		log.Fatal(err)
	}

	iv, err := selectIV(meta.eName, meta.hName, cipherBytes)
	if err != nil {
		return
	}
	plainBytes, err := decrypt(cipherBytes, meta.eName, pHash, iv)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := checkHash(meta.hName, plainBytes, pHash); err != nil {
		log.Fatal(err)
	}

	if err = writeFile(pathname, plainBytes); err != nil {
		log.Fatal(err)
	}
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
	log.Printf("downloadResource: %s", url)
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

	// check for 303 or other status code?

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
		// trim colon from right side of key string
		key := fields[0][:len(fields[0])-1]
		switch {
		case key == "X-Amber-Hash":
			meta.hName = fields[1]
		case key == "X-Amber-Encryption":
			meta.eName = fields[1]
		case key == "Content-Length":
			meta.size = fields[1]
		}
	}
	return
}
