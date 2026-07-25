package effect

import "example.com/datapack/basic-rpg-auto/schema:common"

#PrimaryKey: common.#GameID

#Row: {
    name: string & != ""
    damage: int & common.#NonNegative
}

rows: {
    burn_small: {
        name: "Small Burn"
        damage: 10
    }
    freeze_small: {
        name: "Small Freeze"
        damage: 0
    }
}

_tool: {
    version: 1
    table: {
        namespace: "rpg"
        name: "effect"
    }
    foreignKeys: {}
    formulas: {}
}

_editor: {
    table: {title: "Effects"}
    columns: {}
}

_export: {
    csv: {
        enabled: true
        columns: ["id", "name", "damage"]
    }
}
