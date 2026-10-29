import {expect, test, type Page} from "@playwright/test";

// Exercise the real App and Glide canvases; only the Wails process boundary is stubbed.
test.use({viewport: {width: 1600, height: 1000}, permissions: ["clipboard-read", "clipboard-write"]});

type PasteCall = {path: string; rowIDs: string[]; columns: string[]; values: string[][]};
type ClipboardState = {pastes: PasteCall[]; edits: unknown[]};

declare global {
    interface Window {
        __clipboardTest: ClipboardState;
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
    await selectCell(page, 1, 0);
    await page.getByRole("button", {name: "beta", exact: true}).click();
    await selectCell(page, 2, 1);
    await expect(page.locator(".table-document")).toHaveCount(2);
}

async function paste(page: Page, text: string) {
    await page.evaluate((value) => navigator.clipboard.writeText(value), text);
    await page.keyboard.press("ControlOrMeta+V");
}

async function clipboardState(page: Page) {
    return page.evaluate(() => window.__clipboardTest);
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
                points: {expression: `${name === "alpha" ? 10 : 20} + ${number}`, display: String((name === "alpha" ? 10 : 20) + number), formula: true},
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
        const state: ClipboardState = {pastes: [], edits: []};
        const resultFor = (path: string) => ({table: tables.find((table) => table.path === path), workspace});
        const edit = (argument: {path: string}) => {
            state.edits.push(argument);
            return Promise.resolve(resultFor(argument.path));
        };
        window.__clipboardTest = state;
        (window as unknown as Window & {go: unknown}).go = {main: {App: {
            GetSettings: async () => ({language: "en", theme: "light"}),
            GetExportFormats: async () => [],
            SelectProject: async () => workspace,
            GetTable: async (path: string) => resultFor(path).table,
            Paste: async (argument: PasteCall) => {
                state.pastes.push(argument);
                return resultFor(argument.path);
            },
            UpdateCell: edit, SetCellFormula: edit, SetCellText: edit, ClearCell: edit, RenameRow: edit,
        }}};
    });
    await page.goto("/");
    await page.getByRole("button", {name: "File", exact: true}).click();
    await page.getByLabel("Application menu").getByRole("button", {name: "Open project", exact: true}).click();
});

test("pastes exactly once into the focused tab with retained selections in both grids", async ({page}) => {
    await openBothWithSelections(page);
    await paste(page, "42");
    await expect.poll(() => clipboardState(page)).toEqual({pastes: [{
        path: tablePath("beta"), rowIDs: ["beta2"], columns: ["points"], values: [["42"]],
    }], edits: []});

    // Returning to A must preserve its own cell, not B's position or ownership.
    await page.getByRole("tab", {name: "game.alpha"}).click();
    await page.locator(".tab-panel:visible canvas").first().focus();
    await paste(page, '"updated alpha"');
    await expect.poll(() => clipboardState(page)).toEqual({pastes: [
        {path: tablePath("beta"), rowIDs: ["beta2"], columns: ["points"], values: [["42"]]},
        {path: tablePath("alpha"), rowIDs: ["alpha1"], columns: ["name"], values: [['"updated alpha"']]},
    ], edits: []});
});

test("copies expressions only from the active focused grid after switching tabs", async ({page}) => {
    await openBothWithSelections(page);
    await page.keyboard.press("ControlOrMeta+C");
    await expect.poll(() => page.evaluate(() => navigator.clipboard.readText())).toBe("=20 + 2");
    await page.getByRole("tab", {name: "game.alpha"}).click();
    await page.locator(".tab-panel:visible canvas").first().focus();
    await page.keyboard.press("ControlOrMeta+C");
    await expect.poll(() => page.evaluate(() => navigator.clipboard.readText())).toBe('"alpha 1"');
});

test("leaves inline and formula-bar native editing in control of paste", async ({page}) => {
    await openBothWithSelections(page);
    const canvas = page.locator(".tab-panel:visible canvas").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    await page.mouse.dblclick(box!.x + 225, box!.y + 50);
    const inline = page.locator(".inline-cell-editor:visible");
    await expect(inline).toBeFocused();
    await paste(page, "native inline text");
    await expect(inline).toHaveValue("native inline text");
    expect(await clipboardState(page)).toEqual({pastes: [], edits: []});
    await inline.press("Escape");

    const formula = page.getByRole("textbox", {name: "Formula bar"}).filter({visible: true});
    await formula.click();
    await formula.press("ControlOrMeta+A");
    await paste(page, "native formula text");
    await expect(formula).toHaveValue("native formula text");
    expect(await clipboardState(page)).toEqual({pastes: [], edits: []});
    await formula.press("Escape");
});

test("does not hijack contenteditable text even inside the grid", async ({page}) => {
    await openBothWithSelections(page);
    await page.locator(".tab-panel:visible .table-grid").evaluate((grid) => {
        const editor = document.createElement("div");
        editor.contentEditable = "true";
        editor.setAttribute("aria-label", "Native rich text");
        editor.innerHTML = "<span>original rich text</span>";
        grid.appendChild(editor);
        editor.focus();
        const selection = window.getSelection()!;
        selection.selectAllChildren(editor);
    });
    await page.keyboard.press("ControlOrMeta+C");
    await expect.poll(() => page.evaluate(() => navigator.clipboard.readText())).toBe("original rich text");
    await paste(page, "native rich text");
    await expect(page.getByLabel("Native rich text")).toHaveText("native rich text");
    expect(await clipboardState(page)).toEqual({pastes: [], edits: []});
});

test("does not use retained grid selections while focus is outside the grid or a schema peek hides it", async ({page}) => {
    await openBothWithSelections(page);
    await page.getByRole("button", {name: "Settings", exact: true}).focus();
    await paste(page, "outside");
    await page.keyboard.press("ControlOrMeta+C");
    expect(await page.evaluate(() => navigator.clipboard.readText())).toBe("outside");
    expect(await clipboardState(page)).toEqual({pastes: [], edits: []});

    await page.locator(".schema-row").filter({hasText: "common.#Label"}).click();
    await expect(page.locator(".schema-detail")).toBeVisible();
    await paste(page, "peek");
    expect(await clipboardState(page)).toEqual({pastes: [], edits: []});
});

test("keeps foreign-key cells protected from custom and Glide paste", async ({page}) => {
    await openBothWithSelections(page);
    await selectCell(page, 3, 0);
    await paste(page, "unknown target");
    expect(await clipboardState(page)).toEqual({pastes: [], edits: []});
});


test("preserves the focused grid's single-cell ID and draft paste fallback", async ({page}) => {
    await openBothWithSelections(page);
    await selectCell(page, 0, 0);
    await paste(page, "renamed_beta");
    await expect.poll(() => clipboardState(page)).toEqual({pastes: [], edits: [{
        path: tablePath("beta"), oldID: "beta1", newID: "renamed_beta",
    }]});

    // A draft ID stays local until the row has values; pasting it must not
    // invoke Paste or another cell write in either mounted grid.
    await selectCell(page, 0, 2);
    await paste(page, "draft_beta");
    await expect(page.getByRole("textbox", {name: "Formula bar"}).filter({visible: true})).toHaveValue("draft_beta");
    expect(await clipboardState(page)).toEqual({pastes: [], edits: [{
        path: tablePath("beta"), oldID: "beta1", newID: "renamed_beta",
    }]});
});
