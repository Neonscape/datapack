# Missing struct default precision regression

`internal/compiler/default_precision_test.go` drives actual document loading and
compilation for absent versus explicit struct cells. The boundary test records
the empty evaluator input, unresolved field kind, selected default kind and exact
CUE JSON before decoding. Run:

```bash
go test ./internal/compiler -run 'TestCompileMissingStructFieldDefaultsRetainPrecision|TestMissingStructDefaultPrecisionBoundary' -count=1 -v
```

`run.py` uses the existing `TestCompactConsistencyPipe` App/Service adapter. Build
the binary from the source being tested; do not reuse a pre-fix binary for green
verification. On Linux with the existing WebKit 4.1 build dependencies:

```bash
go test -c -tags webkit2_41 -o /path/to/new-pipe.test .
python3 test/defaultprecision/run.py /path/to/new-pipe.test /path/to/new-minimal
python3 test/defaultprecision/run.py /path/to/new-pipe.test /path/to/new-lifecycle lifecycle
python3 test/defaultprecision/consume.py /path/to/new-lifecycle /path/to/new-receipt.json
```

All output paths must be new. The minimal mode directly asserts the numeric
symptom. The lifecycle mode records six stages: initial, saved, cleared, undo,
redo and reopen. Its `passed` flag covers successful lifecycle operations only;
the separate `consume.py` result is required for data correctness. The same old
binary can complete all operations while its consumer fails. That failure is the
bug, not a successful product test.

Each stage retains source, source/table DTO, Compile/Export responses and actual
CSV. The first App is waited before a different process opens the same project;
the final export is byte-compared with the pre-reopen export. The consumer uses
Python CSV and JSON with Decimal and hardcoded requirements, never floating-point
or producer-supplied expected values. It checks numeric JSON types, missing/empty/
explicit struct distinctions, exact integers and decimals, exponent extremes,
whole-struct defaults, ordinary number spelling, booleans, text, null, empty lists
and save/clear restoration.

Only flat scalar-member structs and primitive lists are product coverage. Optional
struct schema fields and nested containers are outside the documented supported
schema; the existing frontend codec tests preserve optional-label fallback and
quoted fields without claiming those schemas work end to end. JSON here is the
actual struct/list encoding inside CSV, not an independent JSON exporter.

This is a real Go App/Service test over JSON pipes. It does not mount React, start
native Wails IPC or test a GUI. The companion phase4/phase5 harnesses additionally
exercise extracted production frontend functions with synchronous adapters, under
their documented limits. No additional dependencies or services are required.
