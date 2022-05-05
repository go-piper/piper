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
	"math"
)

// ConfigProperty represents a group of properties for config in yaml file.
type ConfigProperty interface {
	// Prefix returns the prefix in yaml config file.
	Prefix() string
}

// Initializer represents an initializer which will be invoked when application is
// initializing. This can be used to change config or do some initial staff.
// Note that the configuration properties may be changed before application was
// initialized, configuration properties should not be wired in your initializers.
// Check StartListener interface if you want a callback before application started.
type Initializer interface {
	// Initialize will be invoked when the application is initializing.
	Initialize(ctx *Context)
}

// Ordered represents the priority of type. This can be used to control the initialization
// order of wired variables. Higher values of `Order` has lower priority.
type Ordered interface {
	// Order get the order for implemented type.
	Order() int
}

const (
	// LowestOrder represents the lowest order in dependency tree.
	LowestOrder = math.MaxInt32
	// HighestOrder represents the highest order in dependency tree.
	HighestOrder = 0
)

// ApplicationProperty defines the property of piper.application section in yaml config.
type ApplicationProperty struct {
	Name string `piper:"name"`
}

func (*ApplicationProperty) Prefix() string {
	return "piper.application"
}
