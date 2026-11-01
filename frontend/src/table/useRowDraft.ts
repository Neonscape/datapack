import {useState} from "react";

type RowDraft = {id:string; values:Record<string,string>};
type DraftTarget = {kind:"id"} | {kind:"cell"; name:string};
type CreateDraft = (id:string,values:Record<string,string>,formulas:Record<string,string>)=>Promise<void>;

// Owns the unsaved row workflow, not grid coordinates, IPC payloads or locale.
// The grid supplies a field identity and a creation command; rejected saves
// retain the exact draft and expose one local error as before.
export function useRowDraft(hasID: (id:string)=>boolean, create:CreateDraft, messages:{duplicate:string;invalid:string}) {
    const [draft,setDraft]=useState<RowDraft>({id:"",values:{}});
    const [error,setError]=useState("");
    async function edit(target:DraftTarget,input:string) {
        const next = target.kind === "id" ? {id:input,values:draft.values} : {id:draft.id,values:{...draft.values,[target.name]:input}};
        if(target.kind === "cell" && input === "") delete next.values[target.name];
        const duplicate=hasID(next.id);
        if(next.id !== "" && (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(next.id)||duplicate)) {
            setDraft(next);setError(duplicate?messages.duplicate:messages.invalid);return;
        }
        const hasValues=Object.values(next.values).some(value=>value!=="");
        if(next.id !== "" && hasValues) {
            const values:Record<string,string>={}, formulas:Record<string,string>={};
            for(const [name,expression] of Object.entries(next.values)) {
                if(expression.startsWith("="))formulas[name]=expression.slice(1).trim();else values[name]=expression;
            }
            try {
                await create(next.id,values,formulas);
                setDraft({id:"",values:{}});setError("");
            }catch(reason){setDraft(next);setError(reason instanceof Error?reason.message:String(reason));}
            return;
        }
        setDraft(next);setError("");
    }
    return {draft,error,edit};
}
