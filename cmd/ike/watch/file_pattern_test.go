package watch_test

import (
	"fmt"

	"github.com/maistra/istio-workspace/cmd/ike/watch"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("File pattern", func() {

	Context("parsing", func() {

		It("should extract regexp", func() {
			// given
			regexpDef := []string{"regex{{my-regexp}}"}

			// when
			parsed := watch.ParseFilePatterns(regexpDef)

			// then
			Expect(parsed).To(ConsistOf(watch.FilePattern{RegExp: "my-regexp"}))
		})
	})

	Context("matching", func() {

		assertThat := func(file, pattern string) {
			parsed := watch.ParseFilePatterns([]string{pattern})
			Expect(parsed.Matches(file)).To(BeTrue())
		}

		assertThatNot := func(file, pattern string) {
			parsed := watch.ParseFilePatterns([]string{pattern})
			Expect(parsed.Matches(file)).To(BeFalse())
		}

		table.DescribeTable(
			"should match paths to simplified regex",
			assertThat,
			file("src/main/resources/Anyfile").matches("**/Anyfile"),
			file("Anyfile").matches("**/**/Anyfile"),
			file("src/Anyfile").matches("*/Anyfile"),
			file("src/test/resources/Anyfile").matches("src/**/Anyfile"),
			file("Anyfile").matches("**/Anyfile"),
			file("Anyfile").matches("Anyfile"),
			file("test_case.py").matches("**/test*.py"),
			file("pkg/test/test_case.py").matches("**/test*.py"),
		)

		table.DescribeTable(
			"should not parse file patterns to regexp",
			assertThatNot,
			file("test/directory/Anyfile").matches("*/Anyfile"),
			file("test/multiple/directory/two/Anyfile").matches("test/multiple/*/Anyfile"),
		)

	})
})

type filePatternProvider func() string

var patternAssertionMsg = "Should match file %s to expression %s"

func file(fileName string) filePatternProvider {
	return filePatternProvider(func() string {
		return fileName
	})
}

func (f filePatternProvider) matches(simplifiedRegExp string) table.TableEntry {
	return table.Entry(fmt.Sprintf(patternAssertionMsg, f(), simplifiedRegExp), f(), simplifiedRegExp)
}
