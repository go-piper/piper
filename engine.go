// Copyright (c) 2022 Vincent Chueng (coolingfall@gmail.com).

package piper

type AppEngine interface {
	// Name of current server engine.
	Name() string

	// Start application engine.
	Start(env *AppEnv) error

	// Stop current application engine.
	Stop()
}

// EngineFunc represents a func to create AppEngine.
type EngineFunc func() AppEngine
