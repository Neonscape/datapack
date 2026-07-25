# 数据表编辑器实现方案草稿

本文档描述当前方案的第一版实现方向。它是实现草稿，不替代数据格式 spec。

## 1. 目标与范围

编辑器管理一个由工程文件组织的数据包。数据包包含：

- 表源文件；
- 表内本地 Schema；
- 可跨表复用的公共 Schema；
- 单元格公式；
- 外键约束；
- 导出配置。

第一版提供完整闭环：

1. 打开工程并显示文件树。
2. 创建和编辑表。
3. 编辑本地 Schema 和公共 Schema。
4. 保存合法的受限 CUE 源文件。
5. 编译并显示错误。
6. 检查公式和外键。
7. 导出 CSV。

第一版暂不提供：

- VCS 适配和结构化合并；
- 文件外部修改自动重载；
- 跨表公式取值；
- 任意 CUE 代码编辑；
- 复杂嵌套数据；
- 结构化自动合并。

导出器采用可扩展设计，CSV 是第一版实现的导出格式。以后可以增加 JSON、二进制、SQLite 或代码生成等导出器。

## 2. 技术栈

| 部分 | 选择 |
|---|---|
| 桌面框架 | Wails 2 |
| 核心语言 | Go |
| 前端语言 | TypeScript |
| 前端框架 | React |
| 表格控件 | Glide Data Grid |
| 数据格式与校验 | CUE Go API |
| 公式引擎 | `expr-lang/expr` |
| 文件输出 | Go 标准库和 CUE AST |

Go 核心不依赖 Wails。Wails 只负责把核心能力暴露给 React 前端，命令行工具也调用同一套 Go 核心。

## 3. 总体结构

```text
React / Glide Data Grid
        |
        | Wails bindings
        v
Go 应用入口
        |
        v
工程核心
  工程状态与快照
  文档模型
  CUE 解析与生成
  Schema 注册与依赖
  编辑命令与事务
  编译器
  公式引擎
  外键索引
  导航索引
  持久化
  导出器
```

所有模块只解释一次 CUE、Schema、公式和外键规则。前端不直接解析 CUE，导出器不重新计算数据。

## 4. 数据模型

系统使用四种主要数据形态。

### 4.1 源文档

`SourceDocument` 表示磁盘上的文件，包含：

- 文件路径；
- CUE AST；
- 格式和语法状态；
- 源码位置；
- 文件摘要。

源文档可能语法正确但业务校验失败。

### 4.2 结构化文档

前端编辑结构化文档，不直接编辑 AST。

```go
type TableDocument struct {
    Identity   TableIdentity
    LocalSchema LocalSchema
    Rows       []Row
    Tool       ToolConfig
    Editor     EditorConfig
    Export     ExportConfig
}

type PublicSchemaDocument struct {
    Package     string
    Definitions []SchemaDefinition
}
```

表文档包含表数据、本地 Schema、公式、外键和其他元信息；公共 Schema 文档包含可复用定义。

### 4.3 工程快照

`WorkspaceSnapshot` 表示某个 revision 下的完整工程状态：

- 工程配置；
- 所有表和公共 Schema；
- 表名和 package 索引；
- Schema 依赖图；
- 主键索引；
- 外键正向和反向索引；
- 公式依赖图；
- 诊断信息；
- 每个文档的保存和编译状态。

编译和导出都针对固定快照执行。

### 4.4 编译结果

`CompiledPackage` 是可以交给导出器的只读数据：

- 已应用默认值；
- 已计算公式；
- 已确定最终类型；
- 已完成 CUE 校验；
- 已完成外键校验；
- 具有稳定的表、行和列顺序。

有错误时不生成可导出的 `CompiledPackage`。

## 5. Schema 模型

Schema 编辑器使用受限的结构化模型，不支持任意 CUE 代码。

### 5.1 类型

第一版支持：

- 字符串；
- 整数；
- 实数；
- 布尔值；
- 枚举；
- 数组；
- 结构体；
- Schema 引用。

选择类型后，编辑器显示对应的属性。

### 5.2 字段属性

字段 Schema 由以下部分组合：

```text
基础类型
可复用 requirements
本字段 requirements
是否允许 null
默认值
外键配置
```

