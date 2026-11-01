import type {TreeNode} from "../types";

// Presentation labels and locale never participate in tree identity.
export function treeNodeKey(node: TreeNode, parentKey = ""): string {
    return `${parentKey}::${node.kind}:${node.name}`;
}

export function initialExpandedNodes(root: TreeNode, defaultPath?: string, defaultDefinition?: string): Set<string> {
    const expanded = new Set([treeNodeKey(root)]);
    if (!defaultPath) return expanded;
    function visit(node: TreeNode, parentKey: string, ancestors: string[]): boolean {
        const key = treeNodeKey(node, parentKey);
        const path = [...ancestors, key];
        if (node.path === defaultPath && (defaultDefinition
            ? node.name === `#${defaultDefinition}` && (node.kind === "schemaDefinition" || node.kind === "localSchema")
            : node.kind === "table" || node.kind === "schemaFile")) {
            path.forEach((ancestor) => expanded.add(ancestor));
            return true;
        }
        return node.children?.some((child) => visit(child, key, path)) ?? false;
    }
    if (!visit(root, "", []) && defaultDefinition) return initialExpandedNodes(root, defaultPath);
    return expanded;
}
