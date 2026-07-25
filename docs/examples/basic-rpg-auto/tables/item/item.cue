package item

import "example.com/datapack/basic-rpg-auto/schema:common"

#PrimaryKey: common.#GameID

_local: {
    #Position: {
        x: number
        y: number
        z: number
    }
}

_schema: {
    localDefinitions: {
        Position: {
            kind: "type"
            type: "struct"
        }
    }
}

#Row: {
    name: string & != ""
    price: int & common.#NonNegative & <9999
    rarity: common.#Rarity
    effectIds: [...common.#GameID]
    position: _local.#Position
}

rows: {
    sword_001: {
        name: "Iron Sword"
        price: 100
        rarity: "N"
        effectIds: ["burn_small"]
        position: {x: 0, y: 1, z: 0}
    }
}

_tool: {
    version: 1
    table: {
        namespace: "rpg"
        name: "item"
    }
    foreignKeys: {
        effectIds: {table: "rpg.effect"}
    }
    formulas: {}
}

_editor: {
    table: {title: "Items"}
    columns: {}
}

_export: {
    csv: {
        enabled: true
        columns: ["id", "name", "price", "rarity", "effectIds", "position"]
        encoding: {
            effectIds: "json"
            position: "json"
        }
    }
}
