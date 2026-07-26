import {useState} from 'react';
import {DataEditor, GridCellKind, type GridCell, type GridColumn, type Item} from '@glideapps/glide-data-grid';
import '@glideapps/glide-data-grid/dist/index.css';
import './App.css';
import {AddColumn, AddRow, DeleteColumn, DeleteRow, Export, GetTable, RenameRow, SelectProject, UpdateCell} from "../wailsjs/go/main/App";

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
    rows: { id: string; values: Record<string, { expression: string; display: string }> }[];
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

function TableGrid({table, error, onAddColumn, onAddRow, onDeleteColumn, onDeleteRow, onRenameRow, onUpdate}: { table: TableDocument; error: CellError | null; onAddColumn: () => void; onAddRow: () => void; onDeleteColumn: () => void; onDeleteRow: (rowID: string) => void; onRenameRow: (rowID: string) => void; onUpdate: (rowID: string, column: string, expression: string) => void }) {
    const [selectedRow, setSelectedRow] = useState<number | null>(null);
    const columns: GridColumn[] = [
        {id: "id", title: "ID", width: 160, style: "highlight"},
        ...table.columns.map((column) => ({id: column.name, title: column.name, width: 170})),
    ];
    const getCellContent = ([column, row]: Item): GridCell => {
        const source = table.rows[row];
        const cell = column === 0 ? {display: source.id, expression: source.id} : source.values[table.columns[column - 1].name];
        return {
            kind: GridCellKind.Text,
            data: cell?.expression ?? "",
            displayData: cell?.display ?? "",
            allowOverlay: column !== 0,
            readonly: column === 0,
            themeOverride: error?.rowID === source.id && error.column === table.columns[column - 1]?.name ? {
                bgCell: "#fff1f1",
                textDark: "#8a3030",
            } : undefined,
        };
    };
    const editCell = ([column, row]: Item, value: GridCell) => {
        if (column === 0 || value.kind !== GridCellKind.Text) {
            return;
        }
        onUpdate(table.rows[row].id, table.columns[column - 1].name, value.data);
    };
    return <div className="table-document">
        <div className="table-document-header">
            <div>
                <div className="table-eyebrow">CUE expression editing</div>
                <h1>{table.namespace}.{table.name}</h1>
            </div>
            <div className="table-actions">
                <button className="table-action" onClick={onAddColumn}>Add column</button>
                <button className="table-action table-action-danger" onClick={onDeleteColumn}>Delete column</button>
                <button className="table-action" onClick={onAddRow}>Add row</button>
                <button className="table-action" disabled={selectedRow === null} onClick={() => {
                    if (selectedRow !== null) {
                        onRenameRow(table.rows[selectedRow].id);
                    }
                }}>Rename ID</button>
                <button className="table-action table-action-danger" disabled={selectedRow === null} onClick={() => {
                    if (selectedRow !== null) {
                        onDeleteRow(table.rows[selectedRow].id);
                    }
                }}>Delete row</button>
            </div>
            <span>{table.rows.length} rows</span>
        </div>
        {error && <div className="table-validation-error"><code>{error.column}</code><span>{error.message}</span></div>}
        <div className="table-grid">
            <DataEditor
                columns={columns}
                getCellContent={getCellContent}
                rows={table.rows.length}
                rowMarkers="clickable-number"
                rowHeight={30}
                headerHeight={34}
                smoothScrollX
                smoothScrollY
                onCellEdited={editCell}
                onCellClicked={([, row]) => setSelectedRow(row)}
            />
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

    async function updateCell(rowID: string, column: string, expression: string) {
        if (!table) {
            return;
        }
        setError("");
        setCellError(null);
        try {
            const result = await UpdateCell({path: table.path, rowID, column, expression});
            setTable(result.table as TableDocument);
            setWorkspace(result.workspace as WorkspaceView);
        } catch (reason) {
            const message = reason instanceof Error ? reason.message : String(reason);
            setError(message);
            setCellError({rowID, column, message});
        }
    }

    async function addRow() {
        if (!table) {
            return;
        }
        const rowID = window.prompt("Row ID");
        if (!rowID) {
            return;
        }
        const expression = window.prompt("Row CUE struct", `{${table.columns.map((column) => `${column.name}: `).join(", ")}}`);
        if (!expression) {
            return;
        }
        setError("");
        setCellError(null);
        try {
            const result = await AddRow({path: table.path, rowID, expression});
            setTable(result.table as TableDocument);
            setWorkspace(result.workspace as WorkspaceView);
        } catch (reason) {
            const message = reason instanceof Error ? reason.message : String(reason);
            setError(message);
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

    async function deleteRow(rowID: string) {
        if (!table || !window.confirm(`Delete row ${rowID}?`)) {
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
                        {table ? <TableGrid table={table} error={cellError} onAddColumn={addColumn} onAddRow={addRow} onDeleteColumn={deleteColumn} onDeleteRow={deleteRow} onRenameRow={renameRow} onUpdate={updateCell}/> : selectedSchema ? <>
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
