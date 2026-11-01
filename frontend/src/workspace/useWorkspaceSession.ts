import {schemaDefinitionContent} from "../schema/definitionContent";
import {useRef, useState} from "react";
import type {SchemaDefinition, SchemaEditorState, TableDocument, TreeNode, WorkspaceView} from "../types";

export type OpenTab =
    | {id: string; kind: "table"; path: string; table: TableDocument}
    | {id: string; kind: "schema"; state: SchemaEditorState};

type SessionState = {workspace: WorkspaceView | null; tabs: OpenTab[]; activeTabId: string | null; selectedSchema: SchemaDefinition | null};
type SessionRequest = {project: number; scope: string; serial: number; focus: number};

export function refreshSchemaState(state: SchemaEditorState, definitions: SchemaEditorState["definitions"], metadata: Pick<SchemaEditorState, "displayName" | "logicalTable" | "package">): SchemaEditorState {
    const unchanged = schemaDefinitionContent(state.definitions) === schemaDefinitionContent(definitions);
    return {...state, ...metadata, definitions,
        selected: unchanged ? state.selected : definitions.some((entry) => entry.name === state.selected) ? state.selected : definitions[0]?.name ?? null,
        creating: unchanged ? state.creating : false};
}

function containsPath(node: TreeNode, path: string): boolean {
    return node.path === path || (node.children ?? []).some(child => containsPath(child, path));
}
function tabPath(tab: OpenTab) { return tab.kind === "table" ? tab.path : tab.state.path; }
function withTabs(state: SessionState, tabs: OpenTab[]): SessionState {
    return {...state, tabs, activeTabId: state.activeTabId && !tabs.some(tab => tab.id === state.activeTabId) ? tabs[0]?.id ?? null : state.activeTabId};
}

