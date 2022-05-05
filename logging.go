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

	"github.com/coolerfall/slago"
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
	Type   string `piper:"type"`
	Layout string `piper:"layout"`
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
		slago.Logger().ResetWriter()
	}

	var config = LoggingProperty{
		Level: slago.DebugLevel.String(),
	}
	if err = ctx.Unmarshal("logging", &config); err != nil {
		return err
	}

	level := slago.ParseLevel(strings.ToUpper(config.Level))
	slago.Logger().SetLevel(level)

	writers := make(map[string]slago.Writer)
	refWriters := make([]string, 0)

	// config logging writters
	for _, w := range config.Writers {
		var writer slago.Writer

		switch w.Type {
		case "console":
			writer, err = l.makeConsoleWriter(w)
		case "file":
			writer, err = l.makeFileWriter(w)
		case "async":
			refWriters = append(refWriters, w.RefWriter)
			continue
		default:
			return errors.New("unkown slago writer")
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

		slago.Logger().AddWriter(slago.NewAsyncWriter(func(o *slago.AsyncWriterOption) {
			o.Ref = writer
		}))
	}

	// add other writers
	for _, w := range writers {
		slago.Logger().AddWriter(w)
	}

	l.initialized = true

	return nil
}

func (l *loggingSystem) makeConsoleWriter(wp WriterProperty) (slago.Writer, error) {
	var encoder slago.Encoder
	if wp.Encoder != nil {
		var err error
		encoder, err = l.makeEncoder(wp.Encoder)
		if err != nil {
			return nil, err
		}
	}

	return slago.NewConsoleWriter(func(o *slago.ConsoleWriterOption) {
		o.Encoder = encoder
	}), nil
}

func (l *loggingSystem) makeFileWriter(wp WriterProperty) (slago.Writer, error) {
	var encoder slago.Encoder
	if wp.Encoder != nil {
		var err error
		encoder, err = l.makeEncoder(wp.Encoder)
		if err != nil {
			return nil, err
		}
	}

	var rollingPolicy slago.RollingPolicy
	if wp.RollingPolicy != nil {
		policy := wp.RollingPolicy
		switch policy.Type {
		case "size-and-time-based":
			rollingPolicy = slago.NewSizeAndTimeBasedRollingPolicy(
				func(o *slago.SizeAndTimeBasedRPOption) {
					o.FilenamePattern = policy.FilenamePattern
					o.MaxFileSize = policy.MaxSize
					o.MaxHistory = policy.MaxHistory
				})
		case "time-based":
			rollingPolicy = slago.NewTimeBasedRollingPolicy(func(o *slago.TimeBasedRPOption) {
				o.FilenamePattern = policy.FilenamePattern
				o.MaxHistory = policy.MaxHistory
			})
		default:
			return nil, errors.New("unkown rolling policy")
		}
	}

	return slago.NewFileWriter(func(o *slago.FileWriterOption) {
		o.Encoder = encoder
		if len(wp.Filename) != 0 {
			o.Filename = wp.Filename
		}
		if rollingPolicy != nil {
			o.RollingPolicy = rollingPolicy
		}
	}), nil
}

func (l *loggingSystem) makeEncoder(ep *EncoderProperty) (slago.Encoder, error) {
	var encoder slago.Encoder

	switch ep.Type {
	case "json":
		encoder = slago.NewJsonEncoder()
	case "pattern":
		encoder = slago.NewPatternEncoder(func(o *slago.PatternEncoderOption) {
			o.Layout = ep.Layout
		})
	default:
		return nil, errors.New("unkown encoder for writer")
	}

	return encoder, nil
}
