package fks
#PrimaryKey: string
#Row: {ids: [...string]}
rows: {alpha: {ids: ({let seed = ["first", "second"], value: seed}).value}}
_tool: {table: {namespace: "probe", name: "fks"}, foreignKeys: {ids: {table: "probe.targets"}}, formulas: {}}
_export: {csv: {enabled: true, columns: ["id", "ids"]}}
