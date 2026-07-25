package common

#GameID: string & =~"^[a-z][a-z0-9_]*$"
#Rarity: "N" | "R" | "SR" | "SSR"
#NonNegative: >=0

_schema: {
    version: 1
    definitions: {
        GameID: {
            kind: "type"
            type: "string"
        }
        Rarity: {
            kind: "type"
            type: "enum"
        }
        NonNegative: {
            kind: "constraint"
            appliesTo: ["integer", "number"]
        }
    }
}
