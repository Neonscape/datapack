import {schemaDefinitionContent} from "../schema/definitionContent";
import type {SchemaDefinition, SchemaEditorState, TableDocument, TreeNode, WorkspaceView} from "../types";

export type OpenTab =
    | {id: string; kind: "table"; path: string; table: TableDocument}
    | {id: string; kind: "schema"; state: SchemaEditorState};

export type SessionState = {workspace: WorkspaceView | null; tabs: OpenTab[]; activeTabId: string | null; selectedSchema: SchemaDefinition | null};

export function refreshSchemaState(state: SchemaEditorState, definitions: SchemaEditorState["definitions"], metadata: Pick<SchemaEditorState, "displayName" | "logicalTable" | "package">): SchemaEditorState {
    const unchanged = schemaDefinitionContent(state.definitions) === schemaDefinitionContent(definitions);
    return {...state, ...metadata, definitions,
        selected: unchanged ? state.selected : definitions.some((entry) => entry.name === state.selected) ? state.selected : definitions[0]?.name ?? null,
        creating: unchanged ? state.creating : false};
}

export function containsPath(node: TreeNode, path: string): boolean {
    return node.path === path || (node.children ?? []).some(child => containsPath(child, path));
}
export function tabPath(tab: OpenTab) { return tab.kind === "table" ? tab.path : tab.state.path; }
export function reconcilePeek(selected: SchemaDefinition | null, view: WorkspaceView | null) {
    return selected ? view?.schemas?.definitions?.find(definition=>definition.id===selected.id) ?? null : null;
}
export function withTabs(state: SessionState, tabs: OpenTab[]): SessionState {
    return {...state, tabs, selectedSchema:reconcilePeek(state.selectedSchema,state.workspace), activeTabId: state.activeTabId && !tabs.some(tab => tab.id === state.activeTabId) ? tabs[0]?.id ?? null : state.activeTabId};
}

