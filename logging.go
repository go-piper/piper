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
	"sync"

	"github.com/coolerfall/lork"
)

var (
	onceLogging sync.Once
	logging     *loggingSystem
)

type loggingSystem struct {
	initialized bool
}

// LoggingProperty defines the property of piper.logging section in yaml config.
type LoggingProperty struct {
	Level   string           `piper:"level"`
	Writers []WriterProperty `piper:"writers"`
}

type WriterProperty struct {
	Name          string                 `piper:"name"`
	RefWriter     string                 `piper:"ref-writer"`
	Type          string                 `piper:"type"`
	Filename      string                 `piper:"filename"`
	Encoder       *EncoderProperty       `piper:"encoder"`
	RollingPolicy *RollingPolicyProperty `piper:"rolling-policy"`
}

type EncoderProperty struct {
	Type    string `piper:"type"`
	Pattern string `piper:"pattern"`
}

type RollingPolicyProperty struct {
	Type            string `piper:"type"`
	FilenamePattern string `piper:"filename-pattern"`
	MaxSize         string `piper:"max-size"`
	MaxHistory      int    `piper:"max-history"`
}

// LoggingSystem gets global logging system to configure.
func LoggingSystem() *loggingSystem {
	onceLogging.Do(func() {
		logging = &loggingSystem{}
	})

	return logging
}

// Initialize logging system with console writer and debug level.
func (l *loggingSystem) Initialize(ctx *Context) (err error) {
	if l.initialized {
		lork.Logger().ResetWriter()
	}

	var config = LoggingProperty{
		Level: lork.DebugLevel.String(),
		Writers: []WriterProperty{
			{Type: "console"},
		},
	}
	if err = ctx.Unmarshal("logging", &config); err != nil {
		return err
	}

	level := lork.ParseLevel(strings.ToUpper(config.Level))
	lork.Logger().SetLevel(level)

	writers := make(map[string]lork.Writer)
	refWriters := make([]string, 0)

	// config logging writters
	for _, w := range config.Writers {
		var writer lork.Writer

		switch w.Type {
		case "console":
			writer, err = l.makeConsoleWriter(w)
		case "file":
			writer, err = l.makeFileWriter(w)
		case "async":
			refWriters = append(refWriters, w.RefWriter)
			continue
		default:
			return errors.New("unkown lork writer")
		}

		if err != nil {
			return err
		}
		writers[w.Name] = writer
	}

	// add asynchronous writers
	for _, name := range refWriters {
		writer, ok := writers[name]
		if !ok {
			return errors.New("no ref writer found: " + name)
		}
		delete(writers, name)

		lork.Logger().AddWriter(lork.NewAsyncWriter(func(o *lork.AsyncWriterOption) {
			o.Ref = writer
		}))
	}

	// add other writers
	for _, w := range writers {
		lork.Logger().AddWriter(w)
	}

	l.initialized = true

	return nil
}

func (l *loggingSystem) makeConsoleWriter(wp WriterProperty) (lork.Writer, error) {
	var encoder lork.Encoder
	if wp.Encoder != nil {
		encoder = l.makeEncoder(wp.Encoder)
	}

	return lork.NewConsoleWriter(func(o *lork.ConsoleWriterOption) {
		o.Encoder = encoder
	}), nil
}

func (l *loggingSystem) makeFileWriter(wp WriterProperty) (lork.Writer, error) {
	var encoder lork.Encoder
	if wp.Encoder != nil {
		encoder = l.makeEncoder(wp.Encoder)
	}

	var rollingPolicy lork.RollingPolicy
	if wp.RollingPolicy != nil {
		policy := wp.RollingPolicy
		switch policy.Type {
		case "size-and-time-based":
			rollingPolicy = lork.NewSizeAndTimeBasedRollingPolicy(
				func(o *lork.SizeAndTimeBasedRPOption) {
					o.FilenamePattern = policy.FilenamePattern
					o.MaxFileSize = policy.MaxSize
					o.MaxHistory = policy.MaxHistory
				})
		case "time-based":
			rollingPolicy = lork.NewTimeBasedRollingPolicy(func(o *lork.TimeBasedRPOption) {
				o.FilenamePattern = policy.FilenamePattern
				o.MaxHistory = policy.MaxHistory
			})
		default:
			return nil, errors.New("unkown rolling policy")
		}
	}

	return lork.NewFileWriter(func(o *lork.FileWriterOption) {
		o.Encoder = encoder
		if len(wp.Filename) != 0 {
			o.Filename = wp.Filename
		}
		if rollingPolicy != nil {
			o.RollingPolicy = rollingPolicy
		}
	}), nil
}

func (l *loggingSystem) makeEncoder(ep *EncoderProperty) lork.Encoder {
	var encoder lork.Encoder

	switch ep.Type {
	case "json":
		encoder = lork.NewJsonEncoder()
	case "pattern":
		encoder = lork.NewPatternEncoder(func(o *lork.PatternEncoderOption) {
			o.Pattern = ep.Pattern
		})
	default:
		encoder = lork.NewPatternEncoder()
	}

	return encoder
}
