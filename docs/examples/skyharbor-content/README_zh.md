# Skyharbor Content 完整学习工程

Skyharbor Content 是 Datapack 的完整学习示例工程。它模拟了一个小型多人游戏的战斗内容配置，包含效果、道具、可操作角色和等级成长数据。

工程规模控制在可以一次读完的范围内，但每张表都侧重演示编辑器的不同能力。请打开 `skyharbor-content.dtproj` 开始学习。

## 学习路径

1. 打开 `common` Schema 文件，查看 `#GameID`、`#Rarity`、`#EffectKind`、`#NonEmptyString`、`#NonNegative` 和 `#Probability`。
2. 打开 `content.effect`，尝试编辑名称、枚举、数字、可空描述和标签数组。
3. 打开 `content.item`，使用 `effectID` 的外键候选值，然后通过单元格右键菜单添加评论。`effectIDs` 数组展示了数组在 Datapack 中是一个完整单元格，而不是多列普通表格数据。
4. 编辑 `position` 结构体。它仍然是一个单元格，其成员由当前表的本地 Schema `_local.#Position` 定义。
5. 查看 `content.item` 的 `sellPrice`。它显示的是公式预览结果，不是写入 `rows` 的固定值；修改 `price` 后，重新编译会更新预览结果。
6. 打开 `content.character`，从 `startingItem` 外键跳转到对应道具。`attack` 列由同一行中的字段计算得到。
7. 打开 `balance.growth` 并切换到手动编译模式。修改 `baseValue`，观察待编译状态，然后主动执行编译。这里的公式使用完整逻辑表名引用其他行。
8. 选中单元格后使用公式栏，尝试粘贴一个矩形数据区域，然后用 Undo/Redo 对比源数据和显示结果的变化。
9. 使用 Schema 编辑器创建或重命名定义。通过工程树观察公共 Schema 文件和表内本地 Schema 定义的区别。
10. 配置项目级导出表选择和表级导出列，然后执行 CSV 导出。导出文件位于 `exports/`；数组和结构体使用 JSON 编码。

应用级功能也可以使用这个工程练习：通过“文件 > 新建工程”和“文件 > 另存为”了解工程生命周期；通过工程树右键菜单创建表或 Schema 文件；通过表和 Schema 的右键菜单执行重命名、移动和删除。

这些操作会改变工程结构，因此不会预先写入示例源文件。功能测试和 Playwright 测试会在临时副本中执行对应流程，不会修改 `docs/examples` 中的原始工程。

## Datapack 与普通表格的区别

Datapack 不是一个额外隐藏了公式层的普通电子表格：

- 表中的 `#Row` 是每一行数据必须满足的结构契约。单元格显示值由 CUE 表达式计算得到，因此显示值可能与源表达式不同。
- 公式保存在 `_tool.formulas` 中，在预览和导出时求值。公式结果不会被复制回 `rows`。
- 外键用于验证目标 ID 是否存在并提供导航，不允许公式通过外键读取目标表的任意字段。
- 公共 Schema 是工程内可复用的共享规则。本地 Schema 只属于一张表，不能被其他表引用。
- 数组和结构体仍然是单个强类型单元格。编辑器会提供结构化编辑界面，CSV 编码方式则需要显式配置。
- 可空字段与空字符串不是同一个概念。清空单元格时，编辑器会根据类型写入安全默认值，而不是简单删除字段。
- 即使项目导出设置排除了某张表的输出文件，执行导出前仍然会完整编译和校验整个工程。

## 功能分布

### `content.effect`

演示：

- 主键和公共类型 Schema；
- 字符串、整数和数字；
- 枚举字段；
- 数字范围约束；
- 可空字符串；
- 字符串数组；
- JSON 数组 CSV 编码。

### `content.item`

演示：

- 标量外键 `effectID`；
- 数组外键 `effectIDs`；
- 本地类型 Schema `_local.#Position`；
- 结构体字段 `position`；
- 可空描述；
- 同行公式 `sellPrice`；
- 数组和结构体的 CSV 编码；
- 外键导航、重命名同步和删除保护。

### `content.character`

演示：

- 本地约束 Schema `_local.#GrowthRate`；
- 公共类型和公共约束的组合；
- nullable 外键；
- 同行公式 `attack`；
- 单元格评论及悬停提示。

### `balance.growth`

演示：

- 本地整数约束 `_local.#Level`；
- 同表跨行公式；
- 完整逻辑表名引用；
- 公式依赖顺序；
- 手动编译和待编译状态。

## 建议实验

初始工程可以完整编译且没有诊断。进行破坏性实验前，建议先使用“另存为”创建副本：

- 将某个 `effectID` 改成不存在的值，然后通过诊断跳转到错误单元格。
- 在仍有道具引用 `burn` 时尝试删除该效果，观察编辑器如何保护被引用行。
- 重命名 `mend`，然后检查 `content.item` 中的外键是否同步更新。
- 将 `chance` 设置为大于 `1` 的值，查看约束诊断，再使用 Undo 撤销。
- 在本地 Schema 仍被列使用时尝试删除它，观察引用保护信息。
- 修改公式使其引用不存在的行，查看公式诊断和定位行为。
- 修改表级 CSV 字段选择，比较导出文件的列变化。
- 暂时取消某张表的项目级导出选择，确认它不会生成文件，但仍然参与完整编译。

## 基础功能覆盖

这个学习工程和自动化测试共同覆盖 Datapack 的基础功能。

示例工程直接覆盖：

- 工程入口和逻辑 namespace；
- 多表工程树和标签页；
- 公共 type Schema 和 constraint Schema；
- 本地 type Schema 和 constraint Schema；
- Schema requirements 和引用导航；
- 字符串、数字、整数、枚举、数组、结构体和 nullable 值；
- 标量外键和数组外键；
- 同行公式和同表跨行公式；
- 单元格评论和 tooltip；
- 编译、诊断和导航；
- 项目级与表级 CSV 配置；
- 数组和结构体的 CSV 序列化；
- 编辑历史和 Undo/Redo。

临时工程测试覆盖：

- 新建工程和另存为；
- 新建、重命名和删除表；
- 新建 Schema 文件；
- 新建、重命名和删除 Schema 定义；
- 新建、重命名和删除行；
- 新建、删除、重命名和移动列；
- 清空单元格和类型默认值；
- 多单元格粘贴；
- 自动编译和手动编译；
- 项目导出表选择和表级导出字段选择；
- 中英文界面；
- light、dark 和 system 主题；
- 全局 Undo/Redo。

## 源文件结构

```text
skyharbor-content.dtproj                    工程入口
cue.mod/module.cue                          CUE module 标识
schema/common.cue                           公共类型和约束
```

自动生成的 `exports/` 目录已从源文件扫描中排除，不属于需要维护的输入数据。
