package common

#GameID: string & =~"^[a-z][a-z0-9_]*$"
#NonEmptyString: string & != ""
#NonNegative: >=0
#Probability: number & >=0 & <=1
#Rarity: "N" | "R" | "SR" | "SSR"
#EffectKind: "damage" | "heal" | "control" | "utility"

_schema: {
    version: 1
    definitions: {
        GameID: {
            kind: "type"
            type: "string"
        }
        NonEmptyString: {
            kind: "type"
            type: "string"
        }
        NonNegative: {
            kind: "constraint"
            appliesTo: ["integer", "number"]
        }
        Probability: {
            kind: "type"
            type: "number"
        }
        Rarity: {
            kind: "type"
            type: "enum"
        }
        EffectKind: {
            kind: "type"
            type: "enum"
        }
    }
}
