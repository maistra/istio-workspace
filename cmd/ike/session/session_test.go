package session

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Session operations", func() {

	Context("route parsing", func() {

		It("should return nil with no error on empty string", func() {
			r, err := parseRoute("")
			Expect(err).ToNot(HaveOccurred())
			Expect(r).To(BeNil())
		})

		It("should error on wrong type format", func() {
			_, err := parseRoute("a=b")
			Expect(err).To(HaveOccurred())
		})

		It("should error on wrong value format", func() {
			_, err := parseRoute("header:a-b")
			Expect(err).To(HaveOccurred())
		})

		It("should return a valid route", func() {
			r, err := parseRoute("header:a=b")
			Expect(err).ToNot(HaveOccurred())
			Expect(r).ToNot(BeNil())

			Expect(r.Type).To(Equal("header"))
			Expect(r.Name).To(Equal("a"))
			Expect(r.Value).To(Equal("b"))
		})
	})
})
