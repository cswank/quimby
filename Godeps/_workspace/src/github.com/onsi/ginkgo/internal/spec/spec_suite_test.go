package spec_test

import (
	. "github.com/cswank/quimby/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cswank/quimby/Godeps/_workspace/src/github.com/onsi/gomega"

	"testing"
)

func TestSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Spec Suite")
}
