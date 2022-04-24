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
	"reflect"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type fooTest struct {
}

type barTest struct {
}

type paramTest struct {
}

type fooDefTest struct {
}

type paramDefTest struct {
}

type primitiveTest struct {
}

type intfTest interface {
	TestIntf()
}

type intfImpl struct {
}

func (*intfImpl) TestIntf() {
}

func newFooTest(_ *paramTest) *fooTest {
	return &fooTest{}
}

func newFooDefTest(_ *paramDefTest) *fooDefTest {
	return &fooDefTest{}
}

func newBarTest(_ *fooTest) *barTest {
	return &barTest{}
}

func newIntfTest() intfTest {
	return &intfImpl{}
}

func newIntfDefaultTest(_ intfTest) *fooTest {
	return &fooTest{}
}

func newPrimitiveTest(_ *string, _ *int, _ chan float32,
	_ map[string]string, _ []string, _ []paramTest) *primitiveTest {
	return &primitiveTest{}
}

func newLazyTest(_ Lazy[*fooTest]) *barTest {
	return &barTest{}
}

func mockDepTree() *depTree {
	return &depTree{
		providers:       make(map[providerKey][]*graphNode),
		unresolvedNodes: make([]*graphNode, 0),
		graphNodes:      make([]*graphNode, 0),
		options:         make(map[string][]*WireOption),
	}
}

func TestDepTree(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "piper dep tree test")
}

var _ = Describe("private func", func() {
	It("build key", func() {
		field, _ := ParseField(&fooTest{})
		key := _depTree().buildKey(field, "")
		Expect(key).To(Equal(providerKey{"github.com/go-piper/piper.fooTest", "", true}))
	})
	It("build uuid", func() {
		container := mockDepTree()
		sUuid := container.buildUuid(&fooTest{})
		fnUuid := container.buildUuid(newFooTest)
		Expect(len(sUuid)).To(Equal(32))
		Expect(sUuid).NotTo(Equal(fnUuid))
	})
})

