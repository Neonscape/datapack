package weapon

import "example.com/datapack/schema-composition/schema:rules"

#PrimaryKey: string & != ""

_local: {
    #WeaponCount: >=1
}

_schema: {
    localDefinitions: {
        WeaponCount: {
            kind: "constraint"
            appliesTo: ["integer", "number"]
        }
    }
}

#Row: {
    name: string & rules.#NonEmpty
    price: rules.#Price
    stackLimit: int & _local.#WeaponCount
}

rows: {
    sword: {name: "Sword", price: 100, stackLimit: 1}
    potion: {name: "Potion", price: 10, stackLimit: 99}
}

_tool: {
    version: 1
    table: {
        namespace: "inventory"
        name: "weapon"
    }
    foreignKeys: {}
    formulas: {}
}

_editor: {
    table: {title: "Weapons"}
    columns: {}
}

_export: {
    csv: {
        enabled: true
        columns: ["id", "name", "price", "stackLimit"]
    }
}
