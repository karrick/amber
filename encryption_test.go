package main

import (
	"fmt"
	"testing"
)

func TestEncryptReturnsErrorWhenUnknownEncryption(t *testing.T) {
	iv := make([]byte, 10)
	actual, err := encrypt(make([]byte, 10), "non-existant-algorithm", "some key", iv)
	if actual != nil {
		t.Errorf("expected: %v, actual: %v", nil, actual)
	}
	if err == nil {
		t.Errorf("expected: %v, actual: %v", fmt.Errorf("unknown encryption algorithm: non-existant-algorithm"), err)
	}
}

func TestEncryptReturnsEncryptionStringOfBytes(t *testing.T) {
	bytes := []byte("just some blob of data")
	iv := make([]byte, 10)
	actual, err := encrypt(bytes, "-", "some key", iv)
	expected := "just some blob of data"
	if string(actual) != expected {
		t.Errorf("expected: %v, actual: %v", expected, string(actual))
	}
	if err != nil {
		t.Errorf("expected: %v, actual: %v", nil, err)
	}
}

////////////////////////////////////////

func TestDecryptReturnsErrorWhenUnknownDecryption(t *testing.T) {
	iv := make([]byte, 10)
	actual, err := decrypt(make([]byte, 10), "non-existant-decryption", "some key", iv)
	if actual != nil {
		t.Errorf("expected: %v, actual: %v", nil, actual)
	}
	if err == nil {
		t.Errorf("expected: %v, actual: %v", fmt.Errorf("unknown decryption algorithm: non-existant-algorithm"), err)
	}
}

func TestDecryptReturnsDecryptionStringOfBytes(t *testing.T) {
	bytes := []byte("just some blob of data")
	iv := make([]byte, 10)
	actual, err := decrypt(bytes, "-", "some key", iv)
	expected := "just some blob of data"
	if string(actual) != expected {
		t.Errorf("expected: %v, actual: %v", expected, string(actual))
	}
	if err != nil {
		t.Errorf("expected: %v, actual: %v", nil, err)
	}
}

////////////////////////////////////////

func TestSelectIVInvalidHashName(t *testing.T) {
	plaintext := []byte("this is a test")
	_, actual := selectIV("aes128", "sha", plaintext)
	expected := fmt.Sprintf("unknown hash: %s", "sha")
	if actual.Error() != expected {
		t.Errorf("expected: %v, actual: %v", expected, actual.Error())
	}
}

func TestSelectIVInvalidEncryptionAlgorithm(t *testing.T) {
	plaintext := []byte("this is a test")

	cases := []map[string]string{
		{
			"eName": "-",
			"msg":   "unknown encryption algorithm: -",
		},
		{
			"eName": "aes",
			"msg":   "unknown encryption algorithm: aes",
		},
	}
	for _, item := range cases {
		_, actual := selectIV(item["eName"], "sha1", plaintext)
		expected := fmt.Sprintf("unknown encryption algorithm: %s", item["eName"])
		if actual.Error() != expected {
			t.Errorf("expected: %v, actual: %v", expected, actual.Error())
		}
	}
}

func TestSelectIV(t *testing.T) {
	plaintext := []byte("this is our blob of plaintext, and its size will be hashed to come up with an iv.")

	cases := []map[string]string{
		{
			"eName":          "aes128", // determines output bit length
			"hName":          "sha1",   // which hash to use
			"ivSizeExpected": "16",
			"ivExpected":     "31643531336330626362653333623265",
		},
		{
			"eName":          "aes192",
			"hName":          "sha1",
			"ivSizeExpected": "24",
			"ivExpected":     "316435313363306263626533336232653734343065356531",
		},
		{
			"eName":          "aes256",
			"hName":          "sha1",
			"ivSizeExpected": "32",
			"ivExpected":     "3164353133633062636265333362326537343430653565313464306232326566",
		},
		{
			"eName":          "aes128",
			"hName":          "sha256",
			"ivSizeExpected": "16",
			"ivExpected":     "35333136636131633564646361386536",
		},
		{
			"eName":          "aes128",
			"hName":          "sha512",
			"ivSizeExpected": "16",
			"ivExpected":     "61346133636436616432376230613539",
		},
	}

	for _, item := range cases {
		iv, err := selectIV(item["eName"], item["hName"], plaintext)
		if err != nil {
			t.Errorf("expected: %v, actual: %v", nil, err)
		}

		if ivSizeActual := fmt.Sprintf("%d", len(iv)); ivSizeActual != item["ivSizeExpected"] {
			t.Errorf("expected: %v, actual: %v", item["ivSizeExpected"], ivSizeActual)
		}

		if ivActual := fmt.Sprintf("%x", iv); ivActual != item["ivExpected"] {
			t.Errorf("expected: %v, actual: %v", item["ivExpected"], ivActual)
		}
	}
}

