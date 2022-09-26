// Copyright (c) 2022 Vincent Cheung (coolingfall@gmail.com).
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
	"embed"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

//go:embed rs_fs_test.yml
var rs embed.FS

var _ = ginkgo.Describe("rs fs", func() {
	ginkgo.It("read", func() {
		rsfs := newResourceFs(rs)
		f, err := rsfs.Open("rs_fs_test.yml")
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(f.Name()).To(gomega.Equal("rs_fs_test.yml"))
	})
})
