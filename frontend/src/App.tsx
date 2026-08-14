import {useEffect, useRef, useState} from 'react';
import {CompactSelection, DataEditor, drawTextCell, GridCellKind, type CustomCell, type CustomRenderer, type DataEditorRef, type GridCell, type GridColumn, type GridSelection, type Item} from '@glideapps/glide-data-grid';
import '@glideapps/glide-data-grid/dist/index.css';
import './App.css';
import {AddColumn, ClearCell, Compile, CreateRow, CreateSchemaDefinition, CreateTable, DeleteColumn, DeleteRow, DeleteSchemaDefinition, Export, GetPublicSchema, GetTable, MoveColumn, Paste, Redo, RenameColumn, RenameRow, RenameSchemaDefinition, SelectProject, SetCellFormula, SetColumnSchema, SetCompileMode, SetSchemaDefinition, Undo, UpdateCell} from "../wailsjs/go/main/App";
import {EventsOn} from "../wailsjs/runtime/runtime";

type LogEntry = {
    time: string;
    level: "debug" | "info" | "warn" | "error";
    message: string;
    fields?: Record<string, unknown>;
};

function LogPanel() {
    const [entries, setEntries] = useState<LogEntry[]>([]);
    const [open, setOpen] = useState(false);
    useEffect(() => {
        const append = (entry: LogEntry) => {
            setEntries((current) => [...current, entry].slice(-500));
        };
        if ((window as Window & {runtime?: {EventsOn?: unknown}}).runtime?.EventsOn) {
            return EventsOn("log:entry", append);
        }
        const fallback = (event: Event) => append((event as CustomEvent<LogEntry>).detail);
        window.addEventListener("log:entry", fallback);
        return () => window.removeEventListener("log:entry", fallback);
    }, []);
    return <section className={`log-panel ${open ? "log-panel-open" : ""}`}>
        <button className="log-panel-toggle" onClick={() => setOpen(!open)} aria-expanded={open}>
            Logs <span>{entries.length}</span>
        </button>
        {open && <div className="log-panel-body">
            <div className="log-panel-toolbar"><span>Application log</span><button onClick={() => setEntries([])}>Clear</button></div>
            <div className="log-list">
                {entries.map((entry, index) => <div className={`log-entry log-${entry.level}`} key={`${entry.time}-${index}`}>
                    <time>{new Date(entry.time).toLocaleTimeString()}</time>
                    <strong>{entry.level.toUpperCase()}</strong>
                    <span>{entry.message}</span>
                    {entry.fields && <code>{JSON.stringify(entry.fields)}</code>}
                </div>)}
                {entries.length === 0 && <div className="log-empty">No log entries yet</div>}
            </div>
        </div>}
    </section>;
}

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
    compile: {
        mode: "auto" | "manual";
        pending: boolean;
        compiledAt?: string;
        diagnostics: {path: string; table?: string; field?: string; rowID?: string; column?: string; message: string}[];
    };
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
    columns: { name: string; optional?: boolean; schema: { kind: string; enumValues?: { expression: string }[]; fields?: {name: string; optional?: boolean}[] }; schemaExpression: string; foreignKeyTable?: string; foreignKeyOptions?: string[] }[];
    rows: { id: string; values: Record<string, { expression: string; display: string; formula?: boolean }> }[];
    diagnostics?: { rowID: string; column?: string; message: string }[];
    localSchema: SchemaDefinitionDocument[];
};

type SchemaDefinitionDocument = {
    name: string;
    kind: "type" | "constraint";
    type?: string;
    appliesTo?: string[];
    schema: { kind: string };
    schemaExpression: string;
};

type PublicSchemaDocument = {
    path: string;
    package: string;
    definitions: SchemaDefinitionDocument[];
};

type SchemaEditorState = {
    path: string;
    local: boolean;
    package?: string;
    logicalTable?: string;
    definitions: SchemaDefinitionDocument[];
    selected: string | null;
};

type SchemaEditResult = {
    schema: PublicSchemaDocument;
    table: TableDocument;
    workspace: WorkspaceView;
};

function schemaDefinitions(workspace: WorkspaceView | null): SchemaDefinition[] {
    return workspace?.schemas?.definitions ?? [];
}

function schemaReferenceOptions(workspace: WorkspaceView | null): {label: string; value: string; kind: "type" | "constraint"}[] {
    const options: {label: string; value: string; kind: "type" | "constraint"}[] = [];
    for (const definition of schemaDefinitions(workspace)) {
        if (definition.id.includes("._local.#")) {
            const name = definition.id.split(".#")[1];
            options.push({label: `#${name}`, value: `#${name}`, kind: definition.kind});
        } else {
            options.push({label: definition.id, value: definition.id, kind: definition.kind});
        }
    }
    options.sort((left, right) => left.label.localeCompare(right.label));
    return options;
}

function Tree({node, depth = 0, onOpenTable, onOpenSchema}: { node: TreeNode; depth?: number; onOpenTable: (path: string) => void; onOpenSchema: (node: TreeNode) => void }) {
    const isFile = node.kind === "tableData" || node.kind === "schemaFile" || node.kind === "schemaDefinition" || node.kind === "damaged";
    const isSchemaNode = node.kind === "schemaFile" || node.kind === "schemaDefinition" || node.kind === "localSchema";
    const openTable = node.kind === "tableData" && node.path ? () => onOpenTable(node.path!) : undefined;
    const openSchema = isSchemaNode && node.path ? () => onOpenSchema(node) : undefined;
    const openAction = openTable ?? openSchema;
    return (
        <div className="tree-node">
            <button className={`tree-row ${isFile ? "tree-file" : ""} ${openAction ? "tree-action" : ""}`} style={{paddingLeft: `${12 + depth * 16}px`}} onClick={openAction} disabled={!openAction}>
                <span className={`node-icon node-${node.kind}`} aria-hidden="true"/>
                <span>{node.name}</span>
            </button>
            {node.children?.map((child) => <Tree key={`${child.kind}:${child.path ?? child.name}:${child.name}`} node={child} depth={depth + 1} onOpenTable={onOpenTable} onOpenSchema={onOpenSchema}/>) }
        </div>
    );
}

