# 样例工程

每个子目录都是一个独立的数据工程，包含完整 `.dtproj`、公共 Schema、表和导出配置示例。

| 工程 | 重点 |
|---|---|
| `basic-rpg-auto` | 公共类型/约束 Schema、本地结构体、数组外键。 |
| `character-growth-manual` | 本表跨行公式、完整表名引用。 |
| `schema-composition` | 类型 Schema、约束 Schema 和字段 requirements 的组合。 |

所有样例工程都有独立 `cue.mod/module.cue`，可以作为 CUE 模块读取；它们的格式约定见 `../dtproj-format-v1.md` 和 `../restricted-cue-schema-v1.md`。自动或手动编译属于用户级设置，见 `../editor-settings-format-v1.md`。
