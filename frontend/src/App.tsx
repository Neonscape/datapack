import {useEffect, useState} from 'react';
import './App.css';
import {AddColumn, ClearCell, Compile, CreateRow, CreateSchemaDefinition, CreateSchemaFile, CreateTable, DeleteColumn, DeleteRow, DeleteSchemaDefinition, DeleteTable, Export, GetExportFormats, GetPublicSchema, GetSettings, GetTable, MoveColumn, Navigate, NewProject, OpenPath, Paste, PrintSchema, Redo, RenameColumn, RenameRow, RenameSchemaDefinition, RenameTable, SaveProjectAs, SelectProject, SetCellComment, SetCellFormula, SetCellText, SetColumnDisplayName, SetColumnSchema, SetColumnWidth, SetCompileMode, SetLanguage, SetSchemaDefinition, SetTableExportSettings, SetTheme, Undo, UpdateCell} from "../wailsjs/go/main/App";
import {schemamodel, workspace as workspaceModel} from "../wailsjs/go/models";
import {useTranslation, type Language} from "./i18n";
import {useTheme, type ThemePreference} from "./theme";
import {LogPanel} from "./LogPanel";
import {schemaDefinitions} from "./schema/refs";
import {SchemaEditor} from "./schema/SchemaEditor";
import {SchemaForm} from "./schema/SchemaForm";
import {ProjectSettings} from "./project/ProjectSettings";
import {TableGrid} from "./table/TableGrid";
import {Tree} from "./tree/Tree";
import {useWorkspaceSession, type OpenTab} from "./workspace/useWorkspaceSession";
import type {CellError, ExportFormat, NavRequest, NavTarget, PublicSchemaDocument, SchemaDefinition, SchemaEditorState, SchemaEditResult, SchemaNode, SchemaUsage, TableDocument, TableExportSettings as TableExportSettingsModel, TreeNode, WorkspaceView} from "./types";

type StructureDialog = {
    mode: "add" | "schema" | "rename" | "move" | "delete" | "displayName";
    name: string;
    second: string;
    schema: string;
    node?: SchemaNode;
    custom?: boolean;
};
type TableDialog = {namespace: string; name: string};

type TreeDialog =
    | {mode: "renameTable"; path: string; namespace: string; name: string}
    | {mode: "deleteTable"; path: string; label: string}
    | {mode: "renameDefinition"; path: string; name: string; newName: string}
    | {mode: "deleteDefinition"; path: string; name: string}
    | {mode: "newSchema"; namespace: string};

