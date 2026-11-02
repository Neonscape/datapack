// Mutations share a single dispatch lane. Reads may proceed concurrently; their
// revision/epoch tags are checked by the session before installing a response.
let tail: Promise<unknown> = Promise.resolve();
let leases = 0;
let kind = "";
const listeners = new Set<() => void>();
const heavy = new Set(["Compile", "Export", "SaveProjectAs", "OpenProject", "SelectProject", "NewProject"]);
export function operationSnapshot() { return {busy: leases > 0, kind}; }
export function subscribeOperations(listener: () => void) { listeners.add(listener); return () => { listeners.delete(listener); }; }
function changed() { for (const listener of listeners) listener(); }
export async function withProjectOperation<T>(name: string, run: () => Promise<T>): Promise<T> {
    leases++; kind = name; changed();
    try { return await run(); } finally { leases--; changed(); }
}
export function orderedCall<T>(name: string, run: () => Promise<T>): Promise<T> {
    if (/^Get|^Print|^Navigate/.test(name)) return run();
    const dispatch = () => {
        const result = tail.then(run);
        tail = result.catch(() => undefined);
        return result;
    };
    return heavy.has(name) ? withProjectOperation(name, dispatch) : dispatch();
}
