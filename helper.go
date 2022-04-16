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
	"strings"
)

var (
	Version   = "dev"
	BuildTime = "unkown"
	GitCommit = "unkown"
)

const (
	resourcesDir  = "resources"
	piper         = "piper"
	keyProfile    = "profile"
	keyConfigName = "config"
)

// Panicf makes panic with format support.
func Panicf(format string, a ...interface{}) {
	err := fmt.Sprintf(format, a...)
	panic(err)
}

// ExpandEnv replaces ${var} or ${var:-def} in the string with environment variables.
func ExpandEnv(s string) string {
	length := len(s)
	if s[0] != '$' || s[1] != '{' || s[length-1] != '}' {
		return s
	}
	s = s[2 : length-1]

	var envName string
	var defVal string
	index := strings.Index(s, ":-")
	if index > 0 {
		envName = s[:index]
		defVal = s[index+2:]
	} else {
		envName = s
	}

	value := os.Getenv(envName)
	if len(value) != 0 {
		return value
	}

	return ExpandEnv(defVal)
}
