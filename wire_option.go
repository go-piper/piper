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
	"strings"
)

const (
	wireIn wireType = iota
	wireOut
)

type wireType int

// WireOption add additional option for the provider when wiring.
type WireOption struct {
	index    int
	required bool
	primary  bool
	lazy     bool
	name     string
	defValue any
	profiles []string
	wireType wireType
}

type applyOptionFunc func(*WireOption)

func applyOption(fn applyOptionFunc) *WireOption {
	opt := &WireOption{
		index:    0,
		required: true,
	}
	fn(opt)
	return opt
}

// In is a convenient func which returns WireOption for in parameter type.
func In() *WireOption {
	return applyOption(func(option *WireOption) {
		option.wireType = wireIn
	})
}

// Out is a convenient func which returns WireOption for out parameter type.
func Out() *WireOption {
	return applyOption(func(option *WireOption) {
		option.wireType = wireOut
	})
}

// NameIn is a convenient func which returns WireOption with name option.
// This func alias to In().Name(name string).
func NameIn(name string) *WireOption {
	return applyOption(func(option *WireOption) {
		option.name = name
		option.wireType = wireIn
	})
}

// Default is a convenient func which returns WireOption with default value option.
// This func alias to In().Default(defVal any).
func Default(defVal any) *WireOption {
	return applyOption(func(option *WireOption) {
		option.defValue = defVal
		option.required = false
	})
}

// Primary is convenient func which returns WireOption with primary option.
// This func means the wired out type will be the first candidate.
// If multiple dependencies was found, then this wired out type will be selected.
func Primary() *WireOption {
	return applyOption(func(option *WireOption) {
		option.primary = true
		option.wireType = wireOut
	})
}

// Active is convenient func which returns WireOption with active option.
// The wired out type will be available when the profiles matches.
func Active(profiles ...string) *WireOption {
	return applyOption(func(option *WireOption) {
		option.profiles = profiles
		option.wireType = wireOut
	})
}

// LazyOut is convenient func which returns WireOption with lazy option.
func LazyOut() *WireOption {
	return applyOption(func(option *WireOption) {
		option.lazy = true
		option.wireType = wireOut
	})
}

// NameOut is a convenient func which returns WireOption with name option and out
// parameter type. This func alias to Out().Name(name string).
func NameOut(name string) *WireOption {
	return applyOption(func(option *WireOption) {
		option.name = name
		option.wireType = wireOut
	})
}

func (o *WireOption) validate() error {
	if o.wireType == wireOut {
		if o.required == false {
			return errors.New("required option of wire out parameter cannot be false")
		}

		if o.defValue != nil {
			return errors.New("default option of wire out parameter cannot exist")
		}
	} else {
		if o.primary {
			return errors.New("primary option of wire in parameter cannot exist")
		}

		if len(o.profiles) != 0 {
			return errors.New("active option of wire in parameter cannot exist")
		}
	}

	return nil
}

func (o *WireOption) isWireIn() bool {
	return o.wireType == wireIn
}

func (o *WireOption) isWireOut() bool {
	return o.wireType == wireOut
}

// Name sets the name in this option.
func (o *WireOption) Name(name string) *WireOption {
	o.name = strings.TrimSpace(name)
	return o
}

// Default sets default value in this option.
func (o *WireOption) Default(defVal any) *WireOption {
	o.defValue = defVal
	o.required = false
	return o
}

// Primary sets primary in this option.
func (o *WireOption) Primary() *WireOption {
	o.primary = true
	return o
}

// Active add actived profiles in this option.
func (o *WireOption) Active(profiles ...string) *WireOption {
	o.profiles = profiles
	return o
}
