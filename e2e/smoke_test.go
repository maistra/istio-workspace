package e2e_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

)

var _ = Describe("Smoke End To End Tests", func() {

	Context("ike command against OpenShift Cluster", func() {

		It("should work", func() {
			test := ""
			Expect(test).To(BeEmpty())
		})

	})

})