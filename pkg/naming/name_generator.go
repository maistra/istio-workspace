package naming

import (
	"crypto/rand"
	"math"
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

// ConcatToMax will cut each section to length based on number of sections to not go beyond max and separate the sections with -.
func ConcatToMax(max int, sections ...string) string {
	sectionLength := (max - len(sections) - 1) / len(sections)
	name := ""

	for i, section := range sections {
		s := section[:int32(math.Min(float64(len(section)), float64(sectionLength)))]
		name = name + "-" + s
		if i+1 != len(sections) {
			sectionLength = (max - len(name) - 1) / (len(sections) - (i + 1))
		}
	}

	return name[1:]
}
