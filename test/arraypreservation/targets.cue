package targets
#PrimaryKey: string
#Row: {label: string}
rows: {first: {label: "第一"}, second: {label: "第二"}, third: {label: "第三"}}
_tool: {table: {namespace: "probe", name: "targets"}, foreignKeys: {}, formulas: {}}
_export: {csv: {enabled: true, columns: ["id", "label"]}}
