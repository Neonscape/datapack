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
        {name: "testAttack", schemaExpression: "_local.#Position", schema: {kind: "reference"}},
    ],
    rows: [{id: "sword", values: {tags: {expression: `["starter", "weapon"]`, display: `["starter", "weapon"]`}, position: {expression: `{x: 4, y: 8}`, display: `{x: 4, y: 8}`}, name: {expression: "20 + 5", display: "25", formula: true}, rarity: {expression: `"N"`, display: "N"}, effectID: {expression: `"burn"`, display: "burn"}}}],
    diagnostics: [{rowID: "sword", column: "effectID", message: "foreign key references missing row rpg.effect.burn"}, {rowID: "", message: "table has an invalid export configuration"}],
};

const publicSchema = {
    path: "/project/schema/common.cue",
    package: "common",
    definitions: [
        {name: "GameID", kind: "type", type: "string", schema: {kind: "string"}, schemaExpression: "string"},
        {name: "NonNegative", kind: "constraint", appliesTo: ["integer", "number"], schema: {kind: "unknown"}, schemaExpression: ">=0"},
    ],
};

const workspace = {
    project: {path: "/project/game.dtproj", root: "/project"},
    settings: {build: {mode: "manual"}},
    tree: {root: {kind: "project", name: "game.dtproj", children: [
        {kind: "tableData", name: "rpg.item", path: table.path},
        {kind: "schemaDefinition", name: "#GameID", path: publicSchema.path},
        {kind: "schemaFile", name: "common", path: publicSchema.path, children: [
            {kind: "schemaDefinition", name: "#GameID", path: publicSchema.path},
            {kind: "schemaDefinition", name: "#NonNegative", path: publicSchema.path},
        ]},
    ]}},
    schemas: {definitions: [
        {id: "common.#GameID", kind: "type", path: publicSchema.path, usages: [{definition: "common.#GameID", path: table.path, table: "rpg.item", field: "#PrimaryKey"}]},
        {id: "common.#NonNegative", kind: "constraint", path: publicSchema.path, usages: []},
        {id: "rpg.item._local.#Position", kind: "type", path: table.path, table: "rpg.item", usages: []},
    ]},
    foreignKeys: {diagnostics: []},
    compile: {mode: "manual", pending: false, compiledAt: "2026-08-13T12:00:00Z", diagnostics: []},
};

test.beforeEach(async ({page}) => {
    await page.addInitScript(({table, workspace, publicSchema}) => {
        const api = {
            SelectProject: () => Promise.resolve(workspace),
            GetTable: () => Promise.resolve(table),
            GetPublicSchema: () => Promise.resolve(publicSchema),
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
            CreateSchemaDefinition: (argument) => {
                publicSchema.definitions.push({name: argument.name, kind: argument.kind, type: argument.type, appliesTo: argument.appliesTo, schema: {kind: "string"}, schemaExpression: argument.schemaExpression});
                return Promise.resolve({schema: publicSchema, table, workspace});
            },
            SetSchemaDefinition: (argument) => {
                const definition = publicSchema.definitions.find((entry) => entry.name === argument.name);
                if (definition) {
                    definition.kind = argument.kind;
                    definition.type = argument.type;
                    definition.appliesTo = argument.appliesTo;
                    definition.schemaExpression = argument.schemaExpression;
                }
                return Promise.resolve({schema: publicSchema, table, workspace});
            },
            RenameSchemaDefinition: (argument) => {
                const definition = publicSchema.definitions.find((entry) => entry.name === argument.oldName);
                if (definition) definition.name = argument.newName;
                return Promise.resolve({schema: publicSchema, table, workspace});
            },
            DeleteSchemaDefinition: (argument) => {
                publicSchema.definitions = publicSchema.definitions.filter((entry) => entry.name !== argument.name);
                return Promise.resolve({schema: publicSchema, table, workspace});
            },
            SetCompileMode: (mode) => Promise.resolve({...workspace, settings: {...workspace.settings, build: {...workspace.settings.build, mode}}, compile: {...workspace.compile, mode}}),
            Compile: () => Promise.resolve({workspace: {...workspace, compile: {...workspace.compile, pending: false}}}),
        };
        // Wails generated bindings resolve methods through this object.
        (window as Window & {go: unknown}).go = {main: {App: api}};
        (window as Window & {__mock: unknown}).__mock = {table, workspace};
    }, {table, workspace, publicSchema});
    await page.goto("/");
    await page.getByRole("button", {name: "File"}).click();
    await page.getByRole("navigation", {name: "Application menu"}).getByRole("button", {name: "Open project"}).click();
    await page.getByRole("button", {name: "rpg.item", exact: true}).click();
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

    await page.mouse.dblclick(box.x + 565, box.y + 50);
    await expect(page.locator(".table-document")).toHaveAttribute("data-inline-editor", "open");
    await expect(page.locator("textarea"), JSON.stringify(canvases)).toBeVisible();

    await page.mouse.click(box.x + 16, box.y + 50, {button: "right"});
    await expect(page.getByRole("button", {name: "Delete row"})).toBeVisible();
    await page.mouse.click(box.x + 225, box.y + 50, {button: "right"});
    await expect(page.getByRole("button", {name: "Delete row"})).toBeHidden();
});

