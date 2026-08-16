import type {SchemaNode, SchemaRequirement, WorkspaceView} from "../types";
import {schemaConstraintReferenceOptions, schemaTypeReferenceOptions} from "./refs";

function predicateOptions(kind: string): {kind: string; label: string}[] {
    if (kind === "string") {
        return [
            {kind: "compare", label: "Compare"},
            {kind: "regex", label: "Regex"},
            {kind: "nonEmpty", label: "Non-empty"},
            {kind: "length", label: "Length"},
            {kind: "reference", label: "Reference"},
            {kind: "custom", label: "Custom"},
        ];
    }
    if (kind === "integer" || kind === "number") {
        return [
            {kind: "compare", label: "Compare"},
            {kind: "reference", label: "Reference"},
            {kind: "custom", label: "Custom"},
        ];
    }
    return [
        {kind: "reference", label: "Reference"},
        {kind: "custom", label: "Custom"},
    ];
}

const predicateLabel = (kind: string) => ({compare: "Compare", regex: "Regex", nonEmpty: "Non-empty", length: "Length", reference: "Reference", custom: "Custom"} as Record<string, string>)[kind] ?? kind;

export function predicateKindsFor(types: string[]): string[] {
    const kinds: string[] = [];
    for (const type of types) {
        for (const option of predicateOptions(type)) {
            if (!kinds.includes(option.kind)) kinds.push(option.kind);
        }
    }
    return kinds;
}

export function schemaNodeWithKind(value: SchemaNode, next: string, typeReferences: string[]): SchemaNode {
    const base: SchemaNode = {kind: next, nullable: value.nullable, default: value.default};
    if (next === "enum") base.enumValues = value.kind === "enum" ? value.enumValues : [];
    if (next === "array") base.element = value.kind === "array" ? value.element : {kind: "string"};
    if (next === "struct") base.fields = value.kind === "struct" ? value.fields : [];
    if (next === "reference") base.reference = value.kind === "reference" ? value.reference : (typeReferences[0] ?? "");
    if (next !== "enum") base.requirements = value.requirements;
    return base;
}

function newRequirement(kind: string): SchemaRequirement {
    switch (kind) {
        case "compare": return {kind, op: ">=", value: ""};
        case "regex": return {kind, op: "=~", pattern: ""};
        case "nonEmpty": return {kind};
        case "length": return {kind, min: 1};
        case "reference": return {kind, reference: ""};
        default: return {kind: "custom", expression: ""};
    }
}

function requirementError(requirement: SchemaRequirement): string {
    switch (requirement.kind) {
        case "compare":
            return requirement.value?.trim() ? "" : "a value is required";
        case "regex":
            if (!requirement.pattern) return "a regex pattern is required";
            try {
                new RegExp(requirement.pattern);
                return "";
            } catch {
                return "invalid regex";
            }
        case "length":
            if (requirement.min != null && (!Number.isInteger(requirement.min) || requirement.min < 0)) return "min length must be a non-negative integer";
            if (requirement.max != null && (!Number.isInteger(requirement.max) || requirement.max < 0)) return "max length must be a non-negative integer";
            return "";
        case "reference":
            return requirement.reference ? "" : "a constraint Schema is required";
        case "custom":
            return requirement.expression?.trim() ? "" : "an expression is required";
        default:
            return "";
    }
}

