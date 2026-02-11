package gslog

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

type Formatter interface {
	Format(EntryDTO) string
}

var bf = &basicFormatter{}

type basicFormatter struct{}

func (bf *basicFormatter) Format(e EntryDTO) string {
	if e.Message == "" && len(e.Context) == 0 {
		return ""
	}
	s := e.Time.Format(time.DateTime)
	s += fmt.Sprintf(" [%v]", e.Level)
	s += fmt.Sprintf(" %v", e.Message)
	keys := []string{}
	hasCtx := len(e.Context) > 0
	for k := range e.Context {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	if hasCtx {
		s += ": "
	}
	for _, key := range keys {
		s += fmt.Sprintf("%v=%v; ", key, e.Context[key])
	}
	s = strings.TrimSuffix(s, "; ")
	if hasCtx {
		s += ""
	}
	return s
}
