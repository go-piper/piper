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
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
)

var (
	onceDepTree sync.Once
	singleton   *depTree
)

type depTree struct {
	providers       map[providerKey][]*graphNode
	unresolvedNodes []*graphNode
	graphNodes      []*graphNode
	options         map[string][]*WireOption
	profile         string
}

// graphNode represents a node in dependencies graph.
type graphNode struct {
	id        string
	name      string
	ctorType  reflect.Type
	ctorValue reflect.Value
	lazyType  reflect.Type
	provided  any

	resolved     bool
	instantiated bool
	isCollection bool
	isLazy       bool
	dependencies []*graphNode
}

// Lazy represents lazy load for field.
type Lazy[T any] func() T

// _depTree will return the singleton global depTree to use.
func _depTree() *depTree {
	onceDepTree.Do(func() {
		singleton = &depTree{
			providers:       make(map[providerKey][]*graphNode),
			unresolvedNodes: make([]*graphNode, 0),
			graphNodes:      make([]*graphNode, 0),
			options:         make(map[string][]*WireOption),
		}
	})

	return singleton
}

// Wire registers field or func provider into global container. Then the container will
// resolve the dependecies for these providers. For example:
//
//  func newA(some SomeType) *TypeA {
//      return &TypeA {
//          ...
//      }
//  }
//
//  piper.Wire(newA)
//
// it also supports to wire multiple providers:
//
//  piper.Wire(newA, &TypeB{}, ...)
func Wire(providers ...any) {
	_depTree().wire(providers...)
}

// WireWithOption wires with options. The order of options keep the same with the order
// of in parameters of provider if it was a func, and wire option should be the last one.
// The field should have only wire out option. For example:
//
//  func newSomething(paramA typeA, paramB typeB) *Something {
//      return &Something{
//          ...
//      }
//  }
//
//  piper.WireWithOption(newSomething, piper.Name("MyA"))
//
// or use can set default value for one in parameter:
//
//  piper.WireWithOption(newSomething, piper.Default(defaultA))
//
// the field provider can only use:
//
//  piper.WireWithOption(&Otherthing{...}, piper.NameOut("MyThing"))
func WireWithOption(provider any, opts ...*WireOption) {
	_depTree().wireWithOption(provider, opts...)
}

// Retrieve gets all the values for the given type with order.
func Retrieve[T any](tp reflect.Type) []T {
	var fields = make([]T, 0)
	for _, v := range _depTree().retrieve(tp) {
		fields = append(fields, v.(T))
	}

	return fields
}

func (dt *depTree) wire(providers ...any) {
	for _, p := range providers {
		dt.wireWithOption(p)
	}
}

func (dt *depTree) wireWithOption(provider any, opts ...*WireOption) {
	if provider == nil {
		panic("pre process provider error: cannot be nil")
	}

	if _, ok := provider.(*WireOption); ok {
		panic("provider cannot be wire option")
	}

	providerType := reflect.TypeOf(provider)
	providerKind := providerType.Kind()
	if providerKind == reflect.Func {
		dt.buildFuncNode(provider, opts...)
	} else {
		dt.buildFieldNode(provider, opts...)
	}
}
func (dt *depTree) validateWireOption(pType reflect.Type,
	actualName string, opts []*WireOption) error {
	if len(opts) == 0 {
		return nil
	}

	var hashOutOption = false
	for i, o := range opts {
		if err := o.validate(); err != nil {
			return err
		}

		if o.wireType == wireOut && i < len(opts)-1 {
			return newWireOutError(actualName)
		}

		if o.isWireIn() {
			o.index = i
		} else {
			hashOutOption = true
		}
	}

	inOptLen := len(opts)
	if hashOutOption {
		inOptLen--
	}

	if inOptLen > pType.NumIn() {
		return newWireInError(actualName)
	}

	return nil
}

func (dt *depTree) splitOptions(opts []*WireOption) ([]*WireOption, *WireOption) {
	inOpts := make([]*WireOption, 0)
	var outOpt *WireOption

	for _, o := range opts {
		if o.isWireIn() {
			inOpts = append(inOpts, o)
		} else if o.isWireOut() {
			outOpt = o
		}
	}

	return inOpts, outOpt
}

func (dt *depTree) nameOptValue(opt *WireOption) string {
	if opt != nil {
		return opt.name
	}

	return ""
}

func (dt *depTree) defaultOptValue(opt *WireOption) any {
	if opt != nil {
		return opt.defValue
	}

	return nil
}

