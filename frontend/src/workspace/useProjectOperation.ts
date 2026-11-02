import {useEffect, useState} from "react";
import {operationSnapshot, subscribeOperations} from "../bridge/ordered";
export type OperationProgress = {id:number;kind:string;phase:string;completed:number;total:number;running:boolean;error?:string};
export function useProjectOperation() {
    const [operation, setOperation] = useState(operationSnapshot);
    const [progress, setProgress] = useState<OperationProgress | null>(null);
    useEffect(() => subscribeOperations(() => setOperation(operationSnapshot())), []);
    useEffect(() => {
        setProgress(null);
        if (!operation.busy) return;
        let active = true, pending = false;
        async function poll() {
            if (pending) return;
            pending = true;
            try {
                // Poll outside the ordered lane so progress is readable while
                // its project RPC is running. Older test bridges may omit it.
                const api = (window as unknown as {go?:{main?:{App?:{GetOperationState?:()=>Promise<OperationProgress>}}}}).go?.main?.App;
                const value = await api?.GetOperationState?.();
                if (active && value?.running) setProgress(value);
            } finally { pending = false; }
        }
        void poll().catch(() => undefined);
        const timer = setInterval(() => { void poll().catch(() => undefined); }, 150);
        return () => { active = false; clearInterval(timer); };
    }, [operation.busy]);
    return {...operation, progress};
}
