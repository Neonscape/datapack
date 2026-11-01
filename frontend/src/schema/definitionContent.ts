import type {SchemaDefinitionDocument} from "../types";

// Presentation metadata is excluded so title refreshes do not reset drafts.
export function schemaDefinitionContent(definitions: readonly SchemaDefinitionDocument[]): string {
    return JSON.stringify(definitions.map((entry) => [entry.name, entry.kind, entry.type, entry.appliesTo, entry.schema, entry.schemaExpression]));
}
