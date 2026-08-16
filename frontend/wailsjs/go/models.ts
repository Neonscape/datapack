export namespace compiler {
	
	export class Diagnostic {
	    path: string;
	    table?: string;
	    field?: string;
	    rowID?: string;
	    column?: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new Diagnostic(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.table = source["table"];
	        this.field = source["field"];
	        this.rowID = source["rowID"];
	        this.column = source["column"];
	        this.message = source["message"];
	    }
	}

}

export namespace cuesource {
	
	export class Diagnostic {
	    severity: string;
	    message: string;
	    line?: number;
	    column?: number;
	
	    static createFrom(source: any = {}) {
	        return new Diagnostic(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.severity = source["severity"];
	        this.message = source["message"];
	        this.line = source["line"];
	        this.column = source["column"];
	    }
	}

}

export namespace document {
	
	export class CSVExport {
	    columns: string[];
	
	    static createFrom(source: any = {}) {
	        return new CSVExport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.columns = source["columns"];
	    }
	}
	export class Cell {
	    expression: string;
	    display: string;
	    formula?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Cell(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.expression = source["expression"];
	        this.display = source["display"];
	        this.formula = source["formula"];
	    }
	}
	export class Column {
	    name: string;
	    optional: boolean;
	    schema: schemamodel.Node;
	    schemaExpression: string;
	    foreignKeyTable?: string;
	    foreignKeyOptions?: string[];
	
