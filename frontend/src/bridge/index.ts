import {orderedCall} from "./ordered";
import {schemamodel, workspace} from "../../wailsjs/go/models";
import type {CellDelta, SchemaNode, ProjectExportSettings, TableExportSettings} from "../types";
import * as api from "../../wailsjs/go/main/App";
import {documentResult, projectSettings, publicSchema, schemaFileResult, schemaResult, tableDocument, workspaceResult, workspaceView} from "./models";

// Arguments remain checked against the generated binding signatures. Only DTO
// results cross into UI models here; no generated files are hand-edited.
function mapped<A extends unknown[], D, U>(name:string, call: (...args:A) => Promise<D>, adapt:(value:D) => U) {
    return (...args:A):Promise<U> => orderedCall(name, () => call(...args)).then(adapt);
}
export const GetTable=mapped("GetTable",api.GetTable,tableDocument);
export const GetPublicSchema=mapped("GetPublicSchema",api.GetPublicSchema,publicSchema);
export const NewProject=mapped("NewProject",api.NewProject,workspaceView);
export const OpenProject=mapped("OpenProject",api.OpenProject,workspaceView);
export const SelectProject=mapped("SelectProject",api.SelectProject,workspaceView);
export const SaveProjectAs=mapped("SaveProjectAs",api.SaveProjectAs,workspaceView);
export const SetCompileMode=mapped("SetCompileMode",api.SetCompileMode,workspaceView);
export const GetProjectSettings=mapped("GetProjectSettings",api.GetProjectSettings,projectSettings);
export const SetProjectExportSettings=(value: ProjectExportSettings) => orderedCall("SetProjectExportSettings", () => api.SetProjectExportSettings(new workspace.ProjectExportSettings(value))).then(projectSettings);
export const Compile=mapped("Compile",api.Compile,workspaceResult);
export const DeleteTable=mapped("DeleteTable",api.DeleteTable,workspaceResult);
export const CreateSchemaFile=mapped("CreateSchemaFile",api.CreateSchemaFile,schemaFileResult);
export const CreateSchemaDefinition=mapped("CreateSchemaDefinition",api.CreateSchemaDefinition,schemaResult);
export const SetSchemaDefinition=mapped("SetSchemaDefinition",api.SetSchemaDefinition,schemaResult);
export const RenameSchemaDefinition=mapped("RenameSchemaDefinition",api.RenameSchemaDefinition,schemaResult);
export const DeleteSchemaDefinition=mapped("DeleteSchemaDefinition",api.DeleteSchemaDefinition,schemaResult);
export const SetDisplayName=mapped("SetDisplayName",api.SetDisplayName,schemaResult);
export const AddColumn=mapped("AddColumn",api.AddColumn,documentResult);
export const ClearCell=mapped("ClearCell",api.ClearCell,documentResult);
export const CreateRow=mapped("CreateRow",api.CreateRow,documentResult);
export const CreateTable=mapped("CreateTable",api.CreateTable,documentResult);
export const DeleteColumn=mapped("DeleteColumn",api.DeleteColumn,documentResult);
export const DeleteRow=mapped("DeleteRow",api.DeleteRow,documentResult);
export const MoveColumn=mapped("MoveColumn",api.MoveColumn,documentResult);
export const Paste=mapped("Paste",api.Paste,documentResult);
export const Redo=historyCall("Redo");
export const RenameColumn=mapped("RenameColumn",api.RenameColumn,documentResult);
export const RenameRow=mapped("RenameRow",api.RenameRow,documentResult);
export const RenameTable=mapped("RenameTable",api.RenameTable,documentResult);
export const SetCellComment=mapped("SetCellComment",api.SetCellComment,documentResult);
export const SetCellFormula=mapped("SetCellFormula",api.SetCellFormula,documentResult);
export const SetCellText=cellCall("SetCellText");
export const SetColumnDisplayName=mapped("SetColumnDisplayName",api.SetColumnDisplayName,documentResult);
export const SetColumnSchema=mapped("SetColumnSchema",api.SetColumnSchema,documentResult);
export const SetColumnWidth=mapped("SetColumnWidth",api.SetColumnWidth,documentResult);
export const SetTableExportSettings=(value: TableExportSettings) => orderedCall("SetTableExportSettings", () => api.SetTableExportSettings(new workspace.TableExportSettings(value))).then(documentResult);
export const Undo=historyCall("Undo");
export const UpdateCell=cellCall("UpdateCell");
export const PrintSchema=(value: SchemaNode) => api.PrintSchema(new schemamodel.Node(value));
export const Export=() => orderedCall("Export", () => api.Export()).then(value=>({...value,workspace:value.workspace?.project?.path ? workspaceView(value.workspace) : undefined}));
export {GetExportFormats, GetSettings, LogPath, Navigate, OpenPath, SelectProjectExportDirectory, SetLanguage, SetTheme} from "../../wailsjs/go/main/App";

// Older test/runtime bridges retain their full-document contract. The native
// current bridge returns deltas only after the safe literal transaction commits.
function hasMethod(name:string) {return typeof (window as unknown as {go?:{main?:{App?:Record<string,unknown>}}}).go?.main?.App?.[name] === "function";}
function compactResult(value:workspace.CellEditResult) {return {...documentResult({table:value.table!,workspace:value.workspace}),delta:value.delta as CellDelta | undefined};}
function cellCall(name:"UpdateCell"|"SetCellText") {
    return (update:workspace.CellUpdate) => orderedCall(name, async()=>{
        if(hasMethod(name+"Delta"))return compactResult(await (name === "UpdateCell" ? api.UpdateCellDelta(update) : api.SetCellTextDelta(update)));
        return {...documentResult(await api[name](update)),delta:undefined};
    });
}
function historyCall(name:"Undo"|"Redo") {
    return () => orderedCall(name, async()=>{
        if(hasMethod(name+"Delta"))return compactResult(await (name === "Undo" ? api.UndoDelta() : api.RedoDelta()));
        return {...documentResult(await api[name]()),delta:undefined};
    });
}
