package naming

import (
	"crypto/rand"
	"math/big"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyz")

// RandName generates random alphabetical name which can be used as application or namespace name in Openshift.
//
// Don't forget to seed before using this function, e.g. rand.Seed(time.Now().UTC().UnixNano())
// otherwise you will always get the same value.
func RandName(length int) string {
	if length == 0 {
		return ""
	}

	if length > 58 {
		length = 58
	}

	b := make([]rune, length)
	for i := range b {
		ri, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		b[i] = letters[ri.Int64()]
	}

	return string(b)
}
