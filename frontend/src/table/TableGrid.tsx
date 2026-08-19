import {useEffect, useRef, useState} from "react";
import {CompactSelection, DataEditor, drawTextCell, GridCellKind, type CustomCell, type CustomRenderer, type DataEditorRef, type DrawCellCallback, type GridCell, type GridColumn, type GridMouseEventArgs, type GridSelection, type Item} from "@glideapps/glide-data-grid";
import "@glideapps/glide-data-grid/dist/index.css";
import {useTranslation} from "../i18n";
import {useTheme} from "../theme";
import {TableExportSettings as TableExportSettingsPanel} from "./TableExportSettings";
import type {ExportFormat, NavRequest, TableDocument, TableExportSettings as TableExportSettingsModel} from "../types";

type DraftRow = { id: string; values: Record<string, string> };
type EnumCell = CustomCell<{kind: "enum"; value: string; options: string[]}>;
type InlineEditor = { column: number; row: number; value: string; bounds: {x: number; y: number; width: number; height: number} };
type ForeignKeyEditor = InlineEditor & {options: string[]; target: string};
type ArrayEditor = {column: number; row: number; entries: string[]; pasted: string};
type StructEditor = {column: number; row: number; values: Record<string, string>};
type CommentEditor = {rowID: string; column: string; value: string};
type CommentTooltip = {text: string; left: number; top: number};

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

