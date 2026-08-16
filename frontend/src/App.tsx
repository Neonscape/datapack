import {useEffect, useState} from 'react';
import './App.css';
import {AddColumn, ClearCell, Compile, CreateRow, CreateSchemaDefinition, CreateSchemaFile, CreateTable, DeleteColumn, DeleteRow, DeleteSchemaDefinition, DeleteTable, Export, GetPublicSchema, GetTable, MoveColumn, Navigate, Paste, PrintSchema, Redo, RenameColumn, RenameRow, RenameSchemaDefinition, RenameTable, SelectProject, SetCellFormula, SetColumnSchema, SetCompileMode, SetSchemaDefinition, Undo, UpdateCell} from "../wailsjs/go/main/App";
import {schemamodel} from "../wailsjs/go/models";
import {LogPanel} from "./LogPanel";
import {schemaDefinitions} from "./schema/refs";
import {SchemaEditor} from "./schema/SchemaEditor";
import {SchemaForm} from "./schema/SchemaForm";
import {TableGrid} from "./table/TableGrid";
import {Tree} from "./tree/Tree";
import type {CellError, NavRequest, NavTarget, PublicSchemaDocument, SchemaDefinition, SchemaEditorState, SchemaEditResult, SchemaNode, SchemaUsage, TableDocument, TreeNode, WorkspaceView} from "./types";

type StructureDialog = {
    mode: "add" | "schema" | "rename" | "move" | "delete";
    name: string;
    second: string;
    schema: string;
    initial: string;
    node?: SchemaNode;
    custom?: boolean;
};
type TableDialog = {namespace: string; name: string};

type TreeDialog =
    | {mode: "renameTable"; path: string; namespace: string; name: string}
    | {mode: "deleteTable"; path: string; label: string}
    | {mode: "renameDefinition"; path: string; name: string; newName: string}
    | {mode: "deleteDefinition"; path: string; name: string};

