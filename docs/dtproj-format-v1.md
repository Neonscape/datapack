# dtproj 格式 v1

`.dtproj` 是数据工程的 JSON 入口文件。用户通过它打开工程；它决定哪些目录包含表、哪些目录包含公共 Schema。

## 完整示例

```json
{
  "version": 1,
  "tableRoots": ["tables"],
  "schemaRoots": ["schema"],
  "exclude": ["exports", ".git"],
  "export": {
    "directory": "exports",
    "targets": ["csv"]
  }
}
```

## 字段

| 字段 | 必填 | 说明 |
|---|---:|---|
| `version` | 是 | 工程格式版本。v1 固定为 `1`。 |
| `tableRoots` | 是 | 相对工程根目录的表目录列表。 |
| `schemaRoots` | 是 | 相对工程根目录的公共 Schema 目录列表。 |
| `exclude` | 否 | 扫描时忽略的相对路径。 |
| `export.directory` | 是 | 导出产物目录。该目录不作为源文件维护。 |
| `export.targets` | 是 | 启用的导出器名称。v1 支持 `csv`。 |

编译模式、自动编译等待时间和其他本机编辑体验选项不属于工程文件，见 `editor-settings-format-v1.md`。这些设置不会进入版本控制。

## 扫描规则

- `tableRoots` 下的 `.cue` 文件按表文件解析，必须包含 `_tool.table`、`#Row` 和 `rows`。
- `schemaRoots` 下的 `.cue` 文件按公共 Schema 文件解析，不能包含 `rows` 或 `_tool.table`。
- 同一个文件不能同时属于表目录和 Schema 目录。
- 目录结构不决定表的 namespace。表的完整逻辑名称完全由 `_tool.table.namespace` 和 `_tool.table.name` 决定。
- 工程内 CUE package 不允许重复；一个公共 Schema 文件只能包含一个 package。
