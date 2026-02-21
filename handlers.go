package gslog

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
)

type messageHandler struct {
	writter       io.Writer
	formatter     Formatter
	minLevel      Level
	maxLevel      Level
	tagsRequired  []string
	tagsForbidden []string
	autoFields    map[string]any
}

type Handler interface {
	Handle(logEntry) error
	ForceShutdown()
}

func NewHandler(w io.Writer) *messageHandler {
	mh := messageHandler{
		writter:  w,
		minLevel: INFO,
		maxLevel: FATAL,
	}
	switch mh.writter {
	case os.Stderr, os.Stdout:
		mh.formatter = bf
	}
	return &mh
}

func (mh *messageHandler) WithMinimumLevel(lv Level) *messageHandler {
	mh.minLevel = lv
	return mh
}

func (mh *messageHandler) WithMaximumLevel(lv Level) *messageHandler {
	mh.maxLevel = lv
	return mh
}

func (mh *messageHandler) WithTagsRequired(tags ...string) *messageHandler {
	mh.tagsRequired = tags
	return mh
}

func (mh *messageHandler) WithTagsForbidden(tags ...string) *messageHandler {
	mh.tagsForbidden = tags
	return mh
}

func (mh *messageHandler) WithFormatter(f Formatter) *messageHandler {
	mh.formatter = f
	return mh
}

func (mh *messageHandler) WithAutoFields(fields map[string]any) *messageHandler {
	if mh.autoFields == nil {
		mh.autoFields = make(map[string]any)
	}
	mh.autoFields = fields
	return mh
}

func (mh *messageHandler) Validate() error {
	if mh.writter == nil {
		return fmt.Errorf("writer was not set")
	}
	return nil
}

func (mh *messageHandler) Handle(m logEntry) error {
	fmt.Println("start handler")
	if mh.writter == nil {
		return fmt.Errorf("no writer")
	}
	if m.level.isLessThan(mh.minLevel) {
		return fmt.Errorf("below miminum level")
	}
	if mh.maxLevel.isLessThan(m.level) {
		return fmt.Errorf("above maixmum level")
	}
	maps.Copy(m.context, mh.autoFields)

	dto, err := packFields(m)
	if err != nil {
		return fmt.Errorf("message lost: %w", err)
	}

	data, err := json.Marshal(&dto)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	switch mh.formatter {
	case nil:
		data = append(data, []byte("\n")...)
	default:
		data = []byte(mh.formatter.Format(dto) + "\n")
	}
	if _, err := mh.writter.Write(data); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	if m.level.Importance >= FATAL.Importance {
		mh.ForceShutdown()
	}
	return nil
}

func (mh *messageHandler) ForceShutdown() {
	signalShutdown = true
}
