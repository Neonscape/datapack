import type {document, projecttree, schemaregistry, workspace} from "../../wailsjs/go/models";
import type {ProjectSettings, PublicSchemaDocument, SchemaDefinitionDocument, TableDocument, TreeNode, WorkspaceView} from "../types";

function schemaKind(kind: string): "type" | "constraint" {
    if (kind === "type" || kind === "constraint") return kind;
    throw new Error(`Unsupported Schema kind: ${kind}`);
}
function buildMode(mode: string): "auto" | "manual" {
    if (mode === "auto" || mode === "manual") return mode;
    throw new Error(`Unsupported compile mode: ${mode}`);
}
function treeKind(kind: string): TreeNode["kind"] {
    switch (kind) {
        case "project": case "namespace": case "group": case "table": case "schemaFile": case "localSchema": case "schemaDefinition": case "damaged": return kind;
        default: throw new Error(`Unsupported project tree kind: ${kind}`);
    }
}
function treeNode(value: projecttree.Node): TreeNode {
    return {...value,kind:treeKind(value.kind),children:value.children?.map(treeNode)};
}
function definition(value: document.SchemaDefinition): SchemaDefinitionDocument {
    return {...value,kind:schemaKind(value.kind)};
}
export function tableDocument(value?: document.Table): TableDocument {
    return {...value,path:value?.path ?? "",namespace:value?.namespace ?? "",name:value?.name ?? "",columns:value?.columns ?? [],rows:value?.rows ?? [],localSchema:(value?.localSchema ?? []).map(definition)};
}
export function publicSchema(value?: document.PublicSchema): PublicSchemaDocument {
    return {...value,path:value?.path ?? "",package:value?.package ?? "",definitions:(value?.definitions ?? []).map(definition)};
}
export function workspaceView(value: workspace.View): WorkspaceView {
    // Wails project selection returns a zero View on cancellation. Preserve its
    // empty path sentinel; callers keep the currently open project unchanged.
    if (!value?.project?.path) return {project:{path:"",root:""},settings:{build:{mode:"manual"}},tree:{root:{kind:"project",name:""}},schemas:{definitions:[]},foreignKeys:{},compile:{mode:"manual",pending:false,diagnostics:[]},canUndo:false,canRedo:false};
    return {...value,settings:{...value.settings,build:{...value.settings.build,mode:buildMode(value.settings.build.mode)}},tree:{...value.tree,root:treeNode(value.tree.root)},schemas:{...value.schemas,definitions:(value.schemas?.definitions ?? []).map(indexedDefinition)},compile:{...value.compile,mode:buildMode(value.compile.mode),diagnostics:value.compile.diagnostics ?? []}};
}
function indexedDefinition(value: schemaregistry.IndexedDefinition) {
    return {...value,kind:schemaKind(value.kind)};
}
export function documentResult(value: {table:document.Table; workspace:workspace.View}) {
    return {...value,table:tableDocument(value.table),workspace:workspaceView(value.workspace)};
}
export function workspaceResult<T extends {workspace:workspace.View}>(value: T) {
    return {...value,workspace:workspaceView(value.workspace)};
}
export function schemaResult(value: workspace.SchemaEditResult) {
    return {...value,schema:publicSchema(value.schema),table:tableDocument(value.table),workspace:workspaceView(value.workspace)};
}
export function schemaFileResult(value: workspace.SchemaFileCreateResult) {
    return {...value,schema:publicSchema(value.schema),workspace:workspaceView(value.workspace)};
}
export function projectSettings(value: workspace.ProjectSettings): ProjectSettings {
    return {...value,formats:value.formats ?? [],tables:value.tables ?? [],workspace:workspaceView(value.workspace)};
}
