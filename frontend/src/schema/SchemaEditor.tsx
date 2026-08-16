import {useEffect, useState} from "react";
import {PrintSchema} from "../../wailsjs/go/main/App";
import {schemamodel} from "../../wailsjs/go/models";
import type {SchemaEditorState, SchemaNode, SchemaUsage, WorkspaceView} from "../types";
import {schemaTypeReferenceOptions} from "./refs";
import {SchemaForm, predicateKindsFor, schemaNodeWithKind} from "./SchemaForm";

type SchemaDraft = {name: string; kind: "type" | "constraint"; type: string; appliesTo: string; expression: string; node: SchemaNode; custom: boolean};

export function SchemaEditor({state, usages, onOpenUsage, onSelect, workspace, onCreate, onApply, onRename, onDelete}: {
    state: SchemaEditorState;
    usages: (name: string) => SchemaUsage[];
    onOpenUsage: (usage: SchemaUsage) => void;
    onSelect: (name: string) => void;
    workspace: WorkspaceView | null;
    onCreate: (name: string, kind: "type" | "constraint", type: string, appliesTo: string[], expression: string) => Promise<void>;
    onApply: (name: string, kind: "type" | "constraint", type: string, appliesTo: string[], expression: string) => Promise<void>;
    onRename: (oldName: string, newName: string) => Promise<void>;
    onDelete: (name: string) => Promise<void>;
}) {
    const [draft, setDraft] = useState<SchemaDraft | null>(null);
    const [creating, setCreating] = useState(false);
    const [renaming, setRenaming] = useState(false);
    const [renameValue, setRenameValue] = useState("");
    const [confirmingDelete, setConfirmingDelete] = useState(false);
    const [busy, setBusy] = useState(false);
    const [editorError, setEditorError] = useState("");
    const typeReferences = schemaTypeReferenceOptions(workspace).map((entry) => entry.value);

    useEffect(() => {
        setCreating(false);
        setRenaming(false);
        setConfirmingDelete(false);
        setEditorError("");
        if (state.creating) {
            setDraft({name: "", kind: "type", type: "string", appliesTo: "", expression: "string", node: {kind: "string"}, custom: false});
            setCreating(true);
            return;
        }
        const definition = state.definitions.find((entry) => entry.name === state.selected) ?? null;
            setDraft(definition ? {
            name: definition.name,
            kind: definition.kind,
            type: definition.type ?? "",
            appliesTo: (definition.appliesTo ?? []).join(", "),
            expression: definition.schemaExpression,
            node: definition.schema,
            custom: false,
        } : null);
    }, [state.selected, state.definitions, state.creating]);

    function startCreate() {
        setDraft({name: "", kind: "type", type: "string", appliesTo: "", expression: "string", node: {kind: "string"}, custom: false});
        setCreating(true);
        setRenaming(false);
        setConfirmingDelete(false);
        setEditorError("");
    }

    async function submit() {
        if (!draft) return;
        setBusy(true);
        setEditorError("");
        try {
            const appliesTo = draft.appliesTo.split(",").map((entry) => entry.trim()).filter((entry) => entry !== "");
            const expression = draft.custom ? draft.expression : await PrintSchema(draft.node as unknown as schemamodel.Node);
            if (creating) await onCreate(draft.name, draft.kind, draft.type, appliesTo, expression);
            else await onApply(draft.name, draft.kind, draft.type, appliesTo, expression);
        } catch (reason) {
            setEditorError(reason instanceof Error ? reason.message : String(reason));
        } finally {
            setBusy(false);
        }
    }

    async function submitRename() {
        if (!draft || !renameValue.trim() || renameValue.trim() === draft.name) {
            setRenaming(false);
            return;
        }
        setBusy(true);
        setEditorError("");
        try {
            await onRename(draft.name, renameValue.trim());
            setRenaming(false);
        } catch (reason) {
            setEditorError(reason instanceof Error ? reason.message : String(reason));
        } finally {
            setBusy(false);
        }
    }

    const selectedUsages = state.selected ? usages(state.selected) : [];
    return (
        <section className="schema-editor">
            <aside className="schema-editor-sidebar">
                <div className="schema-editor-title">{state.local ? "Local Schema" : "Public Schema"}</div>
                <div className="schema-editor-file">{state.path}</div>
                <div className="schema-definition-list">
                    {state.definitions.map((definition) => <button
                        className={`schema-definition-row ${state.selected === definition.name ? "schema-definition-row-selected" : ""}`}
                        key={definition.name}
                        onClick={() => onSelect(definition.name)}
                    >
                        <span>#{definition.name}</span>
                        <span className={`schema-kind-dot schema-kind-${definition.kind}`} aria-hidden="true"/>
                    </button>)}
                    {state.definitions.length === 0 && <div className="schema-editor-empty">No definitions</div>}
                </div>
                <button className="schema-editor-create" onClick={startCreate}>New definition</button>
            </aside>
            <div className="schema-editor-form">
                {draft ? <>
                    <div className="schema-editor-row">
                        <label>Name<input aria-label="Name" value={draft.name} disabled={!creating} onChange={(event) => setDraft({...draft, name: event.target.value})}/></label>
                        <label>Kind<select aria-label="Kind" value={draft.kind} onChange={(event) => setDraft({...draft, kind: event.target.value as "type" | "constraint", node: event.target.value === "type" ? schemaNodeWithKind(draft.node, draft.type, typeReferences) : {...draft.node, kind: "unknown"}})}>
                            <option value="type">type</option>
                            <option value="constraint">constraint</option>
                        </select></label>
                    </div>
                    {draft.kind === "type" ? (
                        <label>Type<select aria-label="Type" value={draft.type} onChange={(event) => setDraft({...draft, type: event.target.value, node: schemaNodeWithKind(draft.node, event.target.value, typeReferences)})}>
                            {["string", "integer", "number", "boolean", "enum", "array", "struct", "reference"].map((option) => <option key={option} value={option}>{option}</option>)}
                        </select></label>
                    ) : (
                        <label>Applies to<input aria-label="Applies to" value={draft.appliesTo} placeholder="integer, number" onChange={(event) => setDraft({...draft, appliesTo: event.target.value})}/></label>
                    )}
                    <div className="schema-mode-toggle">
                        <button type="button" className={!draft.custom ? "schema-mode-active" : ""} onClick={() => setDraft({...draft, custom: false})}>Structured</button>
                        <button type="button" className={draft.custom ? "schema-mode-active" : ""} onClick={() => setDraft({...draft, custom: true})}>Custom expression</button>
                    </div>
                    {draft.custom
                        ? <label>Schema expression<textarea aria-label="Schema expression" rows={6} value={draft.expression} onChange={(event) => setDraft({...draft, expression: event.target.value})}/></label>
                        : <SchemaForm value={draft.node} onChange={(node) => setDraft({...draft, node})} workspace={workspace} fixedKind={draft.kind === "type" ? draft.type : "unknown"} predicateKinds={draft.kind === "constraint" ? predicateKindsFor(draft.appliesTo.split(",").map((entry) => entry.trim()).filter(Boolean)) : undefined}/>}
                    <div className="schema-editor-actions">
                        <button disabled={busy} onClick={submit}>{creating ? "Create" : "Apply"}</button>
                        {creating ? <button onClick={() => { setCreating(false); setDraft(null); }}>Cancel</button> : <>
                            {renaming ? <>
                                <input className="schema-rename-input" value={renameValue} onChange={(event) => setRenameValue(event.target.value)}/>
                                <button disabled={busy} onClick={submitRename}>Confirm rename</button>
                                <button onClick={() => setRenaming(false)}>Cancel</button>
                            </> : <button onClick={() => { setRenameValue(draft.name); setRenaming(true); }}>Rename</button>}
                            <button className={confirmingDelete ? "schema-delete-danger" : ""} onClick={async () => {
                                if (!confirmingDelete) {
                                    setConfirmingDelete(true);
                                    return;
                                }
                                setConfirmingDelete(false);
                                setBusy(true);
                                setEditorError("");
                                try {
                                    await onDelete(draft.name);
                                } catch (reason) {
                                    setEditorError(reason instanceof Error ? reason.message : String(reason));
                                } finally {
                                    setBusy(false);
                                }
                            }}>{confirmingDelete ? "Confirm delete" : "Delete definition"}</button>
                        </>}
                    </div>
                    {editorError && <div className="schema-editor-error">{editorError}</div>}
                </> : <div className="schema-editor-empty">Select or create a definition</div>}
                <div className="schema-editor-usages">
                    <div className="schema-editor-usages-title">Usages</div>
                    {selectedUsages.map((usage) => <button className="usage-row usage-row-action" key={`${usage.path}:${usage.field}`} onClick={() => onOpenUsage(usage)}>
                        <span>{usage.table ?? usage.path}</span>
                        <code>{usage.field}</code>
                    </button>)}
                    {state.selected && selectedUsages.length === 0 && <div className="schema-editor-empty">No usages</div>}
                </div>
            </div>
        </section>
    );
}
