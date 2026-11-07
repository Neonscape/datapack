package main

import (
	"bufio"
	"datapack/internal/document"
	"datapack/internal/workspace"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// TestCompactConsistencyPipe runs actual App/Service methods and JSON encoding.
// It intentionally does not start Wails IPC, mount React or open a native GUI.
// A separate process per reopen provides process-lifetime isolation.
func TestCompactConsistencyPipe(t *testing.T) {
	if os.Getenv("COMPACT_PIPE") != "1" {
		t.Skip("explicit pipe harness only")
	}
	app := &App{workspace: workspace.NewService(os.Getenv("COMPACT_SETTINGS"))}
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 4096), 4*1024*1024)
	for scanner.Scan() {
		var req struct {
			ID     int               `json:"id"`
			Method string            `json:"method"`
			Args   []json.RawMessage `json:"args"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			t.Fatal(err)
		}
		var value any
		var err error
		var path string
		var update workspace.CellUpdate
		var formula workspace.FormulaSet
		if len(req.Args) > 0 {
			switch req.Method {
			case "OpenProject", "GetTable", "ReadSourceTable":
				err = json.Unmarshal(req.Args[0], &path)
			case "SetCellFormula":
				err = json.Unmarshal(req.Args[0], &formula)
			default:
				err = json.Unmarshal(req.Args[0], &update)
			}
		}
		if err == nil {
			switch req.Method {
			case "OpenProject":
				value, err = app.OpenProject(path)
			case "GetTable":
				value, err = app.GetTable(path)
			case "ReadSourceTable":
				value, err = document.LoadTable(path)
			case "UpdateCellDelta":
				value, err = app.UpdateCellDelta(update)
			case "SetCellTextDelta":
				value, err = app.SetCellTextDelta(update)
			case "ClearCell":
				value, err = app.ClearCell(update)
			case "SetCellFormula":
				value, err = app.SetCellFormula(formula)
			case "UndoDelta":
				value, err = app.UndoDelta()
			case "RedoDelta":
				value, err = app.RedoDelta()
			case "Compile":
				value, err = app.Compile()
			case "Export":
				value, err = app.Export()
			default:
				err = fmt.Errorf("unsupported method %q", req.Method)
			}
		}
		response := map[string]any{"id": req.ID, "value": value}
		if err != nil {
			response["error"] = err.Error()
		}
		data, e := json.Marshal(response)
		if e != nil {
			t.Fatal(e)
		}
		fmt.Printf("COMPACT_JSON %s\n", data)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
}
