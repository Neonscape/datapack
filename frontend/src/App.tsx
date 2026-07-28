import {useState} from 'react';
import {DataEditor, GridCellKind, type GridCell, type GridColumn, type Item} from '@glideapps/glide-data-grid';
import '@glideapps/glide-data-grid/dist/index.css';
import './App.css';
import {AddColumn, ClearCell, CreateRow, DeleteColumn, DeleteRow, Export, GetTable, MoveColumn, RenameColumn, RenameRow, SelectProject, SetCellFormula, UpdateCell} from "../wailsjs/go/main/App";

type TreeNode = {
    kind: "project" | "group" | "directory" | "table" | "tableData" | "localSchema" | "schemaFile" | "schemaDefinition" | "damaged";
    name: string;
    path?: string;
    children?: TreeNode[];
};

type WorkspaceView = {
    project: {
        path: string;
        root: string;
    };
    settings: {
        build: {
            mode: "auto" | "manual";
            autoCompileDebounceSeconds?: number;
        };
    };
    tree: {
        root: TreeNode;
    };
    schemas: SchemaIndex;
    foreignKeys: ForeignKeySummary;
};

type SchemaUsage = {
    definition: string;
    path: string;
    table?: string;
    field: string;
};

type SchemaDefinition = {
    id: string;
    kind: "type" | "constraint";
    appliesTo?: string[];
    path: string;
    table?: string;
    usages?: SchemaUsage[];
};

type SchemaDiagnostic = {
    path: string;
    table?: string;
    field: string;
    message: string;
};

type SchemaIndex = {
    definitions: SchemaDefinition[];
    diagnostics?: SchemaDiagnostic[];
};

type ForeignKeySummary = {
    diagnostics?: { rowID: string; column: string; message: string }[];
};

type TableDocument = {
    path: string;
    namespace: string;
    name: string;
    columns: { name: string }[];
    rows: { id: string; values: Record<string, { expression: string; display: string; formula?: boolean }> }[];
    diagnostics?: { rowID: string; column?: string; message: string }[];
};

function Tree({node, depth = 0, onOpenTable}: { node: TreeNode; depth?: number; onOpenTable: (path: string) => void }) {
    const isFile = node.kind === "tableData" || node.kind === "schemaFile" || node.kind === "schemaDefinition" || node.kind === "damaged";
    const openTable = node.kind === "tableData" && node.path ? () => onOpenTable(node.path!) : undefined;
    return (
        <div className="tree-node">
            <button className={`tree-row ${isFile ? "tree-file" : ""} ${openTable ? "tree-action" : ""}`} style={{paddingLeft: `${12 + depth * 16}px`}} onClick={openTable} disabled={!openTable}>
                <span className={`node-icon node-${node.kind}`} aria-hidden="true"/>
                <span>{node.name}</span>
            </button>
            {node.children?.map((child) => <Tree key={`${child.kind}:${child.path ?? child.name}`} node={child} depth={depth + 1} onOpenTable={onOpenTable}/>) }
        </div>
    );
}

type CellError = { rowID: string; column: string; message: string };
type DraftRow = { id: string; values: Record<string, string> };

