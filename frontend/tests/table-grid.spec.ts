import {expect, test} from "@playwright/test";

test.use({viewport: {width: 1920, height: 1080}});

const table = {
    path: "/project/tables/item.cue",
    namespace: "rpg",
    name: "item",
    columns: [
        {name: "tags", schema: {kind: "array"}},
        {name: "effectID", schema: {kind: "string"}, foreignKeyTable: "rpg.effect", foreignKeyOptions: ["burn", "freeze"]},
        {name: "name", schema: {kind: "string"}},
        {name: "position", schema: {kind: "struct", fields: [{name: "x"}, {name: "y"}, {name: "label", optional: true}]}},
        {name: "rarity", schema: {kind: "enum", enumValues: [{expression: `"N"`}, {expression: `"R"`}]}},
    ],
    rows: [{id: "sword", values: {tags: {expression: `["starter", "weapon"]`, display: `["starter", "weapon"]`}, position: {expression: `{x: 4, y: 8}`, display: `{x: 4, y: 8}`}, name: {expression: "20 + 5", display: "25", formula: true}, rarity: {expression: `"N"`, display: "N"}, effectID: {expression: `"burn"`, display: "burn"}}}],
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

    await page.mouse.click(box.x + 565, box.y + 50);
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

    await page.mouse.dblclick(box.x + 565, box.y + 50);

    await expect(page.locator(".table-document")).toHaveAttribute("data-inline-editor", "open");
    await expect(page.locator("textarea")).toHaveValue("20 + 5");
    expect(pageErrors).toEqual([]);
});

test("offers existing IDs while allowing manual foreign-key input", async ({page}) => {
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;

    await page.mouse.click(box.x + 395, box.y + 50);

    const input = page.getByRole("combobox", {name: "Select rpg.effect ID"});
    await expect(input).toHaveValue("burn");
    expect(await page.locator("datalist option").evaluateAll((options) => options.map((option) => (option as HTMLOptionElement).value))).toEqual(["burn", "freeze"]);
    await input.fill("unreleased_effect");
    await input.press("Enter");
    await expect(page.locator(".table-document")).toHaveAttribute("data-inline-editor", "closed");
});

test("edits, reorders, and adds array items", async ({page}) => {
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;

    await page.mouse.click(box.x + 225, box.y + 50);
    await expect(page.getByRole("dialog", {name: "Edit tags"})).toBeVisible();
    await page.getByRole("button", {name: "Move item 2 up"}).click();
    await page.getByRole("textbox", {name: "Paste array items"}).fill(`"rare"\n"equipped"`);
    await page.getByRole("button", {name: "Add pasted items"}).click();
    await expect(page.getByRole("textbox", {name: "Array item 1"})).toHaveValue(`"weapon"`);
    await expect(page.getByRole("textbox", {name: "Array item 4"})).toHaveValue(`"equipped"`);
    await page.getByRole("button", {name: "Apply"}).click();
    await expect(page.locator(".table-document")).toHaveAttribute("data-inline-editor", "closed");
});

test("edits flat struct fields", async ({page}) => {
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;

    await page.mouse.click(box.x + 735, box.y + 50);
    await expect(page.getByRole("dialog", {name: "Edit position"})).toBeVisible();
    await expect(page.getByRole("textbox", {name: "Struct field x"})).toHaveValue("4");
    await page.getByRole("textbox", {name: "Struct field y"}).fill("12");
    await page.getByRole("textbox", {name: "Struct field label"}).fill(`"spawn"`);
    await page.getByRole("button", {name: "Apply"}).click();
    await expect(page.locator(".table-document")).toHaveAttribute("data-inline-editor", "closed");
});
