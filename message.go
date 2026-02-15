package gslog

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

const (
	fl_message   = "message"
	fl_timestamp = "timestamp"
	fl_level     = "level"
	fl_file      = "file"
	fl_funcName  = "func"
	fl_lineNum   = "line"
	fl_tags      = "tags"
	fl_sensitive = "sensitive"
)

type logEntry struct {
	message   string
	timestamp time.Time
	level     Level
	file      string
	funcName  string
	lineNum   int
	tags      []string
	sensitive map[string]int
	context   map[string]any
}

func packFields(m logEntry) (EntryDTO, error) {
	if m.context == nil {
		m.context = make(map[string]any)
	}
	if m.level.Importance <= DEBUG.Importance {
		pc, file, line, ok := runtime.Caller(3)
		if ok {
			m.file = file
			m.lineNum = line
			m.funcName = runtime.FuncForPC(pc).Name()
		}
		m.context[fl_file] = m.file
		m.context[fl_funcName] = m.funcName
		m.context[fl_lineNum] = m.lineNum
	}
	for key, paranoia := range m.sensitive {
		if key == "" {
			return EntryDTO{}, fmt.Errorf("empty key provided as sensetive")
		}
		if val, ok := m.context[key]; !ok {
			return EntryDTO{}, fmt.Errorf("key %v is set as sensetive, but was not found in fields", key)
		} else {
			m.context[key] = obfuscate(val, paranoia)
		}
	}
	e := EntryDTO{
		Message: m.message,
		Time:    m.timestamp,
		Level:   m.level.Label,
		Context: m.context,
	}

	return e, nil
}

func newMessageSafe(text string, args ...any) (logEntry, error) {
	m := logEntry{}
	m.context = make(map[string]any)
	m.timestamp = time.Now()
	m.message = text
	key := ""
	for i, arg := range args {
		switch i % 2 {
		case 0:
			switch arg := arg.(type) {
			case string:
				key = arg
			default:
				return logEntry{}, fmt.Errorf("expect string as a key on even positions of arguments")
			}
		case 1:
			m.context[key] = arg
		}
	}
	return m, nil
}

func mustNewMessage(text string, args ...any) logEntry {
	m, err := newMessageSafe(text, args...)
	if err != nil {
		panic(err)
	}
	return m
}

func NewMessage(text string, fields ...any) logEntry {
	return mustNewMessage(text, fields...)
}

func NewMessageSafe(text string, fields ...any) (logEntry, error) {
	return newMessageSafe(text, fields...)
}

func (m logEntry) WithLevel(level Level) logEntry {
	m.level = level
	return m
}

func (m logEntry) WithTags(tags ...string) logEntry {
	m.tags = tags
	return m
}

func (m logEntry) WithSensitive(key string, paranoia int) logEntry {
	if m.sensitive == nil {
		m.sensitive = make(map[string]int)
	}
	m.sensitive[key] = paranoia
	return m
}

func obfuscate(val any, paranoia int) string {
	obfs := ""
	for range strings.SplitSeq(fmt.Sprintf("%v", val), "") {
		obfs += "?"
	}
	return obfs
}

func AutoField(key string, val any, valueFunc ...func(any) any) (string, any) {
	if len(valueFunc) < 1 {
		return key, val
	}
	return key, valueFunc[0](val)
}

type EntryDTO struct {
	Time    time.Time      `json:"time"`
	Level   string         `json:"level"`
	Message string         `json:"message"`
	Context map[string]any `json:"context,omitempty"`
}