func (dt *depTree) requiredOptValue(opt *WireOption) bool {
	return opt != nil && opt.required
}

func (dt *depTree) buildFieldNode(provider any, opts ...*WireOption) {
	if len(opts) > 1 || len(opts) == 1 &&
		(!opts[0].isWireOut() || opts[0].validate() != nil) {
		panic("pre process instantiated provider error: only one " +
			"wire out option can be passed to wire")
	}

	fieldType := reflect.TypeOf(provider)
	kind := fieldType.Kind()
	if kind != reflect.Ptr && kind != reflect.Chan &&
		kind != reflect.Map && kind != reflect.Slice {
		panic("pre process instantiated provider error: " +
			"only non nil pointer type, chan, map, slice can be passed to wire")
	}

	if (kind == reflect.Ptr && fieldType.Elem().Kind() != reflect.Struct ||
		kind == reflect.Chan || kind == reflect.Map ||
		kind == reflect.Slice && fieldType.Elem().Kind() != reflect.Struct) &&
		len(opts) == 0 {
		panic("pre process instantiated provider error: " +
			"primitive type should be provided with 'WireWithOption'")
		return
	}

	field, err := ParseField(provider)
	if err != nil {
		Panicf("invalid field: %v", field)
		return
	}

	var alias string
	if len(opts) == 1 {
		alias = opts[0].name
	}
	key := dt.buildKey(field, alias)
	savedNodes := dt.providers[key]
	if savedNodes == nil {
		savedNodes = make([]*graphNode, 0)
	}

	uuid := dt.buildUuid(provider)
	dt.options[uuid] = opts

	savedNodes = append(savedNodes, &graphNode{
		id:           uuid,
		name:         field.ActualName(),
		resolved:     true,
		instantiated: true,
		provided:     provider,
	})
	dt.providers[key] = savedNodes
}

func (dt *depTree) buildFuncNode(provider any, opts ...*WireOption) {
	providerType := reflect.TypeOf(provider)
	if providerType.NumOut() != 1 {
		Panicf("pre process provider error: no or more than "+
			"one out parameter for the given provider: %v", providerType)
		return
	}

	fn, err := ParseFunc(provider)
	if err != nil {
		Panicf("parse func error: %v", err)
		return
	}

	if err := dt.validateWireOption(providerType, fn.ActualName(), opts); err != nil {
		Panicf("validate wire option error: %v", err)
		return
	}

	var alias string
	_, outOpt := dt.splitOptions(opts)
	if outOpt != nil {
		alias = outOpt.name
	}

	key := dt.buildKey(fn.OutParam, alias)
	savedNodes := dt.providers[key]
	if savedNodes == nil {
		savedNodes = make([]*graphNode, 0)
	}

	uuid := dt.buildUuid(provider)
	dt.options[uuid] = opts

	newNode := &graphNode{
		id:        uuid,
		name:      fn.ActualName(),
		resolved:  false,
		ctorType:  fn.FuncType,
		ctorValue: fn.FuncValue,
	}
	savedNodes = append(savedNodes, newNode)
	dt.providers[key] = savedNodes

	// save unresolved node
	dt.unresolvedNodes = append(dt.unresolvedNodes, newNode)
}

func (dt *depTree) buildKey(field *Field, alias string) providerKey {
	return providerKey{
		isPointer: field.IsPointer,
		name:      field.ActualName(),
		alias:     alias,
	}
}

func (dt *depTree) buildUuid(provider any) string {
	pValue := reflect.ValueOf(provider)
	md5Hash := md5.New()
	md5Hash.Write([]byte(fmt.Sprint(pValue.Pointer())))

	return hex.EncodeToString(md5Hash.Sum(nil))
}

func (dt *depTree) active(node *graphNode) bool {
	opts := dt.options[node.id]
	_, outOpt := dt.splitOptions(opts)

	if outOpt == nil || len(outOpt.profiles) == 0 {
		return true
	}

	for _, p := range outOpt.profiles {
		if p == dt.profile {
			return true
		}
	}

	return false
}