type CellError = { rowID: string; column: string; message: string };
type DraftRow = { id: string; values: Record<string, string> };
type EnumCell = CustomCell<{kind: "enum"; value: string; options: string[]}>;
type InlineEditor = { column: number; row: number; value: string; bounds: {x: number; y: number; width: number; height: number} };
type ForeignKeyEditor = InlineEditor & {options: string[]; target: string};
type ArrayEditor = {column: number; row: number; entries: string[]; pasted: string};
type StructEditor = {column: number; row: number; values: Record<string, string>};

function arrayEntries(expression: string) {
    const source = expression.trim();
    if (!source.startsWith("[") || !source.endsWith("]")) return [];
    const entries: string[] = [];
    let start = 1;
    let depth = 0;
    let quote = "";
    let escaped = false;
    for (let index = 1; index < source.length - 1; index++) {
        const character = source[index];
        if (quote) {
            if (!escaped && character === quote) quote = "";
            escaped = !escaped && character === "\\";
            continue;
        }
        if (character === `"` || character === "'") quote = character;
        else if (character === "[" || character === "{") depth++;
        else if (character === "]" || character === "}") depth--;
        else if (character === "," && depth === 0) {
            const entry = source.slice(start, index).trim();
            if (entry) entries.push(entry);
            start = index + 1;
        }
    }
    const entry = source.slice(start, -1).trim();
    if (entry) entries.push(entry);
    return entries;
}

function structValues(expression: string, fieldNames: string[]) {
    const source = expression.trim();
    const values: Record<string, string> = {};
    if (!source.startsWith("{") || !source.endsWith("}")) return values;
    let start = 1;
    let depth = 0;
    let quote = "";
    let escaped = false;
    const entries: string[] = [];
    for (let index = 1; index < source.length - 1; index++) {
        const character = source[index];
        if (quote) {
            if (!escaped && character === quote) quote = "";
            escaped = !escaped && character === "\\";
            continue;
        }
        if (character === `"` || character === "'") quote = character;
        else if (character === "[" || character === "{") depth++;
        else if (character === "]" || character === "}") depth--;
        else if ((character === "," || character === "\n") && depth === 0) {
            entries.push(source.slice(start, index).trim());
            start = index + 1;
        }
    }
    entries.push(source.slice(start, -1).trim());
    for (const entry of entries) {
        const separator = entry.indexOf(":");
        if (separator < 0) continue;
        const name = entry.slice(0, separator).trim();
        if (fieldNames.includes(name)) values[name] = entry.slice(separator + 1).trim();
    }
    return values;
}

const enumRenderer: CustomRenderer<EnumCell> = {
    kind: GridCellKind.Custom,
    isMatch: (cell): cell is EnumCell => (cell as EnumCell).data.kind === "enum",
    draw: (args, cell) => drawTextCell(args, cell.data.value),
    provideEditor: (cell) => ({editor: (props) => <select autoFocus value={props.value.data.value} onChange={(event) => props.onChange({...props.value, data: {...props.value.data, value: event.target.value}})} onBlur={() => props.onFinishedEditing(props.value)}>
        {cell.data.options.map((option) => <option key={option} value={option}>{option}</option>)}
    </select>, disablePadding: true}),
};

