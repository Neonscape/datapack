package arrays
import "list"
#PrimaryKey: string
#Row: {a: *["默认", "保留"] | [...string], n: *[9007199254740993, 1e-400, 1e400] | [...number], q: {"exact-value": *9007199254740993.0 | number, enabled: *true | bool}}
_seed: ["旧值", "two"]
rows: {
 reference: {a: _seed}
 conjunction: {a: ["旧值", "two"] & [...string]}
 concat: {a: list.Concat([_seed, ["tail"]])}
 comprehension: {a: [for v in _seed {v}]}
 ellipsis: {a: ["旧值", ...string] & _seed}
 comment: {a: ["旧值", // retained source comment
 "two"]}
 raw: {a: [#"raw \" comma, slash\\"#, "two"]}
 multiline: {a: ["""
    中文, first
    second "quote"
    """, "two"]}
 literal: {a: ["中文", "引号\"逗号,反斜杠\\", "tail"]}
 empty: {a: []}
 defaulted: {}
}
_tool: {table: {namespace: "probe", name: "arrays"}, foreignKeys: {}, formulas: {}}
_export: {csv: {enabled: true, columns: ["id", "a", "n", "q"]}}
