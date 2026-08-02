import {expect, test} from "@playwright/test";

const table = {
    path: "/project/tables/item.cue",
    namespace: "rpg",
    name: "item",
    columns: [
        {name: "name", schema: {kind: "string"}},
        {name: "rarity", schema: {kind: "enum", enumValues: [{expression: `"N"`}, {expression: `"R"`}]}},
    ],
    rows: [{id: "sword", values: {name: {expression: "20 + 5", display: "25", formula: true}, rarity: {expression: `"N"`, display: "N"}}}],
};

const workspace = {
    project: {path: "/project/game.dtproj", root: "/project"},
    settings: {build: {mode: "manual"}},
    tree: {root: {kind: "project", name: "game.dtproj", children: [{kind: "tableData", name: "rpg.item", path: table.path}]}},
    schemas: {definitions: []},
    foreignKeys: {diagnostics: []},
};

test.beforeEach(async ({page}) => {
    await page.addInitScript(({table, workspace}) => {
        const api = {
            SelectProject: () => Promise.resolve(workspace),
            GetTable: () => Promise.resolve(table),
            UpdateCell: () => Promise.resolve({table, workspace}),
            SetCellFormula: () => Promise.resolve({table, workspace}),
            CreateRow: () => Promise.resolve({table, workspace}),
            ClearCell: () => Promise.resolve({table, workspace}),
            DeleteRow: () => Promise.resolve({table, workspace}),
            AddColumn: () => Promise.resolve({table, workspace}),
            DeleteColumn: () => Promise.resolve({table, workspace}),
            MoveColumn: () => Promise.resolve({table, workspace}),
            RenameColumn: () => Promise.resolve({table, workspace}),
            RenameRow: () => Promise.resolve({table, workspace}),
            Paste: () => Promise.resolve({table, workspace, canUndo: true, canRedo: false}),
            Undo: () => Promise.resolve({table, workspace, canUndo: false, canRedo: true}),
            Redo: () => Promise.resolve({table, workspace, canUndo: true, canRedo: false}),
            Export: () => Promise.resolve({files: []}),
        };
        // Wails generated bindings resolve methods through this object.
        (window as Window & {go: unknown}).go = {main: {App: api}};
    }, {table, workspace});
    await page.goto("/");
    await page.locator("header").getByRole("button", {name: "Open project"}).click();
    await page.getByRole("button", {name: "rpg.item"}).click();
});

test("supports pointer editing, row menus, and selected-column actions", async ({page}) => {
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    const gridSize = await page.locator(".table-grid").evaluate((element) => {
        const rect = element.getBoundingClientRect();
        return {width: rect.width, height: rect.height, canvasCount: element.querySelectorAll("canvas").length};
    });
    expect(box, JSON.stringify(gridSize)).not.toBeNull();
    if (!box) return;
    const canvases = await page.locator("canvas").evaluateAll((elements) => elements.map((element) => {
        const rect = element.getBoundingClientRect();
        const style = getComputedStyle(element);
        return {x: rect.x, y: rect.y, width: rect.width, height: rect.height, display: style.display, zIndex: style.zIndex, pointerEvents: style.pointerEvents};
    }));

    // Header then the first data row. The marker is the left-most grid column.
    await page.mouse.click(box.x + 225, box.y + 15);
    await expect(page.getByRole("button", {name: "Rename column"})).toBeEnabled();
    await expect(page.getByRole("button", {name: "Move column"})).toBeEnabled();
    await expect(page.getByRole("button", {name: "Delete column"})).toBeEnabled();

    await page.mouse.click(box.x + 225, box.y + 50);
    await expect(page.getByRole("button", {name: "Rename column"})).toBeDisabled();
    await expect(page.locator(".table-document")).toHaveAttribute("data-inline-editor", "open");
    await expect(page.locator("textarea"), JSON.stringify(canvases)).toBeVisible();

    await page.mouse.click(box.x + 16, box.y + 50, {button: "right"});
    await expect(page.getByRole("button", {name: "Delete row"})).toBeVisible();
    await page.mouse.click(box.x + 225, box.y + 50, {button: "right"});
    await expect(page.getByRole("button", {name: "Delete row"})).toBeHidden();
});

test("edits formula cells with the application inline editor", async ({page}) => {
    const pageErrors: Error[] = [];
    page.on("pageerror", (error) => pageErrors.push(error));
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;

    await page.mouse.dblclick(box.x + 225, box.y + 50);

    await expect(page.locator(".table-document")).toHaveAttribute("data-inline-editor", "open");
    await expect(page.locator("textarea")).toHaveValue("20 + 5");
    expect(pageErrors).toEqual([]);
});
