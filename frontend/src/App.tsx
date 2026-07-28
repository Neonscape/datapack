import {useState} from 'react';
import {DataEditor, drawTextCell, GridCellKind, type CustomCell, type CustomRenderer, type GridCell, type GridColumn, type Item} from '@glideapps/glide-data-grid';
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
    columns: { name: string; schema: { kind: string; enumValues?: { expression: string }[] } }[];
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
type EnumCell = CustomCell<{kind: "enum"; value: string; options: string[]}>;

const enumRenderer: CustomRenderer<EnumCell> = {
    kind: GridCellKind.Custom,
    isMatch: (cell): cell is EnumCell => (cell as EnumCell).data.kind === "enum",
    draw: (args, cell) => drawTextCell(args, cell.data.value),
    provideEditor: (cell) => ({editor: (props) => <select autoFocus value={props.value.data.value} onChange={(event) => props.onChange({...props.value, data: {...props.value.data, value: event.target.value}})} onBlur={() => props.onFinishedEditing(props.value)}>
        {cell.data.options.map((option) => <option key={option} value={option}>{option}</option>)}
    </select>, disablePadding: true}),
};

function TableGrid({table, onAddColumn, onDeleteColumn, onDeleteRow, onMoveColumn, onRenameColumn, onSaveInput}: { table: TableDocument; onAddColumn: () => void; onDeleteColumn: () => void; onDeleteRow: (rowID: string) => void; onMoveColumn: () => void; onRenameColumn: () => void; onSaveInput: (rowID: string, column: string, input: string) => Promise<void> }) {
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
        const columnDefinition = table.columns[column - 1];
        if (column !== 0 && columnDefinition.schema.kind === "enum" && !cell?.formula) {
            const options = columnDefinition.schema.enumValues?.map((entry) => entry.expression.replace(/^"|"$/g, "")) ?? [];
            return {kind: GridCellKind.Custom, allowOverlay: true, copyData: cell?.display ?? "", data: {kind: "enum", value: cell?.display ?? options[0] ?? "", options}, themeOverride: error ? {bgCell: "#fff1f1", textDark: "#8a3030"} : isDraft(row) ? {bgCell: "#f7fafb", textDark: "#51616b"} : undefined} as EnumCell;
        }
        if (column !== 0 && columnDefinition.schema.kind === "boolean" && !cell?.formula) {
            return {kind: GridCellKind.Boolean, allowOverlay: false, data: cell?.expression === "true", themeOverride: error ? {bgCell: "#fff1f1", textDark: "#8a3030"} : isDraft(row) ? {bgCell: "#f7fafb", textDark: "#51616b"} : undefined};
        }
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
        const enumValue = value as EnumCell;
        const input = value.kind === GridCellKind.Custom && enumValue.data.kind === "enum" ? JSON.stringify(enumValue.data.value) : value.kind === GridCellKind.Boolean ? value.data ? "true" : "false" : value.kind === GridCellKind.Text ? value.data.trim() : null;
        if (input === null) return;
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
                <button className="table-action" onClick={onAddColumn}>Add column</button>
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
                customRenderers={[enumRenderer]}
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

type StructureDialog = {
    mode: "add" | "rename" | "move" | "delete";
    name: string;
    second: string;
    schema: string;
    initial: string;
};

function App() {
    const [workspace, setWorkspace] = useState<WorkspaceView | null>(null);
    const [selectedSchema, setSelectedSchema] = useState<SchemaDefinition | null>(null);
    const [table, setTable] = useState<TableDocument | null>(null);
    const [cellError, setCellError] = useState<CellError | null>(null);
    const [error, setError] = useState("");
    const [notice, setNotice] = useState("");
    const [dialog, setDialog] = useState<StructureDialog | null>(null);

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

    async function submitStructureDialog() {
        if (!table || !dialog) return;
        setError("");
        setCellError(null);
        try {
            let result;
            if (dialog.mode === "add") {
                result = await AddColumn({path: table.path, name: dialog.name, schemaExpression: dialog.schema, initialExpression: dialog.initial});
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
                        {table ? <TableGrid table={table} onAddColumn={() => setDialog({mode: "add", name: "", second: "", schema: "string", initial: `""`})} onDeleteColumn={() => setDialog({mode: "delete", name: "", second: "", schema: "", initial: ""})} onDeleteRow={deleteRow} onMoveColumn={() => setDialog({mode: "move", name: "", second: "0", schema: "", initial: ""})} onRenameColumn={() => setDialog({mode: "rename", name: "", second: "", schema: "", initial: ""})} onSaveInput={saveCellInput}/> : selectedSchema ? <>
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
            {dialog && <div className="dialog-backdrop" role="presentation" onMouseDown={() => setDialog(null)}>
                <form className="structure-dialog" onMouseDown={(event) => event.stopPropagation()} onSubmit={(event) => { event.preventDefault(); submitStructureDialog(); }}>
                    <h2>{dialog.mode === "add" ? "Add column" : dialog.mode === "rename" ? "Rename column" : dialog.mode === "move" ? "Move column" : "Delete column"}</h2>
                    <label>Column name<input autoFocus value={dialog.name} onChange={(event) => setDialog({...dialog, name: event.target.value})}/></label>
                    {dialog.mode === "add" && <>
                        <label>CUE Schema<input value={dialog.schema} onChange={(event) => setDialog({...dialog, schema: event.target.value})}/></label>
                        <label>Initial CUE value<input value={dialog.initial} onChange={(event) => setDialog({...dialog, initial: event.target.value})}/></label>
                    </>}
                    {dialog.mode === "rename" && <label>New name<input value={dialog.second} onChange={(event) => setDialog({...dialog, second: event.target.value})}/></label>}
                    {dialog.mode === "move" && <label>New zero-based position<input type="number" min="0" value={dialog.second} onChange={(event) => setDialog({...dialog, second: event.target.value})}/></label>}
                    {dialog.mode === "delete" && <p>This removes the column, all row values, formulas, foreign-key metadata, and CSV export configuration that refer to it.</p>}
                    <div className="dialog-actions"><button type="button" onClick={() => setDialog(null)}>Cancel</button><button className={dialog.mode === "delete" ? "dialog-danger" : ""} type="submit">{dialog.mode === "delete" ? "Delete column" : "Apply"}</button></div>
                </form>
            </div>}
        </main>
    )
}

export default App
