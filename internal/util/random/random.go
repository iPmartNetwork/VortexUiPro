// Package random provides utilities for generating cryptographically secure random strings and numbers.
package random

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
)

// character sequences
var (
	numSeq      [10]rune
	lowerSeq    [26]rune
	upperSeq    [26]rune
	numLowerSeq [36]rune
	numUpperSeq [36]rune
	allSeq      [62]rune
)

func init() {
	for i := 0; i < 10; i++ {
		numSeq[i] = rune('0' + i)
	}
	for i := 0; i < 26; i++ {
		lowerSeq[i] = rune('a' + i)
		upperSeq[i] = rune('A' + i)
	}
	copy(numLowerSeq[:], numSeq[:])
	copy(numLowerSeq[len(numSeq):], lowerSeq[:])

	copy(numUpperSeq[:], numSeq[:])
	copy(numUpperSeq[len(numSeq):], upperSeq[:])

	copy(allSeq[:], numSeq[:])
	copy(allSeq[len(numSeq):], lowerSeq[:])
	copy(allSeq[len(numSeq)+len(lowerSeq):], upperSeq[:])
}

// Seq generates a random string of length n containing all alphanumeric characters.
func Seq(n int) string {
	runes := make([]rune, n)
	for i := range n {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(allSeq))))
		if err != nil {
			panic("crypto/rand failed: " + err.Error())
		}
		runes[i] = allSeq[idx.Int64()]
	}
	return string(runes)
}

// NumLower generates a random string of length n containing digits and lowercase letters only.
func NumLower(n int) string {
	runes := make([]rune, n)
	for i := range n {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(numLowerSeq))))
		if err != nil {
			panic("crypto/rand failed: " + err.Error())
		}
		runes[i] = numLowerSeq[idx.Int64()]
	}
	return string(runes)
}

// Num generates a random integer between 0 and n-1.
func Num(n int) int {
	bn := big.NewInt(int64(n))
	r, err := rand.Int(rand.Reader, bn)
	if err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return int(r.Int64())
}

// Base64Bytes returns n cryptographically-random bytes encoded as standard base64.
// Used for Shadowsocks 2022 keys.
func Base64Bytes(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return base64.StdEncoding.EncodeToString(b)
}

// Hex generates a random hex string of length n.
func Hex(n int) string {
	const hexChars = "0123456789abcdef"
	runes := make([]rune, n)
	for i := range n {
		idx, err := rand.Int(rand.Reader, big.NewInt(16))
		if err != nil {
			panic("crypto/rand failed: " + err.Error())
		}
		runes[i] = rune(hexChars[idx.Int64()])
	}
	return string(runes)
}
