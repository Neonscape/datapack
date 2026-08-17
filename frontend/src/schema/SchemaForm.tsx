import {useTranslation} from "../i18n";
import type {SchemaNode, SchemaRequirement, WorkspaceView} from "../types";
import {schemaConstraintReferenceOptions, schemaTypeReferenceOptions} from "./refs";

type Translate = (key: string, params?: Record<string, string | number>) => string;

function predicateOptions(kind: string, t: Translate): {kind: string; label: string}[] {
    if (kind === "string") {
        return [
            {kind: "compare", label: t("schemaForm.compare")},
            {kind: "regex", label: t("schemaForm.regex")},
            {kind: "nonEmpty", label: t("schemaForm.nonEmpty")},
            {kind: "length", label: t("schemaForm.length")},
            {kind: "reference", label: t("schemaForm.reference")},
            {kind: "custom", label: t("schemaForm.custom")},
        ];
    }
    if (kind === "integer" || kind === "number") {
        return [
            {kind: "compare", label: t("schemaForm.compare")},
            {kind: "reference", label: t("schemaForm.reference")},
            {kind: "custom", label: t("schemaForm.custom")},
        ];
    }
    return [
        {kind: "reference", label: t("schemaForm.reference")},
        {kind: "custom", label: t("schemaForm.custom")},
    ];
}

const predicateLabelKeys: Record<string, string> = {compare: "schemaForm.compare", regex: "schemaForm.regex", nonEmpty: "schemaForm.nonEmpty", length: "schemaForm.length", reference: "schemaForm.reference", custom: "schemaForm.custom"};

function predicateLabel(kind: string, t: Translate): string {
    const key = predicateLabelKeys[kind];
    return key ? t(key) : kind;
}

