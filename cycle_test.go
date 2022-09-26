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
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

type cycleA struct {
}

type cycleB struct {
}

type cycleC struct {
}

func newCycleA(_ *cycleB) *cycleA {
	return &cycleA{}
}

func newCycleB(_ *cycleC) *cycleB {
	return &cycleB{}
}

func newCycleC(_ *cycleA) *cycleC {
	return &cycleC{}
}

var _ = ginkgo.Describe("cycle dependency", func() {
	ginkgo.It("error detected", func() {
		cycleAfn, _ := ParseFunc(newCycleA)
		cycleBfn, _ := ParseFunc(newCycleB)
		cycleCfn, _ := ParseFunc(newCycleC)

		nodeC := &graphNode{
			name:      cycleCfn.ActualName(),
			resolved:  false,
			ctorType:  cycleCfn.FuncType,
			ctorValue: cycleCfn.FuncValue,
		}

		nodeB := &graphNode{
			name:      cycleBfn.ActualName(),
			resolved:  false,
			ctorType:  cycleBfn.FuncType,
			ctorValue: cycleBfn.FuncValue,
		}

		nodeA := &graphNode{
			name:      cycleAfn.ActualName(),
			resolved:  false,
			ctorType:  cycleAfn.FuncType,
			ctorValue: cycleAfn.FuncValue,
		}

		depChain := make([]*graphNode, 0)
		depChain = append(depChain, nodeA)
		depChain = append(depChain, nodeB)
		depChain = append(depChain, nodeC)

		err := cycleDependencyCheck(depChain, nodeA)
		gomega.Expect(err.Error()).To(gomega.Equal("cycle dependencies found: \n\t" +
			"github.com/go-piper/piper.newCycleA\n\t" +
			"depends on github.com/go-piper/piper.newCycleB\n\t" +
			"depends on github.com/go-piper/piper.newCycleC\n\t" +
			"depends on github.com/go-piper/piper.newCycleA"))
	})
})