function TableGrid({table, onDeleteColumn, onDeleteRow, onMoveColumn, onRenameColumn, onSaveInput}: { table: TableDocument; onDeleteColumn: () => void; onDeleteRow: (rowID: string) => void; onMoveColumn: () => void; onRenameColumn: () => void; onSaveInput: (rowID: string, column: string, input: string) => Promise<void> }) {
    const [draft, setDraft] = useState<DraftRow>({id: "", values: {}});
    const [draftError, setDraftError] = useState("");
    const [contextRow, setContextRow] = useState<number | null>(null);
    const [menu, setMenu] = useState<{x: number; y: number} | null>(null);
    const columns: GridColumn[] = [
        {id: "id", title: "ID", width: 160, style: "highlight"},
        ...table.columns.map((column) => ({id: column.name, title: column.name, width: 170})),
    ];
    const isDraft = (row: number) => row === table.rows.length;
    const rowFor = (row: number): TableDocument["rows"][number] => isDraft(row) ? {id: draft.id, values: Object.fromEntries(Object.entries(draft.values).map(([name, expression]) => [name, {expression, display: expression, formula: false}]))} : table.rows[row];
    const errorFor = (rowID: string, column: string) => table.diagnostics?.find((diagnostic) => diagnostic.rowID === rowID && (!diagnostic.column || diagnostic.column === column));
    const getCellContent = ([column, row]: Item): GridCell => {
        const source = rowFor(row);
        const cell = column === 0 ? {display: source.id, expression: source.id} : source.values[table.columns[column - 1].name];
        const error = isDraft(row) ? draftError : errorFor(source.id, table.columns[column - 1]?.name ?? "");
        return {
            kind: GridCellKind.Text,
            data: cell?.formula ? `=${cell.expression}` : cell?.expression ?? "",
            displayData: cell?.formula ? `fx ${cell.display}` : cell?.display ?? "",
            allowOverlay: true,
            themeOverride: error ? {
                bgCell: "#fff1f1",
                textDark: "#8a3030",
            } : isDraft(row) ? {bgCell: "#f7fafb", textDark: "#51616b"} : undefined,
        };
    };
    const editCell = async ([column, row]: Item, value: GridCell) => {
        if (value.kind !== GridCellKind.Text) {
            return;
        }
        const input = value.data.trim();
        if (isDraft(row)) {
            const next = column === 0 ? {id: input, values: draft.values} : {id: draft.id, values: {...draft.values, [table.columns[column - 1].name]: input}};
            if (column !== 0 && input === "") delete next.values[table.columns[column - 1].name];
            const duplicate = table.rows.some((existing) => existing.id === next.id);
            if (next.id !== "" && (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(next.id) || duplicate)) {
                setDraft(next);
                setDraftError(duplicate ? "ID already exists" : "ID must be a valid CUE identifier");
                return;
            }
            const hasValues = Object.values(next.values).some((entry) => entry !== "");
            if (next.id !== "" && hasValues) {
                const values: Record<string, string> = {};
                const formulas: Record<string, string> = {};
                for (const [name, expression] of Object.entries(next.values)) {
                    if (expression.startsWith("=")) formulas[name] = expression.slice(1).trim();
                    else values[name] = expression;
                }
                try {
                    await onSaveInput(next.id, "__create__", JSON.stringify({values, formulas}));
                    setDraft({id: "", values: {}});
                    setDraftError("");
                } catch (reason) {
                    setDraft(next);
                    setDraftError(reason instanceof Error ? reason.message : String(reason));
                }
                return;
            }
            setDraft(next);
            setDraftError("");
            return;
        }
        const source = table.rows[row];
        const columnName = column === 0 ? "id" : table.columns[column - 1].name;
        try {
            await onSaveInput(source.id, columnName, input);
        } catch {
            // The parent reports the rejected source syntax alongside the cell.
        }
    };
    return <div className="table-document">
        <div className="table-document-header">
            <div>
                <div className="table-eyebrow">CUE expression editing</div>
                <h1>{table.namespace}.{table.name}</h1>
            </div>
            <div className="table-actions">
                <button className="table-action" onClick={onRenameColumn}>Rename column</button>
                <button className="table-action" onClick={onMoveColumn}>Move column</button>
                <button className="table-action table-action-danger" onClick={onDeleteColumn}>Delete column</button>
            </div>
            <span>{table.rows.length} rows + draft</span>
        </div>
        {draftError && <div className="table-validation-error"><code>draft</code><span>{draftError}</span></div>}
        <div className="table-grid">
            <DataEditor
                columns={columns}
                getCellContent={getCellContent}
                rows={table.rows.length + 1}
                rowMarkers="clickable-number"
                rowHeight={30}
                headerHeight={34}
                smoothScrollX
                smoothScrollY
                onCellEdited={editCell}
                cellActivationBehavior="double-click"
                onCellContextMenu={([column, row], event) => {
                    if (row === table.rows.length || column !== 0) return;
                    event.preventDefault();
                    setContextRow(row);
                    setMenu({x: event.localEventX, y: event.localEventY});
                }}
            />
            {menu !== null && contextRow !== null && <div className="row-context-menu" style={{left: menu.x, top: menu.y}}>
                <button className="row-context-danger" onClick={() => {
                    onDeleteRow(table.rows[contextRow].id);
                    setMenu(null);
                    setContextRow(null);
                }}>Delete row</button>
            </div>}
        </div>
    </div>;
}

