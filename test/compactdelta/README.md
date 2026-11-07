# Compact delta consistency evidence harness

`literal.cue` is an independent scalar fixture. `run.cjs` executes real App methods
from the root `phase4_bridge_test.go` process helper. It uses the installed
TypeScript compiler in memory to execute the actual generated App.js, bridge,
workspace session, and source-extracted App/grid functions. The only RPC transport
replacement is a JSON pipe into the real App/Service; the separately labelled
fault probes replay/corrupt real responses. No numeric response is invented by JS.

This is **not** React mounting, native Wails runtime IPC, a browser or GUI test.
The React state/ref adapter is synchronous and the actual history layout-effect
callback is explicitly invoked at a simulated commit boundary. This does not test
React scheduling, effects, focus or painting. `ReadSourceTable` is an independent
source parse, not a hydrated UI table: formula/default omissions are checked via
real GetTable and CSV export instead of equating incompatible representations.

The controlled-run entry requires an active parent `status.json`, a fresh quota
snapshot and no STOP_WORK; it refuses work at the 2026-11-06 22:58 UTC cleanup
cutoff. `COMPACT_CONTROL_ROOT` selects that existing control directory when running
from a clean patch-replay checkout. It does not create or change supervision.
Do not bypass this gate for an expired run; a later run requires new authorization.

From an authorized isolated checkout, using preinstalled dependencies:

```
go test ./internal/workspace -run '^TestCompactBoundaryHistorySourceAndStacks$' -count=1
go test -c -tags webkit2_41 -o /local/output/compact-pipe.test .
node test/compactdelta/run.cjs "$PWD" /local/output/compact-pipe.test /local/output/new-run
python3 test/compactdelta/consume.py /local/output/new-run /local/output/decimal.json
```

The App test build requires existing `frontend/dist` embed assets (or a separately
authorized normal frontend build). This harness neither installs dependencies nor
runs a GUI. Configure TMPDIR/GOTMPDIR/GOCACHE/XDG paths into the writable task root;
module dependencies are read-only and network downloads disabled.

The matrix has 28 stages and 19 genuine compact mutation responses. It asserts
compact `delta` with no `table`, exact revision transitions, source persistence,
actual frontend merge and fallback, same/different cell edits, text Undo/Redo,
formula/full-table history transitions, empty-text direct bridge versus App ClearCell,
real Export, and reopening in a distinct Linux App process. Every CSV field is
consumed independently by Python CSV/Decimal. `compact_boundary_test.go` additionally
exhausts numeric compact undo/redo stacks and checks exact source restoration.

The malformed epoch/path/row/index/column/base probes are fault injections, not
backend bugs. Path rejection covers an absent path; a forged otherwise-valid delta
for another cached table is outside this assertion. Formula float semantics and
CSV null/empty ambiguity are unchanged. Project epochs are process-local; separate
Linux processes may both start at epoch `1`, and are not expected to share history.