	    static createFrom(source: any = {}) {
	        return new Column(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.optional = source["optional"];
	        this.schema = this.convertValues(source["schema"], schemamodel.Node);
	        this.schemaExpression = source["schemaExpression"];
	        this.foreignKeyTable = source["foreignKeyTable"];
	        this.foreignKeyOptions = source["foreignKeyOptions"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Diagnostic {
	    rowID: string;
	    column?: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new Diagnostic(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rowID = source["rowID"];
	        this.column = source["column"];
	        this.message = source["message"];
	    }
	}
	export class ForeignKey {
	    column: string;
	    table: string;
	
	    static createFrom(source: any = {}) {
	        return new ForeignKey(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.column = source["column"];
	        this.table = source["table"];
	    }
	}
	export class Formula {
	    rowID: string;
	    column: string;
	    expression: string;
	
	    static createFrom(source: any = {}) {
	        return new Formula(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rowID = source["rowID"];
	        this.column = source["column"];
	        this.expression = source["expression"];
	    }
	}
	export class SchemaDefinition {
	    name: string;
	    kind: string;
	    type?: string;
	    appliesTo?: string[];
	    schema: schemamodel.Node;
	    schemaExpression: string;
	
	    static createFrom(source: any = {}) {
	        return new SchemaDefinition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.kind = source["kind"];
	        this.type = source["type"];
	        this.appliesTo = source["appliesTo"];
	        this.schema = this.convertValues(source["schema"], schemamodel.Node);
	        this.schemaExpression = source["schemaExpression"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PublicSchema {
	    path: string;
	    package: string;
	    definitions: SchemaDefinition[];
	
	    static createFrom(source: any = {}) {
	        return new PublicSchema(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.package = source["package"];
	        this.definitions = this.convertValues(source["definitions"], SchemaDefinition);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Row {
	    id: string;
	    values: Record<string, Cell>;
	
	    static createFrom(source: any = {}) {
	        return new Row(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.values = this.convertValues(source["values"], Cell, true);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Table {
	    path: string;
	    package: string;
	    namespace: string;
	    name: string;
	    primaryKey: schemamodel.Node;
	    columns: Column[];
	    rows: Row[];
	    foreignKeys?: ForeignKey[];
	    formulas?: Formula[];
	    diagnostics?: Diagnostic[];
	    csvExport?: CSVExport;
	    localSchema: SchemaDefinition[];
	
	    static createFrom(source: any = {}) {
	        return new Table(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.package = source["package"];
	        this.namespace = source["namespace"];
	        this.name = source["name"];
	        this.primaryKey = this.convertValues(source["primaryKey"], schemamodel.Node);
	        this.columns = this.convertValues(source["columns"], Column);
	        this.rows = this.convertValues(source["rows"], Row);
	        this.foreignKeys = this.convertValues(source["foreignKeys"], ForeignKey);
	        this.formulas = this.convertValues(source["formulas"], Formula);
	        this.diagnostics = this.convertValues(source["diagnostics"], Diagnostic);
	        this.csvExport = this.convertValues(source["csvExport"], CSVExport);
	        this.localSchema = this.convertValues(source["localSchema"], SchemaDefinition);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace foreignkey {
	
	export class Diagnostic {
	    path: string;
	    table: string;
	    rowID: string;
	    column: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new Diagnostic(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.table = source["table"];
	        this.rowID = source["rowID"];
	        this.column = source["column"];
	        this.message = source["message"];
	    }
	}
	export class Summary {
	    diagnostics?: Diagnostic[];
	
	    static createFrom(source: any = {}) {
	        return new Summary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.diagnostics = this.convertValues(source["diagnostics"], Diagnostic);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace navigation {
	
	export class Request {
	    kind: string;
	    path: string;
	    rowID?: string;
	    column?: string;
	    reference?: string;
	
	    static createFrom(source: any = {}) {
	        return new Request(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.path = source["path"];
	        this.rowID = source["rowID"];
	        this.column = source["column"];
	        this.reference = source["reference"];
	    }
	}
	export class Target {
	    kind: string;
	    path: string;
	    local?: boolean;
	    definition?: string;
	    rowID?: string;
	    column?: string;
	    references?: Target[];
	
	    static createFrom(source: any = {}) {
	        return new Target(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.path = source["path"];
	        this.local = source["local"];
	        this.definition = source["definition"];
	        this.rowID = source["rowID"];
	        this.column = source["column"];
	        this.references = this.convertValues(source["references"], Target);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace projecttree {
	
	export class Node {
	    kind: string;
	    name: string;
	    path?: string;
	    diagnostics?: cuesource.Diagnostic[];
	    children?: Node[];
	
	    static createFrom(source: any = {}) {
	        return new Node(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.diagnostics = this.convertValues(source["diagnostics"], cuesource.Diagnostic);
	        this.children = this.convertValues(source["children"], Node);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Tree {
	    root: Node;
	
	    static createFrom(source: any = {}) {
	        return new Tree(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = this.convertValues(source["root"], Node);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace schemamodel {
	
	export class Requirement {
	    kind?: string;
	    op?: string;
	    value?: string;
	    pattern?: string;
	    min?: number;
	    max?: number;
	    reference?: string;
	    expression?: string;
	
	    static createFrom(source: any = {}) {
	        return new Requirement(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.op = source["op"];
	        this.value = source["value"];
	        this.pattern = source["pattern"];
	        this.min = source["min"];
	        this.max = source["max"];
	        this.reference = source["reference"];
	        this.expression = source["expression"];
	    }
	}
	export class Value {
	    expression: string;
	
	    static createFrom(source: any = {}) {
	        return new Value(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.expression = source["expression"];
	    }
	}
	export class Node {
	    kind: string;
	    nullable: boolean;
	    default?: Value;
	    reference?: string;
	    enumValues?: Value[];
	    element?: Node;
	    fields?: Field[];
	    requirements?: Requirement[];
	
	    static createFrom(source: any = {}) {
	        return new Node(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.nullable = source["nullable"];
	        this.default = this.convertValues(source["default"], Value);
	        this.reference = source["reference"];
	        this.enumValues = this.convertValues(source["enumValues"], Value);
	        this.element = this.convertValues(source["element"], Node);
	        this.fields = this.convertValues(source["fields"], Field);
	        this.requirements = this.convertValues(source["requirements"], Requirement);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Field {
	    name: string;
	    optional: boolean;
	    schema: Node;
	
	    static createFrom(source: any = {}) {
	        return new Field(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.optional = source["optional"];
	        this.schema = this.convertValues(source["schema"], Node);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	

}

export namespace schemaregistry {
	
	export class Diagnostic {
	    path: string;
	    table?: string;
	    field: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new Diagnostic(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.table = source["table"];
	        this.field = source["field"];
	        this.message = source["message"];
	    }
	}
	export class Usage {
	    definition: string;
	    path: string;
	    table?: string;
	    field: string;
	
	    static createFrom(source: any = {}) {
	        return new Usage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.definition = source["definition"];
	        this.path = source["path"];
	        this.table = source["table"];
	        this.field = source["field"];
	    }
	}
	export class IndexedDefinition {
	    id: string;
	    kind: string;
	    appliesTo?: string[];
	    path: string;
	    table?: string;
	    usages?: Usage[];
	
	    static createFrom(source: any = {}) {
	        return new IndexedDefinition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.kind = source["kind"];
	        this.appliesTo = source["appliesTo"];
	        this.path = source["path"];
	        this.table = source["table"];
	        this.usages = this.convertValues(source["usages"], Usage);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Index {
	    definitions: IndexedDefinition[];
	    diagnostics?: Diagnostic[];
	
	    static createFrom(source: any = {}) {
	        return new Index(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.definitions = this.convertValues(source["definitions"], IndexedDefinition);
	        this.diagnostics = this.convertValues(source["diagnostics"], Diagnostic);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

export namespace workspace {
	
	export class BuildSettings {
	    mode: string;
	    autoCompileDebounceSeconds?: number;
	
	    static createFrom(source: any = {}) {
	        return new BuildSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.autoCompileDebounceSeconds = source["autoCompileDebounceSeconds"];
	    }
	}
	export class CellUpdate {
	    path: string;
	    rowID: string;
	    column: string;
	    expression: string;
	
	    static createFrom(source: any = {}) {
	        return new CellUpdate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.rowID = source["rowID"];
	        this.column = source["column"];
	        this.expression = source["expression"];
	    }
	}
	export class CompileState {
	    mode: string;
	    pending: boolean;
	    compiledAt?: string;
	    diagnostics: compiler.Diagnostic[];
	
	    static createFrom(source: any = {}) {
	        return new CompileState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.pending = source["pending"];
	        this.compiledAt = source["compiledAt"];
	        this.diagnostics = this.convertValues(source["diagnostics"], compiler.Diagnostic);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class EditorSettings {
	    version: number;
	    build: BuildSettings;
	
	    static createFrom(source: any = {}) {
	        return new EditorSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.build = this.convertValues(source["build"], BuildSettings);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ExportConfig {
	    directory: string;
	    targets: string[];
	
	    static createFrom(source: any = {}) {
	        return new ExportConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.directory = source["directory"];
	        this.targets = source["targets"];
	    }
	}
	export class ProjectConfig {
	    version: number;
	    tableRoots: string[];
	    schemaRoots: string[];
	    exclude?: string[];
	    export: ExportConfig;
	
	    static createFrom(source: any = {}) {
	        return new ProjectConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.tableRoots = source["tableRoots"];
	        this.schemaRoots = source["schemaRoots"];
	        this.exclude = source["exclude"];
	        this.export = this.convertValues(source["export"], ExportConfig);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Project {
	    path: string;
	    root: string;
	    config: ProjectConfig;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.root = source["root"];
	        this.config = this.convertValues(source["config"], ProjectConfig);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class View {
	    project: Project;
	    settings: EditorSettings;
	    tree: projecttree.Tree;
	    schemas: schemaregistry.Index;
	    foreignKeys: foreignkey.Summary;
	    compile: CompileState;
	
	    static createFrom(source: any = {}) {
	        return new View(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project = this.convertValues(source["project"], Project);
	        this.settings = this.convertValues(source["settings"], EditorSettings);
	        this.tree = this.convertValues(source["tree"], projecttree.Tree);
	        this.schemas = this.convertValues(source["schemas"], schemaregistry.Index);
	        this.foreignKeys = this.convertValues(source["foreignKeys"], foreignkey.Summary);
	        this.compile = this.convertValues(source["compile"], CompileState);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CellUpdateResult {
	    table: document.Table;
	    workspace: View;
	
	    static createFrom(source: any = {}) {
	        return new CellUpdateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table = this.convertValues(source["table"], document.Table);
	        this.workspace = this.convertValues(source["workspace"], View);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ColumnAdd {
	    path: string;
	    name: string;
	    schemaExpression: string;
	    initialExpression: string;
	    position?: number;
	
	    static createFrom(source: any = {}) {
	        return new ColumnAdd(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.schemaExpression = source["schemaExpression"];
	        this.initialExpression = source["initialExpression"];
	        this.position = source["position"];
	    }
	}
	export class ColumnAddResult {
	    table: document.Table;
	    workspace: View;
	
	    static createFrom(source: any = {}) {
	        return new ColumnAddResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table = this.convertValues(source["table"], document.Table);
	        this.workspace = this.convertValues(source["workspace"], View);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ColumnDelete {
	    path: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new ColumnDelete(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	    }
	}
	export class ColumnDeleteResult {
	    table: document.Table;
	    workspace: View;
	
	    static createFrom(source: any = {}) {
	        return new ColumnDeleteResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table = this.convertValues(source["table"], document.Table);
	        this.workspace = this.convertValues(source["workspace"], View);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ColumnMove {
	    path: string;
	    name: string;
	    position: number;
	
	    static createFrom(source: any = {}) {
	        return new ColumnMove(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.position = source["position"];
	    }
	}
	export class ColumnMoveResult {
	    table: document.Table;
	    workspace: View;
	
	    static createFrom(source: any = {}) {
	        return new ColumnMoveResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table = this.convertValues(source["table"], document.Table);
	        this.workspace = this.convertValues(source["workspace"], View);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ColumnRename {
	    path: string;
	    oldName: string;
	    newName: string;
	
	    static createFrom(source: any = {}) {
	        return new ColumnRename(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.oldName = source["oldName"];
	        this.newName = source["newName"];
	    }
	}
	export class ColumnRenameResult {
	    table: document.Table;
	    workspace: View;
	
	    static createFrom(source: any = {}) {
	        return new ColumnRenameResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table = this.convertValues(source["table"], document.Table);
	        this.workspace = this.convertValues(source["workspace"], View);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ColumnSchemaSet {
	    path: string;
	    name: string;
	    schemaExpression: string;
	
	    static createFrom(source: any = {}) {
	        return new ColumnSchemaSet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.schemaExpression = source["schemaExpression"];
	    }
	}
	export class ColumnSchemaSetResult {
	    table: document.Table;
	    workspace: View;
	
	    static createFrom(source: any = {}) {
	        return new ColumnSchemaSetResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table = this.convertValues(source["table"], document.Table);
	        this.workspace = this.convertValues(source["workspace"], View);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CompileResult {
	    workspace: View;
	
	    static createFrom(source: any = {}) {
	        return new CompileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workspace = this.convertValues(source["workspace"], View);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class ExportResult {
	    files: string[];
	
	    static createFrom(source: any = {}) {
	        return new ExportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.files = source["files"];
	    }
	}
	export class FormulaSet {
	    path: string;
	    rowID: string;
	    column: string;
	    expression: string;
	
	    static createFrom(source: any = {}) {
	        return new FormulaSet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.rowID = source["rowID"];
	        this.column = source["column"];
	        this.expression = source["expression"];
	    }
	}
	export class FormulaSetResult {
	    table: document.Table;
	    workspace: View;
	
	    static createFrom(source: any = {}) {
	        return new FormulaSetResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table = this.convertValues(source["table"], document.Table);
	        this.workspace = this.convertValues(source["workspace"], View);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class HistoryResult {
	    table: document.Table;
	    workspace: View;
	    canUndo: boolean;
	    canRedo: boolean;
	
	    static createFrom(source: any = {}) {
	        return new HistoryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table = this.convertValues(source["table"], document.Table);
	        this.workspace = this.convertValues(source["workspace"], View);
	        this.canUndo = source["canUndo"];
	        this.canRedo = source["canRedo"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Paste {
	    path: string;
	    rowIDs: string[];
	    columns: string[];
	    values: string[][];
	
	    static createFrom(source: any = {}) {
	        return new Paste(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.rowIDs = source["rowIDs"];
	        this.columns = source["columns"];
	        this.values = source["values"];
	    }
	}
	
	
	export class RowAdd {
	    path: string;
	    rowID: string;
	    expression: string;
	
	    static createFrom(source: any = {}) {
	        return new RowAdd(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.rowID = source["rowID"];
	        this.expression = source["expression"];
	    }
	}
	export class RowAddResult {
	    table: document.Table;
	    workspace: View;
	
	    static createFrom(source: any = {}) {
	        return new RowAddResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table = this.convertValues(source["table"], document.Table);
	        this.workspace = this.convertValues(source["workspace"], View);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RowCreate {
	    path: string;
	    rowID: string;
	    values: Record<string, string>;
	    formulas: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new RowCreate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.rowID = source["rowID"];
	        this.values = source["values"];
	        this.formulas = source["formulas"];
	    }
	}
	export class RowCreateResult {
	    table: document.Table;
	    workspace: View;
	
	    static createFrom(source: any = {}) {
	        return new RowCreateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table = this.convertValues(source["table"], document.Table);
	        this.workspace = this.convertValues(source["workspace"], View);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RowDelete {
	    path: string;
	    rowID: string;
	
	    static createFrom(source: any = {}) {
	        return new RowDelete(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.rowID = source["rowID"];
	    }
	}
	export class RowDeleteResult {
	    table: document.Table;
	    workspace: View;
	
	    static createFrom(source: any = {}) {
	        return new RowDeleteResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table = this.convertValues(source["table"], document.Table);
	        this.workspace = this.convertValues(source["workspace"], View);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RowRename {
	    path: string;
	    oldID: string;
	    newID: string;
	
	    static createFrom(source: any = {}) {
	        return new RowRename(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.oldID = source["oldID"];
	        this.newID = source["newID"];
	    }
	}
	export class RowRenameResult {
	    table: document.Table;
	    workspace: View;
	
	    static createFrom(source: any = {}) {
	        return new RowRenameResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table = this.convertValues(source["table"], document.Table);
	        this.workspace = this.convertValues(source["workspace"], View);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SchemaDefinitionCreate {
	    path: string;
	    name: string;
	    kind: string;
	    type?: string;
	    appliesTo?: string[];
	    schemaExpression: string;
	
	    static createFrom(source: any = {}) {
	        return new SchemaDefinitionCreate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.kind = source["kind"];
	        this.type = source["type"];
	        this.appliesTo = source["appliesTo"];
	        this.schemaExpression = source["schemaExpression"];
	    }
	}
	export class SchemaDefinitionDelete {
	    path: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new SchemaDefinitionDelete(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	    }
	}
	export class SchemaDefinitionRename {
	    path: string;
	    oldName: string;
	    newName: string;
	
	    static createFrom(source: any = {}) {
	        return new SchemaDefinitionRename(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.oldName = source["oldName"];
	        this.newName = source["newName"];
	    }
	}
	export class SchemaDefinitionSet {
	    path: string;
	    name: string;
	    kind: string;
	    type?: string;
	    appliesTo?: string[];
	    schemaExpression: string;
	
	    static createFrom(source: any = {}) {
	        return new SchemaDefinitionSet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.kind = source["kind"];
	        this.type = source["type"];
	        this.appliesTo = source["appliesTo"];
	        this.schemaExpression = source["schemaExpression"];
	    }
	}
	export class SchemaEditResult {
	    schema: document.PublicSchema;
	    table: document.Table;
	    workspace: View;
	
	    static createFrom(source: any = {}) {
	        return new SchemaEditResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schema = this.convertValues(source["schema"], document.PublicSchema);
	        this.table = this.convertValues(source["table"], document.Table);
	        this.workspace = this.convertValues(source["workspace"], View);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TableCreate {
	    namespace: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new TableCreate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.namespace = source["namespace"];
	        this.name = source["name"];
	    }
	}
	export class TableCreateResult {
	    table: document.Table;
	    workspace: View;
	
	    static createFrom(source: any = {}) {
	        return new TableCreateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table = this.convertValues(source["table"], document.Table);
	        this.workspace = this.convertValues(source["workspace"], View);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

