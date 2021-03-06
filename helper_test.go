// Copyright (c) 2019 Anbillon Team (anbillonteam@gmail.com).
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package piper

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHelper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "helper test")
}

var _ = Describe("helper", func() {
	_ = os.Setenv("ENV_C", "gotC")

	It("expand env default", func() {
		s := ExpandEnv("${ENV_A:-${ENV_B:-notexist}}")
		Expect(s).To(Equal("notexist"))
	})
	It("expand env", func() {
		s := ExpandEnv("${ENV_A:-${ENV_C:-notexist}}")
		Expect(s).To(Equal("gotC"))
	})
})
