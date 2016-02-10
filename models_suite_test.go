package quimby_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestQuimby(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Quimby Suite")
}
