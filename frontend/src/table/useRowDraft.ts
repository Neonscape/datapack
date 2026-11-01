import {useRef, useState} from "react";

type RowDraft = {id:string; values:Record<string,string>};
type DraftTarget = {kind:"id"} | {kind:"cell"; name:string};
type CreateDraft = (id:string,values:Record<string,string>,formulas:Record<string,string>)=>Promise<void>;

// Owns the unsaved row workflow, not grid coordinates, IPC payloads or locale.
// The grid supplies a field identity and a creation command; rejected saves
// retain the exact draft and expose one local error as before.
export function useRowDraft(hasID: (id:string)=>boolean, create:CreateDraft, messages:{duplicate:string;invalid:string}) {
    const [draft,setDraft]=useState<RowDraft>({id:"",values:{}});
    const [error,setError]=useState("");
    const current=useRef(draft);
    const saving=useRef(false);
    const publish=(next:RowDraft)=>{current.current=next;setDraft(next);};
    async function edit(target:DraftTarget,input:string) {
        const draft=current.current;
        const next = target.kind === "id" ? {id:input,values:draft.values} : {id:draft.id,values:{...draft.values,[target.name]:input}};
        if(target.kind === "cell" && input === "") delete next.values[target.name];
        publish(next);
        const duplicate=hasID(next.id);
        if(next.id !== "" && (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(next.id)||duplicate)) {
            setError(duplicate?messages.duplicate:messages.invalid);return;
        }
        setError("");
        // Keep edits locally while the single create is in flight. A later
        // edit retries explicitly; settling an old request never replaces it.
        if(saving.current)return;
        const hasValues=Object.values(next.values).some(value=>value!=="");
        if(next.id !== "" && hasValues) {
            const values:Record<string,string>={}, formulas:Record<string,string>={};
            for(const [name,expression] of Object.entries(next.values)) {
                if(expression.startsWith("="))formulas[name]=expression.slice(1).trim();else values[name]=expression;
            }
            saving.current=true;
            try {
                await create(next.id,values,formulas);
                if(current.current===next){publish({id:"",values:{}});setError("");}
                else if(current.current.id===next.id)setError(messages.duplicate);
            }catch(reason){setError(reason instanceof Error?reason.message:String(reason));}
            finally{saving.current=false;}
            return;
        }
        setError("");
    }
    return {draft,error,edit};
}