公共或本地 Schema 可以作为基础类型，也可以放入 `requirements[]` 作为附加限制。两者可以同时存在。

例如：

```text
price
  基础类型：整数
  requirements：common.#Positive
  本地限制：小于 9999
  允许 null：否
```

`requirements[]` 只支持结构化限制，不支持任意代码。第一版可以支持：

- 大于、大于等于；
- 小于、小于等于；
- 字符串非空；
- 正则表达式；
- 最小和最大长度；
- 数组唯一性；
- 对类型兼容的可复用 Schema 引用。

每个限制有适用类型。整数限制不能放到字符串字段上。类型不兼容时，Schema 修改不能应用。

### 5.3 结构体

结构体定义由 Schema 编辑器维护。

支持：

- 字段内匿名结构体；
- 表本地命名结构体；
- 公共命名结构体。

第一版的最终结构保持扁平：

- 结构体成员可以是基础值、枚举或外键；
- 结构体成员不能是数组或结构体；
- 数组元素不能是结构体；
- 整个结构体不能设置公式。

Schema 分为两类：

- 类型 Schema：带有一个完整基础类型；字段最多使用一个类型 Schema。
- 约束 Schema：不提供基础类型，只声明适用类型和附加限制；字段可以使用多个。

字段也可以直接选择 CUE 支持的内置类型：字符串、整数、实数、布尔值、枚举、数组或结构体。引用类型 Schema 是另一种类型来源，不是内置类型。

类型 Schema 和约束 Schema 都通过文件中的明确字段区分，编辑器不从 CUE 表达式猜测其类别。

### 5.4 Schema 引用

Schema 引用使用完整逻辑名称，例如：

```text
common.#NonNegativeInt
rpg.item._local.#Position
```

引用图包含所有基础类型引用、requirements 引用、结构体成员引用和数组元素引用。

第一版禁止直接或间接循环引用：

```text
A -> A
A -> B -> A
A -> B -> C -> A
```

新增或修改引用时，核心先在候选状态中检查依赖图。发现循环则拒绝应用。

依赖方向固定如下：

- 公共 Schema 可以引用公共约束 Schema；
- 本地 Schema 可以引用当前表本地 Schema 或公共 Schema；
- 表字段可以引用本地或公共 Schema；
- 公共 Schema、其他表和其他表字段不能引用本地 Schema；
- 类型 Schema 不能引用另一个类型 Schema；
- 所有 Schema 不能直接或间接循环引用。

删除任何仍被引用的类型 Schema 或约束 Schema 时，直接拒绝操作并列出引用位置。重命名 Schema 时更新全部引用。

## 6. 工程树和文档打开

工程树同时展示物理文件和逻辑结构：

```text
game-data
├── Tables
│   ├── economy.item
│   │   ├── table
│   │   └── local_schema
│   │       ├── #PrimaryKey
│   │       ├── #Position
│   │       └── #Row
│   └── combat.effect
│       ├── table
│       └── local_schema
│           ├── #PrimaryKey
│           └── #Row
├── Schemas
│   └── common
│       ├── #GameID
│       ├── #Positive
│       └── #Rarity
└── Project Settings
```

`table` 和 `local_schema` 是同一表文件的两个逻辑视图，不是两个文件。

- 双击 `table` 打开 Glide Data Grid；
- 双击 `local_schema` 打开本地 Schema 编辑器；
- 双击公共 Schema 定义打开公共 Schema 编辑器并选中该定义；
- `exports/` 不在源文件树中作为可编辑内容显示。

文档身份使用 `DocumentRef`：

```text
table:rpg.item
local-schema:rpg.item
public-schema:game-data/schema/common#Positive
project-settings
```

## 7. 主要 Go 模块

### 7.1 `workspace`

负责工程生命周期和当前快照：

- 打开和关闭工程；
- 扫描工程入口和目录；
- 管理 revision；
- 发布新快照；
- 选择编译范围；
- 读取并应用用户级编辑器设置中的编译模式。

对上层提供少量高层操作，不让前端自己编排解析、索引和编译步骤。

### 7.2 `projecttree`

根据工程配置和快照生成工程树。它负责：

- 表目录；
- 公共 Schema 目录；
- 表下的 `table` 和 `local_schema` 视图；
- 定义级 Schema 节点；
- 损坏文件占位；
- dirty、错误和待编译状态。

