// encryption
package main

import (
	"crypto/rc4"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

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

const (
	RC4_TRASH_BYTES = 256
)

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
	case strings.HasPrefix(algorithm, "rc4"):
		return encryptRC4(blob, key)
	}
	return nil, fmt.Errorf("unknown encryption algorithm: %s", algorithm)
}

func encryptRC4(blob []byte, key string) ([]byte, error) {
	rc4Key := make([]byte, 256)
	copy(rc4Key, []byte(key))
	c, err := newPrimedRC4Cipher(rc4Key)
	if err != nil {
		return nil, err
	}
	c.XORKeyStream(blob, blob)
	c.Reset()
	return blob, nil
}

func newPrimedRC4Cipher(key []byte) (c *rc4.Cipher, err error) {
	if len(key) == 0 || len(key) > 256 {
		err = fmt.Errorf("key length must be between 1 and 256 bytes")
		return
	}
	c, err = rc4.NewCipher(key)
	if err != nil {
		return
	}
	// primeRC4 encrypts random blob of data to throw away
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	junk := make([]byte, RC4_TRASH_BYTES)
	for i := 0; i < len(junk); i++ {
		junk[i] = byte(r.Intn(256))
	}
	c.XORKeyStream(junk, junk)
	return
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
