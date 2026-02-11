package gslog

var FATAL = Level{"FATAL", 9999}
var ERROR = Level{"ERROR", 9000}
var WARN = Level{"WARN", 5000}
var INFO = Level{"INFO", 2000}
var DEBUG = Level{"DEBUG", 42}
var TRACE = Level{"TRACE", 10}

type Level struct {
	Label      string
	Importance int
}

func (lv Level) isLessThan(comparison Level) bool {
	return lv.Importance < comparison.Importance
}

func defaultLevels() map[Level]bool {
	m := make(map[Level]bool)
	for _, lv := range []Level{
		FATAL,
		ERROR,
		WARN,
		INFO,
		DEBUG,
		TRACE,
	} {
		m[lv] = true
	}
	return m
}

// panic if importance is not 1 > x > 9999
func NewLevel(label string, importance int) Level {
	if importance > 9999 {
		panic("maximum importance is 9999")
	}
	if importance < 1 {
		panic("minimum importance is 1")
	}
	return Level{Label: label, Importance: importance}
}
