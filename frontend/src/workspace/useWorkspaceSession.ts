import {useState} from "react";
import type {PublicSchemaDocument, SchemaDefinition, SchemaEditorState, TableDocument, WorkspaceView} from "../types";

export type OpenTab =
    | {id: string; kind: "table"; path: string; table: TableDocument}
    | {id: string; kind: "schema"; state: SchemaEditorState};

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
        updateSchemaTabById, applyProjectView, addTableTab, addSchemaTab,
    };
}

export type WorkspaceSession = ReturnType<typeof useWorkspaceSession>;