工程树是派生数据，不是独立数据源。

### 7.3 `document`

负责统一文档身份、加载状态和结构化文档访问。

### 7.4 `cuesource`

负责：

- 解析受限 CUE；
- 将 AST 转换为结构化文档；
- 将结构化文档稳定地生成 CUE；
- 保持固定分区、行列顺序和格式；
- 建立源位置到表格位置的映射；
- 拒绝编辑器不支持的 CUE 写法。

它不负责公式计算和外键校验。

### 7.5 `schemamodel`

负责 `SchemaNode`、`SchemaDefinition`、`Requirement` 和 Schema 引用模型。

```go
type SchemaNode struct {
    Kind         SchemaKind
    Nullable     bool
    Default      *Value
    Requirements []Requirement
    Reference    *SchemaReference
    Element      *SchemaNode
    Fields       []SchemaField
}
```

### 7.6 `schemaregistry`

负责：

- 注册本地和公共 Schema；
- 解析 Schema 引用；
- 检查引用是否存在；
- 检查类型兼容性；
- 检查循环引用；
- 建立 Schema 依赖和使用位置索引；
- 计算 Schema 修改的影响范围。

### 7.7 `commands`

所有编辑通过命令执行，不由前端直接修改工程对象。

表数据命令：

```text
SetCellValue
ClearCell
SetCellFormula
RemoveCellFormula
AddRow
DeleteRow
RenamePrimaryKey
PasteRange
```

表结构命令：

```text
AddColumn
DeleteColumn
RenameColumn
MoveColumn
SetColumnSchema
SetForeignKey
RemoveForeignKey
```

Schema 命令：

```text
CreateSchemaDefinition
DeleteSchemaDefinition
RenameSchemaDefinition
SetSchemaType
SetSchemaDefault
SetSchemaNullable
SetSchemaRequirements
AddStructField
DeleteStructField
SetSchemaReference
```

### 7.8 `transaction`

负责把命令变成一次完整的编辑事务：

1. 检查命令前置条件。
2. 基于当前快照创建候选状态。
3. 修改候选文档。
4. 更新候选 Schema、主键、外键和公式索引。
5. 生成所有受影响的源文件。
6. 重新解析生成结果。
7. 检查文件语法和受限格式。
8. 按当前编译模式执行增量编译或标记待编译。
9. 安全写入源文件。
10. 发布新快照。
11. 更新撤销记录和前端视图。

业务校验错误允许保存，但语法错误和受限格式错误阻止保存。

### 7.9 `formula`

负责：

- 解析公式 AST；
- 解析本表行和字段引用；
- 建立依赖图；
- 检查自引用和循环；
- 计算公式；
- 应用默认值作为引用输入；
- 将结果转换为目标字段类型；
- 拒绝公式读取外键字段或外键数组。

第一版公式不支持跨表取值。

### 7.10 `foreignkey`

负责：

- 读取 `_tool.foreignKeys`；
- 建立主键索引；
- 检查外键字段类型；
- 检查标量和数组外键；
- 建立反向引用索引；
- 提供删除保护和主键改名更新；
- 为编辑器提供目标表主键列表。

外键只检查目标主键是否存在，不读取目标行的其他字段。

### 7.11 `compiler`

提供两个入口：

```go
CompileAffected(snapshot, changes) CompileResult
CompileAll(snapshot) CompileResult
```

增量编译用于自动编译。手动编译只调用完整编译。导出前始终调用完整编译。

编译顺序：

1. 解析工程和文档。
2. 检查工程、package、表名和配置。
3. 解析 Schema 并检查 Schema 依赖。
4. 构造字段类型和最终规则。
5. 检查固定值和公式互斥。
6. 解析公式并建立依赖图。
7. 计算公式和默认值。
8. 构造最终行数据。
9. 使用 CUE 与 `#Row` 统一校验。
10. 检查外键。
11. 无错误时生成 `CompiledPackage`。

### 7.12 `navigation`

负责统一的“跳转到定义”和“查找引用”。

支持跳转：

- Schema 引用到 Schema 定义；
- 本地 Schema 引用到当前表的 `local_schema`；
- 公共 Schema 引用到公共 Schema 定义；
- 外键值到目标表的主键单元格；
- 公式引用到目标单元格；
- 主键到所有外键引用位置。

