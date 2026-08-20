package character

import "example.com/datapack/skyharbor-content/schema:common"

#PrimaryKey: common.#GameID

_local: {
    #GrowthRate: int & >=1 & <=2
}

_schema: {
    localDefinitions: {
        GrowthRate: {kind: "constraint", appliesTo: ["integer"]}
    }
}

#Row: {
    name: common.#NonEmptyString
    level: int & common.#NonNegative & >=1
    baseAttack: int & common.#NonNegative
    growthRate: int & _local.#GrowthRate
    attack: int & common.#NonNegative
    startingItem: common.#GameID | null
    description: string | null
}

rows: {
    captain_lyra: {
        name: "Captain Lyra"
        level: 4
        baseAttack: 32
        growthRate: 2
        startingItem: "ember_blade"
        description: "A patient leader of the sky patrol."
    }
    scout_ren: {
        name: "Scout Ren"
        level: 2
        baseAttack: 18
        growthRate: 1
        startingItem: "field_kit"
        description: null
    }
}

_tool: {
    version: 1
    table: {namespace: "content", name: "character"}
    foreignKeys: {startingItem: {table: "content.item"}}
    formulas: {
        captain_lyra: {attack: "baseAttack * growthRate"}
        scout_ren: {attack: "baseAttack * growthRate"}
    }
}

_editor: {
    table: {title: "Playable Characters"}
    columns: {name: {title: "Character Name"}, level: {title: "Level"}, baseAttack: {title: "Base Attack"}, growthRate: {title: "Growth Rate"}, attack: {title: "Computed Attack"}, startingItem: {title: "Starting Item"}, description: {title: "Description"}}
    cellComments: {
        captain_lyra: {attack: "Previewed from baseAttack and growthRate."}
    }
}

_export: {
    csv: {
        enabled: true
        columns: ["id", "name", "level", "baseAttack", "growthRate", "attack", "startingItem", "description"]
    }
}
