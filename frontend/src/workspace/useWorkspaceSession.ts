import {useRef, useState} from "react";
import type {SchemaDefinition, SchemaEditorState, TableDocument, WorkspaceView} from "../types";
import {containsPath, tabPath, withTabs, reconcilePeek, type OpenTab, type SessionState} from "./sessionState";
import {readTab, type TabReader} from "./tabRead";
export {refreshSchemaState} from "./sessionState";
export type {OpenTab} from "./sessionState";

type SessionRequest = {project: number; scope: string; serial: number; focus: number; view: number};

export function useWorkspaceSession() {
    const [state, setState] = useState<SessionState>({workspace:null,tabs:[],activeTabId:null,selectedSchema:null});
    const generation = useRef(0);
    const requests = useRef(new Map<string, number>());
    const focus = useRef(0);
    const publication = useRef(0);
    const projectPath = useRef("");
    const epoch = useRef<string>();
    const revision = useRef(-1);
    function accepts(tag: {projectEpoch?:string}, value?:number) {
        return (!tag.projectEpoch || !epoch.current || tag.projectEpoch === epoch.current) && (value === undefined || value >= revision.current);
    }
    const {workspace, tabs, activeTabId, selectedSchema} = state;
    const activeTab = tabs.find(tab => tab.id === activeTabId) ?? null;
    const table = activeTab?.kind === "table" ? activeTab.table : null;
    const schemaDocument = activeTab?.kind === "schema" ? activeTab.state : null;

    function beginRequest(scope: string): SessionRequest {
        const serial = (requests.current.get(scope) ?? 0) + 1;
        requests.current.set(scope, serial);
        return {project:generation.current,scope,serial,view:publication.current,focus:scope.startsWith("open:") ? ++focus.current : focus.current};
    }
    function isCurrent(request: SessionRequest) {
        return request.project === generation.current && request.serial === requests.current.get(request.scope)
            && (request.scope !== "history" || request.view === publication.current);
    }
    function publishWorkspace(view: WorkspaceView, request?: SessionRequest) {
        if (request && !isCurrent(request)) return false;
        if (!accepts(view, view.revision)) return false;
        if (view.revision !== undefined) revision.current = view.revision;
        if (projectPath.current && view.project?.path && projectPath.current !== view.project.path) return false;
        const project = generation.current;
        ++publication.current;
        if (request) request.view = publication.current;
        setState(current => {
            if (project !== generation.current || request && !isCurrent(request)) return current;
            return {...current,workspace:view,selectedSchema:reconcilePeek(current.selectedSchema,view)};
        });
        return true;
    }
    function invalidateDocument(path: string) {
        for (const kind of ["table","schema"]) {
            const scope = `open:${kind}:${path}`;
            requests.current.set(scope,(requests.current.get(scope) ?? 0)+1);
        }
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
    function updateTable(document: TableDocument) {
        if (!accepts(document, document.workspaceRevision)) return;
        invalidateDocument(document.path);
        setState(current => ({...current,tabs:current.tabs.map(tab => tab.kind === "table" && tab.path === document.path ? {...tab,table:document} : tab)}));
    }
    function updateSchema(value: SchemaEditorState) {
        updateSchemaTabById(`schema:${value.path}`,value);
    }
    function updateSchemaTabById(id: string, value: SchemaEditorState) {
        setState(current => ({...current,tabs:current.tabs.map(tab => tab.id === id && tab.kind === "schema" ? {...tab,state:value} : tab)}));
    }
    function updateDisplayName(path: string, definition: string, displayName: string) {
        invalidateDocument(path);
        const definitions = (entries: TableDocument["localSchema"]) => entries.map(entry => entry.name === definition ? {...entry,displayName} : entry);
        setState(current => ({...current,tabs:current.tabs.map(tab => {
            if (tabPath(tab) !== path) return tab;
            if (tab.kind === "table") return {...tab,table:definition ? {...tab.table,localSchema:definitions(tab.table.localSchema)} : {...tab.table,displayName}};
            return {...tab,state:definition ? {...tab.state,definitions:definitions(tab.state.definitions)} : {...tab.state,displayName}};
        })}));
    }
    function applyProjectView(view: WorkspaceView, selected: SchemaDefinition | null) {
        ++generation.current;
        ++publication.current;
        projectPath.current = view.project.path;
        epoch.current = view.projectEpoch;
        revision.current = view.revision ?? -1;
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
        if (!accepts(document, document.workspaceRevision)) return;
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
                const result = results.get(tab.id);
                if (result?.kind === "table" && !accepts(result.table, result.table.workspaceRevision)) return [tab];
                return [result ?? tab];
            });
            return withTabs(current,next);
        });
    }
    async function refreshOpenTabs(request: SessionRequest, view: WorkspaceView, reader: TabReader) {
        const opened = tabs;
        const refreshed: OpenTab[] = [];
        for (const tab of opened) {
            if (!isCurrent(request)) return;
            try {
                refreshed.push(await readTab(tab, reader));
            } catch {
                // A read failure does not prove deletion. applyRefresh closes a
                // tab only when its source is absent from the new project tree.
            }
        }
        applyRefresh(request, opened, refreshed, view);
    }
    function renameTable(path: string, document: TableDocument) {
        invalidateDocument(path);
        if (document.path !== path) invalidateDocument(document.path);
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
        invalidateDocument(path);
        setState(current => withTabs(current,current.tabs.filter(tab => tabPath(tab) !== path)));
    }
    function applyDefinitionResult(path: string, tableDefinitions: SchemaEditorState["definitions"], publicDefinitions: SchemaEditorState["definitions"], selected?: string) {
        invalidateDocument(path);
        setState(current => ({...current,tabs:current.tabs.map(tab => {
            if (tab.kind !== "schema" || tab.state.path !== path) return tab;
            const definitions = tab.state.local ? tableDefinitions : publicDefinitions;
            return {...tab,state:{...tab.state,definitions,selected:selected ?? definitions[0]?.name ?? null,creating:false}};
        })}));
    }
    return {workspace,tabs,activeTabId,selectedSchema,activeTab,table,schemaDocument,publishWorkspace,selectSchema,activateTab,closeTab,updateTable,updateSchema,updateSchemaTabById,updateDisplayName,applyProjectView,addTableTab,addSchemaTab,beginRequest,isCurrent,refreshOpenTabs,renameTable,closeDocument,applyDefinitionResult};
}

export type WorkspaceSession = ReturnType<typeof useWorkspaceSession>;
