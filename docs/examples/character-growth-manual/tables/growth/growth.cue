package growth

import "example.com/datapack/character-growth-manual/schema:common"

#PrimaryKey: common.#LevelID

#Row: {
    baseAttack: int & common.#NonNegative
    attack: int & common.#NonNegative
}

rows: {
    level_01: {baseAttack: 10}
    level_10: {baseAttack: 20}
    level_11: {baseAttack: 25}
}

_tool: {
    version: 1
    table: {
        namespace: "combat"
        name: "growth"
    }
    foreignKeys: {}
    formulas: {
        level_01: {
            attack: "baseAttack * 2"
        }
        level_10: {
            attack: "combat.growth.level_01.attack + baseAttack"
        }
        level_11: {
            attack: "combat.growth.level_10.attack + 5"
        }
    }
}

_editor: {
    table: {title: "Character Growth"}
    columns: {}
}

_export: {
    csv: {
        enabled: true
        columns: ["id", "baseAttack", "attack"]
    }
}
