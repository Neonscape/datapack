import {expect, test, type Page} from "@playwright/test";

test.use({viewport: {width: 1600, height: 1000}});
type ChineseCall = {method: string; argument: Record<string, unknown>};
declare global {interface Window {__chineseCalls: ChineseCall[]}}

async function cell(page: Page, column: number, double = false) {
    const canvas = page.locator(".tab-panel:visible canvas").first();
    await expect(canvas).toBeVisible();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    const x = box!.x + 225 + (column - 1) * 170;
    if (double) await page.mouse.dblclick(x, box!.y + 50);
    else await page.mouse.click(x, box!.y + 50);
    return box!;
}

test.beforeEach(async ({page}) => {
    await page.addInitScript(() => {
        const path = "/project/tables/item/item.cue";
        const table = {
            path, namespace: "probe", name: "item", title: "中文商品探针",
            columns: [
                {name: "名称", schemaExpression: "common.#名称", schema: {kind: "reference", reference: "common.#名称"}, effectiveSchema: {kind: "string"}},
                {name: "售价", schemaExpression: "int & >=0", schema: {kind: "integer"}},
                {name: "实售价", schemaExpression: "int", schema: {kind: "integer"}},
                {name: "效果ID", schemaExpression: "string", schema: {kind: "string"}, foreignKeyTable: "probe.effect", foreignKeyOptions: ["burn", "mend"]},
            ],
            rows: [{id: "one", values: {
                名称: {expression: '"赤砂短剑"', display: "赤砂短剑"}, 售价: {expression: "100", display: "100"},
                实售价: {expression: "ceil(售价 * 0.8)", display: "80", formula: true}, 效果ID: {expression: '"burn"', display: "burn"},
            }}],
            localSchema: [], cellComments: {one: {售价: "实售价按标价乘折扣向上取整。"}},
            diagnostics: [{rowID: "one", column: "效果ID", message: "中文外键定位探针"}],
        };
        const workspace = {
            project: {path: "/project/probe.dtproj", root: "/project"}, settings: {build: {mode: "manual"}},
            tree: {root: {kind: "project", name: "probe.dtproj", children: [{kind: "table", name: "item", path}]}},
            schemas: {definitions: [{id: "common.#名称", path: "/project/schema/common.cue", kind: "type", usages: []}], diagnostics: []},
            foreignKeys: {diagnostics: []}, compile: {mode: "manual", pending: false, diagnostics: []}, canUndo: false, canRedo: false,
        };
        window.__chineseCalls = [];
        const record = (method: string) => async (argument: Record<string, unknown>) => {
            window.__chineseCalls.push({method, argument});
            return {table, workspace};
        };
        (window as unknown as Window & {go: unknown}).go = {main: {App: {
            GetSettings: async () => ({language: "en", theme: "light"}), GetExportFormats: async () => [],
            SelectProject: async () => workspace, GetTable: async () => table,
            SetCellText: record("SetCellText"), UpdateCell: record("UpdateCell"), SetCellFormula: record("SetCellFormula"), RenameColumn: record("RenameColumn"),
        }}};
    });
    await page.goto("/");
    await page.getByRole("button", {name: "File", exact: true}).click();
    await page.getByLabel("Application menu").getByRole("button", {name: "Open project", exact: true}).click();
    await page.getByRole("button", {name: "item", exact: true}).click();
});

test("edits actual Chinese field names and formula text", async ({page}) => {
    await cell(page, 1, true);
    const editor = page.locator(".inline-cell-editor:visible");
    await expect(editor).toHaveValue("赤砂短剑");
    await editor.fill("霜月长剑");
    await editor.press("Enter");
    await expect.poll(() => page.evaluate(() => window.__chineseCalls)).toEqual([{method: "SetCellText", argument: {path: "/project/tables/item/item.cue", rowID: "one", column: "名称", expression: "霜月长剑"}}]);
    await cell(page, 3);
    await expect(page.locator(".formula-bar-cell:visible")).toHaveText("实售价 · one");
    const formula = page.getByRole("textbox", {name: "Formula bar"}).filter({visible: true});
    await expect(formula).toHaveValue("=ceil(售价 * 0.8)");
    await formula.click();
    await formula.fill("=ceil(售价 * 0.6)");
    await formula.press("Enter");
    await expect.poll(() => page.evaluate(() => window.__chineseCalls.at(-1))).toEqual({method: "SetCellFormula", argument: {path: "/project/tables/item/item.cue", rowID: "one", column: "实售价", expression: "ceil(售价 * 0.6)"}});
});

test("renames a Chinese column through the column dialog", async ({page}) => {
    const box = await cell(page, 2);
    await page.mouse.click(box.x + 395, box.y + 15, {button: "right"});
    await page.getByRole("button", {name: "Rename column", exact: true}).click();
    await page.getByLabel("New name", {exact: true}).fill("标价");
    await page.getByRole("button", {name: "Apply", exact: true}).click();
    await expect.poll(() => page.evaluate(() => window.__chineseCalls)).toEqual([{method: "RenameColumn", argument: {path: "/project/tables/item/item.cue", oldName: "售价", newName: "标价"}}]);
});

test("focuses a Chinese diagnostic and edits its foreign key", async ({page}) => {
    await page.getByRole("button", {name: /one\.效果ID.*中文外键定位探针/}).click();
    await expect(page.locator(".formula-bar-cell:visible")).toHaveText("效果ID · one");
    await cell(page, 4, true);
    await page.getByRole("combobox", {name: "Select probe.effect ID"}).selectOption("mend");
    await expect.poll(() => page.evaluate(() => window.__chineseCalls)).toEqual([{method: "UpdateCell", argument: {path: "/project/tables/item/item.cue", rowID: "one", column: "效果ID", expression: '"mend"'}}]);
});
