import type {TreeNode} from "../types";

export function Tree({node, depth = 0, namespace = "", parentKey = "", collapsed, onToggleCollapse, onOpenTable, onOpenSchema, onOpenProjectSettings, onContextMenu}: { node: TreeNode; depth?: number; namespace?: string; parentKey?: string; collapsed: Set<string>; onToggleCollapse: (key: string) => void; onOpenTable: (path: string) => void; onOpenSchema: (node: TreeNode) => void; onOpenProjectSettings: () => void; onContextMenu: (node: TreeNode, position: {x: number; y: number}, namespace: string) => void }) {
    const key = `${parentKey}::${node.kind}:${node.name}`;
    const isFile = node.kind === "table" || node.kind === "schemaFile" || node.kind === "schemaDefinition" || node.kind === "damaged";
    const isSchemaNode = node.kind === "schemaFile" || node.kind === "schemaDefinition" || node.kind === "localSchema";
    const openTable = node.kind === "table" && node.path ? () => onOpenTable(node.path!) : undefined;
    const openSchema = isSchemaNode && node.path ? () => onOpenSchema(node) : undefined;
    const openProjectSettings = node.kind === "project" ? onOpenProjectSettings : undefined;
    const openAction = openProjectSettings ?? openTable ?? openSchema;
    const childNamespace = node.kind === "namespace" ? node.name : namespace;
    const hasChildren = (node.children?.length ?? 0) > 0;
    const isCollapsed = hasChildren && collapsed.has(key);
    return (
        <div className="tree-node">
            <div className={`tree-row ${isFile ? "tree-file" : ""} ${openAction ? "tree-action" : ""}`} style={{paddingLeft: `${12 + depth * 16}px`}}>
                {hasChildren
                    ? <button className="tree-toggle" type="button" tabIndex={-1} onClick={() => onToggleCollapse(key)}>{isCollapsed ? "▸" : "▾"}</button>
                    : <span className="tree-toggle tree-toggle-spacer" aria-hidden="true"/>}
                <button className={`tree-label ${openAction ? "tree-action" : ""}`} type="button" onClick={openAction} onContextMenu={(event) => {
                    event.preventDefault();
                    onContextMenu(node, {x: event.clientX, y: event.clientY}, namespace);
                }}>
                    <span className={`node-icon node-${node.kind}`} aria-hidden="true"/>
                    <span>{node.name}</span>
                </button>
            </div>
            {!isCollapsed && node.children?.map((child) => <Tree key={`${key}::${child.kind}:${child.name}`} node={child} depth={depth + 1} namespace={childNamespace} parentKey={key} collapsed={collapsed} onToggleCollapse={onToggleCollapse} onOpenTable={onOpenTable} onOpenSchema={onOpenSchema} onOpenProjectSettings={onOpenProjectSettings} onContextMenu={onContextMenu}/>) }
        </div>
    );
}
