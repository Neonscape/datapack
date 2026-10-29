import {expect, test, type Page} from "@playwright/test";
import type {NavRequest} from "../src/types";

// Exercise the real App and Glide grids; only the Wails process boundary is stubbed.
test.use({viewport: {width: 1600, height: 1000}});

declare global {
    interface Window {
        __navigationTest: {calls: NavRequest[]};
    }
}

const tablePath = (name: string) => `/project/tables/${name}.cue`;

async function selectCell(page: Page, column: number, row: number) {
    const canvas = page.locator(".tab-panel:visible canvas").first();
    await expect(canvas).toBeVisible();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    await page.mouse.click(box!.x + (column === 0 ? 85 : 225 + (column - 1) * 170), box!.y + 50 + row * 30);
    await expect.poll(() => page.locator(".tab-panel:visible .table-grid").evaluate((grid) => grid.contains(document.activeElement))).toBe(true);
}

async function openBothWithSelections(page: Page) {
    await page.getByRole("button", {name: "alpha", exact: true}).click();
    await selectCell(page, 0, 0);
    await page.getByRole("button", {name: "beta", exact: true}).click();
    await selectCell(page, 2, 1);
    await expect(page.locator(".table-document")).toHaveCount(2);
}

// Synthetic cancelable events let us check native ownership without opening
// browser devtools in the deliberately unhandled cases.
async function dispatchF12(page: Page, alreadyPrevented = false) {
    return page.evaluate((prevented) => {
        const event = new KeyboardEvent("keydown", {key: "F12", code: "F12", bubbles: true, cancelable: true});
        if (prevented) event.preventDefault();
        (document.activeElement ?? document.body).dispatchEvent(event);
        return event.defaultPrevented;
    }, alreadyPrevented);
}

async function navigationCalls(page: Page) {
    return page.evaluate(() => window.__navigationTest.calls);
}

test.beforeEach(async ({page}) => {
    await page.addInitScript(() => {
        const makeTable = (name: string) => ({
            path: `/project/tables/${name}.cue`, namespace: "game", name,
            columns: [
                {name: "name", schemaExpression: "string", schema: {kind: "string"}},
                {name: "points", schemaExpression: "int", schema: {kind: "integer"}},
                {name: "target", schemaExpression: "string", schema: {kind: "string"}, foreignKeyTable: "game.alpha", foreignKeyOptions: ["alpha1", "alpha2"]},
            ],
            rows: [1, 2].map((number) => ({id: `${name}${number}`, values: {
                name: {expression: JSON.stringify(`${name} ${number}`), display: `${name} ${number}`},
                points: {expression: `10 + ${number}`, display: String(10 + number), formula: true},
                target: {expression: '"alpha1"', display: "alpha1"},
            }})),
            diagnostics: [], localSchema: [],
        });
        const tables = [makeTable("alpha"), makeTable("beta")];
        const workspace = {
            project: {path: "/project/game.dtproj", root: "/project"},
            settings: {build: {mode: "manual"}},
            tree: {root: {kind: "project", name: "game.dtproj", children: tables.map((table) => ({kind: "table", name: table.name, path: table.path}))}},
            schemas: {definitions: [{id: "common.#Label", kind: "type", path: "/project/schema/common.cue", usages: []}], diagnostics: []},
            foreignKeys: {diagnostics: []},
            compile: {mode: "manual", pending: false, diagnostics: []},
            canUndo: false, canRedo: false,
        };
        window.__navigationTest = {calls: []};
        (window as unknown as Window & {go: unknown}).go = {main: {App: {
            GetSettings: async () => ({language: "en", theme: "light"}),
            GetExportFormats: async () => [],
            SelectProject: async () => workspace,
            GetTable: async (path: string) => tables.find((table) => table.path === path),
            Navigate: async (request: NavRequest) => {
                window.__navigationTest.calls.push(request);
                // A reference popover records the request without moving or
                // refocusing either grid, so retained selections remain testable.
                return {...request, references: [{kind: "cell", path: request.path, rowID: request.rowID, column: request.column}]};
            },
        }}};
    });
    await page.goto("/");
    await page.getByRole("button", {name: "File", exact: true}).click();
    await page.getByLabel("Application menu").getByRole("button", {name: "Open project", exact: true}).click();
});

