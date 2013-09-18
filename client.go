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
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	RC4_TRASH_BYTES = 256
)

////////////////////////////////////////
// types
////////////////////////////////////////

// client download
//
// NOTE: seems like N2R is wasted effort when most resources not
// served by favorite remove
//
// send N2R to favorite remove to resolve
// if code == 200
//   done
// if code == 404
//   find_and_download(urn)

// func find_and_download(urn)
//   send N2Ls to favorite remove to resolve
//   if code == 404
//     resource does not exist
//   if code == 303
//     for _,url := range urls {
//       send GET to url
//       if err == nil {
//         if cHash fails validation {
//           optionally send bad cHash message to server
//           this message signed by client and sig ver by server to prevent DOS
//           SERVER may in remove resource if hash invalid,
//           which must be protected against DOS attack
//         }
//         break
//       }
//     }
//     if content == nill {
//       still could not find it
//     }

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
		// 	// FIXME: key must be 16, 24, or 32 bytes for AES-128, -192, or -256
		// 	c,err = aes.NewCipher([]byte(key))
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	eblob = make([]byte, len(blob)) // padding?
		// 	// TODO
		// 	return eblob, nil
	case strings.HasPrefix(algorithm, "rc4"):
		return encryptRC4(blob, key)
	}
	return nil, fmt.Errorf("unknown encryption algorithm: %s", algorithm)
}

func encryptRC4(blob []byte, key string) ([]byte, error) {
	// key must be between 1 and 256 bytes
	rc4Key := make([]byte, 256) // ???
	copy(rc4Key, []byte(key))   // ???
	c, err := rc4.NewCipher(rc4Key)
	if err != nil {
		return nil, err
	}
	// throw away at least 256 bytes of data
	junk := make([]byte, RC4_TRASH_BYTES)
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
	meta.pHash, err = computeHash(meta.hName, plainBytes)
	if err != nil {
		return
	}
	iv, err := selectIV(meta.eName, meta.hName, plainBytes)
	if err != nil {
		return
	}
	cipherBytes, err := encrypt(plainBytes, meta.eName, meta.pHash, iv)
	if err != nil {
		return
	}
	meta.cHash, err = computeHash(meta.hName, cipherBytes)
	if err != nil {
		return
	}
	url := urlFromRemoteAndResource(rem, meta.cHash)
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
	log.Printf("pHash: %s\n", meta.pHash)
	log.Printf("upload response:\n%v", string(out))
	return
}

func resourceFromUrl(url string) (cHash string) {
	i := strings.LastIndex(url, "/")
	if i == -1 {
		return ""
	}
	return url[i+1:]
}

func doDownload(rem remote, urn, pathname, pHash string) {
	var meta metadata
	var err error
	var resource string

	i := strings.LastIndex(urn, ":")
	switch {
	case i == -1:
		err = fmt.Errorf("cannot find colon: %v", urn)
		return
	default:
		resource = urn[i+1:]
	}

	urls, err := resolveUrls(urn, rem)
	if err != nil {
		log.Fatal(err)
	}

	meta, cipherBytes, err := downloadResourceFromUrls(urls)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := checkHash(meta.hName, cipherBytes, resource); err != nil {
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

func downloadResourceFromUrls(urls []string) (meta metadata, blob []byte, err error) {
	var last_err error
	for _, url := range urls {
		meta, blob, err = downloadResource(url)
		if err == nil {
			return
		}
		log.Println(err)
		last_err = err
	}
	err = last_err
	return
}

func downloadResource(url string) (meta metadata, blob []byte, err error) {
	log.Printf("downloadResource: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	meta.eName, err = mustLookupHeader(resp.Header, "X-Amber-Encryption")
	if err != nil {
		meta.eName = "-"
		err = nil
	}
	meta.hName, err = mustLookupHeader(resp.Header, "X-Amber-Hash")
	if err != nil {
		return
	}
	blob, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
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

	// check for 303 or other status code?

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
	meta.cHash = urn[i+1:]
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
