import {expect, test} from "@playwright/test";

test.use({viewport: {width: 1920, height: 1080}});

const table = {
    path: "/project/tables/item.cue",
    namespace: "rpg",
    name: "item",
    columns: [
        {name: "tags", schemaExpression: "[...string]", schema: {kind: "array"}},
        {name: "effectID", schemaExpression: "string", schema: {kind: "string"}, foreignKeyTable: "rpg.effect", foreignKeyOptions: ["burn", "freeze"]},
        {name: "name", schemaExpression: "string", schema: {kind: "string"}},
        {name: "position", schemaExpression: "{x: int, y: int, label?: string}", schema: {kind: "struct", fields: [{name: "x"}, {name: "y"}, {name: "label", optional: true}]}},
        {name: "rarity", schemaExpression: `"N" | "R"`, schema: {kind: "enum", enumValues: [{expression: `"N"`}, {expression: `"R"`}]}},
    ],
    rows: [{id: "sword", values: {tags: {expression: `["starter", "weapon"]`, display: `["starter", "weapon"]`}, position: {expression: `{x: 4, y: 8}`, display: `{x: 4, y: 8}`}, name: {expression: "20 + 5", display: "25", formula: true}, rarity: {expression: `"N"`, display: "N"}, effectID: {expression: `"burn"`, display: "burn"}}}],
    diagnostics: [{rowID: "sword", column: "effectID", message: "foreign key references missing row rpg.effect.burn"}, {rowID: "", message: "table has an invalid export configuration"}],
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
            CreateTable: () => Promise.resolve({table, workspace}),
            ClearCell: () => Promise.resolve({table, workspace}),
            DeleteRow: () => Promise.resolve({table, workspace}),
            AddColumn: () => Promise.resolve({table, workspace}),
            DeleteColumn: () => Promise.resolve({table, workspace}),
            MoveColumn: () => Promise.resolve({table, workspace}),
            RenameColumn: () => Promise.resolve({table, workspace}),
            SetColumnSchema: () => Promise.resolve({table, workspace}),
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
    await page.getByRole("button", {name: "File"}).click();
    await page.getByRole("navigation", {name: "Application menu"}).getByRole("button", {name: "Open project"}).click();
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
    await page.mouse.click(box.x + 225, box.y + 15, {button: "right"});
    await expect(page.getByRole("button", {name: "Add to left"})).toBeVisible();
    await expect(page.getByRole("button", {name: "Edit schema"})).toBeVisible();

    await page.mouse.click(box.x + 565, box.y + 50);
    await expect(page.locator(".table-document")).toHaveAttribute("data-inline-editor", "open");
    await expect(page.locator("textarea"), JSON.stringify(canvases)).toBeVisible();

    await page.mouse.click(box.x + 16, box.y + 50, {button: "right"});
    await expect(page.getByRole("button", {name: "Delete row"})).toBeVisible();
    await page.mouse.click(box.x + 225, box.y + 50, {button: "right"});
    await expect(page.getByRole("button", {name: "Delete row"})).toBeHidden();
});

test("edits formula cells with the application inline editor", async ({page}) => {
    const pageErrors: Error[] = [];
    const calls: string[] = [];
    page.on("pageerror", (error) => pageErrors.push(error));
    await page.evaluate(() => {
        const api = (window as Window & {go: {main: {App: Record<string, (...args: unknown[]) => unknown>}}}).go.main.App;
        const setFormula = api.SetCellFormula;
        const updateCell = api.UpdateCell;
        api.SetCellFormula = (...args) => { (window as Window & {formulaCalls: string[]}).formulaCalls.push("formula"); return setFormula(...args); };
        api.UpdateCell = (...args) => { (window as Window & {formulaCalls: string[]}).formulaCalls.push("value"); return updateCell(...args); };
        (window as Window & {formulaCalls: string[]}).formulaCalls = [];
    });
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;

    await page.mouse.dblclick(box.x + 565, box.y + 50);

    await expect(page.locator(".table-document")).toHaveAttribute("data-inline-editor", "open");
    await expect(page.locator("textarea")).toHaveValue("=20 + 5");
    await page.mouse.click(box.x + 1000, box.y + 200);
    await expect(page.locator(".table-document")).toHaveAttribute("data-inline-editor", "closed");
    expect(await page.evaluate(() => (window as Window & {formulaCalls: string[]}).formulaCalls)).toEqual(["formula"]);
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

test("shows table diagnostics and focuses their cells", async ({page}) => {
    await expect(page.getByRole("region", {name: "Table diagnostics"})).toBeVisible();
    const diagnostic = page.getByRole("button", {name: /sword\.effectID.*foreign key references missing row/});
    await diagnostic.click();
    await expect(page.locator(".table-diagnostics-title")).toContainText("2");
    await expect(page.getByRole("button", {name: /table.*invalid export configuration/})).toBeDisabled();
});

test("edits the selected column Schema", async ({page}) => {
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;
    await page.mouse.click(box.x + 565, box.y + 15, {button: "right"});
    await page.getByRole("button", {name: "Edit schema"}).click();
    await expect(page.getByRole("heading", {name: "Edit column schema"})).toBeVisible();
    const schema = page.getByRole("textbox", {name: "CUE Schema"});
    await schema.fill("int & >=20");
    await page.getByRole("button", {name: "Apply"}).click();
    await expect(page.getByRole("heading", {name: "Edit column schema"})).toBeHidden();
});

test("creates and opens a table from the project explorer", async ({page}) => {
    await page.getByRole("button", {name: "File"}).click();
    await page.getByRole("button", {name: "New table"}).click();
    await expect(page.getByRole("heading", {name: "Create table"})).toBeVisible();
    await page.getByRole("textbox", {name: "Namespace"}).fill("combat");
    await page.getByRole("textbox", {name: "Table name"}).fill("skill");
    await page.getByRole("button", {name: "Create", exact: true}).click();
    await expect(page.getByRole("heading", {name: "Create table"})).toBeHidden();
});

test("groups history commands under Edit", async ({page}) => {
    await page.getByRole("button", {name: "Edit"}).click();
    await expect(page.getByRole("button", {name: /Undo/})).toBeEnabled();
    await expect(page.getByRole("button", {name: /Redo/})).toBeEnabled();
    await page.keyboard.press("Escape");
    await expect(page.getByRole("button", {name: /Undo/})).toBeHidden();
});
