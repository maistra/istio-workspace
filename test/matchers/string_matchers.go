package matchers

import (
	"fmt"
	"strings"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

// BeLetter succeeds if actual is a letter
func BeLetter() types.GomegaMatcher {
	return &beLetterMatcher{OnlyContain("abcdefghijklmnopqrstuvwxyz").(*onlyContainsMatcher)}
}

type beLetterMatcher struct {
	*onlyContainsMatcher
}

func (matcher *beLetterMatcher) Match(actual interface{}) (success bool, err error) {
	if actual == nil {
		return true, nil
	}

	char, ok := actual.(uint8)

	if !ok {
		return false, fmt.Errorf("expected a character (uint8). Got:\n%s", format.Object(actual, 1))
	}

	return matcher.onlyContainsMatcher.Match(strings.ToLower(string(char)))
}

func (matcher *beLetterMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to be a letter")
}

func (matcher *beLetterMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to be a letter")
}

// StartWithLetter succeeds if actual starts with letter
func StartWithLetter() types.GomegaMatcher {
	return &startsWithLetterMatcher{BeLetter().(*beLetterMatcher)}
}

type startsWithLetterMatcher struct {
	*beLetterMatcher
}

func (matcher *startsWithLetterMatcher) Match(actual interface{}) (success bool, err error) {
	if actual == nil {
		return true, nil
	}

	s, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("expected string. Got:\n%s", format.Object(actual, 1))
	}
	return matcher.beLetterMatcher.Match(s[0])
}

// EndWithLetter succeeds if actual starts with letter
func EndWithLetter() types.GomegaMatcher {
	return &endsWithLetterMatcher{BeLetter().(*beLetterMatcher)}
}

type endsWithLetterMatcher struct {
	*beLetterMatcher
}

func (matcher *endsWithLetterMatcher) Match(actual interface{}) (success bool, err error) {
	if actual == nil {
		return true, nil
	}

	s, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("expected string. Got:\n%s", format.Object(actual, 1))
	}

	return matcher.beLetterMatcher.Match(s[len(s)-1])
}

// OnlyContain ensures that
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
