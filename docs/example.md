
# 完整数据包实例 v0.2

这是一个小型 RPG 道具数据包。它覆盖工程入口、逻辑命名空间、基础表结构、全局规则、数组外键、扁平数组、扁平结构体、可选字段、编辑器信息、单元格公式和 CSV 导出。

数组只能放原始值；结构体只能有一层，成员也只能是原始值。数组中不能放结构体，结构体中也不能继续嵌套数组或结构体。

## 1. 目录

```text
game-data/
    game-data.dtproj
    cue.mod/
        module.cue
    schema/
        common/
            common.cue
    tables/
        effect/
            effect.cue
        item/
            item.cue
    exports/
        effect.csv
        item.csv
```

`schema/` 和 `tables/` 是源数据。`exports/` 是自动生成结果，不能手改，也不应作为源数据提交或维护。

`game-data.dtproj` 是 JSON 工程入口。它不逐一登记表，而是声明表目录入口、Schema 入口、排除规则、工程格式版本和导出配置，具体字段见 `dtproj-format-v1.md`。编译模式属于用户级编辑器设置，见 `editor-settings-format-v1.md`。

本例中的两张表分别使用完整逻辑名称 `rpg.effect` 和 `rpg.item`。逻辑名称独立于文件路径；引用保存逻辑名称，而不是相对路径。

## 2. 模块信息

文件：`cue.mod/module.cue`

```cue
module: "game-data"
language: {
    version: "v0.9.0"
}
```

## 3. 可复用规则

文件：`schema/common/common.cue`

```cue
package common

#GameID: string & =~"^[a-z][a-z0-9_]*$"

#NonNegativeInt: int & >=0
#Percent:        int & >=0 & <=100

#Rarity: "N" | "R" | "SR" | "SSR"
#Element: "none" | "fire" | "ice" | "lightning"
```

这里不保存业务行数据，只放需要被多张表复用的规则。

## 4. 效果表

文件：`tables/effect/effect.cue`

```cue
package effect

import "game-data/schema/common"

#PrimaryKey: common.#GameID

_local: {
    #DurationSeconds: int & >=0 & <=3600
}

#Row: {
    name:            string & != ""
    element:         common.#Element
    damage:          common.#NonNegativeInt
    durationSeconds: _local.#DurationSeconds
    isDebuff:        bool
    description?:    string
}

rows: {
    [ID=#PrimaryKey]: #Row

    burn_small: {
        name:            "Small Burn"
        element:         "fire"
        damage:          10
        durationSeconds: 5
        isDebuff:        true
        description:     "Deals fire damage over time."
    }

    freeze_small: {
        name:            "Small Freeze"
        element:         "ice"
        damage:          0
        durationSeconds: 2
        isDebuff:        true
    }
}

_tool: {
    version: 1
    table: {
        namespace: "rpg"
        name:      "effect"
    }
    foreignKeys: {}
    formulas: {}
}

_editor: {
    table: {title: "效果表"}
    columns: {
        id: {
            title: "ID"
            kind: "primaryKey"
            readonly: true
            order: 0
        }
        name: {
            title: "名称"
            comment: "效果的显示名称。"
            order: 1
        }
        element: {
            title: "元素"
            editor: "select"
            order: 2
        }
        damage: {
            title: "伤害"
            editor: "number"
            order: 3
        }
        durationSeconds: {
            title: "持续秒数"
            editor: "number"
            order: 4
        }
        isDebuff: {
            title: "是否减益"
            editor: "checkbox"
            order: 5
        }
        description: {
            title: "说明"
            comment: "可为空。"
            order: 6
        }
    }
}

_export: {
    csv: {
        enabled: true
        columns: ["id", "name", "element", "damage", "durationSeconds", "isDebuff", "description"]
    }
}
```

## 5. 道具表

文件：`tables/item/item.cue`

