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

func (s *applicationConfigLoader) Order() int {
	return math.MinInt32 + 1
}

func (s *applicationConfigLoader) Load(ctx *Context) error {
	configName := ctx.ConfigName()
	profile := ctx.Profile()
	if len(profile) != 0 {
		configName += "-" + profile
	}

	if err := s.readConfig(ctx, configName, false); err != nil {
		return err
	}

	// initialize logging after application config loaded
	return LoggingSystem().Initialize(ctx)
}

func (s *applicationConfigLoader) readConfig(ctx *Context,
	configName string, rollback bool) (err error) {
	if err = ctx.mergeWith(configName); err == nil {
		return nil
	}

	// if no default config file found, just return error
	if rollback {
		return s.readError(ctx)
	}

	if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		return s.readError(ctx)
	} else {
		return s.readConfig(ctx, ctx.ConfigName(), true)
	}
}

func (s *applicationConfigLoader) readError(ctx *Context) error {
	var profileConfigErrMsg = ""
	if len(ctx.Profile()) != 0 {
		profileConfigErrMsg = fmt.Sprintf(" or %v-%v.yml", ctx.ConfigName(), ctx.Profile())
	}
	return errors.New(
		fmt.Sprintf("no %v.yml%v config file found in resources, "+
			"at least one config file should be presented",
			ctx.ConfigName(), profileConfigErrMsg))
}
