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
)

// Piper defines a piper application.
type Piper struct {
	cmdLine *cmdLine
}

type Option struct {
	ShowVersion bool
	Description string
	Banner      Banner
	EngineFunc  EngineFunc
	ResourceFs  embed.FS
}

func init() {
	Wire(&ApplicationProperty{})
}

// NewPiper creates a new piper to run application.
func NewPiper(options ...func(*Option)) *Piper {
	opt := &Option{}

	for _, f := range options {
		f(opt)
	}

	cli := newCmdLine(newAppEnv(opt.ResourceFs), NotNil(opt.EngineFunc, "EngineFunc is nil"),
		NotEmpty(opt.Description))
	banner := opt.Banner
	if banner == nil {
		banner = NewDefaultBanner()
	}
	cli.Init(banner)

	return &Piper{
		cmdLine: cli,
	}
}

func (p *Piper) Run() {
	p.cmdLine.Execute()
}

func (p *Piper) Execute() error {
	return p.cmdLine.Execute()
}