var _ = Describe("public func", func() {
	It("wire no dependency", func() {
		depTree := mockDepTree()
		depTree.wire(newFooTest)
		err := depTree.resolveDependencies()
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(Equal(
			"no dependency of *github.com/go-piper/piper.paramTest " +
				"found for github.com/go-piper/piper.newFooTest",
		))
	})
	It("wire multiple dependencies", func() {
		depTree := mockDepTree()
		depTree.wire(&paramTest{}, &paramTest{})
		depTree.wire(newFooTest)
		err := depTree.resolveDependencies()
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(Equal(
			"more than one dependencies of *github.com/go-piper/piper.paramTest " +
				"found for github.com/go-piper/piper.newFooTest"))
	})
	It("wire multiple dependencies with primary", func() {
		depTree := mockDepTree()
		depTree.wire(&paramTest{})
		depTree.wireWithOption(&paramTest{}, Primary())
		depTree.wire(newFooTest)
		err := depTree.resolveDependencies()
		Expect(err).To(BeNil())
	})
	It("wire with default option", func() {
		depTree := mockDepTree()
		depTree.wireWithOption(newFooDefTest, Default(&paramDefTest{}))
		err := depTree.resolveDependencies()
		Expect(err).To(BeNil())
	})
	It("wire with error default option", func() {
		depTree := mockDepTree()
		depTree.wireWithOption(newFooDefTest, Default(&paramTest{}))
		err := depTree.resolveDependencies()
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(
			Equal("the default value *github.com/go-piper/piper.paramTest" +
				"\n\tis not match *github.com/go-piper/piper.paramDefTest" +
				"\n\tin github.com/go-piper/piper.newFooDefTest"))
	})
	It("wire with func default option", func() {
		depTree := mockDepTree()
		depTree.wireWithOption(newIntfDefaultTest, Default(newIntfTest()))
		err := depTree.resolveDependencies()
		Expect(err).To(BeNil())
	})
	It("wire without name", func() {
		depTree := mockDepTree()
		depTree.wire(&paramTest{})
		depTree.wireWithOption(newFooTest)
		depTree.wireWithOption(newBarTest, NameIn("myfoo"))
		err := depTree.resolveDependencies()
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(Equal(
			"no dependency of *github.com/go-piper/piper.fooTest(myfoo) " +
				"found for github.com/go-piper/piper.newBarTest"))
	})
	It("wire with name", func() {
		depTree := mockDepTree()
		depTree.wire(&paramTest{})
		depTree.wireWithOption(newFooTest, NameOut("foo"))
		depTree.wireWithOption(newBarTest, NameIn("foo"))
		err := depTree.resolveDependencies()
		Expect(err).To(BeNil())
	})
	It("wire with profiles", func() {
		depTree := mockDepTree()
		depTree.profile = "test"
		depTree.wireWithOption(newFooTest, Active("dev", "test"))
		depTree.wire(&paramTest{}, newBarTest)
		err := depTree.resolveDependencies()
		Expect(err).To(BeNil())
	})
	It("wire with error profiles", func() {
		depTree := mockDepTree()
		depTree.profile = "dev"
		depTree.wireWithOption(newFooTest, Active("test"))
		depTree.wire(&paramTest{}, newBarTest)
		err := depTree.resolveDependencies()
		Expect(err).NotTo(BeNil())
	})
	It("get struct", func() {
		depTree := mockDepTree()
		depTree.wire(&paramTest{})
		_ = depTree.resolveDependencies()

		result := depTree.retrieve(reflect.TypeOf((*paramTest)(nil)))
		paramTests := make([]*paramTest, 0)
		for _, v := range result {
			paramTests = append(paramTests, v.(*paramTest))
		}
		Expect(1).To(Equal(len(paramTests)))
	})
	It("get interface", func() {
		depTree := mockDepTree()
		depTree.wire(newIntfTest())
		_ = depTree.resolveDependencies()

		result := depTree.retrieve(reflect.TypeOf((*intfTest)(nil)))
		intfTests := make([]intfTest, 0)
		for _, v := range result {
			intfTests = append(intfTests, v.(intfTest))
		}
		Expect(1).To(Equal(len(intfTests)))
	})
	It("wire primitive type", func() {
		depTree := mockDepTree()
		var primitiveStr = "this is string"
		var primitiveInt = 2
		var primitiveChan = make(chan float32)
		var primitiveMap = make(map[string]string)
		var primitiveSlice = make([]string, 0)
		var structSlice = make([]paramTest, 2)
		depTree.wireWithOption(&primitiveStr, NameOut("myString"))
		depTree.wireWithOption(&primitiveInt, NameOut("myInt"))
		depTree.wireWithOption(primitiveChan, NameOut("myChan"))
		depTree.wireWithOption(primitiveMap, NameOut("myMap"))
		depTree.wireWithOption(primitiveSlice, NameOut("mySlice"))
		depTree.wire(structSlice)
		depTree.wireWithOption(newPrimitiveTest,
			NameIn("myString"), NameIn("myInt"), NameIn("myChan"),
			NameIn("myMap"), NameIn("mySlice"))
		err := depTree.resolveDependencies()
		Expect(err).To(BeNil())

		result := depTree.retrieve(reflect.TypeOf((*primitiveTest)(nil)))
		primitiveTests := make([]*primitiveTest, 0)
		for _, v := range result {
			primitiveTests = append(primitiveTests, v.(*primitiveTest))
		}
		Expect(1).To(Equal(len(primitiveTests)))
	})
	It("wire lazy", func() {
		depTree := mockDepTree()
		depTree.wire(&paramTest{})
		depTree.wireWithOption(newFooTest, LazyOut())
		depTree.wire(newLazyTest)
		err := depTree.resolveDependencies()
		Expect(err).To(BeNil())
	})
})
