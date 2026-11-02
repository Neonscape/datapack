import type {CellDelta, TableDocument} from "../types";
export function patchCell(table:TableDocument, delta:CellDelta):TableDocument | null {
    if (table.path !== delta.path || table.projectEpoch !== delta.projectEpoch || table.workspaceRevision !== delta.baseRevision || delta.revision <= delta.baseRevision || table.rows[delta.rowIndex]?.id !== delta.rowID || !table.columns.some(column=>column.name === delta.column)) return null;
    const rows = table.rows.slice();
    const row = rows[delta.rowIndex];
    rows[delta.rowIndex] = {...row, values:{...row.values,[delta.column]:delta.cell}};
    return {...table, workspaceRevision:delta.revision, rows};
}
