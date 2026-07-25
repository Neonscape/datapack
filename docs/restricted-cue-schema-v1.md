# 受限 CUE Schema 格式 v1

本格式用于公共 Schema 文件和表中的本地 Schema。编辑器生成并管理这些内容，不承诺编辑任意合法 CUE。

## 1. Schema 分类

每个命名 Schema 都在 `_schema.definitions` 中显式声明类别：

- `type`：类型 Schema。它带有一个完整基础类型，可以附带限制。
- `constraint`：约束 Schema。它不提供基础类型，只声明可适用的类型和限制。

编辑器不从 CUE 表达式推断类别。

## 2. 公共 Schema 文件

```cue
package common

#GameID: string & =~"^[a-z][a-z0-9_]*$"
#NonNegative: >=0

_schema: {
    version: 1
    definitions: {
        GameID: {
            kind: "type"
            type: "string"
        }
        NonNegative: {
            kind: "constraint"
            appliesTo: ["integer", "number"]
        }
    }
}
```

规则：

- 一个文件只能有一个 CUE package。
- 每个 `#Definition` 都必须在 `_schema.definitions` 中登记。
- `type` Schema 的 CUE definition 必须提供完整基础类型。
- `constraint` Schema 的 CUE definition 不能提供基础类型，`appliesTo` 必须非空。
- 公共 Schema 只能引用公共约束 Schema，不能引用本地 Schema 或另一个类型 Schema。

## 3. 本地 Schema

本地 Schema 位于表文件的 `_local` 中，并在 `_schema.localDefinitions` 中登记。

```cue
_local: {
    #Discount: >=0
    #Position: {
        x: number
        y: number
    }
}

_schema: {
    localDefinitions: {
        Discount: {
            kind: "constraint"
            appliesTo: ["integer", "number"]
        }
        Position: {
            kind: "type"
            type: "struct"
        }
    }
}
```

本地 Schema 只能被当前表的 `#Row`、`#PrimaryKey`、本地 Schema 和字段 requirements 使用。其他表和公共 Schema 不得引用它。

## 4. 字段定义

`#Row` 的每个成员是一列。字段使用 CUE 的 `&` 组合基础类型、类型 Schema、约束 Schema 和内联限制。

```cue
#Row: {
    name: string & != ""
    price: int & common.#NonNegative & <9999
    rarity: common.#Rarity
    position: _local.#Position
}
```

字段的类型来源只能是以下之一：

- CUE 内置类型：`string`、`int`、`number`、`bool`；
- 内联枚举；
- 内联数组或结构体；
- 一个类型 Schema 引用。

字段可以额外组合零个或多个约束 Schema 和内联限制。类型不兼容或组合后不可满足时，Schema 编辑器拒绝应用修改。

## 5. 可空和默认值

`null` 表示缺失，空字符串是实际字符串。

```cue
#Row: {
    description: string | null
    maxStack: *1 | int & >=1
}
```

- `string | null` 允许缺失；
- `*1 | int & >=1` 在缺失时使用默认值 `1`；
- 没有默认值且最终值为 `null` 时，是否允许由字段定义决定；
- 新增必填字段时，编辑器必须要求默认值或为所有已有行填写实际值。

## 6. 结构限制

- 数组元素只能是基础值、枚举或外键值；
- 结构体成员只能是基础值、枚举或外键值；
- 不支持数组嵌套数组；
- 不支持数组元素为结构体；
- 不支持结构体嵌套结构体或数组；
- 类型 Schema 和约束 Schema 都禁止直接或间接循环引用。

## 7. 删除和改名

- 删除仍被引用的类型 Schema 或约束 Schema 时，编辑器拒绝操作并列出引用位置。
- 重命名 Schema 时，编辑器更新全部 Schema、字段和 requirement 引用。
- 删除列或结构体成员时，编辑器删除对应行数据、公式、外键和导出配置。

## 8. CUE import 别名

公共 Schema 的物理目录名不决定其 CUE package。表文件导入公共 Schema 时，必须在 import path 后使用 `:package` 选择目标 package：

```cue
import "example.com/game-data/schema:common"
```

这样表字段可以稳定写作 `common.#GameID`，而不受公共 Schema 所在目录名影响。
