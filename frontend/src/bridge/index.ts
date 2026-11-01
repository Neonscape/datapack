import {schemamodel, workspace} from "../../wailsjs/go/models";
import type {SchemaNode, ProjectExportSettings, TableExportSettings} from "../types";
import * as api from "../../wailsjs/go/main/App";
import {documentResult, projectSettings, publicSchema, schemaFileResult, schemaResult, tableDocument, workspaceResult, workspaceView} from "./models";

// Arguments remain checked against the generated binding signatures. Only DTO
// results cross into UI models here; no generated files are hand-edited.
function mapped<A extends unknown[], D, U>(call: (...args:A) => Promise<D>, adapt:(value:D) => U) {
    return (...args:A):Promise<U> => call(...args).then(adapt);
}
export const GetTable=mapped(api.GetTable,tableDocument);
export const GetPublicSchema=mapped(api.GetPublicSchema,publicSchema);
export const NewProject=mapped(api.NewProject,workspaceView);
export const OpenProject=mapped(api.OpenProject,workspaceView);
export const SelectProject=mapped(api.SelectProject,workspaceView);
export const SaveProjectAs=mapped(api.SaveProjectAs,workspaceView);
export const SetCompileMode=mapped(api.SetCompileMode,workspaceView);
export const GetProjectSettings=mapped(api.GetProjectSettings,projectSettings);
export const SetProjectExportSettings=(value: ProjectExportSettings) => api.SetProjectExportSettings(new workspace.ProjectExportSettings(value)).then(projectSettings);
export const Compile=mapped(api.Compile,workspaceResult);
export const DeleteTable=mapped(api.DeleteTable,workspaceResult);
export const CreateSchemaFile=mapped(api.CreateSchemaFile,schemaFileResult);
export const CreateSchemaDefinition=mapped(api.CreateSchemaDefinition,schemaResult);
export const SetSchemaDefinition=mapped(api.SetSchemaDefinition,schemaResult);
export const RenameSchemaDefinition=mapped(api.RenameSchemaDefinition,schemaResult);
export const DeleteSchemaDefinition=mapped(api.DeleteSchemaDefinition,schemaResult);
export const SetDisplayName=mapped(api.SetDisplayName,schemaResult);
export const AddColumn=mapped(api.AddColumn,documentResult);
export const ClearCell=mapped(api.ClearCell,documentResult);
export const CreateRow=mapped(api.CreateRow,documentResult);
export const CreateTable=mapped(api.CreateTable,documentResult);
export const DeleteColumn=mapped(api.DeleteColumn,documentResult);
export const DeleteRow=mapped(api.DeleteRow,documentResult);
export const MoveColumn=mapped(api.MoveColumn,documentResult);
export const Paste=mapped(api.Paste,documentResult);
export const Redo=mapped(api.Redo,documentResult);
export const RenameColumn=mapped(api.RenameColumn,documentResult);
export const RenameRow=mapped(api.RenameRow,documentResult);
export const RenameTable=mapped(api.RenameTable,documentResult);
export const SetCellComment=mapped(api.SetCellComment,documentResult);
export const SetCellFormula=mapped(api.SetCellFormula,documentResult);
export const SetCellText=mapped(api.SetCellText,documentResult);
export const SetColumnDisplayName=mapped(api.SetColumnDisplayName,documentResult);
export const SetColumnSchema=mapped(api.SetColumnSchema,documentResult);
export const SetColumnWidth=mapped(api.SetColumnWidth,documentResult);
export const SetTableExportSettings=(value: TableExportSettings) => api.SetTableExportSettings(new workspace.TableExportSettings(value)).then(documentResult);
export const Undo=mapped(api.Undo,documentResult);
export const UpdateCell=mapped(api.UpdateCell,documentResult);
export const PrintSchema=(value: SchemaNode) => api.PrintSchema(new schemamodel.Node(value));
export {Export, GetExportFormats, GetSettings, LogPath, Navigate, OpenPath, SelectProjectExportDirectory, SetLanguage, SetTheme} from "../../wailsjs/go/main/App";