func (dt *depTree) resolveNode(nodeToResolve *graphNode, depChain []*graphNode) error {
	// if the node is resolved, do nothing
	if nodeToResolve.resolved {
		return nil
	}

	ctorType := nodeToResolve.ctorType
	numIn := ctorType.NumIn()
	// if no input parameters, means this node has no dependencies
	if numIn == 0 {
		nodeToResolve.resolved = true
		return nil
	}

	opts := dt.options[nodeToResolve.id]
	inOpts, _ := dt.splitOptions(opts)

	for i := 0; i < ctorType.NumIn(); i++ {
		inType := ctorType.In(i)
		kind := inType.Kind()
		isLazy := false
		originInType := inType

		if kind == reflect.Ptr {
			kind = inType.Elem().Kind()
		} else if kind == reflect.Func {
			lazyOut := originInType.Out(0)
			fn := reflect.FuncOf([]reflect.Type{}, []reflect.Type{lazyOut}, false)
			// TODO: any other way to know 'Lazy'?
			fullPkg := fmt.Sprintf("%s.%s", originInType.PkgPath(), originInType.Name())
			if strings.HasPrefix(fullPkg, "github.com/go-piper/piper.Lazy") &&
				inType.ConvertibleTo(fn) {
				isLazy = true
				inType = lazyOut
			}
		}

		field, err := ParseFieldType(inType)
		if err != nil {
			return errors.New(fmt.Sprintf("%s: %s", err, nodeToResolve.name))
		}

		var inOpt *WireOption
		if i < len(inOpts) {
			inOpt = inOpts[i]
		}
		key := dt.buildKey(field, dt.nameOptValue(inOpt))
		nodes, ok := dt.providers[key]
		if ok && len(nodes) != 0 {
			if kind == reflect.Slice {
				collectionNode := &graphNode{
					ctorType:     inType,
					resolved:     true,
					isCollection: true,
				}
				for _, node := range nodes {
					// if the node is not active in current profile, ignore
					if !dt.active(node) {
						continue
					}

					// resolve child node
					if err := dt.resolveChildNode(node, append(depChain,
						nodeToResolve)); err != nil {
						return err
					}
					collectionNode.dependencies = append(
						collectionNode.dependencies, node)
				}
				nodeToResolve.dependencies = append(nodeToResolve.dependencies, collectionNode)
			} else {
				var primaryNode *graphNode
				if len(nodes) > 1 {
					for _, n := range nodes {
						opts := dt.options[n.id]
						_, outOpt := dt.splitOptions(opts)
						if outOpt != nil && outOpt.primary {
							primaryNode = n
							break
						}
					}

					// no primary node found
					if primaryNode == nil {
						return newMultiDepError(key, nodeToResolve.name)
					}
				} else {
					primaryNode = nodes[0]
				}

				if !dt.active(primaryNode) {
					return newNoDepError(key, nodeToResolve.name)
				}

				if err := dt.resolveChildNode(primaryNode, append(depChain,
					nodeToResolve)); err != nil {
					return err
				}
				_, outOpts := dt.splitOptions(dt.options[primaryNode.id])
				if isLazy {
					if outOpts == nil || !outOpts.lazy {
						return errors.New(fmt.Sprintf("lazy options needed %s",
							primaryNode.name))
					}
					primaryNode.isLazy = isLazy
					primaryNode.lazyType = originInType
				} else if kind != reflect.Func {
					if outOpts != nil && outOpts.lazy {
						return errors.New(fmt.Sprintf(
							"lazy options need 'Lazy' to  call %s", primaryNode.name))
					}
				}
				nodeToResolve.dependencies = append(nodeToResolve.dependencies, primaryNode)
			}
		} else {
			// dependency is not found, try to get default
			defVal := dt.defaultOptValue(inOpt)
			if dt.requiredOptValue(inOpt) || defVal == nil {
				return newNoDepError(key, nodeToResolve.name)
			}

			defValType := reflect.TypeOf(defVal)
			if defValType.Kind() == reflect.Func {
				return errors.New("default value cannot be func")
			}

			defField, err := ParseField(defVal)
			if err != nil {
				return err
			}

			var typeMatched bool
			if inType.Kind() != reflect.Interface && (inType.Kind() == reflect.Ptr &&
				inType.Elem().Kind() != reflect.Interface) {
				typeMatched = defValType == inType
			} else {
				typeMatched = defValType.ConvertibleTo(inType)
			}

			if !defField.Equal(field) && !typeMatched {
				return newDefValueMismatchError(defField, field, nodeToResolve.name)
			}

			nodeToResolve.dependencies = append(nodeToResolve.dependencies, &graphNode{
				id:           dt.buildUuid(defVal),
				name:         key.name,
				resolved:     true,
				instantiated: true,
				provided:     defVal,
			})
		}
	}

	nodeToResolve.resolved = true

	return nil
}