```cue
package item

import "game-data/schema/common"

#PrimaryKey: common.#GameID

_local: {
    #Tag: string & =~"^[a-z][a-z0-9_]*$"
    #EffectID: common.#GameID

    #Position: {
        x: number
        y: number
        z: number
    }
    #Stats: {
        attack:  common.#NonNegativeInt
        defense: common.#NonNegativeInt
        speed:   number & >=0
    }
    #Sale: {
        discountAmount: common.#NonNegativeInt
        taxPercent:     common.#Percent
    }
}

#Row: {
    name:         string & != ""
    description?: string
    price:        common.#NonNegativeInt
    maxStack:     int & >=1 & <=999
    rarity:       common.#Rarity
    element:      common.#Element
    isSellable:   bool

    tags: [..._local.#Tag]

    effectIds: [..._local.#EffectID]

    position: _local.#Position
    stats:    _local.#Stats
    sale:     _local.#Sale
    sellPrice:  int & >=0
    displayName: string
}

rows: {
    [ID=#PrimaryKey]: #Row

    sword_001: {
        name: "Iron Sword"
        description: "A basic sword for new adventurers."
        price: 100
        maxStack: 1
        rarity: "N"
        element: "none"
        isSellable: true
        tags: ["weapon", "starter"]
        effectIds: ["burn_small"]
        position: {
            x: 0
            y: 1
            z: 0
        }
        stats: {
            attack: 12
            defense: 0
            speed: 0
        }
        sale: {
            discountAmount: 10
            taxPercent: 5
        }
    }

    bow_001: {
        name: "Hunter Bow"
        price: 300
        maxStack: 1
        rarity: "R"
        element: "ice"
        isSellable: true
        tags: ["weapon", "ranged"]
        effectIds: ["freeze_small"]
        position: {
            x: 2.5
            y: 1
            z: 0
        }
        stats: {
            attack: 24
            defense: 0
            speed: 0.15
        }
        sale: {
            discountAmount: 0
            taxPercent: 5
        }
    }
}

_tool: {
    version: 1
    table: {
        namespace: "rpg"
        name:      "item"
    }
    foreignKeys: {
        effectIds: {
            table: "rpg.effect"
        }
    }
    formulas: {
        sword_001: {
            sellPrice:  "price - sale.discountAmount"
            displayName: "name + \" [\" + rarity + \"]\""
        }
        bow_001: {
            sellPrice:  "price - sale.discountAmount"
            displayName: "name + \" [\" + rarity + \"]\""
        }
    }
}

_editor: {
    table: {title: "道具表"}
    columns: {
        id: {
            title: "ID"
            kind: "primaryKey"
            readonly: true
            order: 0
        }
        name: {
            title: "名称"
            comment: "不能为空。"
            order: 1
        }
        description: {
            title: "说明"
            comment: "可选字段。"
            order: 2
        }
        price: {
            title: "原价"
            editor: "number"
            order: 3
        }
        maxStack: {
            title: "最大堆叠数"
            editor: "number"
            order: 4
        }
        rarity: {
            title: "稀有度"
            editor: "select"
            order: 5
        }
        element: {
            title: "元素"
            editor: "select"
            order: 6
        }
        isSellable: {
            title: "可出售"
            editor: "checkbox"
            order: 7
        }
        tags: {
            title: "标签"
            kind: "list"
            comment: "一维文本列表；可逐项增删和排序。"
            editor: "list"
            order: 8
        }
        effectIds: {
            title: "附加效果"
            kind: "list"
            comment: "一维效果 ID 列表；每一项均引用效果表。"
            editor: "refList"
            order: 9
        }
        position: {
            title: "摆放坐标"
            kind: "struct"
            comment: "一层结构体，编辑时展开 x/y/z。"
            editor: "struct"
            order: 10
        }
        stats: {
            title: "属性加成"
            kind: "struct"
            comment: "一层结构体，包含 attack/defense/speed。"
            editor: "struct"
            order: 11
        }
        sale: {
            title: "售价规则"
            kind: "struct"
            comment: "一层结构体，包含折扣金额和税率。"
            editor: "struct"
            order: 12
        }
        sellPrice: {
            title: "折后价格"
            comment: "本例中由每个单元格自己的公式计算。"
            order: 13
        }
        displayName: {
            title: "导出显示名"
            order: 14
        }
    }
}

_export: {
    csv: {
        enabled: true
        columns: [
            "id", "name", "description", "price", "maxStack", "rarity", "element",
            "isSellable", "tags", "effectIds", "position", "stats", "sale", "sellPrice",
            "displayName",
        ]
        encoding: {
            tags:      "json"
            effectIds: "json"
            position:  "json"
            stats:     "json"
            sale:      "json"
        }
    }
}
```

## 6. 外键保护和错误定位

`effectIds` 是数组外键。它的目标表是 `rpg.effect`，数组基数由 `#Row` 中的 `[..._local.#EffectID]` 推导，不在 `_tool.foreignKeys` 中重复声明。系统需要记录到具体数组项：

