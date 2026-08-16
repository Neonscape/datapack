import type {SchemaDefinition, WorkspaceView} from "../types";

export function schemaDefinitions(workspace: WorkspaceView | null): SchemaDefinition[] {
    return workspace?.schemas?.definitions ?? [];
}

export function schemaTypeReferenceOptions(workspace: WorkspaceView | null): {label: string; value: string}[] {
    return schemaDefinitions(workspace).filter((definition) => definition.kind === "type").map((definition) => ({
        label: definition.id,
        value: definition.id.includes("._local.#") ? `_local.#${definition.id.split(".#")[1]}` : definition.id,
    }));
}

export function schemaConstraintReferenceOptions(workspace: WorkspaceView | null): {label: string; value: string}[] {
    return schemaDefinitions(workspace).filter((definition) => definition.kind === "constraint").map((definition) => ({
        label: definition.id,
        value: definition.id.includes("._local.#") ? `_local.#${definition.id.split(".#")[1]}` : definition.id,
    }));
}
