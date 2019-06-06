package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Operations for test scenario generator", func() {

	Context("basic sub generators", func() {

		Context("deploymentconfig", func() {
			It("should be created if entry is correct DeploymentType", func() {
				obj := DeploymentConfig(Entry{"test", "DeploymentConfig"})
				Expect(obj).ToNot(BeNil())
			})

			It("should not be created if entry is not correct DeploymentType", func() {
				obj := DeploymentConfig(Entry{"test", "X"})
				Expect(obj).To(BeNil())
			})
		})
		Context("deployment", func() {
			It("should be created if entry is correct DeploymentType", func() {
				obj := Deployment(Entry{"test", "Deployment"})
				Expect(obj).ToNot(BeNil())
			})

			It("should not be created if entry is not correct DeploymentType", func() {
				obj := Deployment(Entry{"test", "X"})
				Expect(obj).To(BeNil())
			})
		})
	})

})
