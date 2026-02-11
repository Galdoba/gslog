package gslog

import (
	"io"
	"os"
)

type Configuration struct {
	BufferSize            int
	Writer                io.Writer
	Levels                map[Level]bool
	MinimumLevel          Level
	RoutineTimeoutSeconds int
	MaximumRoutines       int
	ConditionalWriters    map[string]ConditionFunc
}

func DefaultConfiguration() Configuration {
	return Configuration{
		BufferSize:            1024,
		Writer:                os.Stderr,
		Levels:                defaultLevels(),
		MinimumLevel:          INFO,
		RoutineTimeoutSeconds: 30,
		MaximumRoutines:       1,
	}
}

type ConditionFunc func(string) bool