export function predicateKindsFor(types: string[], t: Translate): string[] {
    const kinds: string[] = [];
    for (const type of types) {
        for (const option of predicateOptions(type, t)) {
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

function requirementError(requirement: SchemaRequirement, t: Translate): string {
    switch (requirement.kind) {
        case "compare":
            return requirement.value?.trim() ? "" : t("schemaForm.compareValueRequired");
        case "regex":
            if (!requirement.pattern) return t("schemaForm.regexPatternRequired");
            try {
                new RegExp(requirement.pattern);
                return "";
            } catch {
                return t("schemaForm.invalidRegex");
            }
        case "length":
            if (requirement.min != null && (!Number.isInteger(requirement.min) || requirement.min < 0)) return t("schemaForm.minLength");
            if (requirement.max != null && (!Number.isInteger(requirement.max) || requirement.max < 0)) return t("schemaForm.maxLength");
            return "";
        case "reference":
            return requirement.reference ? "" : t("schemaForm.constraintRequired");
        case "custom":
            return requirement.expression?.trim() ? "" : t("schemaForm.expressionRequired");
        default:
            return "";
    }
}

function RequirementInput({requirement, baseType, constraintOptions, onChange, t}: {requirement: SchemaRequirement; baseType: string; constraintOptions: {label: string; value: string}[]; onChange: (requirement: SchemaRequirement) => void; t: Translate}) {
    if (requirement.kind === "compare") {
        return <span className="schema-req-input">
            <select aria-label={t("schemaForm.comparisonOperator")} value={requirement.op} onChange={(event) => onChange({...requirement, op: event.target.value})}>
                {[">=", ">", "<=", "<", "==", "!="].map((op) => <option key={op} value={op}>{op}</option>)}
            </select>
            <input aria-label={t("schemaForm.comparisonValue")} value={requirement.value ?? ""} placeholder={baseType === "string" ? `"value"` : "0"} onChange={(event) => onChange({...requirement, value: event.target.value})}/>
        </span>;
    }
    if (requirement.kind === "regex") {
        return <input aria-label={t("schemaForm.regexPattern")} value={requirement.pattern ?? ""} placeholder="^[a-z]+$" onChange={(event) => onChange({...requirement, pattern: event.target.value})}/>;
    }
    if (requirement.kind === "nonEmpty") {
        return <input aria-label={t("schemaForm.nonEmpty")} disabled placeholder={t("schemaForm.nonEmptyNoInput")}/>;
    }
    if (requirement.kind === "length") {
        const mode = requirement.max != null ? "max" : "min";
        return <span className="schema-req-input">
            <select aria-label={t("schemaForm.lengthBound")} value={mode} onChange={(event) => onChange(event.target.value === "max" ? {...requirement, min: undefined, max: 1} : {...requirement, min: 1, max: undefined})}>
                <option value="min">{t("schemaForm.minLengthOption")}</option>
                <option value="max">{t("schemaForm.maxLengthOption")}</option>
            </select>
            <input aria-label={t("schemaForm.lengthValue")} type="number" min="0" value={mode === "max" ? (requirement.max ?? 1) : (requirement.min ?? 1)} onChange={(event) => onChange(mode === "max" ? {...requirement, max: Number(event.target.value)} : {...requirement, min: Number(event.target.value)})}/>
        </span>;
    }
    if (requirement.kind === "reference") {
        return <select aria-label={t("schemaForm.constraintReference")} value={requirement.reference ?? ""} onChange={(event) => onChange({...requirement, reference: event.target.value})}>
            <option value="">{t("schemaForm.chooseConstraint")}</option>
            {constraintOptions.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
        </select>;
    }
    return <input aria-label={t("schemaForm.customExpression")} value={requirement.expression ?? ""} placeholder=">=0" onChange={(event) => onChange({...requirement, expression: event.target.value})}/>;
}

export function SchemaForm({value, onChange, workspace, fixedKind, predicateKinds}: {value: SchemaNode; onChange: (node: SchemaNode) => void; workspace: WorkspaceView | null; fixedKind?: string; predicateKinds?: string[]}) {
    const {t} = useTranslation();
    const kind = fixedKind ?? value.kind ?? "unknown";
    const typeReferences = schemaTypeReferenceOptions(workspace);
    const constraintReferences = schemaConstraintReferenceOptions(workspace);
    const options = predicateKinds ? predicateKinds.map((entry) => ({kind: entry, label: predicateLabel(entry, t)})) : predicateOptions(kind, t);

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
        {fixedKind === undefined && <label className="schema-form-row">{t("schemaForm.baseType")}
            <select aria-label={t("schemaForm.baseType")} value={kind} onChange={(event) => setKind(event.target.value)}>
                {["string", "integer", "number", "boolean", "enum", "array", "struct", "reference"].map((option) => <option key={option} value={option}>{option}</option>)}
            </select>
        </label>}
        {kind === "reference" && <label className="schema-form-row">{t("schemaForm.typeSchema")}
            <select aria-label={t("schemaForm.typeSchema")} value={value.reference ?? ""} onChange={(event) => onChange({...value, reference: event.target.value})}>
                <option value="">{t("schemaForm.chooseType")}</option>
                {typeReferences.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
            </select>
        </label>}
        {kind === "enum" && <div className="schema-form-block">
            <div className="schema-form-block-title">{t("schemaForm.enumValues")}</div>
            {(value.enumValues ?? []).map((enumValue, index) => <div className="schema-enum-row" key={index}>
                <input value={enumValue.expression} onChange={(event) => onChange({...value, enumValues: (value.enumValues ?? []).map((entry, position) => position === index ? {expression: event.target.value} : entry)})}/>
                <button aria-label={t("schemaForm.deleteEnumValue", {index: index + 1})} onClick={() => onChange({...value, enumValues: (value.enumValues ?? []).filter((_, position) => position !== index)})}>×</button>
            </div>)}
            <button type="button" className="schema-form-add" onClick={() => onChange({...value, enumValues: [...(value.enumValues ?? []), {expression: `""`}]})}>{t("schemaForm.addEnumValue")}</button>
        </div>}
        {kind === "array" && <label className="schema-form-row">{t("schemaForm.elementType")}
            <select aria-label={t("schemaForm.elementType")} value={value.element?.kind ?? "string"} onChange={(event) => onChange({...value, element: {kind: event.target.value}})}>
                {["string", "integer", "number", "boolean", "reference"].map((option) => <option key={option} value={option}>{option}</option>)}
            </select>
        </label>}
        {kind === "array" && value.element?.kind === "reference" && <label className="schema-form-row">{t("schemaForm.elementSchema")}
            <select aria-label={t("schemaForm.elementSchema")} value={value.element.reference ?? ""} onChange={(event) => onChange({...value, element: {...value.element!, reference: event.target.value}})}>
                <option value="">{t("schemaForm.chooseType")}</option>
                {typeReferences.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
            </select>
        </label>}
        {kind === "struct" && <div className="schema-form-block">
            <div className="schema-form-block-title">{t("schemaForm.fields")}</div>
            {(value.fields ?? []).map((field, index) => <div className="schema-enum-row" key={index}>
                <input aria-label={t("schemaForm.fieldName", {index: index + 1})} value={field.name} placeholder="name" onChange={(event) => onChange({...value, fields: (value.fields ?? []).map((entry, position) => position === index ? {...entry, name: event.target.value} : entry)})}/>
                <select aria-label={t("schemaForm.fieldType", {index: index + 1})} value={field.schema.kind} onChange={(event) => onChange({...value, fields: (value.fields ?? []).map((entry, position) => position === index ? {...entry, schema: {kind: event.target.value}} : entry)})}>
                    {["string", "integer", "number", "boolean", "enum", "reference"].map((option) => <option key={option} value={option}>{option}</option>)}
                </select>
                <label className="schema-field-optional"><input type="checkbox" checked={!!field.optional} onChange={(event) => onChange({...value, fields: (value.fields ?? []).map((entry, position) => position === index ? {...entry, optional: event.target.checked} : entry)})}/>{t("schemaForm.optional")}</label>
                <button aria-label={t("schemaForm.deleteField", {index: index + 1})} onClick={() => onChange({...value, fields: (value.fields ?? []).filter((_, position) => position !== index)})}>×</button>
            </div>)}
            <button type="button" className="schema-form-add" onClick={() => onChange({...value, fields: [...(value.fields ?? []), {name: "", schema: {kind: "string"}, optional: false}]})}>{t("schemaForm.addField")}</button>
        </div>}
        {kind !== "enum" && <div className="schema-form-block">
            <div className="schema-form-block-title">{t("schemaForm.conditions")}</div>
            {requirements.map((requirement, index) => {
                const error = requirementError(requirement, t);
                return <div className="schema-req-row" key={index}>
                    <select aria-label={t("schemaForm.conditionType")} value={requirement.kind ?? "custom"} onChange={(event) => updateRequirement(index, newRequirement(event.target.value))}>
                        {options.map((option) => <option key={option.kind} value={option.kind}>{option.label}</option>)}
                    </select>
                    <RequirementInput requirement={requirement} baseType={kind} constraintOptions={constraintReferences} onChange={(next) => updateRequirement(index, next)} t={t}/>
                    <button aria-label={t("schemaForm.deleteCondition")} onClick={() => removeRequirement(index)}>×</button>
                    {error && <span className="schema-req-error">{error}</span>}
                </div>;
            })}
            <button type="button" className="schema-form-add" onClick={addRequirement}>{t("schemaForm.addCondition")}</button>
        </div>}
        <div className="schema-form-row schema-form-flags">
            <label><input type="checkbox" checked={!!value.nullable} onChange={(event) => onChange({...value, nullable: event.target.checked})}/>{t("schemaForm.allowNull")}</label>
            <label className="schema-default">{t("schemaForm.defaultValue")} <input value={value.default?.expression ?? ""} placeholder={t("schemaForm.optionalPlaceholder")} onChange={(event) => onChange({...value, default: event.target.value ? {expression: event.target.value} : undefined})}/></label>
        </div>
    </div>;
}