前端交互：

- Ctrl+左键：跳转到定义；
- F12：跳转到定义；
- 右键菜单：跳转到定义、查找引用。

核心使用统一位置模型：

```go
type SemanticLocation struct {
    Document DocumentRef
    Table    string
    Schema   SchemaID
    Row      string
    Field    FieldPath
    Item     *int
}
```

导航目标返回文档和具体位置。前端负责打开文档、展开树节点、选中目标并滚动到位置。

### 7.13 `diagnostics`

统一管理错误和警告。诊断至少包含：

```text
严重级别
错误代码
消息
文件
表或 Schema
行
字段路径
数组下标
目标位置
```

业务错误可以随文件保存，但会阻止编译结果和导出结果产生。

### 7.14 `persistence`

负责：

- 生成临时文件；
- 重新解析生成内容；
- 原子替换单个文件；
- 多文件操作的备份和回滚；
- 防止旧 revision 覆盖新 revision；
- 判断文件是否已被其他程序修改。

第一版不主动监听外部文件变化，也不自动重载外部文件。

### 7.15 `export`

导出器只接收 `CompiledPackage`：

```go
type Exporter interface {
    Name() string
    Validate(pkg *CompiledPackage, config ExportConfig) []Diagnostic
    Build(pkg *CompiledPackage, config ExportConfig) ([]Artifact, error)
}
```

第一版实现 `CSVExporter`。导出器不重新计算公式、不应用默认值、不检查外键。

## 8. 前端模块

### 8.1 `AppShell`

负责窗口布局、工具栏、文档标签、错误面板、编译按钮和导出按钮。

### 8.2 `ProjectExplorer`

负责显示和操作工程树。它只发送结构化命令，不直接访问文件系统。

### 8.3 `DocumentManager`

负责打开文档、标签、当前活动文档、未提交草稿和导航目标。

### 8.4 `TableEditor`

使用 Glide Data Grid，负责：

- 单元格编辑；
- 行和列操作；
- 复制粘贴；
- 数组和一层结构体编辑；
- 公式设置；
- 外键选择；
- 默认值和错误显示；
- 单元格和数组元素定位。

表格编辑器不保存完整工程副本。它只保存当前视图、选区和正在输入的临时文本。

### 8.5 `SchemaEditor`

本地和公共 Schema 使用同一套编辑器。

界面结构：

```text
Schema 定义树
        |
        v
Schema 属性编辑器
        |
        v
基础类型 / requirements / 默认值 / 可空 / 结构体成员
```

Schema 表单修改只存在于前端草稿中。用户点击“应用”后才：

1. 发送一个 Schema 编辑命令。
2. 后端生成候选文档。
3. 检查 Schema 引用和循环。
4. 重新解析文件。
5. 编译受影响表。
6. 通过文件结构检查后保存。
7. 发布新快照。

业务数据错误不会阻止 Schema 应用，但会在诊断面板中显示。

### 8.6 `RequirementEditor`

根据基础类型显示可用限制。用户只能选择结构化 Requirement，并填写参数。

例如整数只显示：

```text
大于
大于等于
小于
小于等于
可复用整数 Schema
```

字符串只显示：

```text
非空
正则
最小长度
最大长度
可复用字符串 Schema
```

### 8.7 `SchemaUsagePanel`

显示当前公共或本地 Schema 的使用位置，支持点击跳转。

### 8.8 `DiagnosticsPanel`

显示语法错误、Schema 错误、公式错误、外键错误和导出错误。点击诊断后跳转到具体表格单元格、Schema 节点或数组元素。

### 8.9 `BuildPanel`

显示当前编译模式、待编译状态、编译进度、诊断数量和导出结果。

## 9. 编译模式

第一版提供两种用户级编辑器设置：

```text
自动编译
手动编译
```

### 9.1 自动编译

表格编辑命令提交并保存后：

1. 生成新的工程 revision。
2. 标记受影响文档。
3. 按编辑器设置中的等待时间 debounce；默认 10 秒后执行增量编译。
4. 只发布最新 revision 的编译结果。
5. 更新单元格错误、Schema 错误和工程状态。

连续修改时，可以取消或丢弃过期的编译任务。

