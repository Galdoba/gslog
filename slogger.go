package gslog

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

var defaultSlogger = &logger{}

func init() {
	err := errors.New("gslog not initialized")
	defaultSlogger, err = New(DefaultConfiguration())
	if err != nil {
		panic(fmt.Errorf("gslog init failed: %w", err))
	}
	bf = &basicFormatter{}
}

type logger struct {
	inputChannel chan logEntry
	wg           sync.WaitGroup
	writer       io.Writer
	mu           sync.Mutex
	running      bool
	minimumLevel Level
	autofields   map[string]any
}

func New(cfg Configuration) (*logger, error) {
	l := &logger{
		inputChannel: make(chan logEntry, cfg.BufferSize),
		writer:       cfg.Writer,
		running:      true,
		minimumLevel: cfg.MinimumLevel,
	}
	l.start(cfg.MaximumRoutines)
	return l, nil
}

func (l *logger) start(routines int) {
	l.wg.Add(routines)
	go func() {
		defer l.shutdown()
		for msg := range l.inputChannel {
			l.mu.Lock()

			if _, err := l.handleMessage(msg); err != nil {
				fmt.Println("here be event creation!", err)
			}

			l.mu.Unlock()
		}
	}()
}

func (l *logger) handleMessage(m logEntry) (EntryDTO, error) {
	if m.level.isLessThan(l.minimumLevel) {
		return EntryDTO{}, nil
	}
	if l.writer == nil {
		return EntryDTO{}, fmt.Errorf("message lost: writer is not set")
	}
	dto, err := l.packFields(m)
	if err != nil {
		return EntryDTO{}, fmt.Errorf("message lost: %w", err)
	}

	data, err := json.Marshal(&dto)
	if err != nil {
		return dto, fmt.Errorf("failed to marshal message: %w", err)
	}
	data = append(data, []byte("\n")...)
	if _, err := l.writer.Write(data); err != nil {
		return dto, fmt.Errorf("failed to write message: %w", err)
	}
	if m.level.Importance >= FATAL.Importance {
		fmt.Fprintf(os.Stderr, "It's over 9000!!!\n")
		panic(bf.Format(dto))
	}
	return dto, nil
}

func (l *logger) shutdown() {
	l.running = false
	close(l.inputChannel)
	time.Sleep(time.Microsecond * 100)
}

////Global Logger Functions

func Fatal(msg string, context ...any) {
	defaultSlogger.handleMessage(NewMessage(msg, context...).WithLevel(FATAL))
}

func Error(msg string, context ...any) {
	defaultSlogger.handleMessage(NewMessage(msg, context...).WithLevel(ERROR))
}

func Warn(msg string, context ...any) {
	defaultSlogger.handleMessage(NewMessage(msg, context...).WithLevel(WARN))
}

func Info(msg string, context ...any) {
	defaultSlogger.handleMessage(NewMessage(msg, context...).WithLevel(INFO))
}

func Debug(msg string, context ...any) {
	defaultSlogger.handleMessage(NewMessage(msg, context...).WithLevel(DEBUG))
}

func Trace(msg string, context ...any) {
	defaultSlogger.handleMessage(NewMessage(msg, context...).WithLevel(TRACE))
}

////Global logger control functions

func SetWriter(w io.Writer) {
	defaultSlogger.writer = w
}