test("F12 navigates exactly once from the active focused grid and retains each tab's selection", async ({page}) => {
    await openBothWithSelections(page);
    expect(await dispatchF12(page)).toBe(true);
    await expect.poll(() => navigationCalls(page)).toEqual([
        {kind: "formula-reference", path: tablePath("beta"), rowID: "beta2", column: "points"},
    ]);

    await page.getByRole("tab", {name: "game.alpha"}).click();
    await page.locator(".tab-panel:visible canvas").first().focus();
    expect(await dispatchF12(page)).toBe(true);
    await expect.poll(() => navigationCalls(page)).toEqual([
        {kind: "formula-reference", path: tablePath("beta"), rowID: "beta2", column: "points"},
        {kind: "primary-key-references", path: tablePath("alpha"), rowID: "alpha1"},
    ]);
});

test("F12 leaves inline textarea, formula input, and foreign-key select in control", async ({page}) => {
    await openBothWithSelections(page);
    const canvas = page.locator(".tab-panel:visible canvas").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    await page.mouse.dblclick(box!.x + 395, box!.y + 80);
    const inline = page.locator(".inline-cell-editor:visible");
    await expect(inline).toBeFocused();
    expect(await dispatchF12(page)).toBe(false);
    await expect(inline).toBeFocused();
    expect(await navigationCalls(page)).toEqual([]);
    await inline.press("Escape");

    const formula = page.getByRole("textbox", {name: "Formula bar"}).filter({visible: true});
    await formula.click();
    await expect(formula).toBeFocused();
    expect(await dispatchF12(page)).toBe(false);
    expect(await navigationCalls(page)).toEqual([]);
    await formula.press("Escape");

    // Use a fresh canvas position after editing overlays have closed.
    const selectBox = await canvas.boundingBox();
    expect(selectBox).not.toBeNull();
    await page.mouse.dblclick(selectBox!.x + 565, selectBox!.y + 50);
    const select = page.locator(".foreign-key-cell-editor select:visible");
    await expect(select).toBeFocused();
    expect(await dispatchF12(page)).toBe(false);
    await expect(select).toBeFocused();
    expect(await navigationCalls(page)).toEqual([]);
});

test("F12 leaves contenteditable text and already-handled events alone inside the grid", async ({page}) => {
    await openBothWithSelections(page);
    expect(await dispatchF12(page, true)).toBe(true);
    expect(await navigationCalls(page)).toEqual([]);

    await page.locator(".tab-panel:visible .table-grid").evaluate((grid) => {
        const editor = document.createElement("div");
        editor.contentEditable = "true";
        editor.setAttribute("aria-label", "Native rich text");
        editor.innerHTML = "<span>rich text</span>";
        grid.appendChild(editor);
        editor.focus();
    });
    await expect(page.getByLabel("Native rich text")).toBeFocused();
    expect(await dispatchF12(page)).toBe(false);
    expect(await navigationCalls(page)).toEqual([]);
});

test("F12 ignores retained selections outside the grid and while schema peek hides it", async ({page}) => {
    await openBothWithSelections(page);
    await page.getByRole("button", {name: "Settings", exact: true}).focus();
    expect(await dispatchF12(page)).toBe(false);
    expect(await navigationCalls(page)).toEqual([]);

    await page.locator(".schema-row").filter({hasText: "common.#Label"}).click();
    await expect(page.locator(".schema-detail")).toBeVisible();
    expect(await dispatchF12(page)).toBe(false);
    expect(await navigationCalls(page)).toEqual([]);
});

test("F12 is not intercepted for plain values or the draft row without navigation targets", async ({page}) => {
    await openBothWithSelections(page);
    await selectCell(page, 1, 0);
    expect(await dispatchF12(page)).toBe(false);
    expect(await navigationCalls(page)).toEqual([]);
    await selectCell(page, 0, 2);
    expect(await dispatchF12(page)).toBe(false);
    expect(await navigationCalls(page)).toEqual([]);
});
