export type LogEntry = {
    time: string;
    level: "debug" | "info" | "warn" | "error";
    message: string;
    fields?: Record<string, unknown>;
};

export type TreeNode = {
    kind: "project" | "namespace" | "group" | "table" | "schemaFile" | "localSchema" | "schemaDefinition" | "damaged";
    name: string;
    path?: string;
    diagnostics?: {severity: string; message: string; line?: number; column?: number}[];
    children?: TreeNode[];
};

export type WorkspaceView = {
    project: {
        path: string;
        root: string;
    };
    settings: {
        build: {
            mode: "auto" | "manual";
            autoCompileDebounceSeconds?: number;
        };
    };
    tree: {
        root: TreeNode;
    };
    schemas: SchemaIndex;
    foreignKeys: ForeignKeySummary;
    compile: {
        mode: "auto" | "manual";
        pending: boolean;
        compiledAt?: string;
        diagnostics: {path: string; table?: string; field?: string; rowID?: string; column?: string; key?: string; args?: string[]; message: string}[];
    };
    canUndo: boolean;
    canRedo: boolean;
};

export type ProjectExportTable = {
    id: string;
    namespace: string;
    name: string;
    path: string;
    selected: boolean;
    configured: boolean;
    columns?: string[];
};

export type ProjectSettings = {
    project: {
        path: string;
        root: string;
        tableRoots: string[];
        schemaRoots: string[];
        exclude?: string[];
    };
    export: {
        directory: string;
        targets: string[];
        tables?: Record<string, boolean>;
    };
    formats: ExportFormat[];
    tables: ProjectExportTable[];
    workspace: WorkspaceView;
};

export type ProjectExportSettings = {
    directory: string;
    targets: string[];
    tables: Record<string, boolean>;
};

export type ExportFormat = {
    id: string;
    labels: Record<string, string>;
    supportsColumns: boolean;
    extensions?: string[];
};

export type TableExportSettings = {
    path: string;
    formats: Record<string, {enabled: boolean; columns?: string[]}>;
};

export type SchemaUsage = {
    definition: string;
    path: string;
    table?: string;
    field: string;
};

export type SchemaDefinition = {
    id: string;
    kind: "type" | "constraint";
    appliesTo?: string[];
    path: string;
    table?: string;
    usages?: SchemaUsage[];
    schema?: SchemaNode;
};

export type SchemaDiagnostic = {
    path: string;
    table?: string;
    field: string;
    key?: string;
    args?: string[];
    message: string;
};

export type SchemaIndex = {
    definitions: SchemaDefinition[];
    diagnostics?: SchemaDiagnostic[];
};

export type ForeignKeySummary = {
    diagnostics?: { path: string; table?: string; rowID: string; column: string; key?: string; args?: string[]; message: string }[];
};

export type SchemaRequirement = {
    kind?: string;
    op?: string;
    value?: string;
    pattern?: string;
    min?: number;
    max?: number;
    reference?: string;
    expression?: string;
};

export type SchemaNode = {
    kind: string;
    nullable?: boolean;
    default?: {expression: string};
    reference?: string;
    enumValues?: {expression: string}[];
    element?: SchemaNode;
    fields?: {name: string; optional?: boolean; schema: SchemaNode}[];
    requirements?: SchemaRequirement[];
};

export type TableDocument = {
    path: string;
    namespace: string;
    name: string;
    columns: { name: string; displayName?: string; width?: number; optional?: boolean; schema: SchemaNode; effectiveSchema?: SchemaNode; schemaExpression: string; foreignKeyTable?: string; foreignKeyOptions?: string[] }[];
    rows: { id: string; values: Record<string, { expression: string; display: string; formula?: boolean }> }[];
    diagnostics?: { rowID: string; column?: string; message: string }[];
    localSchema: SchemaDefinitionDocument[];
    exports?: Record<string, {enabled: boolean; columns?: string[]}>;
    cellComments?: Record<string, Record<string, string>>;
};

export type NavRequest = {
    kind: "schema-reference" | "foreign-key" | "formula-reference" | "primary-key-references";
    path: string;
    rowID?: string;
    column?: string;
    reference?: string;
};

export type NavTarget = {
    kind: string;
    path: string;
    local?: boolean;
    definition?: string;
    rowID?: string;
    column?: string;
    references?: NavTarget[];
};

export type SchemaDefinitionDocument = {
    name: string;
    kind: "type" | "constraint";
    type?: string;
    appliesTo?: string[];
    schema: SchemaNode;
    schemaExpression: string;
};

export type PublicSchemaDocument = {
    path: string;
    package: string;
    definitions: SchemaDefinitionDocument[];
};

export type SchemaEditorState = {
    path: string;
    local: boolean;
    package?: string;
    logicalTable?: string;
    definitions: SchemaDefinitionDocument[];
    selected: string | null;
    creating?: boolean;
};

export type SchemaEditResult = {
    schema: PublicSchemaDocument;
    table: TableDocument;
    workspace: WorkspaceView;
};

export type CellError = { rowID: string; column: string; message: string };
