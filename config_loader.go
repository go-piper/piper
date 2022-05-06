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
	"math"

	"github.com/spf13/viper"
)

// ConfigLoader defines the interface for config loader which will load
// configurations from local files or remote server.
type ConfigLoader interface {
	Ordered

	// Load config with specified profile.
	Load(ctx *Context) error
}

func init() {
	Wire(&applicationConfigLoader{})
}

// applicationConfigLoader is a config loader to load configuration in application-xx.yml.
type applicationConfigLoader struct {
}

func (l *applicationConfigLoader) Order() int {
	return math.MinInt32 + 1
}

func (l *applicationConfigLoader) Load(ctx *Context) error {
	if err := l.readConfig(ctx, ctx.ConfigName(), false); err != nil {
		return err
	}

	// initialize logging after application config loaded
	return LoggingSystem().Initialize(ctx)
}

func (l *applicationConfigLoader) readConfig(ctx *Context, name string, complete bool) error {
	err := ctx.mergeWith(name)
	if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		return err
	}

	// if config loaded completely, just return
	if complete {
		return nil
	}

	return l.readConfig(ctx, fmt.Sprintf("%s-%s", name, ctx.Profile()), true)
}
