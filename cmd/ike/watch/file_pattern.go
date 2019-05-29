package watch

import (
	"os"
	"regexp"
	"strings"
)

const (
	regexpDefinitionPrefix         = "regex{{"
	regexpDefinitionSuffix         = "}}"
	anyPathWildcard                = "**"
	anyNameWildcard                = "*"
	anyNameRegExpUntilDirSeparator = "^[^/]+"
	anythingRegExp                 = ".*"
	twoStarsReplacement            = "<two-stars-replacement>"
	endOfLineRegExp                = "$"
	directorySeparator             = string(os.PathSeparator)
)

// FilePatterns is an alias type representing slice of FilePattern
type FilePatterns []FilePattern

// FilePattern contains regular expression that matches a file
type FilePattern struct {
	RegExp string
}

// Matches checks if the given string (representing path to a file) contains a substring that matches regular expression
// defined by this matcher
func (matcher *FilePattern) Matches(filename string) bool {
	exp, err := regexp.Compile(matcher.RegExp)
	if err != nil {
		return false
	}
	return exp.MatchString(filename)
}

// Matches iterates over all patterns and returns first successful match or false if none patterns matched
func (f *FilePatterns) Matches(filename string) bool {
	for _, matcher := range *f {
		if matcher.Matches(filename) {
			return true
		}
	}
	return false
}

// ParseFilePatterns takes the given patterns and parses to an array of FilePattern instances
func ParseFilePatterns(filePatterns []string) FilePatterns {
	patterns := make([]FilePattern, 0, len(filePatterns))
	for _, pattern := range filePatterns {
		patterns = append(patterns, FilePattern{
			RegExp: parseFilePattern(strings.TrimSpace(pattern)),
		})
	}
	return patterns
}

func parseFilePattern(pattern string) string {

	// if it is regex{{...}} then just return the content
	if strings.HasPrefix(pattern, regexpDefinitionPrefix) && strings.HasSuffix(pattern, regexpDefinitionSuffix) {
		return pattern[len(regexpDefinitionPrefix) : len(pattern)-len(regexpDefinitionSuffix)]
	}

	// if not, then transform the pattern to regexp
	slashIndex := strings.LastIndexAny(pattern, directorySeparator)

	path := transformPathPatternToRegExp(pattern[:slashIndex+1])
	fileName := transformFilenamePatternToRegExp(pattern[slashIndex+1:])

	expr := path + fileName

	if strings.HasSuffix(expr, directorySeparator) {
		expr += anythingRegExp
	} else {
		expr += endOfLineRegExp
	}
	return expr
}

func transformPathPatternToRegExp(path string) string {
	for strings.HasPrefix(path, anyPathWildcard+"/") {
		path = path[strings.Index(path, "/")+1:]
	}
	path = escapeDots(path)
	path = strings.Replace(path, anyPathWildcard, twoStarsReplacement, -1)
	path = strings.Replace(path, anyNameWildcard, anyNameRegExpUntilDirSeparator, -1)
	return strings.Replace(path, twoStarsReplacement, anythingRegExp, -1)
}

func transformFilenamePatternToRegExp(fileName string) string {
	fileName = escapeDots(fileName)

	if strings.HasPrefix(fileName, anyNameWildcard) {
		newPrefix := anythingRegExp
		return newPrefix + replaceAnyNameWildcards(fileName[1:])
	}
	return replaceAnyNameWildcards(fileName)

}

func escapeDots(s string) string {
	return strings.Replace(s, ".", "\\.", -1)
}

func replaceAnyNameWildcards(s string) string {
	return strings.Replace(s, anyNameWildcard, anythingRegExp, -1)
}
