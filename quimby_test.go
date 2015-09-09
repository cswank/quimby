package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Quimby", func() {
	BeforeEach(func() {
		go main()
	})
	It("just getting started", func() {
		Expect(1).To(Equal(1))
	})
})
