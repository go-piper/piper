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
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

// Field defines the reflection information after parsing field.
type Field struct {
	Package      string
	Name         string
	IsPointer    bool
	IsCollection bool
}

// ParseField parses the given interface to Field.
func ParseField(field any) (*Field, error) {
	fieldType := reflect.TypeOf(field)
	if fieldType.Kind() == reflect.Func {
		return nil, errors.New("param is not a valid field")
	}
	return ParseFieldType(fieldType)
}

// ParseFieldType parses the given field type to Field.
func ParseFieldType(fieldType reflect.Type) (*Field, error) {
	realType := fieldType
	kind := fieldType.Kind()
	if kind == reflect.Ptr || kind == reflect.Slice {
		realType = fieldType.Elem()
	} else if kind == reflect.Func &&
		len(realType.PkgPath()) == 0 && len(realType.Name()) == 0 {
		return nil, errors.New("cannot parse closure function")
	}

	name := realType.Name()
	// chan has no name
	if len(name) == 0 {
		name = realType.String()
	}

	return &Field{
		Package:      realType.PkgPath(),
		Name:         name,
		IsPointer:    fieldType.Kind() == reflect.Ptr,
		IsCollection: fieldType.Kind() == reflect.Slice,
	}, nil
}

// Equal checks if the given Field is equal current.
func (f *Field) Equal(other *Field) bool {
	return f.Name == other.Name &&
		f.IsPointer == other.IsPointer &&
		f.IsCollection == f.IsCollection
}

// String returns a string representation of the field.
// For example: *[]path/to/pkg.DemoStruct.
func (f *Field) String() string {
	ptrChar := ""
	arrayChar := ""
	if f.IsPointer {
		ptrChar = "*"
	}
	if f.IsCollection {
		arrayChar = "[]"
	}
	return fmt.Sprintf("%s%s%s.%s", ptrChar, arrayChar, f.Package, f.Name)
}

func (f *Field) ActualName() string {
	return fmt.Sprintf("%s.%s", f.Package, f.Name)
}

// Func defines the refletion information after parsing func.
type Func struct {
	Package   string
	Name      string
	InParam   []*Field
	OutParam  *Field
	FuncType  reflect.Type
	FuncValue reflect.Value
}

// ParseFunc parse func interface into Func with reflect.
func ParseFunc(fn any) (*Func, error) {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		return nil, errors.New("the given interface is not function")
	}

	if fnType.NumOut() == 0 {
		return nil, errors.New("the given function has no output parameters")
	}

	var inParams = make([]*Field, 0)
	for i := 0; i < fnType.NumIn(); i++ {
		field, err := ParseFieldType(fnType.In(i))
		if err != nil {
			return nil, err
		}
		inParams = append(inParams, field)
	}

	outField, err := ParseFieldType(fnType.Out(0))
	if err != nil {
		return nil, err
	}

	fptr := reflect.ValueOf(fn).Pointer()
	pkgName, funcName := splitFuncName(runtime.FuncForPC(fptr).Name())

	return &Func{
		Package:   pkgName,
		Name:      funcName,
		InParam:   inParams,
		OutParam:  outField,
		FuncType:  fnType,
		FuncValue: reflect.ValueOf(fn),
	}, nil
}

func splitFuncName(fn string) (pkgName string, funcName string) {
	if len(fn) == 0 {
		return
	}

	index := 0
	if i := strings.LastIndex(fn, "/"); i >= 0 {
		index = i
	}
	if i := strings.Index(fn[index:], "."); i >= 0 {
		index += i
	}
	return fn[:index], fn[index+1:]
}

// ActualName returns a string representation of the func.
// For example: path/to/pkg.DemoFunc.
func (f *Func) ActualName() string {
	return fmt.Sprintf("%s.%s", f.Package, f.Name)
}