### 9.2 手动编译

编辑命令提交并保存后：

1. 生成新的工程 revision。
2. 更新文档和文件树。
3. 标记工程为“待编译”。
4. 不执行公式计算、完整 CUE 校验和外键检查。

用户点击“编译”按钮或使用快捷键后，执行 `CompileAll`。手动编译第一版不提供增量编译按钮；底部状态栏显示上一次编译结果已过期。

编译模式和等待时间来自用户级 `settings.json`，不保存在 `.dtproj`，也不随工程提交。

### 9.3 导出

导出不依赖当前编译模式：

1. 固定当前工程快照。
2. 强制执行 `CompileAll`。
3. 有任意错误则终止，不更新导出产物。
4. 无错误时生成 `CompiledPackage`。
5. 调用目标导出器。
6. 将所有产物写入临时目录。
7. 所有产物成功后再替换正式导出目录。

导出目录属于生成结果，不提交版本控制，也不作为源文件编辑。

## 10. 关键事务流程

### 10.1 打开工程

```text
读取工程文件
  -> 扫描表和 Schema 入口
  -> 构造工程树
  -> 解析公共 Schema
  -> 解析表身份、本地 Schema 和表数据
  -> 建立 Schema 依赖图
  -> 建立主键和外键索引
  -> 建立公式依赖图
  -> 完整编译
  -> 显示工程树和诊断
```

### 10.2 编辑表格单元格

```text
输入过程只保留在前端
  -> 用户确认单元格
  -> 发送 SetCellValue
  -> 构造候选快照
  -> 生成并重新解析 CUE
  -> 结构正确则保存
  -> 自动模式：增量编译
  -> 手动模式：标记待编译
  -> 发布新快照
```

非法业务值可以保存，但不能导出。

### 10.3 应用 Schema 修改

```text
打开 local_schema 或公共 Schema
  -> 修改表单草稿
  -> 点击应用
  -> 发送 Schema 命令
  -> 检查引用目标、类型兼容和循环
  -> 生成并重新解析 CUE
  -> 找出受影响表
  -> 自动模式：增量编译
  -> 手动模式：标记待编译
  -> 保存源文件
  -> 发布新快照
```

### 10.4 主键改名

```text
检查反向外键索引
  -> 修改目标主键
  -> 更新所有外键值
  -> 更新本表公式中的行引用
  -> 生成所有受影响文件
  -> 重新解析和编译
  -> 多文件保存
  -> 发布新快照
```

### 10.5 公共 Schema 改名

```text
检查 Schema 使用位置
  -> 修改定义名称
  -> 更新所有 Schema 引用
  -> 生成所有受影响文件
  -> 检查引用和循环
  -> 重新解析和编译
  -> 多文件保存
  -> 发布新快照
```

## 11. Wails 接口

Wails 只暴露高层操作：

```go
type App struct {
    workspace *Workspace
}

func (a *App) OpenProject(path string) (WorkspaceView, error)
func (a *App) GetDocument(ref DocumentRef) (DocumentView, error)
func (a *App) Execute(command Command) CommandResult
func (a *App) Compile() CompileResult
func (a *App) Export(request ExportRequest) ExportResult
func (a *App) Navigate(location SemanticLocation) (NavigationTarget, error)
```

前端使用 revision 发送命令。核心发现 revision 过期时拒绝命令并要求前端刷新视图。

大型表格不在每次编辑后传输完整数据。初次打开获取表视图，后续使用 patch；Glide Data Grid 可以按可见区域读取行。

## 12. 建议目录结构

```text
cmd/
    datapack/

internal/
    app/
    workspace/
    projecttree/
    document/
    cuesource/
    schemamodel/
    schemaregistry/
    commands/
    transaction/
    compiler/
    formula/
    foreignkey/
    navigation/
    diagnostics/
    persistence/
    export/
        csv/

frontend/src/
    app/
    project-explorer/
    documents/
    table-editor/
    schema-editor/
    diagnostics/
    build-panel/
    navigation/
    bindings/
```

实现时可以先将多个职责放在较少的 Go package 中，等接口稳定后再拆分。最先需要稳定的核心接口是：

```text
Execute(command)
CompileAll(snapshot)
Export(compiledPackage)
Navigate(location)
```
