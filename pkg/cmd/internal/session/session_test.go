package session_test

import (
	"github.com/maistra/istio-workspace/pkg/cmd/internal/session"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Session operations", func() {

	Context("route parsing", func() {

		It("should return nil with no error on empty string", func() {
			r, err := session.ParseRoute("")
			Expect(err).ToNot(HaveOccurred())
			Expect(r).To(BeNil())
		})

		It("should error on wrong type format", func() {
			_, err := session.ParseRoute("a=b")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("route in wrong format"))
		})

		It("should error on wrong value format", func() {
			_, err := session.ParseRoute("header:a-b")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("route in wrong format"))
		})

		It("should return a valid route", func() {
			r, err := session.ParseRoute("header:a=b")
			Expect(err).ToNot(HaveOccurred())
			Expect(r).ToNot(BeNil())

			Expect(r.Type).To(Equal("header"))
			Expect(r.Name).To(Equal("a"))
			Expect(r.Value).To(Equal("b"))
		})
	})
})
