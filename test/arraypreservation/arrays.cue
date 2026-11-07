package arrays
#PrimaryKey: string
#Row: {a: *["默认", "保留"] | [...string], n: *[9007199254740993, 1e-400, 1e400] | [...number], q: {"exact-value": *9007199254740993.0 | number, enabled: *true | bool}}
rows: {
 reference: {q: {"exact-value": 9007199254740993.0, enabled: true}, a: ({let seed = ["旧值", "two"], value: seed}).value}
 conjunction: {q: {"exact-value": 9007199254740993.0, enabled: true}, a: ["旧值", "two"] & [...string]}
 default_expr: {q: {"exact-value": 9007199254740993.0, enabled: true}, a: *["旧值", "two"] | [...string]}
 comprehension: {q: {"exact-value": 9007199254740993.0, enabled: true}, a: [for v in ["旧值", "two"] {v}]}
 ellipsis: {q: {"exact-value": 9007199254740993.0, enabled: true}, a: ["旧值", ...string] & ["旧值", "two"]}
 comment: {q: {"exact-value": 9007199254740993.0, enabled: true}, a: ["旧值", // retained source comment
 "two"]}
 raw: {q: {"exact-value": 9007199254740993.0, enabled: true}, a: [#"raw \" comma, slash\\"#, "two"]}
 multiline: {q: {"exact-value": 9007199254740993.0, enabled: true}, a: ["""
    中文, first
    second "quote"
    """, "two"]}
 literal: {n: [9007199254740993, 1e-400, 1e400], q: {"exact-value": 9007199254740993.0, enabled: true}, a: ["中文", "引号\"逗号,反斜杠\\", "tail"]}
 empty: {q: {"exact-value": 9007199254740993.0, enabled: true}, a: []}
 defaulted: {q: {"exact-value": 9007199254740993.0, enabled: true}}
}
_tool: {table: {namespace: "probe", name: "arrays"}, foreignKeys: {}, formulas: {}}
_export: {csv: {enabled: true, columns: ["id", "a", "n", "q"]}}
