import type {PublicSchemaDocument, TableDocument} from "../types";
import {refreshSchemaState, type OpenTab} from "./sessionState";

// The session owns request/result ordering; its reader only supplies documents.
// Keeping Wails outside this contract also makes delayed and failed reads real
// inputs to session tests rather than special cases in application handlers.
export type TabReader = {table(path: string): Promise<TableDocument>; schema(path: string): Promise<PublicSchemaDocument>};

export async function readTab(tab: OpenTab, reader: TabReader): Promise<OpenTab> {
    if (tab.kind === "table") return {...tab,table:await reader.table(tab.path)};
    if (tab.state.local) {
        const document = await reader.table(tab.state.path);
        const logical = document.namespace ? `${document.namespace}.${document.name}` : document.name;
        return {...tab,state:refreshSchemaState(tab.state,document.localSchema,{logicalTable:logical,displayName:document.displayName})};
    }
    const document = await reader.schema(tab.state.path);
    return {...tab,state:refreshSchemaState(tab.state,document.definitions ?? [],{package:document.package,displayName:document.displayName})};
}
