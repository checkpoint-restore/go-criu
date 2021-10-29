package main_test

import (
	"testing"
)

func TestMain(t *testing.T) {
	// Do nothing. We just need dummy to make `ginkgo` happy. Without that,
	// `ginkgo` would try to execute the _coverage_test.go _despite_ the
	t.Parallel()
}
