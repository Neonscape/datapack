package common

#LevelID: string & =~"^level_[0-9]{2}$"
#NonNegative: >=0

_schema: {
    version: 1
    definitions: {
        LevelID: {
            kind: "type"
            type: "string"
        }
        NonNegative: {
            kind: "constraint"
            appliesTo: ["integer", "number"]
        }
    }
}
