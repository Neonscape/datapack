/// <reference path="./node-fs.d.ts" />
import {readFileSync} from "node:fs";
import {expect, test, type Page} from "@playwright/test";

test.use({viewport: {width: 1600, height: 1000}});
// Actual example rows, real App/array dialogs; only Wails is mocked here.
// TestYunlanSkillAndTagArraysSaveReopenCompileUndoRedo covers actual CUE I/O.
function section(path: string, name = "rows"): Record<string, Record<string, unknown>> {
    const source = readFileSync(new URL(`../../docs/examples/yunlan-rpg/${path}`, import.meta.url), "utf8");
    const marker = `\n${name}: `;
    const start = source.indexOf(marker);
    const end = source.indexOf("\n\n_", start);
    if (start < 0 || end < 0) throw new Error("sample rows missing");
    return JSON.parse(source.slice(start + marker.length, end)) as Record<string, Record<string, unknown>>;
}
function strings(value: unknown): string[] {
    if (!Array.isArray(value) || !value.every((entry) => typeof entry === "string")) throw new Error("sample string array expected");
    return value as string[];
}
const heroRows = section("tables/content/hero/hero.cue");
const skillRows = section("tables/content/skill/skill.cue");
const heroEditor = section("tables/content/hero/hero.cue", "_editor");
const skillEditor = section("tables/content/skill/skill.cue", "_editor");
function title(editor: Record<string, Record<string, unknown>>, field: string): string {
    const column = editor.columns[field] as {title?: unknown};
    if (typeof column?.title !== "string") throw new Error("sample display name missing");
    return column.title;
}
type State = {values: Record<string, string[]>; calls: {method: string; argument?: Record<string, unknown>}[]};
declare global {interface Window {__yunlanArrays: State; __yunlanHeaderText: string[]}}
async function openArray(page: Page, field: string) {
    const canvas = page.locator(".tab-panel:visible canvas").first();
    await expect(canvas).toBeVisible();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    await page.mouse.dblclick(box!.x + 225, box!.y + 50);
    await expect(page.getByRole("dialog", {name: `Edit ${field}`})).toBeVisible();
}
async function history(page: Page, action: "Undo" | "Redo") {
    await page.getByRole("button", {name: "Edit", exact: true}).click();
    await page.getByRole("button", {name: new RegExp(`^${action}`)}).click();
}

