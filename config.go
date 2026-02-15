package gslog

import (
	"io"
)

type Configuration struct {
	BufferSize            int
	Writer                io.Writer
	Levels                map[Level]bool
	MinimumLevel          Level
	RoutineTimeoutSeconds int
	MaximumRoutines       int
	ConditionalWriters    map[string]ConditionFunc
	Handlers              map[string]Handler
}

func DefaultConfiguration() Configuration {
	cfg := Configuration{
		BufferSize:            1024,
		Levels:                defaultLevels(),
		MinimumLevel:          INFO,
		RoutineTimeoutSeconds: 30,
		MaximumRoutines:       1,
	}
	cfg.Handlers = make(map[string]Handler)
	return cfg
}

type ConditionFunc func(string) bool
