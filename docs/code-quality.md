# Code Quality And Maintenance

This document is the maintenance baseline for Datapack. It describes the current shape of the codebase, the risks worth addressing, and the order in which refactoring should happen. It is not a product specification; use `docs/design.md` and `docs/implementation-draft.md` for product and format decisions.

## Current Baseline

The project has good domain-package boundaries and a thick test suite, but application orchestration has accumulated too many responsibilities in a few files.

| Area | Representative code | Current assessment |
|---|---|---|
| Domain packages | `compiler`, `formula`, `foreignkey`, `schemaregistry`, `exporters`, `logging` | Healthy: small interfaces and focused tests |
| Source parsing and writing | `document`, `cuesource`, `schemamodel` | Usable: complexity mostly comes from restricted CUE AST editing |
| Backend orchestration | `internal/workspace` | Highest maintenance risk: one Service still owns most application workflows |
| Frontend shell | `frontend/src/App.tsx` | High maintenance risk: session state, IPC orchestration, menus, dialogs, and navigation are coupled |
| Table interaction | `frontend/src/table/TableGrid.tsx` | High interaction risk: grid, editors, menus, comments, paste, and display-name headers share one component |
| Model boundary | `frontend/src/types.ts` and generated `frontend/wailsjs/go/models.ts` | Duplicate DTO/view-model definitions cause type assertions and drift risk |

Recent features add more metadata and derived view state to the table boundary:

- `Column.Schema` preserves the source Schema expression.
- `Column.EffectiveSchema` drives the actual editor for named Schema references.
- `Column.DisplayName` is UI metadata and must not replace the stable field name.
- Foreign-key options are derived workspace metadata, not source field values.

These distinctions are correct product behavior, but they increase the need for an explicit metadata boundary instead of adding more logic to `workspace.go` and `TableGrid.tsx`.

## What Is Working Well

- The product boundary is explicit: Datapack edits a restricted CUE project, not arbitrary CUE.
- Source files are the authoritative data; writes validate a candidate before replacing the source.
- `compiler.CompileAll` returns diagnostics or a package, and exporters consume compiled data.
- `cuesource` and `document` intentionally have different strictness: tree scanning can show damaged files, while editing requires strict parsing.
- Foreign-key validation, navigation, rename propagation, and deletion protection are implemented as domain behavior.
- Named Schema references are distinguished from table foreign keys; effective Schema metadata drives the correct editor.
- Field display names are presentation metadata. Stable field names remain the identity used by CUE, formulas, foreign keys, diagnostics, and exports.
- The test suite includes package tests, `test/functional`, `test/edge`, and Playwright coverage.

Do not refactor these strengths away merely to shorten files.

## Priority Risks

### Correctness: Workspace concurrency

`workspace.Service` has mutable state such as `current`, undo/redo stacks, revision counters, compile state, and cached settings. Unlike the logger, it has no mutex. Wails calls can arrive from different goroutines, so concurrent edits, undo, compile, export, or settings changes can race.

This is a correctness issue, not a style preference. Add synchronization at public Service boundaries, then use private locked helpers to avoid recursive locking.

Acceptance signals:

- Concurrent `UpdateCell` and `Undo` cannot corrupt state or history.
- Concurrent `Compile` and `Export` observe a coherent snapshot.
- `go test -race ./...` passes.

### Correctness: swallowed document-load errors

Document collection currently skips `document.LoadTable` and `document.LoadPublicSchema` failures. A damaged file can therefore appear in the tree while silently disappearing from the compile/index input.

Collection must preserve load failures as diagnostics with the source path and, where possible, table or Schema context. “File is damaged” must not become indistinguishable from “file does not exist.”

Acceptance signals:

- Damaged table and Schema files remain visible as diagnostics.
- References to damaged documents report that the target is unavailable, not merely missing.
- Clean projects retain identical compile and export behavior.

### Maintainability: workspace orchestration

The Service currently coordinates:

- Project lifecycle and table lifecycle.
- Cell, row, column, formula, comment, and paste edits.
- Undo/Redo.
- Schema edits, imports, and reference rewriting.
- Compilation, export settings, and exporter dispatch.
- Navigation, foreign-key indexes, user settings, and logging.

Do not change the Wails API first. Split implementation files while preserving the Service type and public method behavior.

### Maintainability: frontend session state

`App.tsx` owns workspace state, tabs, errors, menus, dialogs, navigation, settings, and repeated Wails result handling. New features repeatedly add state and `try/catch` blocks there.

Extract a workspace session layer before introducing a state-management framework. The session layer should own opening, refreshing, updating, closing, and synchronizing tables and Schema tabs.

### Maintainability: frontend model boundary

The frontend has hand-written UI types and generated Wails DTOs. Frequent `as TableDocument` and `as WorkspaceView` assertions hide drift.

Keep UI-only state separate, but define explicit conversions for backend DTOs. Avoid arbitrary casts in `App.tsx`; conversion functions should be the boundary where backend shape and UI shape are checked.

### Maintainability: TableGrid codecs

`TableGrid.tsx` still contains hand-written array and structure expression parsing. This is a risk because CUE string escaping and structured values can become more complex than the local parser expects.

Move these pure functions into a small module with direct tests first. A later backend structured-cell model can replace them without making the grid component larger.

### Performance: full refresh after edits

An edit currently reloads project configuration, scans the tree, reloads documents, rebuilds Schema and foreign-key indexes, and may compile the project. This is acceptable for small projects but should be measured before adding caching or incremental compilation.

First record timing for scan, document loading, indexes, and compilation. Do not introduce incremental build complexity without a benchmark showing the need.

## Refactoring Order

The work should proceed in these vertical steps:

1. Add concurrency protection to `workspace.Service`. Establish public locked boundaries and race tests without changing behavior.
2. Preserve document-load errors in workspace collection and expose them as diagnostics.
3. Split `internal/workspace/workspace.go` by topic while retaining the Service API:
   - `edit.go`: cell, row, column, paste, history.
   - `schema_edit.go`: Schema operations, imports, and reference rewriting.
   - `metadata.go`: diagnostics, effective Schemas, display names, and foreign-key options.
   - `project.go`: project and table lifecycle.
   - `export.go`: project/table export settings and exporter dispatch.
   - `navigation.go`: semantic navigation and target resolution.
4. Extract a frontend workspace session module or hook. Keep `App.tsx` responsible for layout and composition, not every Wails state transition.
5. Move array and structure codecs out of `TableGrid.tsx`, add pure unit tests, and leave the grid responsible for interaction and rendering.

Refactor one step at a time. A physical file split should not also change a public API, data format, and UI behavior in the same commit.

## Explicit Non-Goals

- Do not merge `cuesource` and `document`; their strictness difference is intentional.
- Do not add a state-management framework before extracting the session seam.
- Do not introduce dependency injection, a service locator, or a second exporter abstraction.
- Do not redesign the restricted CUE format during maintenance refactoring.
- Do not implement incremental compilation before collecting performance data.
- Do not replace stable field names with display names. Display names are UI metadata only.
- Do not make CSV exporter behavior depend on field display names; export configuration stores stable field names.

## Quality Gates

Minimum gates for every refactoring step:

```bash
go test ./...
go test -race ./...
npm --prefix frontend run typecheck
npm --prefix frontend run test:e2e
npm --prefix frontend run build
```

For a production desktop change, also run:

```bash
./scripts/build.sh -clean
```

Review should focus on behavior and seams, not line count. A successful refactor means that changing one concern no longer requires opening unrelated parsing, export, navigation, or UI-session code, while the full test and race gates remain green.
