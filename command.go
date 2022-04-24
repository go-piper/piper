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
	"os/signal"
	"reflect"
	"runtime"

	"github.com/spf13/cobra"
)

type cmdLine struct {
	rootCmd    *cobra.Command
	env        *AppEnv
	banner     Banner
	engineFunc EngineFunc
}

func newCmdLine(env *AppEnv, engineFunc EngineFunc, shortDesc string) *cmdLine {
	rootCmd := &cobra.Command{
		Use:           env.cmdName(),
		Short:         shortDesc,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	return &cmdLine{
		rootCmd:    rootCmd,
		env:        env,
		engineFunc: engineFunc,
	}
}

func (c *cmdLine) Init(banner Banner) {
	c.banner = banner

	// create start command
	startCmd := c.newStartCmd()
	startCmd.Flags().StringP(keyProfile, "p", "", "the profile to set")
	err := c.env.vp.BindPFlag(keyProfile, startCmd.Flags().Lookup(keyProfile))
	if err != nil {
		Panicf("initialize command line error %v", err)
	}

	stdOut := c.rootCmd.OutOrStdout()
	c.rootCmd.SetOut(stdOut)
	c.rootCmd.SetErr(stdOut)

	c.rootCmd.AddCommand(c.newVersionCmd(), startCmd)
}

// Execute executes the root command which will start the aplication.
func (c *cmdLine) Execute() error {
	return c.rootCmd.Execute()
}

func (c *cmdLine) newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start to run application",
		RunE: func(_ *cobra.Command, _ []string) error {
			return c.run()
		},
	}
}

func (c *cmdLine) newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: fmt.Sprintf("Show the version of %v", c.env.cmdName()),
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf(""+
				"Go version:        %v\n"+
				"Version:           %v\n"+
				"Git commit:        %v\n"+
				"Build time:             %v\n",
				runtime.Version(), Version, GitCommit, BuildTime)
		},
	}
}

func (c *cmdLine) run() error {
	c.banner.Print()

	if err := _depTree().resolveDependencies(); err != nil {
		return err
	}

	engine := c.engineFunc()

	for _, l := range Retrieve[StartListener](reflect.TypeOf((*StartListener)(nil))) {
		l.OnAppStart()
	}

	c.captureExit(func() {
		for _, l := range Retrieve[StopListener](reflect.TypeOf((*StopListener)(nil))) {
			l.OnAppStop()
			engine.Stop()
		}
	})

	return engine.Start(c.env)
}

func (c *cmdLine) captureExit(stop func()) {
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, os.Kill)
		// block until a signal is received
		<-sig
		stop()
		fmt.Println("process exited with code -1")
		os.Exit(-1)
	}()
}
