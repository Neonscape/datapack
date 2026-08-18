package exporters

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"

	"datapack/internal/compiler"
	"datapack/internal/exporters/formats/csv"
)

//go:embed formats/*/manifest.json
var builtinManifests embed.FS

type Format struct {
	ID              string            `json:"id"`
	Labels          map[string]string `json:"labels"`
	SupportsColumns bool              `json:"supportsColumns"`
	Extensions      []string          `json:"extensions,omitempty"`
}

type Registry struct {
	formats  map[string]Format
	handlers map[string]handler
}

type Result struct {
	Files []string `json:"files"`
}

type handler func(destination string, pkg *compiler.Package) (Result, error)

func Discover() (Registry, error) {
	registry := Registry{formats: map[string]Format{}, handlers: map[string]handler{
		"csv": func(destination string, pkg *compiler.Package) (Result, error) {
			result, err := csv.Export(destination, pkg)
			return Result{Files: result.Files}, err
		},
	}}
	entries, err := fs.Glob(builtinManifests, "formats/*/manifest.json")
	if err != nil {
		return Registry{}, fmt.Errorf("scan exporter manifests: %w", err)
	}
	for _, entry := range entries {
		data, err := builtinManifests.ReadFile(entry)
		if err != nil {
			return Registry{}, fmt.Errorf("read exporter manifest %q: %w", entry, err)
		}
		var format Format
		if err := json.Unmarshal(data, &format); err != nil {
			return Registry{}, fmt.Errorf("decode exporter manifest %q: %w", entry, err)
		}
		if format.ID == "" || len(format.Labels) == 0 {
			return Registry{}, fmt.Errorf("exporter manifest %q is missing id or labels", entry)
		}
		if _, exists := registry.formats[format.ID]; exists {
			return Registry{}, fmt.Errorf("duplicate exporter id %q", format.ID)
		}
		if _, exists := registry.handlers[format.ID]; !exists {
			return Registry{}, fmt.Errorf("exporter manifest %q has no backend implementation", format.ID)
		}
		registry.formats[format.ID] = format
	}
	return registry, nil
}

func (r Registry) Formats() []Format {
	formats := make([]Format, 0, len(r.formats))
	for _, format := range r.formats {
		formats = append(formats, format)
	}
	sort.Slice(formats, func(i, j int) bool { return formats[i].ID < formats[j].ID })
	return formats
}

func (r Registry) Has(id string) bool {
	_, ok := r.formats[id]
	return ok
}

func (r Registry) Format(id string) (Format, bool) {
	format, ok := r.formats[id]
	return format, ok
}

func (r Registry) Export(id, destination string, pkg *compiler.Package) (Result, error) {
	if !r.Has(id) {
		return Result{}, fmt.Errorf("unknown exporter format %q", id)
	}
	return r.handlers[id](destination, pkg)
}
