// Copyright (c) 2022 Vincent Chueng (coolingfall@gmail.com).

package piper

import (
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

func TestPiper(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Piper suite tests")
}