function App() {
    const [workspace, setWorkspace] = useState<WorkspaceView | null>(null);
    const [selectedSchema, setSelectedSchema] = useState<SchemaDefinition | null>(null);
    const [table, setTable] = useState<TableDocument | null>(null);
    const [cellError, setCellError] = useState<CellError | null>(null);
    const [error, setError] = useState("");
    const [notice, setNotice] = useState("");

    async function selectProject() {
        setError("");
        try {
            const view = await SelectProject() as WorkspaceView;
            if (view?.project?.path) {
                setWorkspace(view);
                setSelectedSchema(view.schemas.definitions[0] ?? null);
                setTable(null);
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
            setCellError(null);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
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

    async function addColumn() {
        if (!table) {
            return;
        }
        const name = window.prompt("Column name");
        if (!name) {
            return;
        }
        const schemaExpression = window.prompt("Column CUE Schema", "string");
        if (!schemaExpression) {
            return;
        }
        const initialExpression = window.prompt("Initial CUE value for every row", `""`);
        if (!initialExpression) {
            return;
        }
        setError("");
        setCellError(null);
        try {
            const result = await AddColumn({path: table.path, name, schemaExpression, initialExpression});
            setTable(result.table as TableDocument);
            setWorkspace(result.workspace as WorkspaceView);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function deleteColumn() {
        if (!table) return;
        const name = window.prompt("Column name to delete");
        if (!name || !window.confirm(`Delete column ${name} and all of its values?`)) return;
        setError("");
        setCellError(null);
        try {
            const result = await DeleteColumn({path: table.path, name});
            setTable(result.table as TableDocument);
            setWorkspace(result.workspace as WorkspaceView);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function renameColumn() {
        if (!table) return;
        const oldName = window.prompt("Current column name");
        if (!oldName) return;
        const newName = window.prompt("New column name", oldName);
        if (!newName || newName === oldName) return;
        setError("");
        setCellError(null);
        try {
            const result = await RenameColumn({path: table.path, oldName, newName});
            setTable(result.table as TableDocument);
            setWorkspace(result.workspace as WorkspaceView);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function moveColumn() {
        if (!table) return;
        const name = window.prompt("Column name to move");
        if (!name) return;
        const positionText = window.prompt(`New zero-based position (0-${table.columns.length - 1})`, "0");
        if (positionText === null) return;
        const position = Number(positionText);
        if (!Number.isInteger(position)) {
            setError("Column position must be an integer");
            return;
        }
        setError("");
        setCellError(null);
        try {
            const result = await MoveColumn({path: table.path, name, position});
            setTable(result.table as TableDocument);
            setWorkspace(result.workspace as WorkspaceView);
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

    async function renameRow(oldID: string) {
        if (!table) {
            return;
        }
        const newID = window.prompt("New row ID", oldID);
        if (!newID || newID === oldID) {
            return;
        }
        setError("");
        setCellError(null);
        try {
            const result = await RenameRow({path: table.path, oldID, newID});
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

    return (
        <main className="app-shell">
            <header className="app-header">
                <div className="brand">Datapack</div>
                <div className="header-divider"/>
                <span className="project-name">{workspace?.tree.root.name ?? "No project open"}</span>
                <div className="header-actions">
                    {workspace && <span className={`build-mode build-mode-${workspace.settings.build.mode}`}>{workspace.settings.build.mode === "auto" ? "Auto compile" : "Manual compile"}</span>}
                    {workspace && <button className="export-button" onClick={exportProject}>Export CSV</button>}
                    <button className="open-button" onClick={selectProject}>Open project</button>
                </div>
            </header>

            {workspace ? (
                <section className="workbench">
                    <aside className="project-explorer">
                        <div className="panel-title">Project</div>
                        <div className="tree"><Tree node={workspace.tree.root} onOpenTable={openTable}/></div>
                    </aside>
                    <section className="editor-placeholder">
                        {table ? <TableGrid table={table} onDeleteColumn={deleteColumn} onDeleteRow={deleteRow} onMoveColumn={moveColumn} onRenameColumn={renameColumn} onSaveInput={saveCellInput}/> : selectedSchema ? <>
                            <div className="schema-kind">{selectedSchema.kind}</div>
                            <h1>{selectedSchema.id}</h1>
                            <p>{selectedSchema.path}</p>
                            {selectedSchema.appliesTo && <div className="schema-applies">{selectedSchema.appliesTo.join(", ")}</div>}
                            <div className="usage-list">
                                {(selectedSchema.usages ?? []).map((usage) => <div className="usage-row" key={`${usage.path}:${usage.field}`}>
                                    <span>{usage.table ?? usage.path}</span>
                                    <code>{usage.field}</code>
                                </div>)}
                            </div>
                        </> : <>
                            <h1>No Schema definitions</h1>
                        </>}
                    </section>
                    <aside className="schema-explorer">
                        <div className="panel-title">Schemas</div>
                        <div className="schema-list">
                            {workspace.schemas.definitions.map((definition) => <button
                                className={`schema-row ${selectedSchema?.id === definition.id ? "schema-row-selected" : ""}`}
                                key={definition.id}
                                onClick={() => { setSelectedSchema(definition); setTable(null); setCellError(null); }}
                            >
                                <span className={`schema-kind-dot schema-kind-${definition.kind}`} aria-hidden="true"/>
                                <span>{definition.id}</span>
                                <span className="usage-count">{definition.usages?.length ?? 0}</span>
                            </button>)}
                        </div>
                        {(workspace.schemas.diagnostics?.length ?? 0) > 0 && <div className="schema-diagnostics">
                            {workspace.schemas.diagnostics?.map((diagnostic) => <div className="schema-diagnostic" key={`${diagnostic.path}:${diagnostic.field}:${diagnostic.message}`}>
                                <code>{diagnostic.field}</code>
                                <span>{diagnostic.message}</span>
                            </div>)}
                        </div>}
                        {(workspace.foreignKeys.diagnostics?.length ?? 0) > 0 && <div className="schema-diagnostics">
                            {workspace.foreignKeys.diagnostics?.map((diagnostic) => <div className="schema-diagnostic" key={`${diagnostic.rowID}:${diagnostic.column}:${diagnostic.message}`}>
                                <code>{diagnostic.rowID}.{diagnostic.column}</code>
                                <span>{diagnostic.message}</span>
                            </div>)}
                        </div>}
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
            {error && <div className="error-toast" role="alert">{error}</div>}
            {notice && <div className="notice-toast" role="status">{notice}</div>}
        </main>
    )
}

export default App
