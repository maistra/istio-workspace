package naming

import "math/rand"

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
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
