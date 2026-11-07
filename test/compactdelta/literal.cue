package compact
// Independent literals; changing token lengths must keep later spans intact.
#PrimaryKey: string
#Row: {n: int, f: number, text: *"default" | string, nullable: *null | number | string, marker: bool, spare: int}
rows: {
 alpha: {n: 1, f: 0.5, text: "初始", nullable: null, marker: true, spare: 7}
 beta: {n: -9007199254740993, f: 0.12345678901234567890123456789, text: "untouched", nullable: null, marker: false, spare: -7}
}
_tool: {table: {namespace: "probe", name: "compact"}, foreignKeys: {}, formulas: {}}
_export: {csv: {enabled: true, columns: ["id", "n", "f", "text", "nullable", "marker", "spare"]}}
