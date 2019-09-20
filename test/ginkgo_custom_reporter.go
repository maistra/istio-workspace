package test

import (
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
)

// RunSpecWithJUnitReporter calls custom ginkgo junit reporter
func RunSpecWithJUnitReporter(t *testing.T, description string) { //nolint[:unused]
	junitReporter := reporters.NewJUnitReporter("ginkgo-test-results.xml")
	ginkgo.RunSpecsWithDefaultAndCustomReporters(t, description, []ginkgo.Reporter{junitReporter})
}
