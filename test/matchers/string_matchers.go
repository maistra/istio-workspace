package matchers

import (
	"fmt"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

// OnlyContain ensures that given string contains only characters specified as char.
func OnlyContain(chars string) types.GomegaMatcher {
	allowedChars := make(map[rune]struct{}, len(chars))
	for _, char := range chars {
		allowedChars[char] = struct{}{}
	}
	return &onlyContainsMatcher{chars: allowedChars, asString: chars}
}

type onlyContainsMatcher struct {
	chars    map[rune]struct{}
	asString string
}

func (matcher *onlyContainsMatcher) Match(actual interface{}) (success bool, err error) {
	if actual == nil {
		return true, nil
	}

	s, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("expected string. Got:\n%s", format.Object(actual, 1))
	}

	for _, c := range s {
		if _, contains := matcher.chars[c]; !contains {
			return false, nil
		}
	}

	return true, nil
}

func (matcher *onlyContainsMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to contain any of", matcher.asString)
}

func (matcher *onlyContainsMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to not contain any of", matcher.asString)
}