```text
rpg.item.sword_001.effectIds[0] -> rpg.effect.burn_small
rpg.item.bow_001.effectIds[0]   -> rpg.effect.freeze_small
```

- 重命名 `burn_small` 时，应列出 `rpg.item.sword_001.effectIds[0]`，确认后同步更新。
- 删除 `freeze_small` 时，应阻止删除，并指出 `rpg.item.bow_001.effectIds[0]` 仍在引用它。
- 如果填写 `["not_exist"]`，应报错：`bow_001.effectIds[0] 引用了不存在的 effect：not_exist`。
- 结构体内成员出错时，也要精确定位，例如：`sword_001.stats.attack` 或 `sword_001.sale.taxPercent`。

外键只保证 `effectIds` 中的 ID 存在于效果表，不会把效果表的其他字段复制到道具表，也不提供跨表取值。第一阶段公式不能读取 `effectIds`，也不能通过它访问效果表中的 `damage` 等字段。

## 7. 编辑方式

- 原始值列保持一格一个值。
- `tags`、`effectIds` 显示为简短摘要，例如 `2 项` 或 `[weapon, starter]`；点击后打开列表编辑器。
- `position`、`stats`、`sale` 显示为简短摘要，例如 `x=0, y=1, z=0`；点击后展开成员输入项。
- 不把编辑 JSON 文本作为主要操作方式，避免策划反复处理引号、逗号和类型问题。

`sellPrice` 和 `displayName` 是 `#Row` 中的普通值字段，但本例中的对应单元格不在 `rows` 保存固定值，而是在 `_tool.formulas` 保存公式。预览和导出时，工具链计算公式并在内存中补全这两个字段，再使用 `#Row` 校验最终数据。

因此，计算前的原始 `rows` 对完整 `#Row` 来说可能暂时缺少字段，这是预期状态。项目工具先从 AST 读取源数据和公式，再构造完整数据并调用 CUE Go API 校验；原生 `cue vet` 不能代替这一完整流程。

一个公式也可以引用本表其他行，例如：

```text
rpg.item.sword_001.sellPrice * 2
```

公式可以引用其他公式单元格。工具链会先建立依赖图，按顺序求值；发现循环引用时直接报错。公式结果不写回 `rows`。

## 8. item.csv 的预期结果

数组和扁平结构体默认写成紧凑 JSON，放进一个 CSV 单元格。CSV 会按规则转义其中的双引号。

```csv
id,name,description,price,maxStack,rarity,element,isSellable,tags,effectIds,position,stats,sale,sellPrice,displayName
sword_001,Iron Sword,A basic sword for new adventurers.,100,1,N,none,true,"[""weapon"",""starter""]","[""burn_small""]","{""x"":0,""y"":1,""z"":0}","{""attack"":12,""defense"":0,""speed"":0}","{""discountAmount"":10,""taxPercent"":5}",90,Iron Sword [N]
bow_001,Hunter Bow,,300,1,R,ice,true,"[""weapon"",""ranged""]","[""freeze_small""]","{""x"":2.5,""y"":1,""z"":0}","{""attack"":24,""defense"":0,""speed"":0.15}","{""discountAmount"":0,""taxPercent"":5}",300,Hunter Bow [R]
```

后续如果运行时不想解析 JSON 字符串，可以为一层结构体加“展开导出”选项，例如将 `position` 导出为 `position_x`、`position_y`、`position_z`。这只影响导出结果，不改变源表中“position 是一列”的含义。

## 9. 覆盖范围

| 能力 | 示例位置 |
|---|---|
| 一表一文件 | `tables/effect/effect.cue`、`tables/item/item.cue` |
| 工程入口和自动发现 | `game-data.dtproj` |
| 逻辑命名空间 | `_tool.table` 中的 `rpg.effect`、`rpg.item` |
| 主键和命名规则 | `#PrimaryKey`、`common.#GameID` |
| 本表规则与全局规则 | `_local`、`schema/common/common.cue` |
| 文本、整数、小数、布尔、枚举、可选字段 | 两张表的普通字段 |
| 扁平数组 | `rpg.item.tags` |
| 数组外键 | `rpg.item.effectIds` |
| 扁平结构体 | `position`、`stats`、`sale` |
| 外键改名、删除保护 | 第 6 节 |
| 工具链配置和数组外键 | `_tool.foreignKeys` |
| 列标题、备注、输入控件、顺序 | `_editor.columns` |
| CSV 配置 | `_export.csv` |
| 单元格公式 | `_tool.formulas` 中的 `sellPrice`、`displayName` |
