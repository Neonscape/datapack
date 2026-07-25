package rules

#Price: int & #NonNegative & <9999
#NonNegative: >=0
#AtLeastOne: >=1
#NonEmpty: != ""

_schema: {
    version: 1
    definitions: {
        Price: {
            kind: "type"
            type: "integer"
        }
        NonNegative: {
            kind: "constraint"
            appliesTo: ["integer", "number"]
        }
        AtLeastOne: {
            kind: "constraint"
            appliesTo: ["integer", "number"]
        }
        NonEmpty: {
            kind: "constraint"
            appliesTo: ["string"]
        }
    }
}
