import {schemaDefinitionContent} from "../schema/definitionContent";
import {useState} from "react";
import type {PublicSchemaDocument, SchemaDefinition, SchemaEditorState, TableDocument, WorkspaceView} from "../types";

export type OpenTab =
    | {id: string; kind: "table"; path: string; table: TableDocument}
    | {id: string; kind: "schema"; state: SchemaEditorState};

// A metadata-only refresh keeps selection and creation state, as well as the
// editor's local draft. Semantic changes still reconcile a removed selection.
export function refreshSchemaState(state: SchemaEditorState, definitions: SchemaEditorState["definitions"], metadata: Pick<SchemaEditorState, "displayName" | "logicalTable" | "package">): SchemaEditorState {
    const unchanged = schemaDefinitionContent(state.definitions) === schemaDefinitionContent(definitions);
    return {...state, ...metadata, definitions,
        selected: unchanged ? state.selected : definitions.some((entry) => entry.name === state.selected) ? state.selected : definitions[0]?.name ?? null,
        creating: unchanged ? state.creating : false};
}

export function useWorkspaceSession() {
    const [workspace, setWorkspace] = useState<WorkspaceView | null>(null);
    const [tabs, setTabs] = useState<OpenTab[]>([]);
    const [activeTabId, setActiveTabId] = useState<string | null>(null);
    const [selectedSchema, setSelectedSchema] = useState<SchemaDefinition | null>(null);
    const activeTab = tabs.find((tab) => tab.id === activeTabId) ?? null;
    const table = activeTab?.kind === "table" ? activeTab.table : null;
    const schemaDocument = activeTab?.kind === "schema" ? activeTab.state : null;

    function activateTab(id: string) {
        setActiveTabId(id);
        setSelectedSchema(null);
    }

    function closeTab(id: string) {
        setTabs((current) => {
            const index = current.findIndex((tab) => tab.id === id);
            if (index < 0) return current;
            const next = current.filter((tab) => tab.id !== id);
            if (id === activeTabId) setActiveTabId(next[Math.min(index, next.length - 1)]?.id ?? null);
            return next;
        });
    }

    function updateActiveTable(document: TableDocument) {
        setTabs((current) => current.map((tab) => tab.id === activeTabId && tab.kind === "table" ? {...tab, table: document} : tab));
    }

    function updateActiveSchema(state: SchemaEditorState) {
        setTabs((current) => current.map((tab) => tab.id === activeTabId && tab.kind === "schema" ? {...tab, state} : tab));
    }

    function updateSchemaTabById(id: string, state: SchemaEditorState) {
        setTabs((current) => current.map((tab) => tab.id === id && tab.kind === "schema" ? {...tab, state} : tab));
    }

    function updateDisplayName(path: string, definition: string, displayName: string) {
        const definitions = (entries: TableDocument["localSchema"]) => entries.map((entry) => entry.name === definition ? {...entry, displayName} : entry);
        setTabs((current) => current.map((tab) => {
            if (tab.kind === "table") {
                if (tab.path !== path) return tab;
                return {...tab, table: definition ? {...tab.table, localSchema: definitions(tab.table.localSchema)} : {...tab.table, displayName}};
            }
            if (tab.state.path !== path) return tab;
            return {...tab, state: definition ? {...tab.state, definitions: definitions(tab.state.definitions)} : {...tab.state, displayName}};
        }));
    }

    function applyProjectView(view: WorkspaceView) {
        setWorkspace(view);
        setTabs([]);
        setActiveTabId(null);
        setSelectedSchema(null);
    }

    function addTableTab(path: string, document: TableDocument) {
        const id = `table:${path}`;
        setTabs((current) => current.some((tab) => tab.id === id) ? current : [...current, {id, kind: "table", path, table: document}]);
        activateTab(id);
    }

    function addSchemaTab(path: string, state: SchemaEditorState) {
        const id = `schema:${path}`;
        setTabs((current) => current.some((tab) => tab.id === id) ? current : [...current, {id, kind: "schema", state}]);
        activateTab(id);
    }

    return {
        workspace, setWorkspace, tabs, setTabs, activeTabId, setActiveTabId, selectedSchema, setSelectedSchema,
        activeTab, table, schemaDocument, activateTab, closeTab, updateActiveTable, updateActiveSchema,
        updateSchemaTabById, updateDisplayName, applyProjectView, addTableTab, addSchemaTab,
    };
}

export type WorkspaceSession = ReturnType<typeof useWorkspaceSession>;
