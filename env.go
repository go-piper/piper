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
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// AppEnv represents the environment for the whole lifecycle of application.
type AppEnv struct {
	// just wrap some functions of viper
	vp      *viper.Viper
	cliName string
}

// newAppEnv creates a new instance of AppEnv which can be used to get
// configuration, merge other config or unmarshal configuration to struct.
func newAppEnv() *AppEnv {
	binPath, err := os.Executable()
	if err != nil {
		Panicf("create application environment fail: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		Panicf("create application environment fail: %v", err)
	}

	var binDir string
	var cliName = binPath
	index := strings.LastIndex(binPath, "/")
	if index > 0 {
		cliName = binPath[index+1:]
		binDir = binPath[:index]
	}

	env := &AppEnv{
		vp:      viper.New(),
		cliName: cliName,
	}

	// we currently only support yaml config file
	// TODO: filesystem
	//env.vp.SetFs()
	env.vp.SetConfigType("yml")
	env.vp.AddConfigPath(resourcesDir)
	if wd != binDir {
		env.vp.AddConfigPath(filepath.Join(filepath.Dir(binPath), resourcesDir))
	}
	// this is used for embedded file system
	env.vp.AddConfigPath(fmt.Sprintf("/%s/%s", piper, resourcesDir))

	// set default value
	env.vp.SetDefault(fmt.Sprintf("%s.application.name", piper), fmt.Sprintf("%s-app", piper))

	return env
}

func (c *AppEnv) mergeWith(name string) error {
	c.vp.SetConfigName(name)

	if err := c.vp.MergeInConfig(); err != nil {
		return err
	}

	return c.MergeConfigMap(c.vp.AllSettings())
}

// expandMerge expands environment variables and merges config.
func (c *AppEnv) expandMerge(cfg map[string]any, isChild bool) (map[string]any, error) {
	newCfg := make(map[string]any)

	for k, v := range cfg {
		switch v.(type) {
		case string:
			newCfg[k] = ExpandEnv(v.(string))

		case map[string]any:
			cc, _ := c.expandMerge(v.(map[string]any), true)
			newCfg[k] = cc

		default:
			newCfg[k] = v
		}
	}

	if isChild {
		return newCfg, nil
	}

	return nil, c.vp.MergeConfigMap(newCfg)
}

// MergeConfigMap merges external configuration map into the configuration of application.
func (c *AppEnv) MergeConfigMap(cfg map[string]any) error {
	_, err := c.expandMerge(cfg, false)
	return err
}

// Unmarshal unmarshal configuration with `piper` tag.
func (c *AppEnv) Unmarshal(prefix string, rawVal any) error {
	return c.vp.UnmarshalKey(prefix, rawVal, func(c *mapstructure.DecoderConfig) {
		c.TagName = piper
	})
}

// Profile gets current active profile in command line if existed.
func (c *AppEnv) Profile() string {
	return c.vp.GetString(keyProfile)
}

// ConfigName gets current config name for this application.
func (c *AppEnv) ConfigName() string {
	return c.vp.GetString(keyConfigName)
}

// cmdName returns the command line name of the executed bin.
func (c *AppEnv) cmdName() string {
	return c.cliName
}
