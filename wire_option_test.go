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
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestWireOption(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "wire option test")
}

var _ = Describe("wire option", func() {
	It("in", func() {
		opt := In()
		Expect(*opt).To(Equal(WireOption{
			index:    0,
			required: true,
			wireType: wireIn,
		}))
	})
	It("out", func() {
		opt := Out()
		Expect(*opt).To(Equal(WireOption{
			index:    0,
			required: true,
			wireType: wireOut,
		}))
	})
	It("name", func() {
		opt := Name("test")
		Expect(*opt).To(Equal(WireOption{
			index:    0,
			required: true,
			wireType: wireIn,
			name:     "test",
		}))
	})
	It("default", func() {
		opt := Default("default")
		Expect(*opt).To(Equal(WireOption{
			index:    0,
			required: false,
			wireType: wireIn,
			defValue: "default",
		}))
	})
	It("out name", func() {
		opt := OutName("test")
		Expect(*opt).To(Equal(WireOption{
			index:    0,
			required: true,
			wireType: wireOut,
			name:     "test",
		}))
	})
	It("in with other", func() {
		opt := In().Name("test").Default("default")
		Expect(*opt).To(Equal(WireOption{
			index:    0,
			required: false,
			wireType: wireIn,
			name:     "test",
			defValue: "default",
		}))
	})
})
