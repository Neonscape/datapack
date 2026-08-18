# 功能测试

本目录验证第一版编辑器的**可观察行为**：打开工程、编辑、编译、导出。它不替代 `internal/*/*_test.go` 里的包内单元测试。

## 边界

测试只调用 `workspace.Service`（Wails `app.go` 转发给前端的同一套入口）。

| 测 | 不测 |
|---|---|
| 打开 `docs/examples` 样例、编译、导出 CSV | 把现有 `*_test.go` 搬进来 |
| 空工程走完建表、填行、公式、导出 | 真实 Wails 窗口 / Playwright |
| 诊断可定位，且挡住导出 | 公式白名单、Git 冲突、外部修改摘要（尚未实现） |
| 业务错误可保存；非法 CUE 不写盘 | `app.go` 的文件对话框 |

前端 UI 测试仍在 `frontend/tests/`，后端全部 mock，不证明数据包语义。

## 怎么跑

在仓库根目录：

```bash
go test ./test/functional
go test ./test/functional -count=1
```

`go test ./...` 会一并跑到这里。不需要 Wails、Node 或显示服务器。

## 布局

```text
test/
  README.md
  functional/           # package functional；只导入 workspace 等公开类型
    helpers_test.go
    examples_test.go    # 三个样例工程 → 编译干净 → CSV 与金标准一致
    mvp_loop_test.go    # 空工程走完 design.md §10 闭环
    validation_test.go  # 错误可定位，Export 失败
    persistence_test.go # 写盘语义
  testdata/golden/      # 期望 CSV；字面量，不用实现再算一遍
```

样例工程从 `docs/examples/` **复制到临时目录**再打开，避免测试把 `exports/` 写进文档树。

## 金标准怎么来的

CSV 期望值来自样例源文件和 `docs/design.md`，不是从导出器抄的：

- `character-growth-manual`：`level_01.attack = 10 * 2 = 20`，`level_10.attack = 20 + 20 = 40`，`level_11.attack = 40 + 5 = 45`
- `basic-rpg-auto`：道具行与效果行的固定值；数组/结构体按紧凑 JSON 写入一个单元格，再按 RFC 4180 转义引号
- `schema-composition`：两行武器的字面字段

金标准文件改了，测试就会失败。若实现变了但规格没变，应改代码而不是改金标准。

## 已知产品缺口（测试会绕过）

`workspace.Service` 目前没有设置表级 `_export.csv.columns` 或 `_tool.foreignKeys` 的方法。`CreateTable` 生成的表也不能直接导出。

`mvp_loop_test.go` 在 Service 建表、加列、加行之后，用测试辅助函数把 `_export` 写进源文件，再调用 `Export()`。外键用例则通过夹具 CUE 声明 `_tool.foreignKeys`，仍然只通过 `OpenProject` / `UpdateCell` / `Export` 观察行为。

`CreateSchemaFile` 生成的空 Schema 包也不会出现在工程树里，因此 `CreateSchemaDefinition` 无法以它为路径。`mvp_loop_test.go` 会先写入一个已有定义的公共 Schema 文件，再通过 Service 追加约束。

这不是推荐产品用法，只是当前公开 API 的边界。
