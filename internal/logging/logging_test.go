package logging

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLoggerWritesJSONAndSinks(t *testing.T) {
	var output bytes.Buffer
	logger := New(&output)
	var received Entry
	logger.AddSink(func(entry Entry) { received = entry })
	logger.Info("opened project", map[string]any{"tables": 2})
	var encoded Entry
	if err := json.Unmarshal(output.Bytes(), &encoded); err != nil {
		t.Fatal(err)
	}
	if encoded.Level != Info || encoded.Message != "opened project" || encoded.Fields["tables"] != float64(2) || received.Message != encoded.Message {
		t.Fatalf("entries = %#v, %#v", encoded, received)
	}
}

// The callback completes even when it registers a subsequent listener. The
// bounded child returns after a failing callback, so the regression cannot hang.
func TestSinkCanRegisterAnotherSinkWithoutHoldingLoggerLock(t *testing.T) {
	logger := New(nil)
	done := make(chan struct{})
	var blocked bool
	logger.AddSink(func(Entry) {
		go func() { logger.AddSink(func(Entry) {}); close(done) }()
		select {
		case <-done:
		case <-time.After(200 * time.Millisecond):
			blocked = true
		}
	})
	logger.Info("one", nil)
	<-done
	if blocked {
		t.Fatal("sink registration blocked behind callback's logger lock")
	}
}

func TestSinkSnapshotAndReentrantLog(t *testing.T) {
	var output bytes.Buffer
	logger := New(&output)
	var second []string
	logger.AddSink(func(entry Entry) {
		if entry.Message == "outer" {
			logger.AddSink(func(e Entry) { second = append(second, e.Message) })
			logger.Info("inner", nil)
		}
	})
	logger.Info("outer", nil)
	if !reflect.DeepEqual(second, []string{"inner"}) {
		t.Fatalf("new sink must observe only later entries: %v", second)
	}
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 2 {
		t.Fatal(output.String())
	}
	for _, line := range lines {
		var e Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatal(err)
		}
	}
}
func TestConcurrentLogsKeepWriterLinesWhole(t *testing.T) {
	var output bytes.Buffer
	logger := New(&output)
	var group sync.WaitGroup
	var count atomic.Int64
	logger.AddSink(func(Entry) { count.Add(1) })
	for range 8 {
		group.Go(func() {
			for range 10 {
				logger.Info("entry", map[string]any{"count": 1})
			}
		})
	}
	group.Wait()
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 80 || count.Load() != 80 {
		t.Fatalf("lines/sinks: %d/%d", len(lines), count.Load())
	}
	for _, line := range lines {
		var e Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatal(err)
		}
	}
}
