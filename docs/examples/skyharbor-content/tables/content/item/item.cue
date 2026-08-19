package item

import "example.com/datapack/skyharbor-content/schema:common"

#PrimaryKey: common.#GameID

_local: {
    #Position: {
        x: number
        y: number
        z: number
    }
    #EffectID: common.#GameID
}

_schema: {
    localDefinitions: {
        Position: {kind: "type", type: "struct"}
        EffectID: {kind: "type", type: "string"}
    }
}

#Row: {
    name: common.#NonEmptyString
    rarity: common.#Rarity
    price: int & common.#NonNegative & <10000
    effectID: _local.#EffectID
    effectIDs: [..._local.#EffectID]
    position: _local.#Position
    description: string | null
    sellPrice: int & common.#NonNegative
}

rows: {
    ember_blade: {
        name: "Ember Blade"
        rarity: "R"
        price: 240
        effectID: "burn"
        effectIDs: ["burn"]
        position: {x: 1, y: 0, z: 0}
        description: "A warm blade from the eastern forge."
    }
    field_kit: {
        name: "Field Kit"
        rarity: "N"
        price: 80
        effectID: "mend"
        effectIDs: ["mend", "snare"]
        position: {x: 0, y: 1, z: 0}
        description: "Useful on long expeditions."
    }
}

_tool: {
    version: 1
    table: {namespace: "content", name: "item"}
    foreignKeys: {
        effectID: {table: "content.effect"}
        effectIDs: {table: "content.effect"}
    }
    formulas: {
        ember_blade: {sellPrice: "price * 3 / 4"}
        field_kit: {sellPrice: "price / 2"}
    }
}

_editor: {
    table: {title: "Items"}
    columns: {}
}

_export: {
    csv: {
        enabled: true
        columns: ["id", "name", "rarity", "price", "effectID", "effectIDs", "position", "description", "sellPrice"]
        encoding: {effectIDs: "json", position: "json"}
    }
}
