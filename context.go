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
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Context represents the context for the whole lifecycle of application.
type Context struct {
	// just wrap some functions of viper
	vp      *viper.Viper
	cliName string
}

// newContext creates a new instance of Context which can be used to get
// configuration, merge other config or unmarshal configuration to struct.
func newContext(rs embed.FS) *Context {
	binPath, err := os.Executable()
	if err != nil {
		Panicf("fail to create piper context: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		Panicf("fail to create piper context: %v", err)
	}

	var binDir string
	var cliName = binPath
	index := strings.LastIndex(binPath, "/")
	if index > 0 {
		cliName = binPath[index+1:]
		binDir = binPath[:index]
	}

	env := &Context{
		vp:      viper.New(),
		cliName: cliName,
	}

	// we currently only support yaml config file
	env.vp.SetFs(newResourceFs(rs))
	env.vp.SetConfigType("yml")
	env.vp.AddConfigPath(resourcesDir)
	if wd != binDir {
		env.vp.AddConfigPath(filepath.Join(filepath.Dir(binPath), resourcesDir))
	}
	// this is used for embedded file system
	env.vp.AddConfigPath(fmt.Sprintf("/%s/%s", resourcesDir, piper))

	// set default value
	env.vp.SetDefault(fmt.Sprintf("%s.application.name", piper), fmt.Sprintf("%s-app", piper))

	return env
}

// cmdName returns the command line name of the executed bin.
func (c *Context) cmdName() string {
	return c.cliName
}

func (c *Context) mergeWith(name string) error {
	c.vp.SetConfigName(name)

	if err := c.vp.MergeInConfig(); err != nil {
		return err
	}

	return c.MergeConfigMap(c.vp.AllSettings())
}

// expandMerge expands environment variables and merges config.
func (c *Context) expandMerge(cfg map[string]any, isChild bool) (map[string]any, error) {
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
func (c *Context) MergeConfigMap(cfg map[string]any) error {
	_, err := c.expandMerge(cfg, false)
	return err
}

// Unmarshal unmarshal configuration with `piper` tag.
func (c *Context) Unmarshal(prefix string, rawVal any) error {
	return c.vp.UnmarshalKey(prefix, rawVal, func(c *mapstructure.DecoderConfig) {
		c.TagName = piper
	})
}

// Profile gets current active profile in command line if existed.
func (c *Context) Profile() string {
	return c.vp.GetString(keyProfile)
}

// ConfigName gets current config name for this application.
func (c *Context) ConfigName() string {
	return c.vp.GetString(keyConfigName)
}

// GetString returns the value associated with the key as a string.
func (c *Context) GetString(key string) string {
	return c.vp.GetString(key)
}

// GetBool returns the value associated with the key as a boolean.
func (c *Context) GetBool(key string) bool {
	return c.vp.GetBool(key)
}

// GetInt returns the value associated with the key as an integer.
func (c *Context) GetInt(key string) int {
	return c.vp.GetInt(key)
}

// GetInt32 returns the value associated with the key as an integer.
func (c *Context) GetInt32(key string) int32 {
	return c.vp.GetInt32(key)
}

// GetInt64 returns the value associated with the key as an integer.
func (c *Context) GetInt64(key string) int64 {
	return c.vp.GetInt64(key)
}

// GetUint returns the value associated with the key as an unsigned integer.
func (c *Context) GetUint(key string) uint {
	return c.vp.GetUint(key)
}

// GetUint32 returns the value associated with the key as an unsigned integer.
func (c *Context) GetUint32(key string) uint32 {
	return c.vp.GetUint32(key)
}

// GetUint64 returns the value associated with the key as an unsigned integer.
func (c *Context) GetUint64(key string) uint64 {
	return c.vp.GetUint64(key)
}

// GetFloat64 returns the value associated with the key as a float64.
func (c *Context) GetFloat64(key string) float64 {
	return c.vp.GetFloat64(key)
}

// GetTime returns the value associated with the key as time.
func (c *Context) GetTime(key string) time.Time {
	return c.vp.GetTime(key)
}

// GetDuration returns the value associated with the key as a duration.
func (c *Context) GetDuration(key string) time.Duration {
	return c.vp.GetDuration(key)
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func (c *Context) GetStringSlice(key string) []string {
	return c.vp.GetStringSlice(key)
}

// GetStringMap returns the value associated with the key as a map of interfaces.
func (c *Context) GetStringMap(key string) map[string]interface{} {
	return c.vp.GetStringMap(key)
}

// GetStringMapString returns the value associated with the key as a map of strings.
func (c *Context) GetStringMapString(key string) map[string]string {
	return c.vp.GetStringMapString(key)
}

// GetStringMapStringSlice returns the value associated with the key
// as a map to a slice of strings.
func (c *Context) GetStringMapStringSlice(key string) map[string][]string {
	return c.vp.GetStringMapStringSlice(key)
}