export function useWorkspaceSession() {
    const [state, setState] = useState<SessionState>({workspace:null,tabs:[],activeTabId:null,selectedSchema:null});
    const generation = useRef(0);
    const requests = useRef(new Map<string, number>());
    const focus = useRef(0);
    const {workspace, tabs, activeTabId, selectedSchema} = state;
    const activeTab = tabs.find(tab => tab.id === activeTabId) ?? null;
    const table = activeTab?.kind === "table" ? activeTab.table : null;
    const schemaDocument = activeTab?.kind === "schema" ? activeTab.state : null;

    function beginRequest(scope: string): SessionRequest {
        const serial = (requests.current.get(scope) ?? 0) + 1;
        requests.current.set(scope, serial);
        return {project:generation.current,scope,serial,focus:scope.startsWith("open:") ? ++focus.current : focus.current};
    }
    function isCurrent(request: SessionRequest) {
        return request.project === generation.current && request.serial === requests.current.get(request.scope);
    }
    function publishWorkspace(view: WorkspaceView) {
        setState(current => current.workspace?.project?.path && view.project?.path && current.workspace.project.path !== view.project.path ? current : {...current,workspace:view});
    }
    function selectSchema(value: SchemaDefinition | null) {
        ++focus.current;
        setState(current => ({...current,selectedSchema:value}));
    }
    function activateTab(id: string) {
        ++focus.current;
        setState(current => current.tabs.some(tab => tab.id === id) ? {...current,activeTabId:id,selectedSchema:null} : current);
    }
    function closeTab(id: string) {
        requests.current.set(`open:${id}`, (requests.current.get(`open:${id}`) ?? 0) + 1);
        setState(current => {
            const index = current.tabs.findIndex(tab => tab.id === id);
            if (index < 0) return current;
            const next = current.tabs.filter(tab => tab.id !== id);
            return {...current,tabs:next,activeTabId:current.activeTabId === id ? next[Math.min(index,next.length-1)]?.id ?? null : current.activeTabId};
        });
    }
    function updateActiveTable(document: TableDocument) {
        setState(current => ({...current,tabs:current.tabs.map(tab => tab.kind === "table" && tab.path === document.path ? {...tab,table:document} : tab)}));
    }
    function updateActiveSchema(value: SchemaEditorState) {
        updateSchemaTabById(`schema:${value.path}`,value);
    }
    function updateSchemaTabById(id: string, value: SchemaEditorState) {
        setState(current => ({...current,tabs:current.tabs.map(tab => tab.id === id && tab.kind === "schema" ? {...tab,state:value} : tab)}));
    }
    function updateDisplayName(path: string, definition: string, displayName: string) {
        const definitions = (entries: TableDocument["localSchema"]) => entries.map(entry => entry.name === definition ? {...entry,displayName} : entry);
        setState(current => ({...current,tabs:current.tabs.map(tab => {
            if (tabPath(tab) !== path) return tab;
            if (tab.kind === "table") return {...tab,table:definition ? {...tab.table,localSchema:definitions(tab.table.localSchema)} : {...tab.table,displayName}};
            return {...tab,state:definition ? {...tab.state,definitions:definitions(tab.state.definitions)} : {...tab.state,displayName}};
        })}));
    }
    function applyProjectView(view: WorkspaceView, selected: SchemaDefinition | null) {
        ++generation.current;
        requests.current.clear();
        ++focus.current;
        setState({workspace:view,tabs:[],activeTabId:null,selectedSchema:selected});
    }
    function addTab(tab: OpenTab, request?: SessionRequest) {
        if (request && !isCurrent(request)) return;
        const activate = !request || request.focus === focus.current;
        if (!request) ++focus.current;
        setState(current => {
            if (request && !isCurrent(request)) return current;
            const next = current.tabs.some(entry => entry.id === tab.id) ? current.tabs : [...current.tabs,tab];
            return {...current,tabs:next,...activate ? {activeTabId:tab.id,selectedSchema:null} : {}};
        });
    }
    function addTableTab(path: string, document: TableDocument, request?: SessionRequest) {
        addTab({id:`table:${path}`,kind:"table",path,table:document},request);
    }
    function addSchemaTab(path: string, value: SchemaEditorState, request?: SessionRequest) {
        addTab({id:`schema:${path}`,kind:"schema",state:value},request);
    }
    function applyRefresh(request: SessionRequest, opened: OpenTab[], refreshed: OpenTab[], view: WorkspaceView) {
        if (!isCurrent(request)) return;
        const originals = new Map(opened.map(tab => [tab.id,tab]));
        const results = new Map(refreshed.map(tab => [tab.id,tab]));
        setState(current => {
            if (!isCurrent(request)) return current;
            const next = current.tabs.flatMap(tab => {
                // A close/reopen, edit, selection, rename or new tab after the
                // read started owns its state; a stale snapshot cannot replace it.
                if (originals.get(tab.id) !== tab) return [tab];
                if (!containsPath(view.tree.root,tabPath(tab))) return [];
                return [results.get(tab.id) ?? tab];
            });
            return withTabs(current,next);
        });
    }
    function renameTable(path: string, document: TableDocument) {
        setState(current => {
            const next = current.tabs.map(tab => {
                if (tabPath(tab) !== path) return tab;
                if (tab.kind === "table") return {id:`table:${document.path}`,kind:"table" as const,path:document.path,table:document};
                return {...tab,id:`schema:${document.path}`,state:{...tab.state,path:document.path,displayName:document.displayName,logicalTable:document.namespace ? `${document.namespace}.${document.name}` : document.name}};
            });
            const active = current.tabs.findIndex(tab => tab.id === current.activeTabId);
            return {...current,tabs:next,activeTabId:active < 0 ? current.activeTabId : next[active].id};
        });
    }
    function closeDocument(path: string) {
        setState(current => withTabs(current,current.tabs.filter(tab => tabPath(tab) !== path)));
    }
    function applyDefinitionResult(path: string, tableDefinitions: SchemaEditorState["definitions"], publicDefinitions: SchemaEditorState["definitions"], selected?: string) {
        setState(current => ({...current,tabs:current.tabs.map(tab => {
            if (tab.kind !== "schema" || tab.state.path !== path) return tab;
            const definitions = tab.state.local ? tableDefinitions : publicDefinitions;
            return {...tab,state:{...tab.state,definitions,selected:selected ?? definitions[0]?.name ?? null,creating:false}};
        })}));
    }
    return {workspace,tabs,activeTabId,selectedSchema,activeTab,table,schemaDocument,publishWorkspace,selectSchema,activateTab,closeTab,updateActiveTable,updateActiveSchema,updateSchemaTabById,updateDisplayName,applyProjectView,addTableTab,addSchemaTab,beginRequest,isCurrent,applyRefresh,renameTable,closeDocument,applyDefinitionResult};
}

export type WorkspaceSession = ReturnType<typeof useWorkspaceSession>;
