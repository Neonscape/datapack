# Save As destination preservation

`run.py` drives the opt-in `TestSaveAsServicePipe` in a freshly built workspace test binary. It uses real temporary files, OS symlinks and Service methods, never mocks Save As or simulates SaveDialog clicks. Do not point it at a user project. Each output directory must be new.

```
go test -c -o /tmp/saveas-service.test ./internal/workspace
python3 test/saveaspreservation/run.py --binary /tmp/saveas-service.test --out /tmp/saveas-core
python3 test/saveaspreservation/run.py --binary /tmp/saveas-service.test --out /tmp/saveas-full --mode full
```

Core cases place a dangling symlink at `copy.dtproj` or ordinary `notes.txt`. They require a nonempty error and unchanged source, destination nodes, outside sentinel, current workspace, epoch and both history stacks. A successful overwrite is a red verdict. Before/after inventories use lstat file types, modes, hashes and exact readlink text. Wire logs retain actual responses and history state.

After rejection, exercise Redo/Undo; remove only the collision created by that fixture; retry in the same Service; require a fresh epoch and cleared history. Save As must copy source bytes to different inodes. Compile/Export is consumed with Python CSV, JSON and Decimal against fixed values, including the phase6 exact default and a small array. Another OS process reopens and exports the saved copy; CSV bytes must match. Editing the copy sidecar must not change the original.

Full mode adds ordinary files/directories, live file/directory links, reserved source directory slots, CreateProject shared-helper boundaries, real empty CreateProject directories, legal parent aliases and existing source-symlink rejection. Slot symlinks are occupied nodes; real empty source directories remain allowed for CreateProject. Parent path aliases remain allowed. This does not guarantee safety against concurrent filesystem replacement or crash/power loss.

This is Linux Service validation, not native Wails IPC, GUI, React or cross-platform validation. Explicit native symlink support is required. An opt-in pipe test is skipped by ordinary `go test`; regression tests run normally.