test("uses the column menu for the ID header", async ({page}) => {
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;

    await page.mouse.click(box.x + 85, box.y + 15, {button: "right"});
    await expect(page.getByRole("button", {name: "Add to right"})).toBeEnabled();
    await expect(page.getByRole("button", {name: "Add to left"})).toBeDisabled();
    await expect(page.getByRole("button", {name: "Delete column"})).toBeDisabled();
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
    expect(await page.evaluate(() => (window as Window & {formulaCalls: string[]}).formulaCalls)).toEqual([]);
    expect(pageErrors).toEqual([]);
});

test("offers existing IDs while allowing manual foreign-key input", async ({page}) => {
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;

    await page.mouse.dblclick(box.x + 395, box.y + 50);

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

    await page.mouse.dblclick(box.x + 225, box.y + 50);
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

    await page.mouse.dblclick(box.x + 735, box.y + 50);
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

test("shows local references without prefix and inserts picked references", async ({page}) => {
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;
    await page.mouse.click(box.x + 1075, box.y + 15, {button: "right"});
    await page.getByRole("button", {name: "Edit schema"}).click();
    const schema = page.getByRole("textbox", {name: "CUE Schema"});
    await expect(schema).toHaveValue("#Position");
    const picker = page.getByLabel("Insert Schema reference", {exact: true});
    await expect(picker.locator("option")).toContainText(["Choose a reference", "#Position", "common.#GameID", "common.#NonNegative"]);
    await picker.selectOption("common.#NonNegative");
    await expect(schema).toHaveValue("#Position & common.#NonNegative");
    await page.getByRole("button", {name: "Apply"}).click();
    await expect(page.getByRole("heading", {name: "Edit column schema"})).toBeHidden();
});

test("toggles compile mode between manual and auto", async ({page}) => {
    await expect(page.getByRole("button", {name: "Toggle compile mode"})).toHaveText("Manual compile");
    await expect(page.getByRole("button", {name: "Compile", exact: true})).toBeDisabled();
    await page.getByRole("button", {name: "Toggle compile mode"}).click();
    await expect(page.getByRole("button", {name: "Toggle compile mode"})).toHaveText("Auto compile");
    await expect(page.getByRole("button", {name: "Compile", exact: true})).toBeHidden();
    await page.getByRole("button", {name: "Toggle compile mode"}).click();
    await expect(page.getByRole("button", {name: "Toggle compile mode"})).toHaveText("Manual compile");
});

test("compiles manually after pending edits", async ({page}) => {
    await page.evaluate(() => {
        const mock = (window as Window & {__mock: {table: unknown; workspace: {compile: {pending: boolean}}}}).__mock;
        const api = (window as Window & {go: {main: {App: Record<string, (...args: unknown[]) => unknown>}}}).go.main.App;
        api.UpdateCell = () => Promise.resolve({table: mock.table, workspace: {...mock.workspace, compile: {...mock.workspace.compile, pending: true}}});
    });
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;
    await page.mouse.dblclick(box.x + 565, box.y + 50);
    await page.locator("textarea").fill("Iron Sword");
    await page.mouse.click(box.x + 1000, box.y + 200);
    await expect(page.getByRole("button", {name: "Toggle compile mode"})).toHaveText(/pending/);
    await expect(page.getByRole("button", {name: "Compile", exact: true})).toBeEnabled();
    await page.getByRole("button", {name: "Compile", exact: true}).click();
    await expect(page.getByRole("button", {name: "Compile", exact: true})).toBeDisabled();
    await expect(page.getByRole("button", {name: "Toggle compile mode"})).not.toHaveText(/pending/);
});

test("edits a cell through the formula bar", async ({page}) => {
    await page.evaluate(() => {
        const api = (window as Window & {go: {main: {App: Record<string, (...args: unknown[]) => unknown>}}}).go.main.App;
        const updateCell = api.UpdateCell;
        (window as Window & {formulaBarCalls: {rowID: string; column: string; expression: string}[]}).formulaBarCalls = [];
        api.UpdateCell = (...args) => {
            const argument = args[0] as {rowID: string; column: string; expression: string};
            (window as Window & {formulaBarCalls: {rowID: string; column: string; expression: string}[]}).formulaBarCalls.push(argument);
            return updateCell(...args);
        };
    });
    const formulaBar = page.getByLabel("Formula bar");
    await expect(formulaBar).toBeDisabled();
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;
    await page.mouse.click(box.x + 565, box.y + 50);
    await expect(formulaBar).toBeEnabled();
    await expect(formulaBar).toHaveValue("=20 + 5");
    await formulaBar.click();
    await formulaBar.fill("Steel Sword");
    await formulaBar.blur();
    expect(await page.evaluate(() => (window as Window & {formulaBarCalls: {path: string; rowID: string; column: string; expression: string}[]}).formulaBarCalls)).toEqual([{path: "/project/tables/item.cue", rowID: "sword", column: "name", expression: "Steel Sword"}]);
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

test("keeps navigation visible after creating a draft row", async ({page}) => {
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;
    await page.mouse.dblclick(box.x + 55, box.y + 80);
    await page.locator("textarea").fill("new_row");
    await page.locator("textarea").press("Enter");
    await expect(page.getByRole("button", {name: "rpg.item", exact: true})).toBeVisible();
    await expect(page.getByText("Schemas", {exact: true})).toBeVisible();
});

test("shows backend log entries in the GUI panel", async ({page}) => {
    await page.evaluate(() => window.dispatchEvent(new CustomEvent("log:entry", {detail: {
        time: "2026-08-13T12:00:00Z",
        level: "error",
        message: "workspace refresh failed",
        fields: {path: "tables/item.cue"},
    }})));
    await page.getByRole("button", {name: /Logs/}).click();
    await expect(page.getByText("workspace refresh failed")).toBeVisible();
});

test("edits a public Schema definition from the project explorer", async ({page}) => {
    await page.getByRole("button", {name: "#GameID"}).first().click();
    await expect(page.getByText("Public Schema")).toBeVisible();
    const expression = page.locator(".schema-editor-form textarea");
    await expect(expression).toHaveValue("string");
    await expression.fill(`string & =~"^[a-z]+$"`);
    await page.getByRole("button", {name: "Apply"}).click();
    await expect(expression).toHaveValue(`string & =~"^[a-z]+$"`);
    await expect(page.getByText("Usages")).toBeVisible();
    await expect(page.getByText("#PrimaryKey")).toBeVisible();
});

test("creates and renames a public Schema definition", async ({page}) => {
    await page.getByRole("button", {name: "#GameID"}).first().click();
    await page.getByRole("button", {name: "New definition"}).click();
    await page.getByLabel("Name", {exact: true}).fill("Rarity");
    await page.getByLabel("Type", {exact: true}).selectOption("enum");
    await page.locator(".schema-editor-form textarea").fill(`"N" | "R"`);
    await page.getByRole("button", {name: "Create"}).click();
    await expect(page.getByRole("button", {name: "#Rarity"})).toBeVisible();
    await page.getByRole("button", {name: "Rename"}).click();
    await page.locator(".schema-rename-input").fill("RarityTier");
    await page.getByRole("button", {name: "Confirm rename"}).click();
    await expect(page.getByRole("button", {name: "#RarityTier"})).toBeVisible();
});

test("does not save an unchanged empty cell on blur", async ({page}) => {
    const pageErrors: Error[] = [];
    page.on("pageerror", (error) => pageErrors.push(error));
    await page.evaluate(() => {
        const api = (window as Window & {go: {main: {App: Record<string, (...args: unknown[]) => unknown>}}}).go.main.App;
        const clearCell = api.ClearCell;
        (window as Window & {clearCalls: number}).clearCalls = 0;
        api.ClearCell = (...args) => {
            (window as Window & {clearCalls: number}).clearCalls++;
            return clearCell(...args);
        };
    });
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;

    await page.mouse.dblclick(box.x + 1075, box.y + 50);
    await expect(page.locator("textarea")).toHaveValue("");
    await page.mouse.click(box.x + 1250, box.y + 200);
    await expect(page.locator(".table-document")).toHaveAttribute("data-inline-editor", "closed");
    expect(await page.evaluate(() => (window as Window & {clearCalls: number}).clearCalls)).toBe(0);
    expect(pageErrors).toEqual([]);
});

test("renders a workspace with null Schema collections", async ({page}) => {
    await page.evaluate(() => {
        const api = (window as Window & {go: {main: {App: Record<string, (...args: unknown[]) => unknown>}}}).go.main.App;
        const selectProject = api.SelectProject;
        api.SelectProject = async (...args) => {
            const workspace = await selectProject(...args) as {schemas: Record<string, unknown>};
            return {...workspace, schemas: {...workspace.schemas, definitions: null, diagnostics: null}};
        };
    });
    await page.getByRole("button", {name: "File"}).click();
    await page.getByRole("navigation", {name: "Application menu"}).getByRole("button", {name: "Open project"}).click();
    await expect(page.getByText("Schemas", {exact: true})).toBeVisible();
    await expect(page.locator(".schema-row")).toHaveCount(0);
});
