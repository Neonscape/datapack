package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"datapack/internal/document"
	"datapack/internal/perftrace"
	"datapack/internal/workspace"
)

type editAudit struct {
	Method     string `json:"method"`
	RowID      string `json:"row_id"`
	Column     string `json:"column"`
	Expression string `json:"expression"`
}

var activeEdit *editAudit

type record struct {
	Stage        string           `json:"stage"`
	Iteration    int              `json:"iteration"`
	APISeconds   float64          `json:"api_seconds"`
	JSONSeconds  float64          `json:"json_seconds"`
	TotalSeconds float64          `json:"service_plus_json_seconds"`
	JSONBytes    int              `json:"json_bytes"`
	HeapAlloc    uint64           `json:"heap_alloc_bytes"`
	RSSKiB       string           `json:"process_rss_status"`
	SourceBytes  int64            `json:"source_bytes"`
	SourceSHA256 string           `json:"source_sha256"`
	Spans        []perftrace.Span `json:"nested_spans"`
	Edit         *editAudit       `json:"edit,omitempty"`
}

var output string
var source string

func run(stage string, i int, cpu bool, fn func() (any, error)) {
	fmt.Fprintf(os.Stderr, "START %s %d\n", stage, i)
	perftrace.Reset()
	var profile *os.File
	if cpu {
		profile, _ = os.Create(filepath.Join(output, stage+".cpu.pprof"))
		if profile != nil {
			_ = pprof.StartCPUProfile(profile)
		}
	}
	start := time.Now()
	result, err := fn()
	api := time.Since(start).Seconds()
	if err != nil {
		panic(fmt.Sprintf("%s: %v", stage, err))
	}
	encodeStart := time.Now()
	encoded, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}
	encode := time.Since(encodeStart).Seconds()
	total := time.Since(start).Seconds()
	if cpu && profile != nil {
		pprof.StopCPUProfile()
		profile.Close()
	}
	if stage == "edit_manual" && i == 1 {
		if err := os.WriteFile(filepath.Join(output, "full-edit-response.json"), encoded, 0600); err != nil {
			panic(err)
		}
	}
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	status, _ := os.ReadFile("/proc/self/status")
	var rss []string
	for _, line := range strings.Split(string(status), "\n") {
		if strings.HasPrefix(line, "VmRSS:") || strings.HasPrefix(line, "VmHWM:") {
			rss = append(rss, strings.TrimSpace(line))
		}
	}
	raw, err := os.ReadFile(source)
	if err != nil {
		panic(err)
	}
	digest := sha256.Sum256(raw)
	r := record{stage, i, api, encode, total, len(encoded), memory.HeapAlloc, strings.Join(rss, "; "), int64(len(raw)), hex.EncodeToString(digest[:]), perftrace.Snapshot(), activeEdit}
	f, err := os.OpenFile(filepath.Join(output, "measurements.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	_ = json.NewEncoder(f).Encode(r)
	f.Close()
	fmt.Fprintf(os.Stderr, "DONE %s %d %.3fs + JSON %.3fs (%d bytes)\n", stage, i, api, encode, len(encoded))
	activeEdit = nil
	runtime.GC()
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		_, err = io.Copy(out, in)
		ce := out.Close()
		if err != nil {
			return err
		}
		return ce
	})
}
func checkTable(result workspace.CellUpdateResult, rowID, col, expression string) error {
	if len(result.Table.Rows) != 100000 {
		return fmt.Errorf("wrong row count %d", len(result.Table.Rows))
	}
	for _, row := range result.Table.Rows {
		if row.ID == rowID {
			if row.Values[col].Expression != expression {
				return fmt.Errorf("wrong saved expression %s", row.Values[col].Expression)
			}
			return nil
		}
	}
	return fmt.Errorf("missing edited row")
}
func main() {
	fixture := flag.String("fixture", "", "original fixture project directory")
	out := flag.String("out", "", "new output directory")
	repeats := flag.Int("repeats", 3, "number of observations per phase")
	scenario := flag.String("scenario", "full", "full, edit, open, burst, cold, or saveas")
	openCount := flag.Int("opens", 0, "override open observations; default repeats")
	flag.Parse()
	if *fixture == "" || *out == "" {
		panic("fixture and out required")
	}
	output = *out
	if err := os.MkdirAll(output, 0755); err != nil {
		panic(err)
	}
	if _, err := os.Stat(filepath.Join(output, "measurements.jsonl")); err == nil {
		panic("refuse to overwrite existing run")
	}
	project := filepath.Join(output, "working-project")
	if err := copyTree(*fixture, project); err != nil {
		panic(err)
	}
	projectPath, err := filepath.Abs(filepath.Join(project, "stress-100k.dtproj"))
	if err != nil {
		panic(err)
	}
	source = filepath.Join(filepath.Dir(projectPath), "tables", "items", "items.cue")
	config := filepath.Join(output, "isolated-settings")
	settings := workspace.DefaultEditorSettings()
	settings.Build.Mode = "manual"
	if err := workspace.SaveEditorSettings(config, settings); err != nil {
		panic(err)
	}
	perftrace.Enable()
	var service *workspace.Service
	opens := *repeats
	if *openCount > 0 {
		opens = *openCount
	}
	for i := 1; i <= opens; i++ {
		service = workspace.NewService(config)
		run("open_manual", i, false, func() (any, error) {
			v, e := service.OpenProject(projectPath)
			if e == nil && len(v.Compile.Diagnostics) != 0 {
				e = fmt.Errorf("open diagnostics %v", v.Compile.Diagnostics)
			}
			return v, e
		})
	}
	if *scenario == "open" {
		return
	}
	if *scenario == "saveas" {
		for i := 1; i <= *repeats; i++ {
			target, err := filepath.Abs(filepath.Join(output, fmt.Sprintf("save-as-%d", i), "stress-100k.dtproj"))
			if err != nil {
				panic(err)
			}
			run("save_project_as", i, false, func() (any, error) {
				v, e := service.SaveProjectAs(target)
				if e == nil {
					source = filepath.Join(filepath.Dir(target), "tables", "items", "items.cue")
				}
				return v, e
			})
		}
		return
	}
	if *scenario == "burst" || *scenario == "cold" {
		run("get_manual", 1, false, func() (any, error) { return service.GetTable(source) })
		ids := []string{"item_000001", "item_050000", "item_100000"}
		for i := 1; i <= *repeats; i++ {
			id := ids[(i-1)%3]
			col, expr := "price", fmt.Sprint(900000+i)
			text := false
			if *scenario == "burst" {
				switch i % 4 {
				case 1:
					col, expr = "name", fmt.Sprintf("编辑样本%03d, \"引号\"", i)
					text = true
				case 2:
					col, expr = "enabled", fmt.Sprint(i%8 == 2)
				case 3:
					col, expr = "category", `"potion"`
				}
			}
			method := "UpdateCell"
			if text {
				method = "SetCellText"
			}
			activeEdit = &editAudit{method, id, col, expr}
			run("edit_sequence", i, false, func() (any, error) {
				var r workspace.CellUpdateResult
				var e error
				expected := expr
				if text {
					r, e = service.SetCellText(workspace.CellUpdate{source, id, col, expr})
					expected = document.EncodeStringExpression(expr)
				} else {
					r, e = service.UpdateCell(workspace.CellUpdate{source, id, col, expr})
				}
				if e == nil {
					e = checkTable(r, id, col, expected)
				}
				return r, e
			})
			if *scenario == "cold" {
				run("compile_after_edit", i, false, func() (any, error) {
					r, e := service.Compile()
					if e == nil && len(r.Workspace.Compile.Diagnostics) > 0 {
						e = fmt.Errorf("compile diagnostics")
					}
					return r, e
				})
				activeEdit = &editAudit{"UpdateCell", id, "stock", fmt.Sprint(i + 100)}
				run("edit_before_export", i, false, func() (any, error) {
					return service.UpdateCell(workspace.CellUpdate{source, id, "stock", fmt.Sprint(i + 100)})
				})
				run("export_after_edit", i, false, func() (any, error) { return service.Export() })
			}
		}
		if *scenario == "burst" {
			for i := 1; i <= *repeats; i++ {
				run("undo_sequence", i, false, func() (any, error) { return service.Undo() })
			}
			for i := 1; i <= *repeats; i++ {
				run("redo_sequence", i, false, func() (any, error) { return service.Redo() })
			}
			if stats, ok := any(service).(interface{ ResearchHistoryStats() map[string]int64 }); ok {
				b, _ := json.MarshalIndent(stats.ResearchHistoryStats(), "", "  ")
				_ = os.WriteFile(filepath.Join(output, "history-payload.json"), b, 0600)
			}
			run("export_after_sequence", 1, false, func() (any, error) { return service.Export() })
		}
		return
	}
	if *scenario == "full" {
		for i := 1; i <= *repeats; i++ {
			run("get_manual", i, false, func() (any, error) {
				t, e := service.GetTable(source)
				if e == nil && len(t.Rows) != 100000 {
					e = fmt.Errorf("wrong row count")
				}
				return t, e
			})
		}
	}
	ids := []string{"item_000001", "item_050000", "item_100000"}
	for i := 1; i <= *repeats; i++ {
		id := ids[(i-1)%len(ids)]
		expr := fmt.Sprint(600000 + i)
		activeEdit = &editAudit{"UpdateCell", id, "price", expr}
		run("edit_manual", i, i == 1, func() (any, error) {
			r, e := service.UpdateCell(workspace.CellUpdate{Path: source, RowID: id, Column: "price", Expression: expr})
			if e == nil {
				e = checkTable(r, id, "price", expr)
			}
			return r, e
		})
	}
	if *scenario == "full" {
		for i := 1; i <= *repeats; i++ {
			run("undo_with_table_reload", i, false, func() (any, error) {
				r, e := service.Undo()
				if e != nil {
					return nil, e
				}
				t, e := service.GetTable(source)
				return workspace.CellUpdateResult{Table: t, Workspace: r.Workspace}, e
			})
		}
		for i := 1; i <= *repeats; i++ {
			run("redo_with_table_reload", i, false, func() (any, error) {
				r, e := service.Redo()
				if e != nil {
					return nil, e
				}
				t, e := service.GetTable(source)
				return workspace.CellUpdateResult{Table: t, Workspace: r.Workspace}, e
			})
		}
		for i := 1; i <= *repeats; i++ {
			run("compile", i, false, func() (any, error) {
				r, e := service.Compile()
				if e == nil && len(r.Workspace.Compile.Diagnostics) != 0 {
					e = fmt.Errorf("compile diagnostics %v", r.Workspace.Compile.Diagnostics)
				}
				return r, e
			})
		}
		for i := 1; i <= *repeats; i++ {
			run("export", i, false, func() (any, error) { return service.Export() })
		}
		run("switch_auto", 1, false, func() (any, error) { return service.SetCompileMode("auto") })
		for i := 1; i <= *repeats; i++ {
			id := ids[(i-1)%len(ids)]
			expr := fmt.Sprint(700000 + i)
			activeEdit = &editAudit{"UpdateCell", id, "price", expr}
			run("edit_auto", i, false, func() (any, error) {
				r, e := service.UpdateCell(workspace.CellUpdate{Path: source, RowID: id, Column: "price", Expression: expr})
				if e == nil {
					e = checkTable(r, id, "price", expr)
				}
				return r, e
			})
		}
	}
	f, err := os.Create(filepath.Join(output, "heap.pprof"))
	if err == nil {
		runtime.GC()
		_ = pprof.WriteHeapProfile(f)
		f.Close()
	}
}
