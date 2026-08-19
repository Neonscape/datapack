package growth

import "example.com/datapack/skyharbor-content/schema:common"

#PrimaryKey: common.#GameID

_local: {
    #Level: int & >=1 & <=60
}

_schema: {
    localDefinitions: {
        Level: {kind: "constraint", appliesTo: ["integer"]}
    }
}

#Row: {
    level: int & _local.#Level
    baseValue: int & common.#NonNegative
    multiplier: int & >=1
    value: int & common.#NonNegative
    nextValue: int & common.#NonNegative
}

rows: {
    level_01: {level: 1, baseValue: 10, multiplier: 2}
    level_02: {level: 2, baseValue: 14, multiplier: 2}
    level_03: {level: 3, baseValue: 19, multiplier: 2}
}

_tool: {
    version: 1
    table: {namespace: "balance", name: "growth"}
    foreignKeys: {}
    formulas: {
        level_01: {
            value: "baseValue"
            nextValue: "value * multiplier"
        }
        level_02: {
            value: "balance.growth.level_01.value * multiplier"
            nextValue: "balance.growth.level_01.value"
        }
        level_03: {
            value: "balance.growth.level_02.value * multiplier"
            nextValue: "balance.growth.level_02.value"
        }
    }
}

_editor: {
    table: {title: "Level Growth"}
    columns: {}
}

_export: {
    csv: {enabled: true, columns: ["id", "level", "baseValue", "multiplier", "value", "nextValue"]}
}
