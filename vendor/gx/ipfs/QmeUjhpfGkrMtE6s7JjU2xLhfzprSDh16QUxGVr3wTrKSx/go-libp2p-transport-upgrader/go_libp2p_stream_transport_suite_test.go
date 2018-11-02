package stream_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGoLibp2pStreamTransport(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GoLibp2pStreamTransport Suite")
}