export function TableGrid({table, formats, onAddColumn, onDeleteColumn, onDeleteRow, onEditColumnSchema, onMoveColumn, onRenameColumn, onPaste, onSaveInput, onSaveComment, onSaveExportSettings, onNavigate, focusTarget, onFocusDone}: { table: TableDocument; formats: ExportFormat[]; onAddColumn: (position: number) => void; onDeleteColumn: (column: string) => void; onDeleteRow: (rowID: string) => void; onEditColumnSchema: (column: string, schemaExpression: string) => void; onMoveColumn: (column: string) => void; onRenameColumn: (column: string) => void; onPaste: (row: number, column: number, text: string) => Promise<void>; onSaveInput: (rowID: string, column: string, input: string) => Promise<void>; onSaveComment: (rowID: string, column: string, comment: string) => Promise<void>; onSaveExportSettings: (settings: TableExportSettingsModel) => Promise<void>; onNavigate: (request: NavRequest, position: {x: number; y: number}) => void; focusTarget: {path: string; rowID: string; column: string; nonce: number} | null; onFocusDone: () => void }) {
    const {t} = useTranslation();
    const {grid, dataGridTheme} = useTheme();
    const [draft, setDraft] = useState<DraftRow>({id: "", values: {}});
    const [draftError, setDraftError] = useState("");
    const [contextRow, setContextRow] = useState<number | null>(null);
    const [contextCell, setContextCell] = useState<{column: number; row: number} | null>(null);
    const [menu, setMenu] = useState<{x: number; y: number; fixed?: boolean} | null>(null);
    const [columnMenu, setColumnMenu] = useState<{column: number; x: number; y: number} | null>(null);
    const [selectedColumn, setSelectedColumn] = useState<number | null>(null);
    const [gridSelection, setGridSelection] = useState<GridSelection>({columns: CompactSelection.empty(), rows: CompactSelection.empty(), current: undefined});
    const [inlineEditor, setInlineEditor] = useState<InlineEditor | null>(null);
    const [foreignKeyEditor, setForeignKeyEditor] = useState<ForeignKeyEditor | null>(null);
    const [arrayEditor, setArrayEditor] = useState<ArrayEditor | null>(null);
    const [structEditor, setStructEditor] = useState<StructEditor | null>(null);
    const [commentEditor, setCommentEditor] = useState<CommentEditor | null>(null);
    const [commentTooltip, setCommentTooltip] = useState<CommentTooltip | null>(null);
    const [commentError, setCommentError] = useState("");
    const [selectedCell, setSelectedCell] = useState<{column: number; row: number} | null>(null);
    const [formulaEditing, setFormulaEditing] = useState(false);
    const [formulaValue, setFormulaValue] = useState("");
    const editorRef = useRef<HTMLTextAreaElement | null>(null);
    const foreignKeyEditorRef = useRef<HTMLInputElement | null>(null);
    const formulaBarRef = useRef<HTMLInputElement | null>(null);
    const dataEditorRef = useRef<DataEditorRef | null>(null);
    const gridRef = useRef<HTMLDivElement | null>(null);
    const columns: GridColumn[] = [
        {id: "id", title: "ID", width: 160, style: "highlight"},
        ...table.columns.map((column) => ({id: column.name, title: column.name, width: 170})),
    ];
    const isDraft = (row: number) => row === table.rows.length;
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
                setContextCell(null);
            }
        };
        const closeOnEscape = (event: KeyboardEvent) => {
            if (event.key === "Escape") {
                setMenu(null);
                setContextRow(null);
                setColumnMenu(null);
                setContextCell(null);
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
        const columnDefinition = table.columns[column - 1];
        const cell = source.values[columnDefinition.name];
        if (cell?.formula) return `=${cell.expression}`;
        return columnDefinition.schema.kind === "string" ? cell?.display ?? "" : cell?.expression ?? "";
    };
    const selectedCellLabel = selectedCell ? `${selectedCell.column === 0 ? "id" : table.columns[selectedCell.column - 1].name} · ${rowFor(selectedCell.row).id || t("grid.draft")}` : "";
    useEffect(() => {
        if (formulaEditing || !selectedCell) return;
        setFormulaValue(rawCellValue(selectedCell.column, selectedCell.row));
    }, [selectedCell, table, draft, formulaEditing]);
    const navigationRequestForCell = (column: number, row: number): NavRequest | null => {
        if (isDraft(row)) return null;
        const source = table.rows[row];
        if (column === 0) {
            return {kind: "primary-key-references", path: table.path, rowID: source.id};
        }
        const columnDefinition = table.columns[column - 1];
        const cell = source.values[columnDefinition.name];
        if (cell?.formula) {
            return {kind: "formula-reference", path: table.path, rowID: source.id, column: columnDefinition.name};
        }
        if (columnDefinition.foreignKeyTable && cell) {
            return {kind: "foreign-key", path: table.path, rowID: cell.display, column: columnDefinition.name};
        }
        if (columnDefinition.schema.kind === "reference" && columnDefinition.schema.reference) {
            return {kind: "schema-reference", path: table.path, reference: columnDefinition.schema.reference};
        }
        return null;
    };
    useEffect(() => {
        if (!focusTarget || focusTarget.path !== table.path) return;
        if (!focusTarget.column || focusTarget.column === "#PrimaryKey") {
            onFocusDone();
            return;
        }
        const columnIndex = focusTarget.column === "id" ? 0 : table.columns.findIndex((entry) => entry.name === focusTarget.column) + 1;
        if (columnIndex >= 0) {
            const rowIndex = focusTarget.rowID ? table.rows.findIndex((entry) => entry.id === focusTarget.rowID) : -1;
            if (rowIndex >= 0) {
                setGridSelection({columns: CompactSelection.fromSingleSelection(columnIndex), rows: CompactSelection.fromSingleSelection(rowIndex), current: {cell: [columnIndex, rowIndex], range: {x: columnIndex, y: rowIndex, width: 1, height: 1}, rangeStack: []}});
                dataEditorRef.current?.scrollTo(columnIndex, rowIndex, "both", 24, 48, {hAlign: "center", vAlign: "center"});
                dataEditorRef.current?.focus();
            } else {
                setSelectedColumn(columnIndex === 0 ? null : columnIndex);
                dataEditorRef.current?.scrollTo(columnIndex, 0, "horizontal", 24, 48, {hAlign: "center", vAlign: "start"});
            }
        }
        onFocusDone();
    }, [focusTarget, table]);
    useEffect(() => {
        const navigateShortcut = (event: KeyboardEvent) => {
            if (event.key !== "F12" || !selectedCell) return;
            event.preventDefault();
            const request = navigationRequestForCell(selectedCell.column, selectedCell.row);
            if (!request) return;
            const gridBounds = gridRef.current?.getBoundingClientRect();
            onNavigate(request, {x: gridBounds?.x ?? 0, y: gridBounds?.y ?? 0});
        };
        document.addEventListener("keydown", navigateShortcut);
        return () => document.removeEventListener("keydown", navigateShortcut);
    }, [selectedCell, table]);
    const errorFor = (rowID: string, column: string) => table.diagnostics?.find((diagnostic) => diagnostic.rowID === rowID && (!diagnostic.column || diagnostic.column === column));
    const commentFor = (rowID: string, column: string) => table.cellComments?.[rowID]?.[column] ?? "";
    const cellColumnName = (column: number) => column === 0 ? "id" : table.columns[column - 1]?.name ?? "";
    const drawCell: DrawCellCallback = (args, drawContent) => {
        drawContent();
        const column = cellColumnName(args.col);
        const row = table.rows[args.row];
        if (!row || !commentFor(row.id, column)) return;
        args.ctx.save();
        args.ctx.strokeStyle = grid.commentColor;
        args.ctx.lineWidth = 2;
        args.ctx.strokeRect(args.rect.x + 1, args.rect.y + 1, args.rect.width - 2, args.rect.height - 2);
        args.ctx.restore();
    };
    const handleGridMouseMove = (args: GridMouseEventArgs) => {
        if (args.kind !== "cell") {
            setCommentTooltip(null);
            return;
        }
        const [column, row] = args.location;
        const source = table.rows[row];
        const comment = source ? commentFor(source.id, cellColumnName(column)) : "";
        if (!comment) {
            setCommentTooltip(null);
            return;
        }
        const gridBounds = gridRef.current?.getBoundingClientRect();
        const left = (gridBounds?.left ?? 0) + args.bounds.x + args.localEventX;
        const top = (gridBounds?.top ?? 0) + args.bounds.y + args.localEventY + 12;
        setCommentTooltip({text: comment, left: Math.max(4, Math.min(left, window.innerWidth - 340)), top: Math.max(4, Math.min(top, window.innerHeight - 160))});
    };
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
            return {kind: GridCellKind.Text, data: cell?.expression ?? "[]", displayData: cell?.display ?? "[]", allowOverlay: false, themeOverride: error ? {bgCell: grid.errorCellBg, textDark: grid.errorCellText} : isDraft(row) ? {bgCell: grid.draftCellBg, textDark: grid.draftCellText} : undefined};
        }
        if (column !== 0 && columnDefinition.foreignKeyTable && !cell?.formula) {
            return {kind: GridCellKind.Text, data: cell?.expression ?? "", displayData: cell?.display ?? "", allowOverlay: false, themeOverride: error ? {bgCell: grid.errorCellBg, textDark: grid.errorCellText} : isDraft(row) ? {bgCell: grid.draftCellBg, textDark: grid.draftCellText} : undefined};
        }
        if (column !== 0 && columnDefinition.schema.kind === "enum" && !cell?.formula) {
            const options = columnDefinition.schema.enumValues?.map((entry) => entry.expression.replace(/^"|"$/g, "")) ?? [];
            return {kind: GridCellKind.Custom, allowOverlay: true, copyData: cell?.display ?? "", data: {kind: "enum", value: cell?.display ?? options[0] ?? "", options}, themeOverride: error ? {bgCell: grid.errorCellBg, textDark: grid.errorCellText} : isDraft(row) ? {bgCell: grid.draftCellBg, textDark: grid.draftCellText} : undefined} as EnumCell;
        }
        if (column !== 0 && columnDefinition.schema.kind === "boolean" && !cell?.formula) {
            return {kind: GridCellKind.Boolean, allowOverlay: false, data: cell?.expression === "true", themeOverride: error ? {bgCell: grid.errorCellBg, textDark: grid.errorCellText} : isDraft(row) ? {bgCell: grid.draftCellBg, textDark: grid.draftCellText} : undefined};
        }
        return {
            kind: GridCellKind.Text,
            data: cell?.formula ? `=${cell.expression}` : column !== 0 && columnDefinition.schema.kind === "string" ? cell?.display ?? "" : cell?.expression ?? "",
            displayData: cell?.display ?? "",
            allowOverlay: false,
            themeOverride: error ? {
                bgCell: grid.errorCellBg,
                textDark: grid.errorCellText,
            } : isDraft(row) ? {bgCell: grid.draftCellBg, textDark: grid.draftCellText} : cell?.formula ? {
                bgCell: grid.formulaBg,
                textDark: grid.formulaText,
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
                setDraftError(duplicate ? t("grid.idExists") : t("grid.idInvalid"));
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
                <div className="table-eyebrow">{t("grid.cueExpressionEditing")}</div>
                <h1>{table.namespace}.{table.name}</h1>
            </div>
            <span>{t("grid.rowsDraft", {count: table.rows.length})}</span>
        </div>
        {draftError && <div className="table-validation-error"><code>{t("grid.draft")}</code><span>{draftError}</span></div>}
        {(table.diagnostics?.length ?? 0) > 0 && <section className="table-diagnostics" aria-label={t("grid.tableDiagnostics")}>
            <div className="table-diagnostics-title"><strong>{t("grid.diagnostics")}</strong><span>{table.diagnostics!.length}</span></div>
            <div className="table-diagnostics-list">{table.diagnostics!.map((diagnostic, index) => {
                const location = diagnostic.rowID ? diagnostic.column ? `${diagnostic.rowID}.${diagnostic.column}` : diagnostic.rowID : t("grid.table");
                const navigable = table.rows.some((row) => row.id === diagnostic.rowID) && (!diagnostic.column || table.columns.some((column) => column.name === diagnostic.column));
                return <button key={`${location}:${diagnostic.message}:${index}`} className="table-diagnostic" disabled={!navigable} onClick={() => focusDiagnostic(diagnostic)}><code>{location}</code><span>{diagnostic.message}</span></button>;
            })}</div>
        </section>}
        <div className="formula-bar">
            <span className="formula-bar-cell">{selectedCellLabel}</span>
            <input
                ref={formulaBarRef}
                className="formula-bar-input"
                aria-label={t("grid.formulaBar")}
                value={formulaValue}
                disabled={!selectedCell}
                readOnly={!formulaEditing}
                placeholder={t("grid.selectCell")}
                onChange={(event) => setFormulaValue(event.target.value)}
                onClick={() => {
                    if (!selectedCell || formulaEditing) return;
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
                theme={dataGridTheme}
                drawCell={drawCell}
                onMouseMove={handleGridMouseMove}
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
                    if (event.ctrlKey || event.metaKey) {
                        const request = navigationRequestForCell(column, row);
                        if (request) {
                            const gridBounds = gridRef.current?.getBoundingClientRect();
                            onNavigate(request, {x: (gridBounds?.x ?? 0) + event.localEventX, y: (gridBounds?.y ?? 0) + event.localEventY});
                        }
                    }
                }}
                onCellActivated={([column, row]) => {
                    if (column < 0 || row < 0) return;
                    const rect = dataEditorRef.current?.getBounds(column, row);
                    const gridBounds = gridRef.current?.getBoundingClientRect();
                    const bounds = rect && gridBounds ? {x: rect.x - gridBounds.x, y: rect.y - gridBounds.y, width: rect.width, height: rect.height} : {x: 0, y: 0, width: 100, height: 24};
                    openInlineAt(column, row, bounds);
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
                    if (row === table.rows.length) return;
                    event.preventDefault();
                    setContextRow(column === -1 ? row : null);
                    setContextCell(column >= 0 ? {column, row} : null);
                    const gridBounds = gridRef.current?.getBoundingClientRect();
                    setMenu(column >= 0 && gridBounds
                        ? {x: gridBounds.left + event.bounds.x + event.localEventX, y: gridBounds.top + event.bounds.y + event.localEventY, fixed: true}
                        : {x: event.localEventX, y: event.localEventY});
                }}
            />
            {menu !== null && (contextRow !== null || contextCell !== null) && <div className={`row-context-menu ${menu.fixed ? "cell-comment-menu" : ""}`} style={{left: menu.x, top: menu.y}}>
                {contextCell && <>
                    <button onClick={() => {
                        const rowID = table.rows[contextCell.row].id;
                        const column = cellColumnName(contextCell.column);
                        setCommentError("");
                        setCommentEditor({rowID, column, value: commentFor(rowID, column)});
                        setMenu(null);
                        setContextCell(null);
                    }}>{commentFor(table.rows[contextCell.row].id, cellColumnName(contextCell.column)) ? t("comment.edit") : t("comment.add")}</button>
                    {commentFor(table.rows[contextCell.row].id, cellColumnName(contextCell.column)) && <button className="row-context-danger" onClick={() => {
                        void onSaveComment(table.rows[contextCell.row].id, cellColumnName(contextCell.column), "");
                        setMenu(null);
                        setContextCell(null);
                    }}>{t("comment.clear")}</button>}
                </>}
                {contextRow !== null && <button className="row-context-danger" onClick={() => {
                    onDeleteRow(table.rows[contextRow].id);
                    setMenu(null);
                    setContextRow(null);
                }}>{t("grid.deleteRow")}</button>}
            </div>}
            {columnMenu && <div className="column-context-menu" style={{left: columnMenu.x, top: columnMenu.y}}>
                <button disabled={columnMenu.column === 0} onClick={() => { onAddColumn(columnMenu.column - 1); setColumnMenu(null); }}>{t("grid.addToLeft")}</button>
                <button onClick={() => { onAddColumn(columnMenu.column); setColumnMenu(null); }}>{t("grid.addToRight")}</button>
                <hr/>
                <button disabled={columnMenu.column === 0} onClick={() => { const column = table.columns[columnMenu.column - 1]; onEditColumnSchema(column.name, column.schemaExpression); setColumnMenu(null); }}>{t("grid.editSchema")}</button>
                <button disabled={columnMenu.column === 0} onClick={() => { onRenameColumn(table.columns[columnMenu.column - 1].name); setColumnMenu(null); }}>{t("grid.renameColumn")}</button>
                <button disabled={columnMenu.column === 0} onClick={() => { onMoveColumn(table.columns[columnMenu.column - 1].name); setColumnMenu(null); }}>{t("grid.moveColumn")}</button>
                <button className="column-context-danger" disabled={columnMenu.column === 0} onClick={() => { onDeleteColumn(table.columns[columnMenu.column - 1].name); setColumnMenu(null); }}>{t("grid.deleteColumn")}</button>
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
                <input ref={foreignKeyEditorRef} defaultValue={foreignKeyEditor.value} list={`foreign-key-options-${foreignKeyEditor.column}`} aria-label={t("grid.selectId", {target: foreignKeyEditor.target})} onKeyDown={(event) => {
                    if (event.key === "Enter") finishForeignKeyEdit(event.currentTarget.value);
                    if (event.key === "Escape") setForeignKeyEditor(null);
                }} onBlur={(event) => finishForeignKeyEdit(event.currentTarget.value)}/>
                <datalist id={`foreign-key-options-${foreignKeyEditor.column}`}>{foreignKeyEditor.options.map((option) => <option key={option} value={option}/>)}</datalist>
            </div>}
            {arrayEditor && <div className="array-editor-backdrop" role="presentation" onMouseDown={() => setArrayEditor(null)}>
                <section className="array-editor-dialog" role="dialog" aria-label={t("grid.editColumn", {name: table.columns[arrayEditor.column - 1].name})} onMouseDown={(event) => event.stopPropagation()}>
                    <header><div><span>{t("grid.arrayValues")}</span><h2>{table.columns[arrayEditor.column - 1].name}</h2></div><button type="button" aria-label={t("grid.closeArrayEditor")} onClick={() => setArrayEditor(null)}>x</button></header>
                    <div className="array-entry-list">
                        {arrayEditor.entries.map((entry, index) => <div className="array-entry" key={`${index}:${entry}`}>
                            <code>{index + 1}</code>
                            <input aria-label={t("grid.arrayItem", {index: index + 1})} value={entry} onChange={(event) => setArrayEditor({...arrayEditor, entries: arrayEditor.entries.map((value, position) => position === index ? event.target.value : value)})}/>
                            <button type="button" aria-label={t("grid.moveItemUp", {index: index + 1})} disabled={index === 0} onClick={() => setArrayEditor({...arrayEditor, entries: arrayEditor.entries.map((value, position) => position === index ? arrayEditor.entries[index - 1] : position === index - 1 ? arrayEditor.entries[index] : value)})}>{t("grid.up")}</button>
                            <button type="button" aria-label={t("grid.moveItemDown", {index: index + 1})} disabled={index === arrayEditor.entries.length - 1} onClick={() => setArrayEditor({...arrayEditor, entries: arrayEditor.entries.map((value, position) => position === index ? arrayEditor.entries[index + 1] : position === index + 1 ? arrayEditor.entries[index] : value)})}>{t("grid.down")}</button>
                            <button type="button" aria-label={t("grid.deleteItem", {index: index + 1})} onClick={() => setArrayEditor({...arrayEditor, entries: arrayEditor.entries.filter((_, position) => position !== index)})}>{t("grid.delete")}</button>
                        </div>)}
                    </div>
                    <button className="array-add" type="button" onClick={() => setArrayEditor({...arrayEditor, entries: [...arrayEditor.entries, ""]})}>{t("grid.addItem")}</button>
                    <label className="array-paste">{t("grid.pasteItems")}<textarea aria-label={t("grid.pasteArrayItems")} value={arrayEditor.pasted} placeholder={t("grid.pastePlaceholder")} onChange={(event) => setArrayEditor({...arrayEditor, pasted: event.target.value})}/></label>
                    <div className="dialog-actions"><button type="button" onClick={() => setArrayEditor({...arrayEditor, entries: [...arrayEditor.entries, ...arrayEditor.pasted.replace(/\r/g, "").split("\n").map((entry) => entry.trim()).filter(Boolean)], pasted: ""})} disabled={!arrayEditor.pasted.trim()}>{t("grid.addPastedItems")}</button><button type="button" onClick={() => setArrayEditor(null)}>{t("common.cancel")}</button><button type="button" onClick={saveArray}>{t("common.apply")}</button></div>
                </section>
            </div>}
            {structEditor && <div className="array-editor-backdrop" role="presentation" onMouseDown={() => setStructEditor(null)}>
                <section className="struct-editor-dialog" role="dialog" aria-label={t("grid.editColumn", {name: table.columns[structEditor.column - 1].name})} onMouseDown={(event) => event.stopPropagation()}>
                    <header><div><span>{t("grid.structureValues")}</span><h2>{table.columns[structEditor.column - 1].name}</h2></div><button type="button" aria-label={t("grid.closeStructEditor")} onClick={() => setStructEditor(null)}>x</button></header>
                    <div className="struct-field-list">{(table.columns[structEditor.column - 1].schema.fields ?? []).map((field) => <label key={field.name}>{field.name}<input aria-label={t("grid.structField", {name: field.name})} value={structEditor.values[field.name] ?? ""} placeholder={t("grid.cueExpression")} onChange={(event) => setStructEditor({...structEditor, values: {...structEditor.values, [field.name]: event.target.value}})}/></label>)}</div>
                    <div className="dialog-actions"><button type="button" onClick={() => setStructEditor(null)}>{t("common.cancel")}</button><button type="button" onClick={saveStruct}>{t("common.apply")}</button></div>
                </section>
            </div>}
            {commentTooltip && <div className="cell-comment-tooltip" role="tooltip" style={{left: commentTooltip.left, top: commentTooltip.top}}>{commentTooltip.text}</div>}
        </div>
        <TableExportSettingsPanel table={table} formats={formats} onSave={onSaveExportSettings}/>
        {commentEditor && <div className="dialog-backdrop" role="presentation" onMouseDown={() => setCommentEditor(null)}>
            <form className="structure-dialog" onMouseDown={(event) => event.stopPropagation()} onSubmit={(event) => {
                event.preventDefault();
                void onSaveComment(commentEditor.rowID, commentEditor.column, commentEditor.value).then(() => setCommentEditor(null)).catch((reason) => setCommentError(reason instanceof Error ? reason.message : String(reason)));
            }}>
                <h2>{t("comment.title")}</h2>
                <textarea autoFocus rows={7} aria-label={t("comment.input")} value={commentEditor.value} placeholder={t("comment.placeholder")} onChange={(event) => setCommentEditor({...commentEditor, value: event.target.value})}/>
                {commentError && <div className="schema-editor-error">{commentError}</div>}
                <div className="dialog-actions"><button type="button" onClick={() => setCommentEditor(null)}>{t("common.cancel")}</button><button type="submit">{t("common.apply")}</button></div>
            </form>
        </div>}
    </div>;
}