function App() {
    const {t, language, setLanguage} = useTranslation();
    const {preference, setPreference} = useTheme();
    const {workspace, setWorkspace, selectedSchema, setSelectedSchema, tabs, setTabs, activeTabId, setActiveTabId, activeTab, table, schemaDocument, activateTab, closeTab, updateActiveTable, updateActiveSchema, updateSchemaTabById, addTableTab, addSchemaTab} = useWorkspaceSession();
    const [cellError, setCellError] = useState<CellError | null>(null);
    const [error, setError] = useState("");
    const [notice, setNotice] = useState("");
    const [dialog, setDialog] = useState<StructureDialog | null>(null);
    const [tableDialog, setTableDialog] = useState<TableDialog | null>(null);
    const [settingsOpen, setSettingsOpen] = useState(false);
    const [projectSettingsOpen, setProjectSettingsOpen] = useState(false);
    const [exportFormats, setExportFormats] = useState<ExportFormat[]>([]);
    const [exportResult, setExportResult] = useState<{directory: string; files: string[]} | null>(null);
    const [appMenu, setAppMenu] = useState<"file" | "edit" | null>(null);
    const [focusTarget, setFocusTarget] = useState<{path: string; rowID: string; column: string; nonce: number} | null>(null);
    const [navMenu, setNavMenu] = useState<{x: number; y: number; targets: NavTarget[]} | null>(null);
    const [treeMenu, setTreeMenu] = useState<{node: TreeNode; namespace: string; x: number; y: number} | null>(null);
    const [treeDialog, setTreeDialog] = useState<TreeDialog | null>(null);
    const [collapsedNodes, setCollapsedNodes] = useState<Set<string>>(new Set());

    useEffect(() => {
        void (async () => {
            try {
                const [settings, formats] = await Promise.all([GetSettings() as Promise<{language?: string; theme?: string}>, GetExportFormats() as Promise<ExportFormat[]>]);
                setExportFormats(formats);
                if (settings.language === "en" || settings.language === "zh-CN") setLanguage(settings.language);
                if (settings.theme === "system" || settings.theme === "light" || settings.theme === "dark") setPreference(settings.theme);
            } catch {
                // Settings may be unavailable before a project is open; defaults apply.
            }
        })();
    }, [setLanguage, setPreference]);

    async function changeLanguage(language: Language) {
        setLanguage(language);
        setError("");
        try {
            await SetLanguage(language);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function changeTheme(theme: ThemePreference) {
        setPreference(theme);
        setError("");
        try {
            await SetTheme(theme);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    function diagnosticText(key?: string, args?: string[], message?: string): string {
        if (!key) return message ?? "";
        const params: Record<string, string> = {};
        (args ?? []).forEach((value, index) => { params[String(index)] = value; });
        return t(key, params);
    }

    function closePeek() {
        setSelectedSchema(null);
    }

    async function openTableTab(path: string) {
        const id = `table:${path}`;
        if (tabs.some((tab) => tab.id === id)) {
            activateTab(id);
            return;
        }
        setError("");
        try {
            const document = await GetTable(path) as TableDocument;
            addTableTab(path, document);
            setCellError(null);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function openSchemaTab(path: string, definition: string | null, local: boolean, creating = false) {
        const id = `schema:${path}`;
        const existing = tabs.find((tab) => tab.id === id);
        if (existing && existing.kind === "schema") {
            if (creating) {
                updateSchemaTabById(id, {...existing.state, selected: null, creating: true});
            } else if (definition != null) {
                updateSchemaTabById(id, {...existing.state, selected: definition});
            }
            activateTab(id);
            return;
        }
        setError("");
        try {
            let state: SchemaEditorState;
            if (local) {
                const document = await GetTable(path) as TableDocument;
                const logical = document.namespace ? `${document.namespace}.${document.name}` : document.name;
                state = {path, local: true, logicalTable: logical, definitions: document.localSchema, selected: definition ?? document.localSchema[0]?.name ?? null, creating};
            } else {
                const document = await GetPublicSchema(path) as PublicSchemaDocument;
                const definitions = document.definitions ?? [];
                state = {path, local: false, package: document.package, definitions, selected: definition ?? definitions[0]?.name ?? null, creating};
            }
            addSchemaTab(path, state);
            setCellError(null);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function refreshOpenTabs() {
        const open = tabs;
        const refreshed: OpenTab[] = [];
        for (const tab of open) {
            try {
                if (tab.kind === "table") {
                    const document = await GetTable(tab.path) as TableDocument;
                    refreshed.push({...tab, table: document});
                } else if (tab.state.local) {
                    const document = await GetTable(tab.state.path) as TableDocument;
                    const logical = document.namespace ? `${document.namespace}.${document.name}` : document.name;
                    const definitions = document.localSchema;
                    refreshed.push({...tab, state: {...tab.state, logicalTable: logical, definitions, selected: definitions.some((entry) => entry.name === tab.state.selected) ? tab.state.selected : (definitions[0]?.name ?? null), creating: false}});
                } else {
                    const document = await GetPublicSchema(tab.state.path) as PublicSchemaDocument;
                    const definitions = document.definitions ?? [];
                    refreshed.push({...tab, state: {...tab.state, package: document.package, definitions, selected: definitions.some((entry) => entry.name === tab.state.selected) ? tab.state.selected : (definitions[0]?.name ?? null), creating: false}});
                }
            } catch {
                continue;
            }
        }
        setTabs(refreshed);
        if (!refreshed.some((tab) => tab.id === activeTabId)) {
            setActiveTabId(refreshed[0]?.id ?? null);
            setSelectedSchema(null);
        }
    }

    async function pasteCells(row: number, column: number, text: string) {
        if (!table) return;
        const values = text.replace(/\r/g, "").split("\n").filter((line, index, lines) => line !== "" || index < lines.length - 1).map((line) => line.split("\t"));
        if (values.length === 0 || values.some((line) => line.length !== values[0].length)) {
            setError(t("error.clipboardColumns"));
            return;
        }
        const rowIDs = table.rows.slice(row, row + values.length).map((entry) => entry.id);
        const columns = table.columns.slice(column - 1, column - 1 + values[0].length).map((entry) => entry.name);
        if (rowIDs.length !== values.length || columns.length !== values[0].length) {
            setError(t("error.pasteRange"));
            return;
        }
        setError("");
        try {
            const result = await Paste({path: table.path, rowIDs, columns, values});
            updateActiveTable(result.table as TableDocument);
            setWorkspace(result.workspace as WorkspaceView);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    useEffect(() => {
        const historyShortcut = (event: KeyboardEvent) => {
            if (!activeTab || !(event.ctrlKey || event.metaKey) || (event.target as Element | null)?.closest("textarea, input, select")) return;
            const redo = event.key.toLowerCase() === "y" || (event.shiftKey && event.key.toLowerCase() === "z");
            if (event.key.toLowerCase() !== "z" && !redo) return;
            if (redo ? !workspace?.canRedo : !workspace?.canUndo) return;
            event.preventDefault();
            void applyHistory(redo);
        };
        document.addEventListener("keydown", historyShortcut);
        return () => document.removeEventListener("keydown", historyShortcut);
    }, [activeTab, workspace, tabs]);

    useEffect(() => {
        const closeMenu = (event: PointerEvent) => {
            if (!(event.target as Element | null)?.closest(".app-menu")) setAppMenu(null);
            if (!(event.target as Element | null)?.closest(".nav-menu")) setNavMenu(null);
            if (!(event.target as Element | null)?.closest(".tree-context-menu")) setTreeMenu(null);
        };
        const closeOnEscape = (event: KeyboardEvent) => {
            if (event.key === "Escape") {
                setAppMenu(null);
                setNavMenu(null);
                setTreeMenu(null);
            }
        };
        document.addEventListener("pointerdown", closeMenu);
        document.addEventListener("keydown", closeOnEscape);
        return () => {
            document.removeEventListener("pointerdown", closeMenu);
            document.removeEventListener("keydown", closeOnEscape);
        };
    }, []);

    async function selectProject() {
        setError("");
        try {
            const view = await SelectProject() as WorkspaceView;
            applyProjectView(view);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    function applyProjectView(view: WorkspaceView) {
        if (!view?.project?.path) return;
        setWorkspace(view);
        setSelectedSchema(schemaDefinitions(view)[0] ?? null);
        setTabs([]);
        setActiveTabId(null);
        setCellError(null);
    }

    async function createProject() {
        setError("");
        try {
            applyProjectView(await NewProject() as WorkspaceView);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function saveProjectAs() {
        setError("");
        try {
            applyProjectView(await SaveProjectAs() as WorkspaceView);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function openTable(path: string) {
        await openTableTab(path);
    }

    async function openSchema(node: TreeNode) {
        if (node.kind === "localSchema") {
            const name = node.name.startsWith("#") ? node.name.replace(/^#/, "") : null;
            await openSchemaTab(node.path!, name, true);
        } else {
            const name = node.kind === "schemaDefinition" ? node.name.replace(/^#/, "") : null;
            await openSchemaTab(node.path!, name, false);
        }
    }

    async function openSchemaByPath(path: string, definition: string | null, local: boolean, creating = false) {
        await openSchemaTab(path, definition, local, creating);
    }

    async function navigateTo(request: NavRequest, position: {x: number; y: number}) {
        setError("");
        try {
            const target = await Navigate(request) as NavTarget;
            applyNavigationTarget(target, position);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    function applyNavigationTarget(target: NavTarget, position: {x: number; y: number}) {
        if (target.kind === "schema-reference") {
            setNavMenu(null);
            void openSchemaByPath(target.path, target.definition ?? null, !!target.local);
            return;
        }
        if (target.references && target.references.length > 0) {
            setNavMenu({x: position.x, y: position.y, targets: target.references});
            return;
        }
        focusTableCell(target.path, target.rowID ?? "", target.column ?? "");
    }

    function focusTableCell(path: string, rowID: string, column: string) {
        setNavMenu(null);
        setFocusTarget({path, rowID, column, nonce: Date.now()});
        if (table?.path !== path) {
            void openTable(path);
        }
    }
    function openUsage(usage: SchemaUsage) {
        if (usage.table) {
            focusTableCell(usage.path, "", usage.field === "#PrimaryKey" ? "" : usage.field);
            return;
        }
        const definition = schemaDefinitions(workspace).find((entry) => entry.id === usage.field);
        if (definition) {
            void openSchemaByPath(definition.path, definition.id.split(".#")[1], !!definition.table);
        }
    }

    function jumpToDiagnostic(diagnostic: {path?: string; table?: string; field?: string; rowID?: string; column?: string}) {
        const path = diagnostic.path;
        if (!path) return;
        if (diagnostic.rowID && diagnostic.column) {
            focusTableCell(path, diagnostic.rowID, diagnostic.column);
            return;
        }
        if (diagnostic.column) {
            focusTableCell(path, "", diagnostic.column);
            return;
        }
        if (diagnostic.field) {
            if (diagnostic.field === "#PrimaryKey") {
                void openTableTab(path);
                return;
            }
            if (diagnostic.table) {
                focusTableCell(path, "", diagnostic.field);
                return;
            }
            const definition = schemaDefinitions(workspace).find((entry) => entry.id === diagnostic.field);
            if (definition) {
                void openSchemaTab(definition.path, definition.id.split(".#")[1], false);
                return;
            }
            void openTableTab(path);
            return;
        }
        void openTableTab(path);
    }

    function renameDefinitionAt(path: string, name: string) {
        setTreeDialog({mode: "renameDefinition", path, name, newName: name});
    }

    function deleteDefinitionAt(path: string, name: string) {
        setTreeDialog({mode: "deleteDefinition", path, name});
    }

    function renameTableAt(path: string, name: string, namespace: string) {
        setTreeDialog({mode: "renameTable", path, namespace, name});
    }

    function deleteTableAt(path: string, label: string) {
        setTreeDialog({mode: "deleteTable", path, label});
    }

    function schemaFileInNamespace(node: TreeNode): TreeNode | null {
        const schemas = node.children?.find((child) => child.kind === "group" && child.name === "Schemas");
        return schemas?.children?.find((child) => child.kind === "schemaFile") ?? null;
    }

    async function newDefinitionInNamespace(namespace: string, schemaPath?: string) {
        setError("");
        try {
            if (schemaPath) {
                await openSchemaByPath(schemaPath, null, false, true);
                return;
            }
            const definition = schemaDefinitions(workspace).find((entry) => !entry.table && entry.id.startsWith(`${namespace}.#`));
            if (definition) {
                await openSchemaByPath(definition.path, null, false, true);
                return;
            }
            const result = await CreateSchemaFile(namespace) as {schema: PublicSchemaDocument; workspace: WorkspaceView};
            setWorkspace(result.workspace);
            await openSchemaByPath(result.schema.path, null, false, true);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function submitTreeDialog() {
        if (!treeDialog) return;
        setError("");
        try {
            if (treeDialog.mode === "newSchema") {
                await newDefinitionInNamespace(treeDialog.namespace);
            } else if (treeDialog.mode === "renameTable") {
                const result = await RenameTable({path: treeDialog.path, namespace: treeDialog.namespace, name: treeDialog.name}) as {table: TableDocument; workspace: WorkspaceView};
                setWorkspace(result.workspace);
                const oldId = `table:${treeDialog.path}`;
                const newId = `table:${result.table.path}`;
                setTabs((current) => current.map((tab) => tab.id === oldId && tab.kind === "table" ? {id: newId, kind: "table", path: result.table.path, table: result.table} : tab));
                setActiveTabId((current) => current === oldId ? newId : current);
            } else if (treeDialog.mode === "deleteTable") {
                const result = await DeleteTable({path: treeDialog.path}) as {workspace: WorkspaceView};
                setWorkspace(result.workspace);
                setTabs((current) => current.filter((tab) => !(tab.kind === "table" && tab.path === treeDialog.path)));
                setActiveTabId((current) => current === `table:${treeDialog.path}` ? null : current);
            } else if (treeDialog.mode === "renameDefinition") {
                const result = await RenameSchemaDefinition({path: treeDialog.path, oldName: treeDialog.name, newName: treeDialog.newName}) as SchemaEditResult;
                setWorkspace(result.workspace);
                setTabs((current) => current.map((tab) => {
                    if (tab.kind !== "schema" || tab.state.path !== treeDialog.path) return tab;
                    const state = tab.state.local
                        ? {...tab.state, definitions: result.table.localSchema, selected: treeDialog.newName, creating: false}
                        : {...tab.state, definitions: result.schema.definitions, selected: treeDialog.newName, creating: false};
                    return {...tab, state};
                }));
            } else {
                const result = await DeleteSchemaDefinition({path: treeDialog.path, name: treeDialog.name}) as SchemaEditResult;
                setWorkspace(result.workspace);
                setTabs((current) => current.map((tab) => {
                    if (tab.kind !== "schema" || tab.state.path !== treeDialog.path) return tab;
                    const state = tab.state.local
                        ? {...tab.state, definitions: result.table.localSchema, selected: result.table.localSchema[0]?.name ?? null, creating: false}
                        : {...tab.state, definitions: result.schema.definitions, selected: result.schema.definitions[0]?.name ?? null, creating: false};
                    return {...tab, state};
                }));
            }
            setTreeDialog(null);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    function applySchemaResult(result: SchemaEditResult, selected: string | null) {
        setWorkspace(result.workspace);
        if (!schemaDocument) return;
        if (schemaDocument.local) {
            updateActiveSchema({...schemaDocument, definitions: result.table.localSchema, selected: selected ?? result.table.localSchema[0]?.name ?? null, creating: false});
        } else {
            updateActiveSchema({...schemaDocument, definitions: result.schema.definitions, selected: selected ?? result.schema.definitions[0]?.name ?? null, creating: false});
        }
    }

    async function saveSchemaDefinition(create: boolean, name: string, kind: "type" | "constraint", type: string, appliesTo: string[], expression: string) {
        if (!schemaDocument) return;
        setError("");
        const payload = {path: schemaDocument.path, name, kind, type, appliesTo, schemaExpression: expression};
        const result = create
            ? await CreateSchemaDefinition(payload) as SchemaEditResult
            : await SetSchemaDefinition(payload) as SchemaEditResult;
        applySchemaResult(result, name);
    }

    async function renameSchemaDefinition(oldName: string, newName: string) {
        if (!schemaDocument) return;
        setError("");
        const result = await RenameSchemaDefinition({path: schemaDocument.path, oldName, newName}) as SchemaEditResult;
        applySchemaResult(result, newName);
    }

    async function deleteSchemaDefinition(name: string) {
        if (!schemaDocument) return;
        setError("");
        const result = await DeleteSchemaDefinition({path: schemaDocument.path, name}) as SchemaEditResult;
        applySchemaResult(result, null);
    }

    function schemaUsages(name: string): SchemaUsage[] {
        if (!schemaDocument || !workspace) return [];
        const id = schemaDocument.local
            ? `${schemaDocument.logicalTable}._local.#${name}`
            : `${schemaDocument.package}.#${name}`;
        return schemaDefinitions(workspace).find((definition) => definition.id === id)?.usages ?? [];
    }

    async function saveCellInput(rowID: string, column: string, input: string) {
        if (!table) {
            return;
        }
        setError("");
        setCellError(null);
        try {
            if (column === "__create__") {
                const {values, formulas} = JSON.parse(input);
                const result = await CreateRow({path: table.path, rowID, values, formulas});
                updateActiveTable(result.table as TableDocument);
                setWorkspace(result.workspace as WorkspaceView);
                return;
            }
            if (column === "id") {
                if (!input) throw new Error(t("error.idEmpty"));
                const result = await RenameRow({path: table.path, oldID: rowID, newID: input});
                updateActiveTable(result.table as TableDocument);
                setWorkspace(result.workspace as WorkspaceView);
                return;
            }
            if (input === "") {
                const result = await ClearCell({path: table.path, rowID, column, expression: ""});
                updateActiveTable(result.table as TableDocument);
                setWorkspace(result.workspace as WorkspaceView);
                return;
            }
            const columnDefinition = table.columns.find((entry) => entry.name === column);
            const result = input.startsWith("=")
                ? await SetCellFormula({path: table.path, rowID, column, expression: input.slice(1).trim()})
                : columnDefinition?.schema.kind === "string"
                    ? await SetCellText({path: table.path, rowID, column, expression: input})
                    : await UpdateCell({path: table.path, rowID, column, expression: input});
            updateActiveTable(result.table as TableDocument);
            setWorkspace(result.workspace as WorkspaceView);
        } catch (reason) {
            // The draft editor owns creation errors and must keep its unsaved row.
            // Ordinary cell edits continue to report their error through the app toast.
            if (column === "__create__") throw reason;
            const message = reason instanceof Error ? reason.message : String(reason);
            setError(message);
            setCellError({rowID, column, message});
        }
    }

    async function saveTableExportSettings(settings: TableExportSettingsModel) {
        setError("");
        try {
            const result = await SetTableExportSettings(new workspaceModel.TableExportSettings(settings)) as {table: TableDocument; workspace: WorkspaceView};
            updateActiveTable(result.table);
            setWorkspace(result.workspace);
        } catch (reason) {
            const message = reason instanceof Error ? reason.message : String(reason);
            setError(message);
            throw reason;
        }
    }

    async function saveCellComment(rowID: string, column: string, comment: string) {
        setError("");
        try {
            const result = await SetCellComment({path: table!.path, rowID, column, comment}) as {table: TableDocument; workspace: WorkspaceView};
            updateActiveTable(result.table);
            setWorkspace(result.workspace);
        } catch (reason) {
            const message = reason instanceof Error ? reason.message : String(reason);
            setError(message);
            throw reason;
        }
    }

    async function submitStructureDialog() {
        if (!table || !dialog) return;
        setError("");
        setCellError(null);
        try {
            if (dialog.mode === "displayName") {
                const result = await SetColumnDisplayName({path: table.path, column: dialog.name, displayName: dialog.second}) as {table: TableDocument; workspace: WorkspaceView};
                updateActiveTable(result.table);
                setWorkspace(result.workspace);
                setDialog(null);
                return;
            }
            let schemaExpression = dialog.schema;
            if ((dialog.mode === "add" || dialog.mode === "schema") && !dialog.custom) {
                schemaExpression = await PrintSchema((dialog.node ?? {kind: "string"}) as unknown as schemamodel.Node);
            }
            let result;
            if (dialog.mode === "add") {
                result = await AddColumn({path: table.path, name: dialog.name, schemaExpression, initialExpression: "", position: Number(dialog.second)});
            } else if (dialog.mode === "schema") {
                result = await SetColumnSchema({path: table.path, name: dialog.name, schemaExpression});
            } else if (dialog.mode === "rename") {
                result = await RenameColumn({path: table.path, oldName: dialog.name, newName: dialog.second});
            } else if (dialog.mode === "move") {
                const position = Number(dialog.second);
                if (!Number.isInteger(position)) throw new Error(t("error.columnPosition"));
                result = await MoveColumn({path: table.path, name: dialog.name, position});
            } else {
                result = await DeleteColumn({path: table.path, name: dialog.name});
            }
            updateActiveTable(result.table as TableDocument);
            setWorkspace(result.workspace as WorkspaceView);
            setDialog(null);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function deleteRow(rowID: string) {
        if (!table) {
            return;
        }
        setError("");
        setCellError(null);
        try {
            const result = await DeleteRow({path: table.path, rowID});
            updateActiveTable(result.table as TableDocument);
            setWorkspace(result.workspace as WorkspaceView);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function createTable() {
        if (!tableDialog) return;
        setError("");
        try {
            const result = await CreateTable(tableDialog);
            setWorkspace(result.workspace as WorkspaceView);
            setTabs((current) => [...current, {id: `table:${result.table.path}`, kind: "table", path: result.table.path, table: result.table as TableDocument}]);
            setActiveTabId(`table:${result.table.path}`);
            setSelectedSchema(null);
            setTableDialog(null);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function applyHistory(redo: boolean) {
        setError("");
        try {
            const result = redo ? await Redo() : await Undo();
            setWorkspace(result.workspace as WorkspaceView);
            await refreshOpenTabs();
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function exportProject() {
        setError("");
        setNotice("");
        try {
            const result = await Export() as {directory: string; files: string[]};
            setExportResult(result);
            void OpenPath(result.directory).catch(() => undefined);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function toggleCompileMode() {
        if (!workspace) return;
        setError("");
        try {
            const mode = workspace.compile.mode === "auto" ? "manual" : "auto";
            const view = await SetCompileMode(mode) as WorkspaceView;
            setWorkspace(view);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function compileProject() {
        if (!workspace) return;
        setError("");
        try {
            const result = await Compile() as {workspace: WorkspaceView};
            setWorkspace(result.workspace);
            if (table) {
                updateActiveTable(await GetTable(table.path) as TableDocument);
            }
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    return (
        <main className="app-shell" onContextMenu={(event) => event.preventDefault()}>
            <header className="app-header">
                <div className="brand">Datapack</div>
                <div className="header-divider"/>
                <nav className="app-menubar" aria-label={t("menu.appMenu")}>
                    <div className="app-menu">
                        <button aria-expanded={appMenu === "file"} onClick={() => setAppMenu(appMenu === "file" ? null : "file")}>{t("menu.file")}</button>
                        {appMenu === "file" && <div className="app-menu-popover">
                            <button onClick={() => { setAppMenu(null); void createProject(); }}>{t("menu.newProject")}</button>
                            <button onClick={() => { setAppMenu(null); selectProject(); }}>{t("menu.openProject")}</button>
                            <button disabled={!workspace} onClick={() => { setAppMenu(null); void saveProjectAs(); }}>{t("menu.saveAs")}</button>
                            <hr/>
                            <button disabled={!workspace} onClick={() => { setAppMenu(null); setTableDialog({namespace: "game", name: ""}); }}>{t("menu.newTable")}</button>
                            <hr/>
                            <button disabled={!workspace} onClick={() => { setAppMenu(null); setProjectSettingsOpen(true); }}>{t("projectSettings.title")}</button>
                            <button disabled={!workspace} onClick={() => { setAppMenu(null); void exportProject(); }}>{t("menu.export")}</button>
                        </div>}
                    </div>
                    <div className="app-menu">
                        <button aria-expanded={appMenu === "edit"} onClick={() => setAppMenu(appMenu === "edit" ? null : "edit")}>{t("menu.edit")}</button>
                        {appMenu === "edit" && <div className="app-menu-popover">
                            <button disabled={!workspace?.canUndo} onClick={() => { setAppMenu(null); applyHistory(false); }}>{t("menu.undo")} <kbd>Ctrl+Z</kbd></button>
                            <button disabled={!workspace?.canRedo} onClick={() => { setAppMenu(null); applyHistory(true); }}>{t("menu.redo")} <kbd>Ctrl+Y</kbd></button>
                        </div>}
                    </div>
                </nav>
                <span className="project-name">{workspace?.tree.root.name ?? t("header.noProjectOpen")}</span>
                <div className="header-actions">
                    {workspace && <button className={`build-mode build-mode-${workspace.compile.mode} build-mode-toggle`} aria-label={t("header.toggleCompileMode")} title={t("header.compileModeHint")} onClick={toggleCompileMode}>{workspace.compile.mode === "auto" ? t("header.autoCompile") : t("header.manualCompile")}{workspace.compile.pending ? t("header.pending") : ""}</button>}
                    {workspace && workspace.compile.mode === "manual" && <button className="compile-button" disabled={!workspace.compile.pending} onClick={compileProject}>{t("header.compile")}</button>}
                    {workspace && workspace.compile.diagnostics.length > 0 && <span className="compile-errors" title={t("header.compileDiagnostics")}>{workspace.compile.diagnostics.length}</span>}
                    <button className="export-button" aria-label={t("settings.title")} onClick={() => setSettingsOpen(true)}>{t("settings.title")}</button>
                </div>
            </header>

            {workspace ? (
                <section className="workbench">
                    <aside className="project-explorer">
                        <div className="panel-title">{t("panel.project")}</div>
                        <div className="tree"><Tree node={workspace.tree.root} collapsed={collapsedNodes} onToggleCollapse={(key) => setCollapsedNodes((current) => {
                            const next = new Set(current);
                            if (next.has(key)) next.delete(key); else next.add(key);
                            return next;
                        })} onOpenTable={openTable} onOpenSchema={openSchema} onOpenProjectSettings={() => setProjectSettingsOpen(true)} onContextMenu={(node, position, namespace) => {
                            if (node.kind === "project") {
                                setTreeMenu({node, namespace, x: position.x, y: position.y});
                                return;
                            }
                            if (node.kind === "group" || node.kind === "damaged") return;
                            setTreeMenu({node, namespace, x: position.x, y: position.y});
                        }}/></div>
                    </aside>
                    <section className={`editor-placeholder ${activeTab || selectedSchema ? "editor-table" : ""}`}>
                        {tabs.length > 0 && <div className="tab-bar" role="tablist">
                            {tabs.map((tab) => {
                                const isActive = tab.id === activeTabId && !selectedSchema;
                                const label = tab.kind === "table"
                                    ? `${tab.table.namespace}.${tab.table.name}`
                                    : tab.state.local
                                        ? `${tab.state.logicalTable ?? tab.state.path} ${t("common.local")}`
                                        : tab.state.package ?? tab.state.path;
                                return <div key={tab.id} role="tab" aria-selected={isActive} className={`tab ${isActive ? "tab-active" : ""}`} onClick={() => activateTab(tab.id)}>
                                    <span className="tab-label">{label}</span>
                                    <button className="tab-close" type="button" onClick={(event) => { event.stopPropagation(); closeTab(tab.id); }}>×</button>
                                </div>;
                            })}
                        </div>}
                        {tabs.map((tab) => (
                            <div key={tab.id} className="tab-panel" style={{display: tab.id === activeTabId && !selectedSchema ? "flex" : "none"}}>
                                {tab.kind === "table"
                                     ? <TableGrid key={tab.id} active={tab.id === activeTabId && !selectedSchema} table={tab.table} formats={exportFormats} onAddColumn={(position) => setDialog({mode: "add", name: "", second: String(position), schema: "string", node: {kind: "string"}, custom: false})} onSetColumnDisplayName={(name) => setDialog({mode: "displayName", name, second: tab.table.columns.find((column) => column.name === name)?.displayName ?? "", schema: ""})} onSetColumnWidth={async (column, width) => { const result = await SetColumnWidth({path: tab.table.path, column, width}); updateActiveTable(result.table as TableDocument); setWorkspace(result.workspace as WorkspaceView); }} onDeleteColumn={(name) => setDialog({mode: "delete", name, second: "", schema: ""})} onDeleteRow={deleteRow} onEditColumnSchema={(name, schemaExpression) => setDialog({mode: "schema", name, second: "", schema: schemaExpression.replaceAll("_local.#", "#"), node: tab.table.columns.find((column) => column.name === name)?.schema ?? {kind: "string"}, custom: false})} onMoveColumn={(name) => setDialog({mode: "move", name, second: "0", schema: ""})} onRenameColumn={(name) => setDialog({mode: "rename", name, second: "", schema: ""})} onPaste={pasteCells} onSaveInput={saveCellInput} onSaveComment={saveCellComment} onSaveExportSettings={saveTableExportSettings} onNavigate={navigateTo} onOpenSchemaReference={(reference) => { const localName = reference.startsWith("_local.#") ? reference.slice("_local.".length) : ""; const definition = schemaDefinitions(workspace).find((entry) => entry.id === reference || (localName && entry.id.endsWith("._local." + localName))); if (definition) void openSchemaByPath(definition.path, definition.id.split(".#")[1], !!definition.table); }} focusTarget={focusTarget} onFocusDone={() => setFocusTarget(null)}/>
                                     : <SchemaEditor key={tab.id} state={tab.state} usages={schemaUsages} onOpenUsage={openUsage} workspace={workspace} onSelect={(name) => updateSchemaTabById(tab.id, {...tab.state, selected: name, creating: false})} onCreate={(name, kind, type, appliesTo, expression) => saveSchemaDefinition(true, name, kind, type, appliesTo, expression)} onApply={(name, kind, type, appliesTo, expression) => saveSchemaDefinition(false, name, kind, type, appliesTo, expression)} onRename={renameSchemaDefinition} onDelete={deleteSchemaDefinition} onOpenSchemaReference={(reference) => { const localName = reference.startsWith("_local.#") ? reference.slice("_local.".length) : ""; const definition = schemaDefinitions(workspace).find((entry) => entry.id === reference || (localName && entry.id.endsWith("._local." + localName))); if (definition) void openSchemaByPath(definition.path, definition.id.split(".#")[1], !!definition.table); }}/>}
                            </div>
                        ))}
                        {selectedSchema && <section className="schema-detail">
                            <div className="schema-detail-head">
                                <span className={`schema-kind-dot schema-kind-${selectedSchema.kind}`} aria-hidden="true"/>
                                <h1>{selectedSchema.id}</h1>
                                <button className="peek-close" type="button" aria-label={t("schemaDetail.close")} onClick={closePeek}>×</button>
                            </div>
                            <div className="schema-detail-row"><span>{t("schemaDetail.kind")}</span><code>{selectedSchema.kind}</code></div>
                            <div className="schema-detail-row"><span>{t("schemaDetail.filepath")}</span><code>{selectedSchema.path}</code></div>
                            {selectedSchema.appliesTo && <div className="schema-detail-row"><span>{t("schemaDetail.appliesTo")}</span><code>{selectedSchema.appliesTo.join(", ")}</code></div>}
                            <div className="schema-detail-row"><span>{t("schemaDetail.references")}</span><code>{selectedSchema.usages?.length ?? 0}</code></div>
                            <div className="schema-detail-references">
                                {(selectedSchema.usages ?? []).map((usage) => <button className="reference-row" key={`${usage.path}:${usage.field}`} onClick={() => openUsage(usage)}>
                                    <span>{usage.table ?? usage.path}</span>
                                    <code>{usage.field}</code>
                                </button>)}
                                {(selectedSchema.usages ?? []).length === 0 && <div className="schema-editor-empty">{t("empty.noReferences")}</div>}
                            </div>
                        </section>}
                        {!activeTab && !selectedSchema && <div className="schema-editor-empty">{t("empty.noOpenDocuments")}</div>}
                    </section>
                    <aside className="schema-explorer">
                        <div className="diagnostics-panel">
                            <div className="panel-title">{t("panel.diagnostics")}</div>
                            {(workspace.schemas.diagnostics ?? []).length === 0 && (workspace.foreignKeys.diagnostics ?? []).length === 0 && workspace.compile.diagnostics.length === 0 && <div className="schema-editor-empty">{t("panel.noDiagnostics")}</div>}
                            {(workspace.schemas.diagnostics ?? []).length > 0 && <div className="schema-diagnostics">
                                {(workspace.schemas.diagnostics ?? []).map((diagnostic) => <button className="schema-diagnostic schema-diagnostic-action" key={`${diagnostic.path}:${diagnostic.field}:${diagnostic.message}`} onClick={() => jumpToDiagnostic(diagnostic)}>
                                    <code>{diagnostic.field}</code>
                                    <span>{diagnosticText(diagnostic.key, diagnostic.args, diagnostic.message)}</span>
                                </button>)}
                            </div>}
                            {(workspace.foreignKeys.diagnostics ?? []).length > 0 && <div className="schema-diagnostics">
                                {(workspace.foreignKeys.diagnostics ?? []).map((diagnostic) => <button className="schema-diagnostic schema-diagnostic-action" key={`${diagnostic.path}:${diagnostic.rowID}:${diagnostic.column}:${diagnostic.message}`} onClick={() => jumpToDiagnostic(diagnostic)}>
                                    <code>{diagnostic.rowID}.{diagnostic.column}</code>
                                    <span>{diagnosticText(diagnostic.key, diagnostic.args, diagnostic.message)}</span>
                                </button>)}
                            </div>}
                            {workspace.compile.diagnostics.length > 0 && <div className="schema-diagnostics compile-diagnostics">
                                <div className="schema-diagnostics-title">{t("panel.compileDiagnostics")}</div>
                                {workspace.compile.diagnostics.map((diagnostic) => <button className="schema-diagnostic schema-diagnostic-action" key={`${diagnostic.path}:${diagnostic.rowID ?? ""}:${diagnostic.column ?? ""}:${diagnostic.message}`} onClick={() => jumpToDiagnostic(diagnostic)}>
                                    <code>{[diagnostic.rowID, diagnostic.column].filter(Boolean).join(".") || diagnostic.field || diagnostic.table || t("panel.compile")}</code>
                                    <span>{diagnosticText(diagnostic.key, diagnostic.args, diagnostic.message)}</span>
                                </button>)}
                            </div>}
                        </div>
                        <div className="schema-panel">
                            <div className="panel-title">{t("panel.schemas")}</div>
                            <div className="schema-list">
                                {schemaDefinitions(workspace).map((definition) => <button
                                    className={`schema-row ${selectedSchema?.id === definition.id ? "schema-row-selected" : ""}`}
                                    key={definition.id}
                                    onClick={() => { setCellError(null); setSelectedSchema((current) => current?.id === definition.id ? null : definition); }}
                                >
                                    <span className={`schema-kind-dot schema-kind-${definition.kind}`} aria-hidden="true"/>
                                    <span>{definition.id}</span>
                                    <span className="usage-count">{definition.usages?.length ?? 0}</span>
                                </button>)}
                            </div>
                        </div>
                    </aside>
                </section>
            ) : (
                <section className="welcome">
                    <div className="welcome-mark">DP</div>
                    <h1>{t("welcome.title")}</h1>
                    <p>{t("welcome.body")}</p>
                    <div className="welcome-actions">
                        <button className="open-button" onClick={selectProject}>{t("menu.openProject")}</button>
                        <button className="export-button" onClick={createProject}>{t("menu.newProject")}</button>
                    </div>
                </section>
            )}
            <LogPanel/>
            {navMenu && <div className="nav-menu" style={{left: navMenu.x, top: navMenu.y}} role="menu" aria-label={t("schemaDetail.references")}>
                {navMenu.targets.map((target, index) => <button key={index} onClick={() => focusTableCell(target.path, target.rowID ?? "", target.column ?? "")}>
                    <code>{target.path}</code>
                    <span>{target.rowID}{target.column ? `.${target.column}` : ""}</span>
                </button>)}
            </div>}
            {treeMenu && <div className="nav-menu tree-context-menu" style={{left: treeMenu.x, top: treeMenu.y}} role="menu" aria-label={t("tree.actions")}>
                {treeMenu.node.kind === "project" && <>
                    <button onClick={() => { setTableDialog({namespace: "game", name: ""}); setTreeMenu(null); }}>{t("menu.newTable")}</button>
                    <button onClick={() => { setTreeDialog({mode: "newSchema", namespace: ""}); setTreeMenu(null); }}>{t("tree.newSchema")}</button>
                </>}
                {treeMenu.node.kind === "table" && <>
                    <button onClick={() => { renameTableAt(treeMenu.node.path!, treeMenu.node.name, treeMenu.namespace); setTreeMenu(null); }}>{t("tree.renameTable")}</button>
                    <button className="tree-context-danger" onClick={() => { deleteTableAt(treeMenu.node.path!, treeMenu.namespace ? `${treeMenu.namespace}.${treeMenu.node.name}` : treeMenu.node.name); setTreeMenu(null); }}>{t("tree.deleteTable")}</button>
                </>}
                {treeMenu.node.kind === "localSchema" && treeMenu.node.name === "local_schema" && <button onClick={() => { setTreeMenu(null); void openSchemaByPath(treeMenu.node.path!, null, true, true); }}>{t("tree.newDefinition")}</button>}
                {treeMenu.node.kind === "localSchema" && treeMenu.node.name.startsWith("#") && <>
                    <button onClick={() => { renameDefinitionAt(treeMenu.node.path!, treeMenu.node.name.replace(/^#/, "")); setTreeMenu(null); }}>{t("tree.renameDefinition")}</button>
                    <button className="tree-context-danger" onClick={() => { deleteDefinitionAt(treeMenu.node.path!, treeMenu.node.name.replace(/^#/, "")); setTreeMenu(null); }}>{t("tree.deleteDefinition")}</button>
                </>}
                {treeMenu.node.kind === "schemaDefinition" && <>
                    <button onClick={() => { renameDefinitionAt(treeMenu.node.path!, treeMenu.node.name.replace(/^#/, "")); setTreeMenu(null); }}>{t("tree.renameDefinition")}</button>
                    <button className="tree-context-danger" onClick={() => { deleteDefinitionAt(treeMenu.node.path!, treeMenu.node.name.replace(/^#/, "")); setTreeMenu(null); }}>{t("tree.deleteDefinition")}</button>
                </>}
                {treeMenu.node.kind === "schemaFile" && <>
                    <button onClick={() => { setTreeMenu(null); void openSchemaByPath(treeMenu.node.path!, null, false); }}>{t("tree.openSchema")}</button>
                    <button onClick={() => { setTreeMenu(null); void newDefinitionInNamespace(treeMenu.namespace, treeMenu.node.path); }}>{t("tree.newDefinition")}</button>
                </>}
                {treeMenu.node.kind === "namespace" && <>
                    <button onClick={() => { setTreeMenu(null); setTableDialog({namespace: treeMenu.node.name, name: ""}); }}>{t("menu.newTable")}</button>
                    {schemaFileInNamespace(treeMenu.node)
                        ? <>
                            <button onClick={() => { const schemaFile = schemaFileInNamespace(treeMenu.node); setTreeMenu(null); void openSchemaByPath(schemaFile!.path!, null, false); }}>{t("tree.openSchema")}</button>
                            <button onClick={() => { const schemaFile = schemaFileInNamespace(treeMenu.node); setTreeMenu(null); void newDefinitionInNamespace(treeMenu.node.name, schemaFile!.path); }}>{t("tree.newDefinition")}</button>
                        </>
                        : <button onClick={() => { setTreeMenu(null); void newDefinitionInNamespace(treeMenu.node.name); }}>{t("tree.newSchema")}</button>}
                </>}
            </div>}
            {error && <div className="error-toast" role="alert">{error}</div>}
            {notice && <div className="notice-toast" role="status">{notice}</div>}
            {dialog && <div className="dialog-backdrop" role="presentation" onMouseDown={() => setDialog(null)}>
                <form className="structure-dialog" onMouseDown={(event) => event.stopPropagation()} onSubmit={(event) => { event.preventDefault(); submitStructureDialog(); }}>
                    <h2>{dialog.mode === "add" ? t("column.add") : dialog.mode === "schema" ? t("column.editSchema") : dialog.mode === "rename" ? t("column.rename") : dialog.mode === "move" ? t("column.move") : dialog.mode === "displayName" ? t("column.displayName") : t("column.delete")}</h2>
                    {dialog.mode !== "schema" && dialog.mode !== "displayName" && <label>{t("column.name")}<input autoFocus value={dialog.name} onChange={(event) => setDialog({...dialog, name: event.target.value})}/></label>}
                    {dialog.mode === "schema" && <label>{t("column.name")}<input value={dialog.name} disabled/></label>}
                    {(dialog.mode === "add" || dialog.mode === "schema") && <>
                        <div className="schema-mode-toggle">
                            <button type="button" className={!dialog.custom ? "schema-mode-active" : ""} onClick={() => setDialog({...dialog, custom: false})}>{t("column.structured")}</button>
                            <button type="button" className={dialog.custom ? "schema-mode-active" : ""} onClick={() => setDialog({...dialog, custom: true})}>{t("column.customExpression")}</button>
                        </div>
                        {!dialog.custom
                             ? <SchemaForm value={dialog.node ?? {kind: "string"}} onChange={(node) => setDialog({...dialog, node})} workspace={workspace} onOpenSchemaReference={(reference) => { const localName = reference.startsWith("_local.#") ? reference.slice("_local.".length) : ""; const definition = schemaDefinitions(workspace).find((entry) => entry.id === reference || (localName && entry.id.endsWith("._local." + localName))); if (definition) { setDialog(null); void openSchemaByPath(definition.path, definition.id.split(".#")[1], !!definition.table); } }}/>
                            : <label>{t("column.cueSchema")}<textarea autoFocus value={dialog.schema} onChange={(event) => setDialog({...dialog, schema: event.target.value})}/></label>}
                    </>}
                    {dialog.mode === "rename" && <label>{t("column.newName")}<input value={dialog.second} onChange={(event) => setDialog({...dialog, second: event.target.value})}/></label>}
                    {dialog.mode === "move" && <label>{t("column.newPosition")}<input type="number" min="0" value={dialog.second} onChange={(event) => setDialog({...dialog, second: event.target.value})}/></label>}
                    {dialog.mode === "delete" && <p>{t("column.deleteBody")}</p>}
                    {dialog.mode === "displayName" && <>
                        <label>{t("column.displayName")}<input autoFocus value={dialog.second} placeholder={t("column.displayNamePlaceholder")} onChange={(event) => setDialog({...dialog, second: event.target.value})}/></label>
                        <p className="schema-editor-empty">{t("column.internalName")}: <code>{dialog.name}</code></p>
                    </>}
                    <div className="dialog-actions"><button type="button" onClick={() => setDialog(null)}>{t("common.cancel")}</button><button className={dialog.mode === "delete" ? "dialog-danger" : ""} type="submit">{dialog.mode === "delete" ? t("column.delete") : t("common.apply")}</button></div>
                </form>
            </div>}
            {tableDialog && <div className="dialog-backdrop" role="presentation" onMouseDown={() => setTableDialog(null)}>
                <form className="structure-dialog" onMouseDown={(event) => event.stopPropagation()} onSubmit={(event) => { event.preventDefault(); createTable(); }}>
                    <h2>{t("table.create")}</h2>
                    <label>{t("table.namespace")}<input autoFocus value={tableDialog.namespace} onChange={(event) => setTableDialog({...tableDialog, namespace: event.target.value})}/></label>
                    <label>{t("table.name")}<input value={tableDialog.name} onChange={(event) => setTableDialog({...tableDialog, name: event.target.value})}/></label>
                    <div className="dialog-actions"><button type="button" onClick={() => setTableDialog(null)}>{t("common.cancel")}</button><button type="submit">{t("common.create")}</button></div>
                </form>
            </div>}
            {treeDialog && <div className="dialog-backdrop" role="presentation" onMouseDown={() => setTreeDialog(null)}>
                <form className="structure-dialog" onMouseDown={(event) => event.stopPropagation()} onSubmit={(event) => { event.preventDefault(); submitTreeDialog(); }}>
                    {treeDialog.mode === "renameTable" && <>
                        <h2>{t("table.rename")}</h2>
                        <label>{t("table.namespace")}<input autoFocus value={treeDialog.namespace} onChange={(event) => setTreeDialog({...treeDialog, namespace: event.target.value})}/></label>
                        <label>{t("table.name")}<input value={treeDialog.name} onChange={(event) => setTreeDialog({...treeDialog, name: event.target.value})}/></label>
                        <div className="dialog-actions"><button type="button" onClick={() => setTreeDialog(null)}>{t("common.cancel")}</button><button type="submit">{t("common.apply")}</button></div>
                    </>}
                    {treeDialog.mode === "deleteTable" && <>
                        <h2>{t("table.delete")}</h2>
                        <p>{t("table.deleteBody", {label: treeDialog.label})}</p>
                        <div className="dialog-actions"><button type="button" onClick={() => setTreeDialog(null)}>{t("common.cancel")}</button><button className="dialog-danger" type="submit">{t("table.delete")}</button></div>
                    </>}
                    {treeDialog.mode === "renameDefinition" && <>
                        <h2>{t("definition.rename")}</h2>
                        <label>{t("definition.name")}<input autoFocus value={treeDialog.newName} onChange={(event) => setTreeDialog({...treeDialog, newName: event.target.value})}/></label>
                        <div className="dialog-actions"><button type="button" onClick={() => setTreeDialog(null)}>{t("common.cancel")}</button><button type="submit">{t("common.apply")}</button></div>
                    </>}
                    {treeDialog.mode === "deleteDefinition" && <>
                        <h2>{t("definition.delete")}</h2>
                        <p>{t("definition.deleteBody", {name: `#${treeDialog.name}`})}</p>
                        <div className="dialog-actions"><button type="button" onClick={() => setTreeDialog(null)}>{t("common.cancel")}</button><button className="dialog-danger" type="submit">{t("definition.delete")}</button></div>
                    </>}
                    {treeDialog.mode === "newSchema" && <>
                        <h2>{t("tree.newSchema")}</h2>
                        <label>{t("table.namespace")}<input autoFocus value={treeDialog.namespace} onChange={(event) => setTreeDialog({...treeDialog, namespace: event.target.value})}/></label>
                        <div className="dialog-actions"><button type="button" onClick={() => setTreeDialog(null)}>{t("common.cancel")}</button><button type="submit">{t("common.create")}</button></div>
                    </>}
                </form>
            </div>}
            {settingsOpen && <div className="dialog-backdrop" role="presentation" onMouseDown={() => setSettingsOpen(false)}>
                <form className="structure-dialog" onMouseDown={(event) => event.stopPropagation()} onSubmit={(event) => { event.preventDefault(); setSettingsOpen(false); }}>
                    <h2>{t("settings.title")}</h2>
                    <label>{t("settings.language")}
                        <select value={language} onChange={(event) => changeLanguage(event.target.value as Language)}>
                            <option value="zh-CN">简体中文</option>
                            <option value="en">English</option>
                        </select>
                    </label>
                    <label>{t("settings.theme")}
                        <select value={preference} onChange={(event) => changeTheme(event.target.value as ThemePreference)}>
                            <option value="system">{t("settings.theme.system")}</option>
                            <option value="light">{t("settings.theme.light")}</option>
                            <option value="dark">{t("settings.theme.dark")}</option>
                        </select>
                    </label>
                    <div className="dialog-actions"><button type="submit">{t("common.apply")}</button></div>
                </form>
            </div>}
            {projectSettingsOpen && <ProjectSettings onClose={() => setProjectSettingsOpen(false)} onSaved={(view) => setWorkspace(view)}/>}
            {exportResult && <div className="dialog-backdrop" role="presentation" onMouseDown={() => setExportResult(null)}>
                <form className="structure-dialog" onMouseDown={(event) => event.stopPropagation()} onSubmit={(event) => { event.preventDefault(); void OpenPath(exportResult.directory); }}>
                    <h2>{t("export.title")}</h2>
                    <label>{t("export.directory")}<input value={exportResult.directory} readOnly onFocus={(event) => event.currentTarget.select()}/></label>
                    <ul className="export-files">{exportResult.files.map((file) => <li key={file}>{file}</li>)}</ul>
                    <div className="dialog-actions"><button type="button" onClick={() => setExportResult(null)}>{t("common.close")}</button><button type="submit">{t("export.openDirectory")}</button></div>
                </form>
            </div>}
        </main>
    )
}

export default App
