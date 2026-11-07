package defaults
#PrimaryKey: string
#Row: {
    q: {
        largeI: *9007199254740993 | int
        "exact-value": *9007199254740993.0 | number
        fraction: *9007199254740993.125 | number
        negative: *-9007199254740995.0 | number
        long: *0.12345678901234567890123456789 | number
        huge: *1e400 | number
        tiny: *1e-400 | number
        small: *0.1250 | number
        whole: *2.0 | number
        enabled: *true | bool
        "a:b": *"中文, \"引号\"\\路径\n下一行" | string
        nullable: *null | string
    }
    branch: *{n: 9007199254740993.0, enabled: true} | {n: number, enabled: bool}
    editable: {n: *9007199254740993.0 | number, enabled: *true | bool}
    normal: *0.1250 | number
    flag: *false | bool
    text: *"普通文本" | string
    emptyList: *[] | [...number]
}
rows: {
    missing: {}
    empty: {q: {}}
    explicit: {q: {"exact-value": 9007199254740995.0, enabled: false}}
}
_tool: {table: {namespace: "probe", name: "defaults"}}
_export: {csv: {enabled: true, columns: ["id", "q", "branch", "editable", "normal", "flag", "text", "emptyList"]}}
