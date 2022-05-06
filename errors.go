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
	"fmt"
)

type noDepError struct {
	key      providerKey
	nodeName string
}

func newNoDepError(key providerKey, nodeName string) error {
	return &noDepError{
		key:      key,
		nodeName: nodeName,
	}
}

func (e *noDepError) Error() string {
	return fmt.Sprintf("no dependency of %s found for %s", e.key, e.nodeName)
}

type multiDepError struct {
	key      providerKey
	nodeName string
}

func newMultiDepError(key providerKey, nodeName string) error {
	return &multiDepError{
		key:      key,
		nodeName: nodeName,
	}
}

func (e *multiDepError) Error() string {
	return fmt.Sprintf("more than one dependencies of %s found for %s", e.key, e.nodeName)
}

type defValueMismatchError struct {
	defField, field *Field
	nodeName        string
}

func newDefValueMismatchError(defField *Field, field *Field, nodeName string) error {
	return &defValueMismatchError{
		defField: defField,
		field:    field,
		nodeName: nodeName,
	}
}

func (e *defValueMismatchError) Error() string {
	return fmt.Sprintf("the default value %s\n\tis not match %s\n\tin %s",
		e.defField, e.field, e.nodeName)
}

type wireInError struct {
	providerName string
}

func newWireInError(providerName string) error {
	return &wireInError{
		providerName: providerName,
	}
}

func (e *wireInError) Error() string {
	return fmt.Sprintf("wire in option is more than the in "+
		"parameters in provider: %s", e.providerName)
}

type wireOutError struct {
	providerName string
}

func newWireOutError(providerName string) error {
	return &wireOutError{
		providerName: providerName,
	}
}

func (e *wireOutError) Error() string {
	return fmt.Sprintf("wire out option can only be the last "+
		"option in provider: %s", e.providerName)
}

type appStartError struct {
	err error
}

func newAppStartError(err error) *appStartError {
	return &appStartError{
		err: err,
	}
}

func (e *appStartError) Error() string {
	return "\n" +
		"**********************************" + "\n" +
		"*    application start failed    *" + "\n" +
		"**********************************" + "\n\n" +
		e.err.Error()
}
