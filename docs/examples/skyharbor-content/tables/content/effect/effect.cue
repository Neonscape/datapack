package effect

import "example.com/datapack/skyharbor-content/schema:common"

#PrimaryKey: common.#GameID

#Row: {
    name: common.#NonEmptyString
    kind: common.#EffectKind
    power: int & common.#NonNegative
    chance: common.#Probability
    duration: int & common.#NonNegative
    description: string | null
    tags: [...string]
}

rows: {
    burn: {
        name: "Burn"
        kind: "damage"
        power: 18
        chance: 0.75
        duration: 3
        description: "Deals damage over time."
        tags: ["fire", "damage"]
    }
    mend: {
        name: "Mend"
        kind: "heal"
        power: 25
        chance: 1
        duration: 1
        description: null
        tags: ["support"]
    }
    snare: {
        name: "Snare"
        kind: "control"
        power: 0
        chance: 0.6
        duration: 2
        description: "Slows the target."
        tags: ["control"]
    }
}

_tool: {
    version: 1
    table: {namespace: "content", name: "effect"}
    foreignKeys: {}
    formulas: {}
}

_editor: {
    table: {title: "Combat Effects"}
    columns: {}
}

_export: {
    csv: {
        enabled: true
        columns: ["id", "name", "kind", "power", "chance", "duration", "description", "tags"]
        encoding: {tags: "json"}
    }
}