function RequirementInput({requirement, baseType, constraintOptions, onChange}: {requirement: SchemaRequirement; baseType: string; constraintOptions: {label: string; value: string}[]; onChange: (requirement: SchemaRequirement) => void}) {
    if (requirement.kind === "compare") {
        return <span className="schema-req-input">
            <select aria-label="Comparison operator" value={requirement.op} onChange={(event) => onChange({...requirement, op: event.target.value})}>
                {[">=", ">", "<=", "<", "==", "!="].map((op) => <option key={op} value={op}>{op}</option>)}
            </select>
            <input aria-label="Comparison value" value={requirement.value ?? ""} placeholder={baseType === "string" ? `"value"` : "0"} onChange={(event) => onChange({...requirement, value: event.target.value})}/>
        </span>;
    }
    if (requirement.kind === "regex") {
        return <input aria-label="Regex pattern" value={requirement.pattern ?? ""} placeholder="^[a-z]+$" onChange={(event) => onChange({...requirement, pattern: event.target.value})}/>;
    }
    if (requirement.kind === "nonEmpty") {
        return <input aria-label="Non-empty" disabled placeholder="Non-empty (no input)"/>;
    }
    if (requirement.kind === "length") {
        const mode = requirement.max != null ? "max" : "min";
        return <span className="schema-req-input">
            <select aria-label="Length bound" value={mode} onChange={(event) => onChange(event.target.value === "max" ? {...requirement, min: undefined, max: 1} : {...requirement, min: 1, max: undefined})}>
                <option value="min">Min length</option>
                <option value="max">Max length</option>
            </select>
            <input aria-label="Length value" type="number" min="0" value={mode === "max" ? (requirement.max ?? 1) : (requirement.min ?? 1)} onChange={(event) => onChange(mode === "max" ? {...requirement, max: Number(event.target.value)} : {...requirement, min: Number(event.target.value)})}/>
        </span>;
    }
    if (requirement.kind === "reference") {
        return <select aria-label="Constraint reference" value={requirement.reference ?? ""} onChange={(event) => onChange({...requirement, reference: event.target.value})}>
            <option value="">Choose a constraint Schema</option>
            {constraintOptions.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
        </select>;
    }
    return <input aria-label="Custom expression" value={requirement.expression ?? ""} placeholder=">=0" onChange={(event) => onChange({...requirement, expression: event.target.value})}/>;
}

export function SchemaForm({value, onChange, workspace, fixedKind, predicateKinds}: {value: SchemaNode; onChange: (node: SchemaNode) => void; workspace: WorkspaceView | null; fixedKind?: string; predicateKinds?: string[]}) {
    const kind = fixedKind ?? value.kind ?? "unknown";
    const typeReferences = schemaTypeReferenceOptions(workspace);
    const constraintReferences = schemaConstraintReferenceOptions(workspace);
    const options = predicateKinds ? predicateKinds.map((entry) => ({kind: entry, label: predicateLabel(entry)})) : predicateOptions(kind);

    const setKind = (next: string) => {
        onChange(schemaNodeWithKind(value, next, typeReferences.map((entry) => entry.value)));
    };

    const updateRequirement = (index: number, requirement: SchemaRequirement) => {
        onChange({...value, requirements: (value.requirements ?? []).map((entry, position) => position === index ? requirement : entry)});
    };
    const addRequirement = () => {
        onChange({...value, requirements: [...(value.requirements ?? []), newRequirement(options[0].kind)]});
    };
    const removeRequirement = (index: number) => {
        onChange({...value, requirements: (value.requirements ?? []).filter((_, position) => position !== index)});
    };

    const requirements = value.requirements ?? [];

    return <div className="schema-form">
        {fixedKind === undefined && <label className="schema-form-row">Base type
            <select aria-label="Base type" value={kind} onChange={(event) => setKind(event.target.value)}>
                {["string", "integer", "number", "boolean", "enum", "array", "struct", "reference"].map((option) => <option key={option} value={option}>{option}</option>)}
            </select>
        </label>}
        {kind === "reference" && <label className="schema-form-row">Type Schema
            <select aria-label="Type Schema" value={value.reference ?? ""} onChange={(event) => onChange({...value, reference: event.target.value})}>
                <option value="">Choose a type Schema</option>
                {typeReferences.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
            </select>
        </label>}
        {kind === "enum" && <div className="schema-form-block">
            <div className="schema-form-block-title">Enum values</div>
            {(value.enumValues ?? []).map((enumValue, index) => <div className="schema-enum-row" key={index}>
                <input value={enumValue.expression} onChange={(event) => onChange({...value, enumValues: (value.enumValues ?? []).map((entry, position) => position === index ? {expression: event.target.value} : entry)})}/>
                <button aria-label={`Delete enum value ${index + 1}`} onClick={() => onChange({...value, enumValues: (value.enumValues ?? []).filter((_, position) => position !== index)})}>×</button>
            </div>)}
            <button type="button" className="schema-form-add" onClick={() => onChange({...value, enumValues: [...(value.enumValues ?? []), {expression: `""`}]})}>Add enum value</button>
        </div>}
        {kind === "array" && <label className="schema-form-row">Element type
            <select aria-label="Element type" value={value.element?.kind ?? "string"} onChange={(event) => onChange({...value, element: {kind: event.target.value}})}>
                {["string", "integer", "number", "boolean", "reference"].map((option) => <option key={option} value={option}>{option}</option>)}
            </select>
        </label>}
        {kind === "array" && value.element?.kind === "reference" && <label className="schema-form-row">Element Schema
            <select aria-label="Element Schema" value={value.element.reference ?? ""} onChange={(event) => onChange({...value, element: {...value.element!, reference: event.target.value}})}>
                <option value="">Choose a type Schema</option>
                {typeReferences.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
            </select>
        </label>}
        {kind === "struct" && <div className="schema-form-block">
            <div className="schema-form-block-title">Fields</div>
            {(value.fields ?? []).map((field, index) => <div className="schema-enum-row" key={index}>
                <input aria-label={`Field name ${index + 1}`} value={field.name} placeholder="name" onChange={(event) => onChange({...value, fields: (value.fields ?? []).map((entry, position) => position === index ? {...entry, name: event.target.value} : entry)})}/>
                <select aria-label={`Field type ${index + 1}`} value={field.schema.kind} onChange={(event) => onChange({...value, fields: (value.fields ?? []).map((entry, position) => position === index ? {...entry, schema: {kind: event.target.value}} : entry)})}>
                    {["string", "integer", "number", "boolean", "enum", "reference"].map((option) => <option key={option} value={option}>{option}</option>)}
                </select>
                <label className="schema-field-optional"><input type="checkbox" checked={!!field.optional} onChange={(event) => onChange({...value, fields: (value.fields ?? []).map((entry, position) => position === index ? {...entry, optional: event.target.checked} : entry)})}/>optional</label>
                <button aria-label={`Delete field ${index + 1}`} onClick={() => onChange({...value, fields: (value.fields ?? []).filter((_, position) => position !== index)})}>×</button>
            </div>)}
            <button type="button" className="schema-form-add" onClick={() => onChange({...value, fields: [...(value.fields ?? []), {name: "", schema: {kind: "string"}, optional: false}]})}>Add field</button>
        </div>}
        {kind !== "enum" && <div className="schema-form-block">
            <div className="schema-form-block-title">Conditions (all must hold)</div>
            {requirements.map((requirement, index) => {
                const error = requirementError(requirement);
                return <div className="schema-req-row" key={index}>
                    <select aria-label="Condition type" value={requirement.kind ?? "custom"} onChange={(event) => updateRequirement(index, newRequirement(event.target.value))}>
                        {options.map((option) => <option key={option.kind} value={option.kind}>{option.label}</option>)}
                    </select>
                    <RequirementInput requirement={requirement} baseType={kind} constraintOptions={constraintReferences} onChange={(next) => updateRequirement(index, next)}/>
                    <button aria-label="Delete condition" onClick={() => removeRequirement(index)}>×</button>
                    {error && <span className="schema-req-error">{error}</span>}
                </div>;
            })}
            <button type="button" className="schema-form-add" onClick={addRequirement}>Add condition</button>
        </div>}
        <div className="schema-form-row schema-form-flags">
            <label><input type="checkbox" checked={!!value.nullable} onChange={(event) => onChange({...value, nullable: event.target.checked})}/>Allow null</label>
            <label className="schema-default">Default value <input value={value.default?.expression ?? ""} placeholder="optional" onChange={(event) => onChange({...value, default: event.target.value ? {expression: event.target.value} : undefined})}/></label>
        </div>
    </div>;
}