test("edits Yunlan skill/tag arrays, reopens, compiles, and uses Undo/Redo", async ({page}) => {
    await page.addInitScript(({heroSkills, skillTags, allSkillIDs, displayNames}) => {
        const state: State = {values: {hero: heroSkills, skill: skillTags}, calls: []};
        window.__yunlanArrays = state;
        // Observe the actual canvas draw calls without changing rendered text.
        window.__yunlanHeaderText = [];
        const fillText = CanvasRenderingContext2D.prototype.fillText;
        CanvasRenderingContext2D.prototype.fillText = function (this: CanvasRenderingContext2D, text: string, x: number, y: number, maxWidth?: number) {
            if (!window.__yunlanHeaderText.includes(text)) window.__yunlanHeaderText.push(text);
            if (maxWidth === undefined) fillText.call(this, text, x, y);
            else fillText.call(this, text, x, y, maxWidth);
        };
        const undo: Record<string, string[]>[] = [], redo: Record<string, string[]>[] = [];
        let pending = false;
        const path = (name: string) => `/project/tables/content/${name}/${name}.cue`;
        const view = () => ({
            project: {path: "/project/yunlan-rpg.dtproj", root: "/project"}, settings: {build: {mode: "manual"}},
            tree: {root: {kind: "project", name: "yunlan-rpg.dtproj", children: ["hero", "skill"].map((name) => ({kind: "table", name, path: path(name)}))}},
            schemas: {definitions: [], diagnostics: []}, foreignKeys: {diagnostics: []},
            compile: {mode: "manual", pending, diagnostics: []}, canUndo: undo.length > 0, canRedo: redo.length > 0,
        });
        const table = (tablePath: string) => {
            const name = tablePath === path("hero") ? "hero" : "skill";
            const field = name === "hero" ? "skill_ids" : "tags";
            const expression = `[\n${state.values[name].map((v) => JSON.stringify(v)).join(",\n")},\n]`;
            return {path: tablePath, namespace: "content", name,
                columns: [{name: field, displayName: displayNames[name], schemaExpression: "[...string]", schema: {kind: "array", element: {kind: "string"}}, ...(name === "hero" ? {foreignKeyTable: "content.skill", foreignKeyOptions: allSkillIDs} : {})}],
                rows: [{id: `${name}_01`, values: {[field]: {expression, display: expression}}}], localSchema: [], diagnostics: []};
        };
        (window as unknown as Window & {go: unknown}).go = {main: {App: {
            GetSettings: async () => ({language: "en", theme: "light"}), GetExportFormats: async () => [],
            SelectProject: async () => view(), GetTable: async (p: string) => table(p),
            UpdateCell: async (argument: {path: string; rowID: string; column: string; expression: string}) => {
                state.calls.push({method: "UpdateCell", argument});
                undo.push(structuredClone(state.values)); redo.length = 0;
                const name = argument.path === path("hero") ? "hero" : "skill";
                state.values = {...state.values, [name]: JSON.parse(argument.expression) as string[]}; pending = true;
                return {table: table(argument.path), workspace: view()};
            },
            Compile: async () => {state.calls.push({method: "Compile"}); pending = false; return {workspace: view()};},
            Undo: async () => {state.calls.push({method: "Undo"}); redo.push(structuredClone(state.values)); state.values = undo.pop()!; return {workspace: view(), canUndo: undo.length > 0, canRedo: true};},
            Redo: async () => {state.calls.push({method: "Redo"}); undo.push(structuredClone(state.values)); state.values = redo.pop()!; return {workspace: view(), canUndo: true, canRedo: redo.length > 0};},
        }}};
    }, {heroSkills: strings(heroRows.hero_01.skill_ids), skillTags: strings(skillRows.skill_01.tags), allSkillIDs: Object.keys(skillRows), displayNames: {hero: title(heroEditor, "skill_ids"), skill: title(skillEditor, "tags")}});
    await page.goto("/");
    await page.getByRole("button", {name: "File", exact: true}).click();
    await page.getByLabel("Application menu").getByRole("button", {name: "Open project", exact: true}).click();
    await page.getByRole("button", {name: "hero", exact: true}).click();
    await openArray(page, "skill_ids");
    await expect.poll(() => page.evaluate(() => window.__yunlanHeaderText.some((text) => text.startsWith("技能列表")))).toBe(true);
    await expect(page.getByRole("combobox", {name: "Array item 1", exact: true})).toHaveValue("skill_01");
    await page.getByRole("button", {name: "Move item 2 up"}).click();
    await page.getByRole("button", {name: "Apply", exact: true}).click();
    const hero = ["skill_02", "skill_01", "skill_03", "skill_04"];
    await expect.poll(() => page.evaluate(() => window.__yunlanArrays.values.hero)).toEqual(hero);
    await page.getByRole("tab", {name: /content.hero/}).locator(".tab-close").click();
    await page.getByRole("button", {name: "hero", exact: true}).click();
    await openArray(page, "skill_ids");
    await expect(page.getByRole("combobox", {name: "Array item 1", exact: true})).toHaveValue("skill_02");
    await page.getByRole("button", {name: "Close array editor"}).click();
    await page.getByRole("button", {name: "skill", exact: true}).click();
    await openArray(page, "tags");
    await expect.poll(() => page.evaluate(() => window.__yunlanHeaderText)).toContain("标签 (tags)");
    await page.getByRole("button", {name: "Move item 2 up"}).click();
    await page.getByRole("button", {name: "Add item", exact: true}).click();
    await page.getByRole("textbox", {name: "Array item 3", exact: true}).fill('"教学验证"');
    await page.getByRole("button", {name: "Apply", exact: true}).click();
    const skill = ["战士", "主动", "教学验证"];
    await expect.poll(() => page.evaluate(() => window.__yunlanArrays.values.skill)).toEqual(skill);
    await page.getByRole("button", {name: "Compile", exact: true}).click();
    await expect.poll(() => page.evaluate(() => window.__yunlanArrays.calls.filter((call) => call.method === "Compile").length)).toBe(1);
    await history(page, "Undo");
    await expect.poll(() => page.evaluate(() => window.__yunlanArrays.values.skill)).toEqual(["主动", "战士"]);
    await history(page, "Undo");
    await expect.poll(() => page.evaluate(() => window.__yunlanArrays.values.hero)).toEqual(strings(heroRows.hero_01.skill_ids));
    await history(page, "Redo"); await history(page, "Redo");
    await expect.poll(() => page.evaluate(() => window.__yunlanArrays.values)).toEqual({hero, skill});
    await openArray(page, "tags");
    await expect(page.getByRole("textbox", {name: "Array item 3", exact: true})).toHaveValue('"教学验证"');
    await page.getByRole("button", {name: "Close array editor"}).click();
    const updates = await page.evaluate(() => window.__yunlanArrays.calls.filter((call) => call.method === "UpdateCell"));
    expect(updates).toEqual([
        {method: "UpdateCell", argument: {path: "/project/tables/content/hero/hero.cue", rowID: "hero_01", column: "skill_ids", expression: '["skill_02", "skill_01", "skill_03", "skill_04"]'}},
        {method: "UpdateCell", argument: {path: "/project/tables/content/skill/skill.cue", rowID: "skill_01", column: "tags", expression: '["战士", "主动", "教学验证"]'}},
    ]);
});
