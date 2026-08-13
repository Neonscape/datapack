package logging

import (
	"bytes"
	"encoding/json"
	"testing"
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