function App() {
    const [workspace, setWorkspace] = useState<WorkspaceView | null>(null);
    const [selectedSchema, setSelectedSchema] = useState<SchemaDefinition | null>(null);
    const [table, setTable] = useState<TableDocument | null>(null);
    const [schemaDocument, setSchemaDocument] = useState<SchemaEditorState | null>(null);
    const [cellError, setCellError] = useState<CellError | null>(null);
    const [error, setError] = useState("");
    const [notice, setNotice] = useState("");
    const [dialog, setDialog] = useState<StructureDialog | null>(null);
    const [tableDialog, setTableDialog] = useState<TableDialog | null>(null);
    const [appMenu, setAppMenu] = useState<"file" | "edit" | null>(null);
    const [focusTarget, setFocusTarget] = useState<{path: string; rowID: string; column: string; nonce: number} | null>(null);
    const [navMenu, setNavMenu] = useState<{x: number; y: number; targets: NavTarget[]} | null>(null);
    const [treeMenu, setTreeMenu] = useState<{node: TreeNode; namespace: string; x: number; y: number} | null>(null);
    const [treeDialog, setTreeDialog] = useState<TreeDialog | null>(null);
    const [collapsedNodes, setCollapsedNodes] = useState<Set<string>>(new Set());

    async function pasteCells(row: number, column: number, text: string) {
        if (!table) return;
        const values = text.replace(/\r/g, "").split("\n").filter((line, index, lines) => line !== "" || index < lines.length - 1).map((line) => line.split("\t"));
        if (values.length === 0 || values.some((line) => line.length !== values[0].length)) {
            setError("Clipboard rows must have the same number of columns");
            return;
        }
        const rowIDs = table.rows.slice(row, row + values.length).map((entry) => entry.id);
        const columns = table.columns.slice(column - 1, column - 1 + values[0].length).map((entry) => entry.name);
        if (rowIDs.length !== values.length || columns.length !== values[0].length) {
            setError("Paste range exceeds the existing table");
            return;
        }
        setError("");
        try {
            const result = await Paste({path: table.path, rowIDs, columns, values});
            setTable(result.table as TableDocument);
            setWorkspace(result.workspace as WorkspaceView);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    useEffect(() => {
        const historyShortcut = (event: KeyboardEvent) => {
            if (!table || !(event.ctrlKey || event.metaKey) || (event.target as Element | null)?.closest("textarea, input, select")) return;
            const redo = event.key.toLowerCase() === "y" || (event.shiftKey && event.key.toLowerCase() === "z");
            if (event.key.toLowerCase() !== "z" && !redo) return;
            event.preventDefault();
            void (async () => {
                try {
                    const result = redo ? await Redo(table.path) : await Undo(table.path);
                    setTable(result.table as TableDocument);
                    setWorkspace(result.workspace as WorkspaceView);
                } catch (reason) {
                    setError(reason instanceof Error ? reason.message : String(reason));
                }
            })();
        };
        document.addEventListener("keydown", historyShortcut);
        return () => document.removeEventListener("keydown", historyShortcut);
    }, [table]);

    useEffect(() => {
        const closeMenu = (event: PointerEvent) => {
            if (!(event.target as Element | null)?.closest(".app-menu")) setAppMenu(null);
            if (!(event.target as Element | null)?.closest(".nav-menu")) setNavMenu(null);
            if (!(event.target as Element | null)?.closest(".tree-context-menu")) setTreeMenu(null);
        };
        const closeOnEscape = (event: KeyboardEvent) => {
            if (event.key === "Escape") {
                setAppMenu(null);
                setNavMenu(null);
                setTreeMenu(null);
            }
        };
        document.addEventListener("pointerdown", closeMenu);
        document.addEventListener("keydown", closeOnEscape);
        return () => {
            document.removeEventListener("pointerdown", closeMenu);
            document.removeEventListener("keydown", closeOnEscape);
        };
    }, []);

    async function selectProject() {
        setError("");
        try {
            const view = await SelectProject() as WorkspaceView;
            if (view?.project?.path) {
                setWorkspace(view);
                setSelectedSchema(schemaDefinitions(view)[0] ?? null);
                setTable(null);
                setSchemaDocument(null);
                setCellError(null);
            }
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function openTable(path: string) {
        setError("");
        try {
            setTable(await GetTable(path) as TableDocument);
            setSchemaDocument(null);
            setCellError(null);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function openSchema(node: TreeNode) {
        if (node.kind === "localSchema") {
            const name = node.name.startsWith("#") ? node.name.replace(/^#/, "") : null;
            await openSchemaByPath(node.path!, name, true);
        } else {
            const name = node.kind === "schemaDefinition" ? node.name.replace(/^#/, "") : null;
            await openSchemaByPath(node.path!, name, false);
        }
    }

    async function openSchemaByPath(path: string, definition: string | null, local: boolean, creating = false) {
        setError("");
        try {
            if (local) {
                const document = await GetTable(path) as TableDocument;
                const logical = document.namespace ? `${document.namespace}.${document.name}` : document.name;
                setSchemaDocument({path, local: true, logicalTable: logical, definitions: document.localSchema, selected: definition ?? document.localSchema[0]?.name ?? null, creating});
            } else {
                const document = await GetPublicSchema(path) as PublicSchemaDocument;
                setSchemaDocument({path, local: false, package: document.package, definitions: document.definitions, selected: definition ?? document.definitions[0]?.name ?? null, creating});
            }
            setTable(null);
            setCellError(null);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function navigateTo(request: NavRequest, position: {x: number; y: number}) {
        setError("");
        try {
            const target = await Navigate(request) as NavTarget;
            applyNavigationTarget(target, position);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    function applyNavigationTarget(target: NavTarget, position: {x: number; y: number}) {
        if (target.kind === "schema-reference") {
            setNavMenu(null);
            void openSchemaByPath(target.path, target.definition ?? null, !!target.local);
            return;
        }
        if (target.references && target.references.length > 0) {
            setNavMenu({x: position.x, y: position.y, targets: target.references});
            return;
        }
        focusTableCell(target.path, target.rowID ?? "", target.column ?? "");
    }

    function focusTableCell(path: string, rowID: string, column: string) {
        setNavMenu(null);
        setFocusTarget({path, rowID, column, nonce: Date.now()});
        if (table?.path !== path) {
            void openTable(path);
        }
    }

    function openUsage(usage: SchemaUsage) {
        if (usage.table) {
            focusTableCell(usage.path, "", usage.field === "#PrimaryKey" ? "" : usage.field);
            return;
        }
        const definition = schemaDefinitions(workspace).find((entry) => entry.id === usage.field);
        if (definition) {
            void openSchemaByPath(definition.path, definition.id.split(".#")[1], !!definition.table);
        }
    }

    function renameDefinitionAt(path: string, name: string) {
        setTreeDialog({mode: "renameDefinition", path, name, newName: name});
    }

    function deleteDefinitionAt(path: string, name: string) {
        setTreeDialog({mode: "deleteDefinition", path, name});
    }

    function renameTableAt(path: string, name: string, namespace: string) {
        setTreeDialog({mode: "renameTable", path, namespace, name});
    }

    function deleteTableAt(path: string, label: string) {
        setTreeDialog({mode: "deleteTable", path, label});
    }

    async function newDefinitionInNamespace(namespace: string) {
        setError("");
        try {
            const definition = schemaDefinitions(workspace).find((entry) => !entry.table && entry.id.startsWith(`${namespace}.#`));
            if (definition) {
                await openSchemaByPath(definition.path, null, false, true);
                return;
            }
            const result = await CreateSchemaFile(namespace) as {schema: PublicSchemaDocument; workspace: WorkspaceView};
            setWorkspace(result.workspace);
            await openSchemaByPath(result.schema.path, null, false, true);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function submitTreeDialog() {
        if (!treeDialog) return;
        setError("");
        try {
            if (treeDialog.mode === "renameTable") {
                const result = await RenameTable({path: treeDialog.path, namespace: treeDialog.namespace, name: treeDialog.name}) as {table: TableDocument; workspace: WorkspaceView};
                setWorkspace(result.workspace);
                if (table?.path === treeDialog.path) {
                    setTable(await GetTable(result.table.path) as TableDocument);
                } else {
                    setTable(null);
                }
            } else if (treeDialog.mode === "deleteTable") {
                const result = await DeleteTable({path: treeDialog.path}) as {workspace: WorkspaceView};
                setWorkspace(result.workspace);
                if (table?.path === treeDialog.path) {
                    setTable(null);
                }
            } else if (treeDialog.mode === "renameDefinition") {
                const result = await RenameSchemaDefinition({path: treeDialog.path, oldName: treeDialog.name, newName: treeDialog.newName}) as SchemaEditResult;
                setWorkspace(result.workspace);
                if (schemaDocument && schemaDocument.path === treeDialog.path) {
                    if (schemaDocument.local) {
                        setSchemaDocument({...schemaDocument, definitions: result.table.localSchema, selected: treeDialog.newName});
                    } else {
                        setSchemaDocument({...schemaDocument, definitions: result.schema.definitions, selected: treeDialog.newName});
                    }
                }
            } else {
                const result = await DeleteSchemaDefinition({path: treeDialog.path, name: treeDialog.name}) as SchemaEditResult;
                setWorkspace(result.workspace);
            }
            setTreeDialog(null);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    function applySchemaResult(result: SchemaEditResult, selected: string | null) {
        setWorkspace(result.workspace);
        if (!schemaDocument) return;
        if (schemaDocument.local) {
            setSchemaDocument({...schemaDocument, definitions: result.table.localSchema, selected: selected ?? result.table.localSchema[0]?.name ?? null});
        } else {
            setSchemaDocument({...schemaDocument, definitions: result.schema.definitions, selected: selected ?? result.schema.definitions[0]?.name ?? null});
        }
    }

    async function saveSchemaDefinition(create: boolean, name: string, kind: "type" | "constraint", type: string, appliesTo: string[], expression: string) {
        if (!schemaDocument) return;
        setError("");
        const payload = {path: schemaDocument.path, name, kind, type, appliesTo, schemaExpression: expression};
        const result = create
            ? await CreateSchemaDefinition(payload) as SchemaEditResult
            : await SetSchemaDefinition(payload) as SchemaEditResult;
        applySchemaResult(result, name);
    }

    async function renameSchemaDefinition(oldName: string, newName: string) {
        if (!schemaDocument) return;
        setError("");
        const result = await RenameSchemaDefinition({path: schemaDocument.path, oldName, newName}) as SchemaEditResult;
        applySchemaResult(result, newName);
    }

    async function deleteSchemaDefinition(name: string) {
        if (!schemaDocument) return;
        setError("");
        const result = await DeleteSchemaDefinition({path: schemaDocument.path, name}) as SchemaEditResult;
        applySchemaResult(result, null);
    }

    function schemaUsages(name: string): SchemaUsage[] {
        if (!schemaDocument || !workspace) return [];
        const id = schemaDocument.local
            ? `${schemaDocument.logicalTable}._local.#${name}`
            : `${schemaDocument.package}.#${name}`;
        return schemaDefinitions(workspace).find((definition) => definition.id === id)?.usages ?? [];
    }

    async function saveCellInput(rowID: string, column: string, input: string) {
        if (!table) {
            return;
        }
        setError("");
        setCellError(null);
        try {
            if (column === "__create__") {
                const {values, formulas} = JSON.parse(input);
                const result = await CreateRow({path: table.path, rowID, values, formulas});
                setTable(result.table as TableDocument);
                setWorkspace(result.workspace as WorkspaceView);
                return;
            }
            if (column === "id") {
                if (!input) throw new Error("ID must not be empty");
                const result = await RenameRow({path: table.path, oldID: rowID, newID: input});
                setTable(result.table as TableDocument);
                setWorkspace(result.workspace as WorkspaceView);
                return;
            }
            if (input === "") {
                const result = await ClearCell({path: table.path, rowID, column, expression: ""});
                setTable(result.table as TableDocument);
                setWorkspace(result.workspace as WorkspaceView);
                return;
            }
            const result = input.startsWith("=")
                ? await SetCellFormula({path: table.path, rowID, column, expression: input.slice(1).trim()})
                : await UpdateCell({path: table.path, rowID, column, expression: input});
            setTable(result.table as TableDocument);
            setWorkspace(result.workspace as WorkspaceView);
        } catch (reason) {
            const message = reason instanceof Error ? reason.message : String(reason);
            setError(message);
            setCellError({rowID, column, message});
        }
    }

    async function submitStructureDialog() {
        if (!table || !dialog) return;
        setError("");
        setCellError(null);
        try {
            let schemaExpression = dialog.schema;
            if ((dialog.mode === "add" || dialog.mode === "schema") && !dialog.custom) {
                schemaExpression = await PrintSchema((dialog.node ?? {kind: "string"}) as unknown as schemamodel.Node);
            }
            let result;
            if (dialog.mode === "add") {
                result = await AddColumn({path: table.path, name: dialog.name, schemaExpression, initialExpression: dialog.initial, position: Number(dialog.second)});
            } else if (dialog.mode === "schema") {
                result = await SetColumnSchema({path: table.path, name: dialog.name, schemaExpression});
            } else if (dialog.mode === "rename") {
                result = await RenameColumn({path: table.path, oldName: dialog.name, newName: dialog.second});
            } else if (dialog.mode === "move") {
                const position = Number(dialog.second);
                if (!Number.isInteger(position)) throw new Error("Column position must be an integer");
                result = await MoveColumn({path: table.path, name: dialog.name, position});
            } else {
                result = await DeleteColumn({path: table.path, name: dialog.name});
            }
            setTable(result.table as TableDocument);
            setWorkspace(result.workspace as WorkspaceView);
            setDialog(null);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function deleteRow(rowID: string) {
        if (!table) {
            return;
        }
        setError("");
        setCellError(null);
        try {
            const result = await DeleteRow({path: table.path, rowID});
            setTable(result.table as TableDocument);
            setWorkspace(result.workspace as WorkspaceView);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function createTable() {
        if (!tableDialog) return;
        setError("");
        try {
            const result = await CreateTable(tableDialog);
            setWorkspace(result.workspace as WorkspaceView);
            setTable(result.table as TableDocument);
            setTableDialog(null);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function applyHistory(redo: boolean) {
        if (!table) return;
        setError("");
        try {
            const result = redo ? await Redo(table.path) : await Undo(table.path);
            setTable(result.table as TableDocument);
            setWorkspace(result.workspace as WorkspaceView);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function exportProject() {
        setError("");
        setNotice("");
        try {
            const result = await Export();
            setNotice(`Exported ${result.files.length} CSV file${result.files.length === 1 ? "" : "s"}`);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function toggleCompileMode() {
        if (!workspace) return;
        setError("");
        try {
            const mode = workspace.compile.mode === "auto" ? "manual" : "auto";
            const view = await SetCompileMode(mode) as WorkspaceView;
            setWorkspace(view);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function compileProject() {
        if (!workspace) return;
        setError("");
        try {
            const result = await Compile() as {workspace: WorkspaceView};
            setWorkspace(result.workspace);
            if (table) {
                setTable(await GetTable(table.path) as TableDocument);
            }
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    return (
        <main className="app-shell" onContextMenu={(event) => event.preventDefault()}>
            <header className="app-header">
                <div className="brand">Datapack</div>
                <div className="header-divider"/>
                <nav className="app-menubar" aria-label="Application menu">
                    <div className="app-menu">
                        <button aria-expanded={appMenu === "file"} onClick={() => setAppMenu(appMenu === "file" ? null : "file")}>File</button>
                        {appMenu === "file" && <div className="app-menu-popover">
                            <button onClick={() => { setAppMenu(null); selectProject(); }}>Open project</button>
                            <button disabled={!workspace} onClick={() => { setAppMenu(null); setTableDialog({namespace: "game", name: ""}); }}>New table</button>
                        </div>}
                    </div>
                    <div className="app-menu">
                        <button aria-expanded={appMenu === "edit"} onClick={() => setAppMenu(appMenu === "edit" ? null : "edit")}>Edit</button>
                        {appMenu === "edit" && <div className="app-menu-popover">
                            <button disabled={!table} onClick={() => { setAppMenu(null); applyHistory(false); }}>Undo <kbd>Ctrl+Z</kbd></button>
                            <button disabled={!table} onClick={() => { setAppMenu(null); applyHistory(true); }}>Redo <kbd>Ctrl+Y</kbd></button>
                        </div>}
                    </div>
                </nav>
                <span className="project-name">{workspace?.tree.root.name ?? "No project open"}</span>
                <div className="header-actions">
                    {workspace && <button className={`build-mode build-mode-${workspace.compile.mode} build-mode-toggle`} aria-label="Toggle compile mode" title="Switch between auto and manual compile" onClick={toggleCompileMode}>{workspace.compile.mode === "auto" ? "Auto compile" : "Manual compile"}{workspace.compile.pending ? " · pending" : ""}</button>}
                    {workspace && workspace.compile.mode === "manual" && <button className="compile-button" disabled={!workspace.compile.pending} onClick={compileProject}>Compile</button>}
                    {workspace && workspace.compile.diagnostics.length > 0 && <span className="compile-errors" title="Compile diagnostics">{workspace.compile.diagnostics.length}</span>}
                    {workspace && <button className="export-button" onClick={exportProject}>Export CSV</button>}
                </div>
            </header>

            {workspace ? (
                <section className="workbench">
                    <aside className="project-explorer">
                        <div className="panel-title">Project</div>
                        <div className="tree"><Tree node={workspace.tree.root} collapsed={collapsedNodes} onToggleCollapse={(key) => setCollapsedNodes((current) => {
                            const next = new Set(current);
                            if (next.has(key)) next.delete(key); else next.add(key);
                            return next;
                        })} onOpenTable={openTable} onOpenSchema={openSchema} onContextMenu={(node, position, namespace) => {
                            if (node.kind === "project" || node.kind === "group" || node.kind === "damaged") return;
                            setTreeMenu({node, namespace, x: position.x, y: position.y});
                        }}/></div>
                    </aside>
                    <section className={`editor-placeholder ${table || schemaDocument || selectedSchema ? "editor-table" : ""}`}>
                        {table ? <TableGrid key={table.path} table={table} onAddColumn={(position) => setDialog({mode: "add", name: "", second: String(position), schema: "string", initial: `""`, node: {kind: "string"}, custom: false})} onDeleteColumn={(name) => setDialog({mode: "delete", name, second: "", schema: "", initial: ""})} onDeleteRow={deleteRow} onEditColumnSchema={(name, schemaExpression) => setDialog({mode: "schema", name, second: "", schema: schemaExpression.replaceAll("_local.#", "#"), initial: "", node: table?.columns.find((column) => column.name === name)?.schema ?? {kind: "string"}, custom: false})} onMoveColumn={(name) => setDialog({mode: "move", name, second: "0", schema: "", initial: ""})} onRenameColumn={(name) => setDialog({mode: "rename", name, second: "", schema: "", initial: ""})} onPaste={pasteCells} onSaveInput={saveCellInput} onNavigate={navigateTo} focusTarget={focusTarget} onFocusDone={() => setFocusTarget(null)}/> : schemaDocument ? <SchemaEditor state={schemaDocument} usages={schemaUsages} onOpenUsage={openUsage} workspace={workspace} onSelect={(name) => setSchemaDocument({...schemaDocument, selected: name, creating: false})} onCreate={(name, kind, type, appliesTo, expression) => saveSchemaDefinition(true, name, kind, type, appliesTo, expression)} onApply={(name, kind, type, appliesTo, expression) => saveSchemaDefinition(false, name, kind, type, appliesTo, expression)} onRename={renameSchemaDefinition} onDelete={deleteSchemaDefinition}/> : selectedSchema ? <section className="schema-detail">
                            <div className="schema-detail-head">
                                <span className={`schema-kind-dot schema-kind-${selectedSchema.kind}`} aria-hidden="true"/>
                                <h1>{selectedSchema.id}</h1>
                            </div>
                            <div className="schema-detail-row"><span>Kind</span><code>{selectedSchema.kind}</code></div>
                            <div className="schema-detail-row"><span>Filepath</span><code>{selectedSchema.path}</code></div>
                            {selectedSchema.appliesTo && <div className="schema-detail-row"><span>AppliesTo</span><code>{selectedSchema.appliesTo.join(", ")}</code></div>}
                            <div className="schema-detail-row"><span>References</span><code>{selectedSchema.usages?.length ?? 0}</code></div>
                            <div className="schema-detail-references">
                                {(selectedSchema.usages ?? []).map((usage) => <button className="reference-row" key={`${usage.path}:${usage.field}`} onClick={() => openUsage(usage)}>
                                    <span>{usage.table ?? usage.path}</span>
                                    <code>{usage.field}</code>
                                </button>)}
                                {(selectedSchema.usages ?? []).length === 0 && <div className="schema-editor-empty">No references</div>}
                            </div>
                        </section> : <>
                            <h1>No Schema definitions</h1>
                        </>}
                    </section>
                    <aside className="schema-explorer">
                        <div className="diagnostics-panel">
                            <div className="panel-title">Diagnostics</div>
                            {(workspace.schemas.diagnostics ?? []).length === 0 && (workspace.foreignKeys.diagnostics ?? []).length === 0 && workspace.compile.diagnostics.length === 0 && <div className="schema-editor-empty">No diagnostics</div>}
                            {(workspace.schemas.diagnostics ?? []).length > 0 && <div className="schema-diagnostics">
                                {(workspace.schemas.diagnostics ?? []).map((diagnostic) => <div className="schema-diagnostic" key={`${diagnostic.path}:${diagnostic.field}:${diagnostic.message}`}>
                                    <code>{diagnostic.field}</code>
                                    <span>{diagnostic.message}</span>
                                </div>)}
                            </div>}
                            {(workspace.foreignKeys.diagnostics ?? []).length > 0 && <div className="schema-diagnostics">
                                {(workspace.foreignKeys.diagnostics ?? []).map((diagnostic) => <div className="schema-diagnostic" key={`${diagnostic.rowID}:${diagnostic.column}:${diagnostic.message}`}>
                                    <code>{diagnostic.rowID}.{diagnostic.column}</code>
                                    <span>{diagnostic.message}</span>
                                </div>)}
                            </div>}
                            {workspace.compile.diagnostics.length > 0 && <div className="schema-diagnostics compile-diagnostics">
                                <div className="schema-diagnostics-title">Compile diagnostics</div>
                                {workspace.compile.diagnostics.map((diagnostic) => <div className="schema-diagnostic" key={`${diagnostic.path}:${diagnostic.rowID ?? ""}:${diagnostic.column ?? ""}:${diagnostic.message}`}>
                                    <code>{[diagnostic.rowID, diagnostic.column].filter(Boolean).join(".") || diagnostic.field || diagnostic.table || "compile"}</code>
                                    <span>{diagnostic.message}</span>
                                </div>)}
                            </div>}
                        </div>
                        <div className="schema-panel">
                            <div className="panel-title">Schemas</div>
                            <div className="schema-list">
                                {schemaDefinitions(workspace).map((definition) => <button
                                    className={`schema-row ${selectedSchema?.id === definition.id ? "schema-row-selected" : ""}`}
                                    key={definition.id}
                                    onClick={() => { setSelectedSchema(definition); setTable(null); setSchemaDocument(null); setCellError(null); }}
                                >
                                    <span className={`schema-kind-dot schema-kind-${definition.kind}`} aria-hidden="true"/>
                                    <span>{definition.id}</span>
                                    <span className="usage-count">{definition.usages?.length ?? 0}</span>
                                </button>)}
                            </div>
                        </div>
                    </aside>
                </section>
            ) : (
                <section className="welcome">
                    <div className="welcome-mark">DP</div>
                    <h1>Open a data project</h1>
                    <p>Choose a <code>.dtproj</code> file to inspect its tables and public Schema files.</p>
                    <button className="open-button open-button-large" onClick={selectProject}>Open project</button>
                </section>
            )}
            <LogPanel/>
            {navMenu && <div className="nav-menu" style={{left: navMenu.x, top: navMenu.y}} role="menu" aria-label="References">
                {navMenu.targets.map((target, index) => <button key={index} onClick={() => focusTableCell(target.path, target.rowID ?? "", target.column ?? "")}>
                    <code>{target.path}</code>
                    <span>{target.rowID}{target.column ? `.${target.column}` : ""}</span>
                </button>)}
            </div>}
            {treeMenu && <div className="nav-menu tree-context-menu" style={{left: treeMenu.x, top: treeMenu.y}} role="menu" aria-label="Project tree actions">
                {treeMenu.node.kind === "table" && <>
                    <button onClick={() => { renameTableAt(treeMenu.node.path!, treeMenu.node.name, treeMenu.namespace); setTreeMenu(null); }}>Rename table</button>
                    <button className="tree-context-danger" onClick={() => { deleteTableAt(treeMenu.node.path!, treeMenu.namespace ? `${treeMenu.namespace}.${treeMenu.node.name}` : treeMenu.node.name); setTreeMenu(null); }}>Delete table</button>
                </>}
                {treeMenu.node.kind === "localSchema" && treeMenu.node.name === "local_schema" && <button onClick={() => { setTreeMenu(null); void openSchemaByPath(treeMenu.node.path!, null, true, true); }}>New definition</button>}
                {treeMenu.node.kind === "localSchema" && treeMenu.node.name.startsWith("#") && <>
                    <button onClick={() => { renameDefinitionAt(treeMenu.node.path!, treeMenu.node.name.replace(/^#/, "")); setTreeMenu(null); }}>Rename definition</button>
                    <button className="tree-context-danger" onClick={() => { deleteDefinitionAt(treeMenu.node.path!, treeMenu.node.name.replace(/^#/, "")); setTreeMenu(null); }}>Delete definition</button>
                </>}
                {treeMenu.node.kind === "schemaDefinition" && <>
                    <button onClick={() => { renameDefinitionAt(treeMenu.node.path!, treeMenu.node.name.replace(/^#/, "")); setTreeMenu(null); }}>Rename definition</button>
                    <button className="tree-context-danger" onClick={() => { deleteDefinitionAt(treeMenu.node.path!, treeMenu.node.name.replace(/^#/, "")); setTreeMenu(null); }}>Delete definition</button>
                </>}
                {treeMenu.node.kind === "namespace" && <button onClick={() => { setTreeMenu(null); void newDefinitionInNamespace(treeMenu.node.name); }}>New schema</button>}
            </div>}
            {error && <div className="error-toast" role="alert">{error}</div>}
            {notice && <div className="notice-toast" role="status">{notice}</div>}
            {dialog && <div className="dialog-backdrop" role="presentation" onMouseDown={() => setDialog(null)}>
                <form className="structure-dialog" onMouseDown={(event) => event.stopPropagation()} onSubmit={(event) => { event.preventDefault(); submitStructureDialog(); }}>
                    <h2>{dialog.mode === "add" ? "Add column" : dialog.mode === "schema" ? "Edit column schema" : dialog.mode === "rename" ? "Rename column" : dialog.mode === "move" ? "Move column" : "Delete column"}</h2>
                    {dialog.mode !== "schema" && <label>Column name<input autoFocus value={dialog.name} onChange={(event) => setDialog({...dialog, name: event.target.value})}/></label>}
                    {dialog.mode === "schema" && <label>Column name<input value={dialog.name} disabled/></label>}
                    {(dialog.mode === "add" || dialog.mode === "schema") && <>
                        <div className="schema-mode-toggle">
                            <button type="button" className={!dialog.custom ? "schema-mode-active" : ""} onClick={() => setDialog({...dialog, custom: false})}>Structured</button>
                            <button type="button" className={dialog.custom ? "schema-mode-active" : ""} onClick={() => setDialog({...dialog, custom: true})}>Custom expression</button>
                        </div>
                        {!dialog.custom
                            ? <SchemaForm value={dialog.node ?? {kind: "string"}} onChange={(node) => setDialog({...dialog, node})} workspace={workspace}/>
                            : <label>CUE Schema<textarea autoFocus value={dialog.schema} onChange={(event) => setDialog({...dialog, schema: event.target.value})}/></label>}
                    </>}
                    {dialog.mode === "add" && <label>Initial CUE value<input value={dialog.initial} onChange={(event) => setDialog({...dialog, initial: event.target.value})}/></label>}
                    {dialog.mode === "rename" && <label>New name<input value={dialog.second} onChange={(event) => setDialog({...dialog, second: event.target.value})}/></label>}
                    {dialog.mode === "move" && <label>New zero-based position<input type="number" min="0" value={dialog.second} onChange={(event) => setDialog({...dialog, second: event.target.value})}/></label>}
                    {dialog.mode === "delete" && <p>This removes the column, all row values, formulas, foreign-key metadata, and CSV export configuration that refer to it.</p>}
                    <div className="dialog-actions"><button type="button" onClick={() => setDialog(null)}>Cancel</button><button className={dialog.mode === "delete" ? "dialog-danger" : ""} type="submit">{dialog.mode === "delete" ? "Delete column" : "Apply"}</button></div>
                </form>
            </div>}
            {tableDialog && <div className="dialog-backdrop" role="presentation" onMouseDown={() => setTableDialog(null)}>
                <form className="structure-dialog" onMouseDown={(event) => event.stopPropagation()} onSubmit={(event) => { event.preventDefault(); createTable(); }}>
                    <h2>Create table</h2>
                    <label>Namespace<input autoFocus value={tableDialog.namespace} onChange={(event) => setTableDialog({...tableDialog, namespace: event.target.value})}/></label>
                    <label>Table name<input value={tableDialog.name} onChange={(event) => setTableDialog({...tableDialog, name: event.target.value})}/></label>
                    <div className="dialog-actions"><button type="button" onClick={() => setTableDialog(null)}>Cancel</button><button type="submit">Create</button></div>
                </form>
            </div>}
            {treeDialog && <div className="dialog-backdrop" role="presentation" onMouseDown={() => setTreeDialog(null)}>
                <form className="structure-dialog" onMouseDown={(event) => event.stopPropagation()} onSubmit={(event) => { event.preventDefault(); submitTreeDialog(); }}>
                    {treeDialog.mode === "renameTable" && <>
                        <h2>Rename table</h2>
                        <label>Namespace<input autoFocus value={treeDialog.namespace} onChange={(event) => setTreeDialog({...treeDialog, namespace: event.target.value})}/></label>
                        <label>Table name<input value={treeDialog.name} onChange={(event) => setTreeDialog({...treeDialog, name: event.target.value})}/></label>
                        <div className="dialog-actions"><button type="button" onClick={() => setTreeDialog(null)}>Cancel</button><button type="submit">Apply</button></div>
                    </>}
                    {treeDialog.mode === "deleteTable" && <>
                        <h2>Delete table</h2>
                        <p>This removes table <code>{treeDialog.label}</code> and its source files. This cannot be undone.</p>
                        <div className="dialog-actions"><button type="button" onClick={() => setTreeDialog(null)}>Cancel</button><button className="dialog-danger" type="submit">Delete table</button></div>
                    </>}
                    {treeDialog.mode === "renameDefinition" && <>
                        <h2>Rename definition</h2>
                        <label>Definition name<input autoFocus value={treeDialog.newName} onChange={(event) => setTreeDialog({...treeDialog, newName: event.target.value})}/></label>
                        <div className="dialog-actions"><button type="button" onClick={() => setTreeDialog(null)}>Cancel</button><button type="submit">Apply</button></div>
                    </>}
                    {treeDialog.mode === "deleteDefinition" && <>
                        <h2>Delete definition</h2>
                        <p>This removes definition <code>#{treeDialog.name}</code> and its registration.</p>
                        <div className="dialog-actions"><button type="button" onClick={() => setTreeDialog(null)}>Cancel</button><button className="dialog-danger" type="submit">Delete definition</button></div>
                    </>}
                </form>
            </div>}
        </main>
    )
}

export default App
