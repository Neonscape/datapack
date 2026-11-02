// Package perftrace contains experiment-only coarse spans, disabled by default.
package perftrace

import (
	"sync"
	"time"
)

type Span struct {
	Name    string  `json:"name"`
	Seconds float64 `json:"seconds"`
}

var mu sync.Mutex
var enabled bool
var spans []Span

func Enable()          { mu.Lock(); enabled = true; spans = nil; mu.Unlock() }
func Reset()           { mu.Lock(); spans = nil; mu.Unlock() }
func Snapshot() []Span { mu.Lock(); defer mu.Unlock(); return append([]Span(nil), spans...) }
func Start(name string) func() {
	if !enabled {
		return func() {}
	}
	start := time.Now()
	return func() { mu.Lock(); spans = append(spans, Span{name, time.Since(start).Seconds()}); mu.Unlock() }
}
