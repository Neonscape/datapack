// Narrow test-only declaration: read existing UTF-8 example files without
// installing @types/node or weakening the strict test-source checks.
declare module "node:fs" {
    export function readFileSync(path: URL, encoding: "utf8"): string;
}