function TableGrid({table, onAddColumn, onDeleteColumn, onDeleteRow, onEditColumnSchema, onMoveColumn, onRenameColumn, onPaste, onSaveInput}: { table: TableDocument; onAddColumn: (position: number) => void; onDeleteColumn: (column: string) => void; onDeleteRow: (rowID: string) => void; onEditColumnSchema: (column: string, schemaExpression: string) => void; onMoveColumn: (column: string) => void; onRenameColumn: (column: string) => void; onPaste: (row: number, column: number, text: string) => Promise<void>; onSaveInput: (rowID: string, column: string, input: string) => Promise<void> }) {
    const [draft, setDraft] = useState<DraftRow>({id: "", values: {}});
    const [draftError, setDraftError] = useState("");
    const [contextRow, setContextRow] = useState<number | null>(null);
    const [menu, setMenu] = useState<{x: number; y: number} | null>(null);
    const [columnMenu, setColumnMenu] = useState<{column: number; x: number; y: number} | null>(null);
    const [selectedColumn, setSelectedColumn] = useState<number | null>(null);
    const [gridSelection, setGridSelection] = useState<GridSelection>({columns: CompactSelection.empty(), rows: CompactSelection.empty(), current: undefined});
    const [inlineEditor, setInlineEditor] = useState<InlineEditor | null>(null);
    const [foreignKeyEditor, setForeignKeyEditor] = useState<ForeignKeyEditor | null>(null);
    const [arrayEditor, setArrayEditor] = useState<ArrayEditor | null>(null);
    const [structEditor, setStructEditor] = useState<StructEditor | null>(null);
    const [selectedCell, setSelectedCell] = useState<{column: number; row: number} | null>(null);
    const [formulaEditing, setFormulaEditing] = useState(false);
    const [formulaValue, setFormulaValue] = useState("");
    const editorRef = useRef<HTMLTextAreaElement | null>(null);
    const foreignKeyEditorRef = useRef<HTMLInputElement | null>(null);
    const formulaBarRef = useRef<HTMLInputElement | null>(null);
    const dataEditorRef = useRef<DataEditorRef | null>(null);
    const gridRef = useRef<HTMLDivElement | null>(null);
    const lastCellBounds = useRef<{cell: Item; bounds: {x: number; y: number; width: number; height: number}} | null>(null);
    const columns: GridColumn[] = [
        {id: "id", title: "ID", width: 160, style: "highlight"},
        ...table.columns.map((column) => ({id: column.name, title: column.name, width: 170})),
    ];
    const isDraft = (row: number) => row === table.rows.length;
    const selectedColumnName = selectedColumn === null ? null : table.columns[selectedColumn - 1]?.name ?? null;
    const columnWidth = (column: GridColumn) => "width" in column ? column.width : 170;
    useEffect(() => {
        const editor = inlineEditor ? editorRef.current : foreignKeyEditorRef.current;
        editor?.focus();
        editor?.select();
    }, [inlineEditor, foreignKeyEditor]);
    useEffect(() => {
        const closeContextMenu = (event: PointerEvent) => {
            const target = event.target as Element | null;
            if (!target?.closest(".row-context-menu, .column-context-menu")) {
	                setMenu(null);
	                setContextRow(null);
	                setColumnMenu(null);
            }
        };
        const closeOnEscape = (event: KeyboardEvent) => {
            if (event.key === "Escape") {
                setMenu(null);
                setContextRow(null);
	                setColumnMenu(null);
            }
        };
        document.addEventListener("pointerdown", closeContextMenu);
        document.addEventListener("keydown", closeOnEscape);
        return () => {
            document.removeEventListener("pointerdown", closeContextMenu);
            document.removeEventListener("keydown", closeOnEscape);
        };
    }, []);
    const rowFor = (row: number): TableDocument["rows"][number] => isDraft(row) ? {id: draft.id, values: Object.fromEntries(Object.entries(draft.values).map(([name, expression]) => [name, {expression, display: expression, formula: false}]))} : table.rows[row];
    const rawCellValue = (column: number, row: number): string => {
        const source = rowFor(row);
        if (column === 0) return source.id;
        const cell = source.values[table.columns[column - 1].name];
        return cell?.formula ? `=${cell.expression}` : cell?.expression ?? "";
    };
    const selectedCellLabel = selectedCell ? `${selectedCell.column === 0 ? "id" : table.columns[selectedCell.column - 1].name} · ${rowFor(selectedCell.row).id || "draft"}` : "";
    useEffect(() => {
        if (formulaEditing || !selectedCell) return;
        setFormulaValue(rawCellValue(selectedCell.column, selectedCell.row));
    }, [selectedCell, table, draft, formulaEditing]);
    const errorFor = (rowID: string, column: string) => table.diagnostics?.find((diagnostic) => diagnostic.rowID === rowID && (!diagnostic.column || diagnostic.column === column));
    const focusDiagnostic = (diagnostic: NonNullable<TableDocument["diagnostics"]>[number]) => {
        const row = table.rows.findIndex((entry) => entry.id === diagnostic.rowID);
        const column = diagnostic.column ? table.columns.findIndex((entry) => entry.name === diagnostic.column) + 1 : 0;
        if (row < 0 || column < 0) return;
        setGridSelection({columns: CompactSelection.fromSingleSelection(column), rows: CompactSelection.fromSingleSelection(row), current: {cell: [column, row], range: {x: column, y: row, width: 1, height: 1}, rangeStack: []}});
        dataEditorRef.current?.scrollTo(column, row, "both", 24, 48, {hAlign: "center", vAlign: "center"});
        dataEditorRef.current?.focus();
    };
    const getCellContent = ([column, row]: Item): GridCell => {
        const source = rowFor(row);
        const cell = column === 0 ? {display: source.id, expression: source.id} : source.values[table.columns[column - 1].name];
        const error = isDraft(row) ? draftError : errorFor(source.id, table.columns[column - 1]?.name ?? "");
        const columnDefinition = table.columns[column - 1];
        if (column !== 0 && ["array", "struct"].includes(columnDefinition.schema.kind) && !cell?.formula) {
            return {kind: GridCellKind.Text, data: cell?.expression ?? "[]", displayData: cell?.display ?? "[]", allowOverlay: false, themeOverride: error ? {bgCell: "#fff1f1", textDark: "#8a3030"} : isDraft(row) ? {bgCell: "#f7fafb", textDark: "#51616b"} : undefined};
        }
        if (column !== 0 && columnDefinition.foreignKeyTable && !cell?.formula) {
            return {kind: GridCellKind.Text, data: cell?.expression ?? "", displayData: cell?.display ?? "", allowOverlay: false, themeOverride: error ? {bgCell: "#fff1f1", textDark: "#8a3030"} : isDraft(row) ? {bgCell: "#f7fafb", textDark: "#51616b"} : undefined};
        }
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
            displayData: cell?.display ?? "",
            allowOverlay: false,
            themeOverride: error ? {
                bgCell: "#fff1f1",
                textDark: "#8a3030",
            } : isDraft(row) ? {bgCell: "#f7fafb", textDark: "#51616b"} : cell?.formula ? {
                bgCell: "#e8f3f5",
                textDark: "#1c5d73",
                baseFontStyle: "700 13px sans-serif",
            } : undefined,
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
    const finishInlineEdit = () => {
        if (!inlineEditor || !editorRef.current) return;
        const value = editorRef.current.value;
        const target = inlineEditor;
        setInlineEditor(null);
        // A blank placeholder cell has no persisted value to clear. Avoid a
        // needless workspace refresh when the user simply dismisses it.
        if (value.trim() === target.value.trim()) return;
        void editCell([target.column, target.row], {kind: GridCellKind.Text, allowOverlay: true, data: value, displayData: value});
    };
    const openInlineAt = (column: number, row: number, bounds: {x: number; y: number; width: number; height: number}) => {
        if (column < 0 || row < 0 || row >= table.rows.length + 1) return;
        const source = rowFor(row);
        const cell = column === 0 ? {expression: source.id, formula: false} : source.values[table.columns[column - 1].name];
        const columnDefinition = table.columns[column - 1];
        if (column !== 0 && columnDefinition.schema.kind === "struct" && !cell?.formula) {
            window.requestAnimationFrame(() => setStructEditor({column, row, values: structValues(cell?.expression ?? "{}", columnDefinition.schema.fields?.map((field) => field.name) ?? [])}));
            return;
        }
        if (column !== 0 && columnDefinition.schema.kind === "array" && !cell?.formula) {
            window.requestAnimationFrame(() => setArrayEditor({column, row, entries: arrayEntries(cell?.expression ?? "[]"), pasted: ""}));
            return;
        }
        if (column !== 0 && columnDefinition.foreignKeyTable && !cell?.formula) {
            window.requestAnimationFrame(() => setForeignKeyEditor({column, row, value: cell?.expression.replace(/^"|"$/g, "") ?? "", bounds, options: columnDefinition.foreignKeyOptions ?? [], target: columnDefinition.foreignKeyTable!}));
            return;
        }
        if (column !== 0 && ["enum", "boolean"].includes(table.columns[column - 1].schema.kind)) return;
        window.setTimeout(() => setInlineEditor({column, row, value: cell?.formula ? `=${cell.expression}` : cell?.expression ?? "", bounds}), 0);
    };
    const finishForeignKeyEdit = (value: string) => {
        if (!foreignKeyEditor) return;
        const target = foreignKeyEditor;
        setForeignKeyEditor(null);
        void editCell([target.column, target.row], {kind: GridCellKind.Text, allowOverlay: true, data: JSON.stringify(value), displayData: value});
    };
    const saveArray = () => {
        if (!arrayEditor) return;
        const target = arrayEditor;
        setArrayEditor(null);
        void editCell([target.column, target.row], {kind: GridCellKind.Text, allowOverlay: true, data: `[${target.entries.filter((entry) => entry.trim()).join(", ")}]`, displayData: ""});
    };
    const saveStruct = () => {
        if (!structEditor) return;
        const target = structEditor;
        const fields = table.columns[target.column - 1].schema.fields ?? [];
        const entries = fields.filter((field) => target.values[field.name]?.trim()).map((field) => `${field.name}: ${target.values[field.name].trim()}`);
        setStructEditor(null);
        void editCell([target.column, target.row], {kind: GridCellKind.Text, allowOverlay: true, data: `{${entries.join(", ")}}`, displayData: ""});
    };
    const selectedRange = () => {
        const current = gridSelection.current?.cell;
        if (!current || current[0] === 0 || current[1] >= table.rows.length) return null;
        const range = gridSelection.current?.range;
        const x = range?.x ?? current[0];
        const y = range?.y ?? current[1];
        const width = range?.width ?? 1;
        const height = range?.height ?? 1;
        if (x === 0 || x + width > table.columns.length + 1 || y + height > table.rows.length) return null;
        return {x, y, width, height};
    };
    useEffect(() => {
        const clipboard = (event: ClipboardEvent) => {
            if ((event.target as Element | null)?.closest("textarea, input, select")) return;
            const range = selectedRange();
            if (!range) return;
            if (event.type === "copy") {
                const text = Array.from({length: range.height}, (_, row) => Array.from({length: range.width}, (_, column) => {
                    const cell = table.rows[range.y + row].values[table.columns[range.x + column - 1].name];
                    return cell?.formula ? `=${cell.expression}` : cell?.expression ?? "";
                }).join("\t")).join("\n");
                event.clipboardData?.setData("text/plain", text);
                event.preventDefault();
            }
            if (event.type === "paste") {
                const text = event.clipboardData?.getData("text/plain");
                if (text === undefined) return;
                event.preventDefault();
                void onPaste(range.y, range.x, text);
            }
        };
        document.addEventListener("copy", clipboard);
        document.addEventListener("paste", clipboard);
        return () => {
            document.removeEventListener("copy", clipboard);
            document.removeEventListener("paste", clipboard);
        };
    }, [gridSelection, table, onPaste]);
    return <div className="table-document" data-inline-editor={inlineEditor || foreignKeyEditor || arrayEditor || structEditor ? "open" : "closed"}>
        <div className="table-document-header">
            <div>
                <div className="table-eyebrow">CUE expression editing</div>
                <h1>{table.namespace}.{table.name}</h1>
            </div>
            <span>{table.rows.length} rows + draft</span>
        </div>
        {draftError && <div className="table-validation-error"><code>draft</code><span>{draftError}</span></div>}
        {(table.diagnostics?.length ?? 0) > 0 && <section className="table-diagnostics" aria-label="Table diagnostics">
            <div className="table-diagnostics-title"><strong>Diagnostics</strong><span>{table.diagnostics!.length}</span></div>
            <div className="table-diagnostics-list">{table.diagnostics!.map((diagnostic, index) => {
                const location = diagnostic.rowID ? diagnostic.column ? `${diagnostic.rowID}.${diagnostic.column}` : diagnostic.rowID : "table";
                const navigable = table.rows.some((row) => row.id === diagnostic.rowID) && (!diagnostic.column || table.columns.some((column) => column.name === diagnostic.column));
                return <button key={`${location}:${diagnostic.message}:${index}`} className="table-diagnostic" disabled={!navigable} onClick={() => focusDiagnostic(diagnostic)}><code>{location}</code><span>{diagnostic.message}</span></button>;
            })}</div>
        </section>}
        <div className="formula-bar">
            <span className="formula-bar-cell">{selectedCellLabel}</span>
            <input
                ref={formulaBarRef}
                className="formula-bar-input"
                aria-label="Formula bar"
                value={formulaValue}
                disabled={!selectedCell}
                readOnly={!formulaEditing}
                placeholder="Select a cell to edit its value"
                onChange={(event) => setFormulaValue(event.target.value)}
                onDoubleClick={() => {
                    if (!selectedCell) return;
                    setInlineEditor(null);
                    setFormulaEditing(true);
                    window.requestAnimationFrame(() => formulaBarRef.current?.select());
                }}
                onBlur={() => {
                    if (!formulaEditing || !selectedCell) return;
                    const target = selectedCell;
                    setFormulaEditing(false);
                    const value = formulaValue.trim();
                    if (value === rawCellValue(target.column, target.row).trim()) return;
                    void editCell([target.column, target.row], {kind: GridCellKind.Text, allowOverlay: true, data: value, displayData: value});
                }}
                onKeyDown={(event) => {
                    if (event.key === "Enter") {
                        event.preventDefault();
                        event.currentTarget.blur();
                    }
                    if (event.key === "Escape") {
                        setFormulaValue(rawCellValue(selectedCell!.column, selectedCell!.row));
                        setFormulaEditing(false);
                        formulaBarRef.current?.blur();
                    }
                }}
            />
        </div>
        <div ref={gridRef} className="table-grid">
            <DataEditor
                ref={dataEditorRef}
                columns={columns}
                getCellContent={getCellContent}
                rows={table.rows.length + 1}
                rowMarkers="clickable-number"
                rowHeight={30}
                headerHeight={34}
                smoothScrollX
                smoothScrollY
                onCellEdited={editCell}
                customRenderers={[enumRenderer]}
                gridSelection={gridSelection}
                onGridSelectionChange={(selection) => {
                    setGridSelection(selection);
                    const selectedColumns = selection.columns.toArray();
                    if (selectedColumns.length > 0) {
                        setSelectedColumn(selectedColumns.length === 1 && selectedColumns[0] > 0 ? selectedColumns[0] : null);
                    }
                }}
                onCellClicked={([column, row], event) => {
                    setSelectedColumn(null);
                    setSelectedCell({column, row});
                    setFormulaEditing(false);
                    const gridBounds = gridRef.current?.getBoundingClientRect();
                    const bounds = {x: event.bounds.x - (gridBounds?.x ?? 0), y: event.bounds.y - (gridBounds?.y ?? 0), width: event.bounds.width, height: event.bounds.height};
                    lastCellBounds.current = {cell: [column, row], bounds};
                    if (column >= 0) openInlineAt(column, row, bounds);
                }}
                onHeaderClicked={(column) => {
                    setSelectedColumn(column === 0 ? null : column);
                    setMenu(null);
                    setContextRow(null);
                }}
                onHeaderContextMenu={(column, event) => {
                    event.preventDefault();
                    setSelectedColumn(column === 0 ? null : column);
                    setColumnMenu({column, x: event.localEventX, y: event.localEventY});
                    setMenu(null);
                    setContextRow(null);
                }}
                onCellContextMenu={([column, row], event) => {
                    if (row === table.rows.length || column !== -1) return;
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
            {columnMenu && <div className="column-context-menu" style={{left: columnMenu.x, top: columnMenu.y}}>
                <button disabled={columnMenu.column === 0} onClick={() => { onAddColumn(columnMenu.column - 1); setColumnMenu(null); }}>Add to left</button>
                <button onClick={() => { onAddColumn(columnMenu.column); setColumnMenu(null); }}>Add to right</button>
                <hr/>
                <button disabled={columnMenu.column === 0} onClick={() => { const column = table.columns[columnMenu.column - 1]; onEditColumnSchema(column.name, column.schemaExpression); setColumnMenu(null); }}>Edit schema</button>
                <button disabled={columnMenu.column === 0} onClick={() => { onRenameColumn(table.columns[columnMenu.column - 1].name); setColumnMenu(null); }}>Rename column</button>
                <button disabled={columnMenu.column === 0} onClick={() => { onMoveColumn(table.columns[columnMenu.column - 1].name); setColumnMenu(null); }}>Move column</button>
                <button className="column-context-danger" disabled={columnMenu.column === 0} onClick={() => { onDeleteColumn(table.columns[columnMenu.column - 1].name); setColumnMenu(null); }}>Delete column</button>
            </div>}
            {inlineEditor && <textarea
                ref={editorRef}
                className="inline-cell-editor"
                defaultValue={inlineEditor.value}
                style={{left: inlineEditor.bounds.x, top: inlineEditor.bounds.y, width: inlineEditor.bounds.width, height: inlineEditor.bounds.height}}
                onBlur={finishInlineEdit}
                onKeyDown={(event) => {
                    if (event.key === "Enter" && !event.shiftKey) {
                        event.preventDefault();
                        finishInlineEdit();
                    }
                    if (event.key === "Escape") setInlineEditor(null);
                }}
            />}
            {foreignKeyEditor && <div className="foreign-key-cell-editor" style={{left: foreignKeyEditor.bounds.x, top: foreignKeyEditor.bounds.y, width: foreignKeyEditor.bounds.width, minHeight: foreignKeyEditor.bounds.height}}>
                <input ref={foreignKeyEditorRef} defaultValue={foreignKeyEditor.value} list={`foreign-key-options-${foreignKeyEditor.column}`} aria-label={`Select ${foreignKeyEditor.target} ID`} onKeyDown={(event) => {
                    if (event.key === "Enter") finishForeignKeyEdit(event.currentTarget.value);
                    if (event.key === "Escape") setForeignKeyEditor(null);
                }} onBlur={(event) => finishForeignKeyEdit(event.currentTarget.value)}/>
                <datalist id={`foreign-key-options-${foreignKeyEditor.column}`}>{foreignKeyEditor.options.map((option) => <option key={option} value={option}/>)}</datalist>
            </div>}
            {arrayEditor && <div className="array-editor-backdrop" role="presentation" onMouseDown={() => setArrayEditor(null)}>
                <section className="array-editor-dialog" role="dialog" aria-label={`Edit ${table.columns[arrayEditor.column - 1].name}`} onMouseDown={(event) => event.stopPropagation()}>
                    <header><div><span>Array values</span><h2>{table.columns[arrayEditor.column - 1].name}</h2></div><button type="button" aria-label="Close array editor" onClick={() => setArrayEditor(null)}>x</button></header>
                    <div className="array-entry-list">
                        {arrayEditor.entries.map((entry, index) => <div className="array-entry" key={`${index}:${entry}`}>
                            <code>{index + 1}</code>
                            <input aria-label={`Array item ${index + 1}`} value={entry} onChange={(event) => setArrayEditor({...arrayEditor, entries: arrayEditor.entries.map((value, position) => position === index ? event.target.value : value)})}/>
                            <button type="button" aria-label={`Move item ${index + 1} up`} disabled={index === 0} onClick={() => setArrayEditor({...arrayEditor, entries: arrayEditor.entries.map((value, position) => position === index ? arrayEditor.entries[index - 1] : position === index - 1 ? arrayEditor.entries[index] : value)})}>Up</button>
                            <button type="button" aria-label={`Move item ${index + 1} down`} disabled={index === arrayEditor.entries.length - 1} onClick={() => setArrayEditor({...arrayEditor, entries: arrayEditor.entries.map((value, position) => position === index ? arrayEditor.entries[index + 1] : position === index + 1 ? arrayEditor.entries[index] : value)})}>Down</button>
                            <button type="button" aria-label={`Delete item ${index + 1}`} onClick={() => setArrayEditor({...arrayEditor, entries: arrayEditor.entries.filter((_, position) => position !== index)})}>Delete</button>
                        </div>)}
                    </div>
                    <button className="array-add" type="button" onClick={() => setArrayEditor({...arrayEditor, entries: [...arrayEditor.entries, ""]})}>Add item</button>
                    <label className="array-paste">Paste items<textarea aria-label="Paste array items" value={arrayEditor.pasted} placeholder={'One CUE expression per line'} onChange={(event) => setArrayEditor({...arrayEditor, pasted: event.target.value})}/></label>
                    <div className="dialog-actions"><button type="button" onClick={() => setArrayEditor({...arrayEditor, entries: [...arrayEditor.entries, ...arrayEditor.pasted.replace(/\r/g, "").split("\n").map((entry) => entry.trim()).filter(Boolean)], pasted: ""})} disabled={!arrayEditor.pasted.trim()}>Add pasted items</button><button type="button" onClick={() => setArrayEditor(null)}>Cancel</button><button type="button" onClick={saveArray}>Apply</button></div>
                </section>
            </div>}
            {structEditor && <div className="array-editor-backdrop" role="presentation" onMouseDown={() => setStructEditor(null)}>
                <section className="struct-editor-dialog" role="dialog" aria-label={`Edit ${table.columns[structEditor.column - 1].name}`} onMouseDown={(event) => event.stopPropagation()}>
                    <header><div><span>Structure values</span><h2>{table.columns[structEditor.column - 1].name}</h2></div><button type="button" aria-label="Close struct editor" onClick={() => setStructEditor(null)}>x</button></header>
                    <div className="struct-field-list">{(table.columns[structEditor.column - 1].schema.fields ?? []).map((field) => <label key={field.name}>{field.name}{field.optional && <span>Optional</span>}<input aria-label={`Struct field ${field.name}`} value={structEditor.values[field.name] ?? ""} placeholder="CUE expression" onChange={(event) => setStructEditor({...structEditor, values: {...structEditor.values, [field.name]: event.target.value}})}/></label>)}</div>
                    <div className="dialog-actions"><button type="button" onClick={() => setStructEditor(null)}>Cancel</button><button type="button" onClick={saveStruct}>Apply</button></div>
                </section>
            </div>}
        </div>
    </div>;
}

type StructureDialog = {
    mode: "add" | "schema" | "rename" | "move" | "delete";
    name: string;
    second: string;
    schema: string;
    initial: string;
};
type TableDialog = {namespace: string; name: string};

type SchemaDraft = {name: string; kind: "type" | "constraint"; type: string; appliesTo: string; expression: string};

function SchemaEditor({state, usages, onSelect, onCreate, onApply, onRename, onDelete}: {
    state: SchemaEditorState;
    usages: (name: string) => SchemaUsage[];
    onSelect: (name: string) => void;
    onCreate: (name: string, kind: "type" | "constraint", type: string, appliesTo: string[], expression: string) => Promise<void>;
    onApply: (name: string, kind: "type" | "constraint", type: string, appliesTo: string[], expression: string) => Promise<void>;
    onRename: (oldName: string, newName: string) => Promise<void>;
    onDelete: (name: string) => Promise<void>;
}) {
    const [draft, setDraft] = useState<SchemaDraft | null>(null);
    const [creating, setCreating] = useState(false);
    const [renaming, setRenaming] = useState(false);
    const [renameValue, setRenameValue] = useState("");
    const [confirmingDelete, setConfirmingDelete] = useState(false);
    const [busy, setBusy] = useState(false);
    const [editorError, setEditorError] = useState("");

    useEffect(() => {
        setCreating(false);
        setRenaming(false);
        setConfirmingDelete(false);
        setEditorError("");
        const definition = state.definitions.find((entry) => entry.name === state.selected) ?? null;
        setDraft(definition ? {
            name: definition.name,
            kind: definition.kind,
            type: definition.type ?? "",
            appliesTo: (definition.appliesTo ?? []).join(", "),
            expression: definition.schemaExpression,
        } : null);
    }, [state.selected, state.definitions]);

    function startCreate() {
        setDraft({name: "", kind: "type", type: "string", appliesTo: "", expression: "string"});
        setCreating(true);
        setRenaming(false);
        setConfirmingDelete(false);
        setEditorError("");
    }

    async function submit() {
        if (!draft) return;
        setBusy(true);
        setEditorError("");
        try {
            const appliesTo = draft.appliesTo.split(",").map((entry) => entry.trim()).filter((entry) => entry !== "");
            if (creating) await onCreate(draft.name, draft.kind, draft.type, appliesTo, draft.expression);
            else await onApply(draft.name, draft.kind, draft.type, appliesTo, draft.expression);
        } catch (reason) {
            setEditorError(reason instanceof Error ? reason.message : String(reason));
        } finally {
            setBusy(false);
        }
    }

    async function submitRename() {
        if (!draft || !renameValue.trim() || renameValue.trim() === draft.name) {
            setRenaming(false);
            return;
        }
        setBusy(true);
        setEditorError("");
        try {
            await onRename(draft.name, renameValue.trim());
            setRenaming(false);
        } catch (reason) {
            setEditorError(reason instanceof Error ? reason.message : String(reason));
        } finally {
            setBusy(false);
        }
    }

    const selectedUsages = state.selected ? usages(state.selected) : [];
    return (
        <section className="schema-editor">
            <aside className="schema-editor-sidebar">
                <div className="schema-editor-title">{state.local ? "Local Schema" : "Public Schema"}</div>
                <div className="schema-editor-file">{state.path}</div>
                <div className="schema-definition-list">
                    {state.definitions.map((definition) => <button
                        className={`schema-definition-row ${state.selected === definition.name ? "schema-definition-row-selected" : ""}`}
                        key={definition.name}
                        onClick={() => onSelect(definition.name)}
                    >
                        <span>#{definition.name}</span>
                        <span className={`schema-kind-dot schema-kind-${definition.kind}`} aria-hidden="true"/>
                    </button>)}
                    {state.definitions.length === 0 && <div className="schema-editor-empty">No definitions</div>}
                </div>
                <button className="schema-editor-create" onClick={startCreate}>New definition</button>
            </aside>
            <div className="schema-editor-form">
                {draft ? <>
                    <div className="schema-editor-row">
                        <label>Name<input aria-label="Name" value={draft.name} disabled={!creating} onChange={(event) => setDraft({...draft, name: event.target.value})}/></label>
                        <label>Kind<select aria-label="Kind" value={draft.kind} onChange={(event) => setDraft({...draft, kind: event.target.value as "type" | "constraint"})}>
                            <option value="type">type</option>
                            <option value="constraint">constraint</option>
                        </select></label>
                    </div>
                    {draft.kind === "type" ? (
                        <label>Type<select aria-label="Type" value={draft.type} onChange={(event) => setDraft({...draft, type: event.target.value})}>
                            {["string", "integer", "number", "boolean", "enum", "array", "struct", "reference"].map((option) => <option key={option} value={option}>{option}</option>)}
                        </select></label>
                    ) : (
                        <label>Applies to<input aria-label="Applies to" value={draft.appliesTo} placeholder="integer, number" onChange={(event) => setDraft({...draft, appliesTo: event.target.value})}/></label>
                    )}
                    <label>Schema expression<textarea aria-label="Schema expression" rows={6} value={draft.expression} onChange={(event) => setDraft({...draft, expression: event.target.value})}/></label>
                    <div className="schema-editor-actions">
                        <button disabled={busy} onClick={submit}>{creating ? "Create" : "Apply"}</button>
                        {creating ? <button onClick={() => { setCreating(false); setDraft(null); }}>Cancel</button> : <>
                            {renaming ? <>
                                <input className="schema-rename-input" value={renameValue} onChange={(event) => setRenameValue(event.target.value)}/>
                                <button disabled={busy} onClick={submitRename}>Confirm rename</button>
                                <button onClick={() => setRenaming(false)}>Cancel</button>
                            </> : <button onClick={() => { setRenameValue(draft.name); setRenaming(true); }}>Rename</button>}
                            <button className={confirmingDelete ? "schema-delete-danger" : ""} onClick={async () => {
                                if (!confirmingDelete) {
                                    setConfirmingDelete(true);
                                    return;
                                }
                                setConfirmingDelete(false);
                                setBusy(true);
                                setEditorError("");
                                try {
                                    await onDelete(draft.name);
                                } catch (reason) {
                                    setEditorError(reason instanceof Error ? reason.message : String(reason));
                                } finally {
                                    setBusy(false);
                                }
                            }}>{confirmingDelete ? "Confirm delete" : "Delete definition"}</button>
                        </>}
                    </div>
                    {editorError && <div className="schema-editor-error">{editorError}</div>}
                </> : <div className="schema-editor-empty">Select or create a definition</div>}
                <div className="schema-editor-usages">
                    <div className="schema-editor-usages-title">Usages</div>
                    {selectedUsages.map((usage) => <div className="usage-row" key={`${usage.path}:${usage.field}`}>
                        <span>{usage.table ?? usage.path}</span>
                        <code>{usage.field}</code>
                    </div>)}
                    {state.selected && selectedUsages.length === 0 && <div className="schema-editor-empty">No usages</div>}
                </div>
            </div>
        </section>
    );
}

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
        };
        const closeOnEscape = (event: KeyboardEvent) => {
            if (event.key === "Escape") setAppMenu(null);
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
        setError("");
        try {
            if (node.kind === "localSchema") {
                const document = await GetTable(node.path!) as TableDocument;
                const logical = document.namespace ? `${document.namespace}.${document.name}` : document.name;
                setSchemaDocument({path: node.path!, local: true, logicalTable: logical, definitions: document.localSchema, selected: document.localSchema[0]?.name ?? null});
            } else {
                const document = await GetPublicSchema(node.path!) as PublicSchemaDocument;
                const name = node.kind === "schemaDefinition" ? node.name.replace(/^#/, "") : null;
                setSchemaDocument({path: node.path!, local: false, package: document.package, definitions: document.definitions, selected: name ?? document.definitions[0]?.name ?? null});
            }
            setTable(null);
            setCellError(null);
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
            let result;
            if (dialog.mode === "add") {
                result = await AddColumn({path: table.path, name: dialog.name, schemaExpression: dialog.schema, initialExpression: dialog.initial, position: Number(dialog.second)});
            } else if (dialog.mode === "schema") {
                result = await SetColumnSchema({path: table.path, name: dialog.name, schemaExpression: dialog.schema});
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
                        <div className="tree"><Tree node={workspace.tree.root} onOpenTable={openTable} onOpenSchema={openSchema}/></div>
                    </aside>
                    <section className={`editor-placeholder ${table ? "editor-table" : ""}`}>
                        {table ? <TableGrid key={table.path} table={table} onAddColumn={(position) => setDialog({mode: "add", name: "", second: String(position), schema: "string", initial: `""`})} onDeleteColumn={(name) => setDialog({mode: "delete", name, second: "", schema: "", initial: ""})} onDeleteRow={deleteRow} onEditColumnSchema={(name, schemaExpression) => setDialog({mode: "schema", name, second: "", schema: schemaExpression.replaceAll("_local.#", "#"), initial: ""})} onMoveColumn={(name) => setDialog({mode: "move", name, second: "0", schema: "", initial: ""})} onRenameColumn={(name) => setDialog({mode: "rename", name, second: "", schema: "", initial: ""})} onPaste={pasteCells} onSaveInput={saveCellInput}/> : schemaDocument ? <SchemaEditor state={schemaDocument} usages={schemaUsages} onSelect={(name) => setSchemaDocument({...schemaDocument, selected: name})} onCreate={(name, kind, type, appliesTo, expression) => saveSchemaDefinition(true, name, kind, type, appliesTo, expression)} onApply={(name, kind, type, appliesTo, expression) => saveSchemaDefinition(false, name, kind, type, appliesTo, expression)} onRename={renameSchemaDefinition} onDelete={deleteSchemaDefinition}/> : selectedSchema ? <>
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
                            {schemaDefinitions(workspace).map((definition) => <button
                                className={`schema-row ${selectedSchema?.id === definition.id ? "schema-row-selected" : ""}`}
                                key={definition.id}
                                onClick={() => { setSelectedSchema(definition); setTable(null); setCellError(null); }}
                            >
                                <span className={`schema-kind-dot schema-kind-${definition.kind}`} aria-hidden="true"/>
                                <span>{definition.id}</span>
                                <span className="usage-count">{definition.usages?.length ?? 0}</span>
                            </button>)}
                        </div>
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
                        {workspace.compile.mode === "manual" && workspace.compile.diagnostics.length > 0 && <div className="schema-diagnostics compile-diagnostics">
                            <div className="schema-diagnostics-title">Compile diagnostics</div>
                            {workspace.compile.diagnostics.map((diagnostic) => <div className="schema-diagnostic" key={`${diagnostic.path}:${diagnostic.rowID ?? ""}:${diagnostic.column ?? ""}:${diagnostic.message}`}>
                                <code>{[diagnostic.rowID, diagnostic.column].filter(Boolean).join(".") || diagnostic.field || diagnostic.table || "compile"}</code>
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
            <LogPanel/>
            {error && <div className="error-toast" role="alert">{error}</div>}
            {notice && <div className="notice-toast" role="status">{notice}</div>}
            {dialog && <div className="dialog-backdrop" role="presentation" onMouseDown={() => setDialog(null)}>
                <form className="structure-dialog" onMouseDown={(event) => event.stopPropagation()} onSubmit={(event) => { event.preventDefault(); submitStructureDialog(); }}>
                    <h2>{dialog.mode === "add" ? "Add column" : dialog.mode === "schema" ? "Edit column schema" : dialog.mode === "rename" ? "Rename column" : dialog.mode === "move" ? "Move column" : "Delete column"}</h2>
                    {dialog.mode !== "schema" && <label>Column name<input autoFocus value={dialog.name} onChange={(event) => setDialog({...dialog, name: event.target.value})}/></label>}
                    {dialog.mode === "schema" && <label>Column name<input value={dialog.name} disabled/></label>}
                    {dialog.mode === "add" && <>
                        <label>CUE Schema<input value={dialog.schema} onChange={(event) => setDialog({...dialog, schema: event.target.value})}/></label>
                        <label>Initial CUE value<input value={dialog.initial} onChange={(event) => setDialog({...dialog, initial: event.target.value})}/></label>
                    </>}
                    {dialog.mode === "schema" && <label>CUE Schema<textarea autoFocus value={dialog.schema} onChange={(event) => setDialog({...dialog, schema: event.target.value})}/></label>}
                    {dialog.mode === "schema" && <label>Insert Schema reference<select aria-label="Insert Schema reference" value="" onChange={(event) => {
                        const option = schemaReferenceOptions(workspace).find((entry) => entry.value === event.target.value);
                        if (!option) return;
                        const current = dialog.schema.trim();
                        setDialog({...dialog, schema: option.kind === "constraint" ? (current ? `${current} & ${option.value}` : option.value) : option.value});
                    }}>
                        <option value="">Choose a reference</option>
                        {schemaReferenceOptions(workspace).map((option) => <option key={`${option.kind}:${option.value}`} value={option.value}>{option.label}</option>)}
                    </select></label>}
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
        </main>
    )
}

export default App
