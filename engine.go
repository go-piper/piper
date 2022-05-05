// Copyright (c) 2022 Vincent Chueng (coolingfall@gmail.com).

package piper

type AppEngine interface {
	// Name the name of current application engine.
	Name() string

	// Start application engine.
	Start(ctx *Context) error

	// Stop current application engine.
	Stop()
}

// EngineFunc represents a func to create AppEngine.
type EngineFunc func() AppEngine
