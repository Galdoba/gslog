package gslog

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/tools/go/analysis/passes/slog"
	"google.golang.org/api/retail/v2"
)

var defaultSlogger = &logger{}
var signalShutdown bool

func init() {
	bf = &basicFormatter{}
	err := errors.New("gslog not initialized")
	defaultSlogger, err = New(DefaultConfiguration())
	if err != nil {
		panic(fmt.Errorf("gslog init failed: %w", err))
	}
<<<<<<< HEAD
	bf = &basicFormatter{}
	defaultSlogger.formatter = bf
=======
>>>>>>> 5c480458ebfc854bd55fe6395e76191ccf406b82
}

type logger struct {
	inputChannel chan logEntry
	wg           sync.WaitGroup
	writer       io.Writer
	mu           sync.Mutex
	// running      bool
	minimumLevel Level
	autofields   map[string]any
<<<<<<< HEAD
	formatter    Formatter
=======
	handlers     map[string]Handler
>>>>>>> 5c480458ebfc854bd55fe6395e76191ccf406b82
}

func New(cfg Configuration) (*logger, error) {
	l := &logger{
		inputChannel: make(chan logEntry, cfg.BufferSize),
		// running:      true,
		handlers: make(map[string]Handler),
	}
	l.handlers[""] = NewHandler(os.Stderr).WithFormatter(bf)
	for k, h := range cfg.Handlers {
		delete(l.handlers, "")
		l.handlers[k] = h
	}

	l.start(cfg.MaximumRoutines)
	return l, nil
}

func (l *logger) start(routines int) {
	fmt.Println("start")
	l.wg.Add(routines)
	go func() {
		defer l.shutdown()
		for msg := range l.inputChannel {
			l.mu.Lock()

			if err := l.processMessage(msg); err != nil {
				fmt.Println("here be event creation!", err)
			}
			if signalShutdown {
				fmt.Fprintf(os.Stderr, "forced shutdown: %v", msg.level)
				l.shutdown()
			}
			l.mu.Unlock()
		}
	}()
}

func (l *logger) processMessage(m logEntry) error {
	fmt.Println("process", m, len(l.handlers))
	processed := 0
	if len(l.handlers) == 0 {
		return fmt.Errorf("message lost: no handlers")
	}
<<<<<<< HEAD
	dto, err := l.packFields(m)
	if err != nil {
		return EntryDTO{}, fmt.Errorf("message lost: %w", err)
	}

	data, err := json.Marshal(&dto)
	if err != nil {
		return dto, fmt.Errorf("failed to marshal message: %w", err)
	}
	switch l.writer {
	case os.Stderr, os.Stdout:
		if l.formatter != nil {
			data = []byte(l.formatter.Format(dto))
		}
	}
	data = append(data, []byte("\n")...)
	if _, err := l.writer.Write(data); err != nil {
		return dto, fmt.Errorf("failed to write message: %w", err)
=======
	reasons := []string{}
	for key, h := range l.handlers {
		fmt.Println("handler", key, h)
		if err := h.Handle(m); err == nil {

			processed++
		} else {
			reasons = append(reasons, err.Error())
		}
>>>>>>> 5c480458ebfc854bd55fe6395e76191ccf406b82
	}
	fmt.Println("processed", processed)
	if processed == 0 {
		return fmt.Errorf("message lost: %v", reasons)
	}
	return nil
}

func (l *logger) shutdown() {
	// l.running = false
	close(l.inputChannel)
	time.Sleep(time.Microsecond * 100)
}

////Global Logger Functions

func Fatal(msg string, context ...any) {
	// defaultSlogger.running = true
	defaultSlogger.processMessage(NewMessage(msg, context...).WithLevel(FATAL))
}

func Error(msg string, context ...any) {
	defaultSlogger.processMessage(NewMessage(msg, context...).WithLevel(ERROR))
}

func Warn(msg string, context ...any) {
	defaultSlogger.processMessage(NewMessage(msg, context...).WithLevel(WARN))
}

func Info(msg string, context ...any) {
	// defaultSlogger.running = true
	defaultSlogger.processMessage(NewMessage(msg, context...).WithLevel(INFO))
}

func Debug(msg string, context ...any) {
	defaultSlogger.processMessage(NewMessage(msg, context...).WithLevel(DEBUG))
}

func Trace(msg string, context ...any) {
	defaultSlogger.processMessage(NewMessage(msg, context...).WithLevel(TRACE))
}

////Global logger control functions

func SetWriter(w io.Writer) {
	if _, ok := defaultSlogger.handlers[""]; ok {
		defaultSlogger.handlers[""] = NewHandler(w)
	}
}

func (l *logger) Fatal(msg string, context ...any) {
	m := NewMessage(msg, context...).WithLevel(FATAL)
	l.processMessage(m)
	panic(m.message)
}

func (l *logger) Error(msg string, context ...any) {
	m := NewMessage(msg, context...).WithLevel(ERROR)
	l.processMessage(m)
}
func (l *logger) Warn(msg string, context ...any) {
	m := NewMessage(msg, context...).WithLevel(WARN)
	l.processMessage(m)
	panic(m)
}

func (l *logger) Info(msg string, context ...any) {
	m := NewMessage(msg, context...).WithLevel(INFO)
	l.processMessage(m)
	panic(m)
}

func (l *logger) Debug(msg string, context ...any) {
	m := NewMessage(msg, context...).WithLevel(DEBUG)
	l.processMessage(m)
	panic(m)
}

func (l *logger) Trace(msg string, context ...any) {
	m := NewMessage(msg, context...).WithLevel(TRACE)
	l.processMessage(m)
	panic(m)
}

func (l *logger) Errorf(format string, args ...any) error {
	err := errors.New(fmt.Sprintf(format, args...))
	m := NewMessage(err.Error()).WithLevel(ERROR)
	if loggErr := l.processMessage(m); loggErr != nil {
		return fmt.Errorf("failed to process error: %w (error=%v)", loggErr, err)
	}
	return err
}

func SetFormatter(f Formatter) {
	defaultSlogger.formatter = f
}