func (dt *depTree) resolveChildNode(node *graphNode, chain []*graphNode) error {
	if !node.resolved {
		if err := cycleDependencyCheck(chain, node); err != nil {
			return err
		}

		if err := dt.resolveNode(node, chain); err != nil {
			return err
		}
	}

	return nil
}

func (dt *depTree) resolveDependencies() error {
	for _, nodeToResolve := range dt.unresolvedNodes {
		if err := dt.resolveNode(nodeToResolve, make([]*graphNode, 0)); err != nil {
			return err
		}
	}

	// clear unused resource
	dt.unresolvedNodes = nil

	for _, nodes := range dt.providers {
		for _, node := range nodes {
			dt.graphNodes = append(dt.graphNodes, node)
			dt.instantiate(node)
		}
	}

	return nil
}

func (dt *depTree) matchType(node *graphNode, fieldType reflect.Type) bool {
	ctorType := node.ctorType
	var outType reflect.Type
	if ctorType != nil {
		outType = ctorType.Out(0)
	} else if node.instantiated {
		outType = reflect.TypeOf(node.provided)
	} else {
		return false
	}

	if outType.Kind() != fieldType.Kind() {
		return false
	}

	// if type to match is not interface, check if type was the same
	if fieldType.Kind() != reflect.Interface &&
		fieldType.Elem().Kind() != reflect.Interface {
		return outType == fieldType
	}

	if outType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	return outType.Implements(fieldType)
}

func (dt *depTree) instantiate(node *graphNode) {
	if !node.resolved || node.instantiated {
		return
	}

	for _, dep := range node.dependencies {
		if dep.instantiated {
			continue
		}
		dt.instantiate(dep)
	}

	in := make([]reflect.Value, 0)
	if node.isCollection {
		for _, child := range node.dependencies {
			if child.instantiated {
				continue
			}
			dt.instantiate(child)
		}
		node.instantiated = true
		return
	} else {
		// the number of in parameters is equal to number of dependencies
		numIn := node.ctorType.NumIn()
		for i := 0; i < numIn; i++ {
			depNode := node.dependencies[i]
			if !depNode.isCollection {
				in = append(in, reflect.ValueOf(depNode.provided))
				continue
			}

			collectionIn := reflect.New(depNode.ctorType).Elem()
			for _, child := range depNode.dependencies {
				childKind := reflect.TypeOf(child.provided).Kind()
				childValue := reflect.ValueOf(child.provided)
				if childKind == reflect.Slice {
					collectionIn = reflect.AppendSlice(collectionIn, childValue)
				} else {
					collectionIn = reflect.Append(collectionIn, childValue)
				}
			}
			in = append(in, collectionIn)
		}
	}

	if node.isLazy {
		node.provided = dt.lazyInstantiate(node, in)
	} else {
		// instantiates the node with parameters
		out := node.ctorValue.Call(in)
		node.provided = out[0].Interface()
	}
	node.instantiated = true
}

func (dt *depTree) lazyInstantiate(node *graphNode, in []reflect.Value) any {
	ctorValue := node.ctorValue
	var instantiated = false
	var value reflect.Value
	proxyFn := reflect.MakeFunc(node.lazyType, func(args []reflect.Value) []reflect.Value {
		if !instantiated {
			value = reflect.ValueOf(ctorValue.Call(in)[0].Interface())
		}

		return []reflect.Value{value}
	})

	lazyFn := reflect.New(node.lazyType)
	lazyFn.Elem().Set(proxyFn)

	return lazyFn.Elem().Interface()
}

func (dt *depTree) retrieve(tp reflect.Type) []any {
	var fields = make([]any, 0)
	var orderedFields = make([]any, 0)

	for _, node := range dt.graphNodes {
		if dt.matchType(node, tp) {
			if !node.instantiated {
				dt.instantiate(node)
			}
			// check again after instantiating
			if node.instantiated {
				if _, ok := node.provided.(Ordered); ok {
					orderedFields = append(orderedFields, node.provided)
				} else {
					fields = append(fields, node.provided)
				}
			}
		}
	}

	sort.Slice(orderedFields, func(i, j int) bool {
		return orderedFields[i].(Ordered).Order() < orderedFields[j].(Ordered).Order()
	})

	return append(orderedFields, fields...)
}
