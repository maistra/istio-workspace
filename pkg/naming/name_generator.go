package naming

import "math/rand"

var letters = []rune("abcdefghijklmnopqrstuvwxyz")
var alphaNumeric = []rune("abcdefghijklmnopqrstuvwxyz0987654321-")

// RandName generates random alphanumeric name which can be used as application or namespace name in Openshift.
// Generated name follows spec:
//    Must be an a lower case alphanumeric (a-z, and 0-9) string with a maximum length of 58 characters,
//    where the first character is a letter (a-z), and the '-'
//    character is allowed anywhere except the first or last character.
// Don't forget to seed before using this function, e.g. rand.Seed(time.Now().UTC().UnixNano())
// otherwise you will always get the same value
func RandName(length int) string {
	if length == 0 {
		return ""
	}

	if length > 58 {
		length = 58
	}

	b := make([]rune, length)
	for i := range b {
		b[i] = alphaNumeric[rand.Intn(len(alphaNumeric))]
	}
	b[0] = letters[rand.Intn(len(letters))]
	b[length-1] = letters[rand.Intn(len(letters))]
	return string(b)
}
