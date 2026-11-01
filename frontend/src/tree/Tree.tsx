import type {TreeNode} from "../types";
import {useTranslation} from "../i18n";
import {displayLabel} from "./displayLabel";
import {treeNodeKey} from "./expansion";

export function Tree({node, depth = 0, namespace = "", parentKey = "", expanded, onToggleCollapse, onOpenTable, onOpenSchema, onOpenProjectSettings, onContextMenu}: { node: TreeNode; depth?: number; namespace?: string; parentKey?: string; expanded: Set<string>; onToggleCollapse: (key: string) => void; onOpenTable: (path: string) => void; onOpenSchema: (node: TreeNode) => void; onOpenProjectSettings: () => void; onContextMenu: (node: TreeNode, position: {x: number; y: number}, namespace: string) => void }) {
    const {t} = useTranslation();
    const key = treeNodeKey(node, parentKey);
    const label = node.kind === "group" && node.name === "Tables" ? t("tree.tables")
        : node.kind === "group" && node.name === "Schemas" ? t("tree.schemas")
        : node.kind === "localSchema" && node.name === "local_schema" ? t("tree.localSchema") : displayLabel(node.displayName, node.name);
    const isFile = node.kind === "table" || node.kind === "schemaFile" || node.kind === "schemaDefinition" || node.kind === "damaged";
    const isSchemaNode = node.kind === "schemaFile" || node.kind === "schemaDefinition" || node.kind === "localSchema";
    const openTable = node.kind === "table" && node.path ? () => onOpenTable(node.path!) : undefined;
    const openSchema = isSchemaNode && node.path ? () => onOpenSchema(node) : undefined;
    const openProjectSettings = node.kind === "project" ? onOpenProjectSettings : undefined;
    const openAction = openProjectSettings ?? openTable ?? openSchema;
    const childNamespace = node.kind === "namespace" ? node.name : namespace;
    const hasChildren = (node.children?.length ?? 0) > 0;
    const isCollapsed = hasChildren && !expanded.has(key);
    return (
        <div className="tree-node">
            <div className={`tree-row ${isFile ? "tree-file" : ""} ${openAction ? "tree-action" : ""}`} style={{paddingLeft: `${12 + depth * 16}px`}}>
                {hasChildren
                    ? <button className="tree-toggle" type="button" aria-label={t(isCollapsed ? "tree.expand" : "tree.collapse", {name: label})} aria-expanded={!isCollapsed} onClick={() => onToggleCollapse(key)}>{isCollapsed ? "▸" : "▾"}</button>
                    : <span className="tree-toggle tree-toggle-spacer" aria-hidden="true"/>}
                <button className={`tree-label ${openAction ? "tree-action" : ""}`} type="button" onClick={openAction} onContextMenu={(event) => {
                    event.preventDefault();
                    onContextMenu(node, {x: event.clientX, y: event.clientY}, namespace);
                }}>
                    <span className={`node-icon node-${node.kind}`} aria-hidden="true"/>
                    <span title={label}>{label}</span>
                </button>
            </div>
            {!isCollapsed && node.children?.map((child) => <Tree key={treeNodeKey(child, key)} node={child} depth={depth + 1} namespace={childNamespace} parentKey={key} expanded={expanded} onToggleCollapse={onToggleCollapse} onOpenTable={onOpenTable} onOpenSchema={onOpenSchema} onOpenProjectSettings={onOpenProjectSettings} onContextMenu={onContextMenu}/>) }
        </div>
    );
}
