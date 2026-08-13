// Package logging provides small structured application logging without making
// business packages depend on terminal, file, or GUI details.
package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Level string

const (
	Debug Level = "debug"
	Info  Level = "info"
	Warn  Level = "warn"
	Error Level = "error"
)

type Entry struct {
	Time    time.Time      `json:"time"`
	Level   Level          `json:"level"`
	Message string         `json:"message"`
	Fields  map[string]any `json:"fields,omitempty"`
}

// Sink receives an immutable log entry. A sink must not block application work.
type Sink func(Entry)

type Logger struct {
	mu     sync.RWMutex
	writer io.Writer
	sinks  []Sink
}

func New(writer io.Writer) *Logger {
	return &Logger{writer: writer}
}

// NewUserLogger writes line-oriented JSON entries to the user configuration
// directory and mirrors readable output to stderr.
func NewUserLogger(configRoot string) (*Logger, string, error) {
	directory := filepath.Join(configRoot, "datapack", "logs")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, "", fmt.Errorf("create log directory: %w", err)
	}
	path := filepath.Join(directory, "datapack.log")
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, "", fmt.Errorf("open log file: %w", err)
	}
	return New(io.MultiWriter(os.Stderr, file)), path, nil
}

func (l *Logger) AddSink(sink Sink) {
	if sink == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.sinks = append(l.sinks, sink)
}

func (l *Logger) Log(level Level, message string, fields map[string]any) {
	entry := Entry{Time: time.Now().UTC(), Level: level, Message: message, Fields: fields}
	l.mu.Lock()
	defer l.mu.Unlock()
	writer := l.writer
	sinks := append([]Sink(nil), l.sinks...)
	if writer != nil {
		encoded, err := json.Marshal(entry)
		if err == nil {
			_, _ = fmt.Fprintln(writer, string(encoded))
		}
	}
	for _, sink := range sinks {
		sink(entry)
	}
}

func (l *Logger) Debug(message string, fields map[string]any) { l.Log(Debug, message, fields) }
func (l *Logger) Info(message string, fields map[string]any)  { l.Log(Info, message, fields) }
func (l *Logger) Warn(message string, fields map[string]any)  { l.Log(Warn, message, fields) }
func (l *Logger) Error(message string, fields map[string]any) { l.Log(Error, message, fields) }
