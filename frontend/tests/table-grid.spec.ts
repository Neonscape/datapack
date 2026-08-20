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
        {name: "testAttack", schemaExpression: "_local.#Position", schema: {kind: "reference", reference: "_local.#Position"}, effectiveSchema: {kind: "struct", fields: [{name: "x"}, {name: "y"}]}},
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
        {kind: "namespace", name: "rpg", children: [
            {kind: "group", name: "Tables", children: [
                {kind: "table", name: "item", path: table.path, children: [
                    {kind: "localSchema", name: "local_schema", path: table.path},
                ]},
            ]},
        ]},
        {kind: "namespace", name: "common", children: [
            {kind: "group", name: "Schemas", children: [
                {kind: "schemaFile", name: "common.cue", path: publicSchema.path, children: [
                    {kind: "schemaDefinition", name: "#GameID", path: publicSchema.path},
                    {kind: "schemaDefinition", name: "#NonNegative", path: publicSchema.path},
                ]},
            ]},
        ]},
    ]}},
    schemas: {definitions: [
        {id: "common.#GameID", kind: "type", path: publicSchema.path, usages: [{definition: "common.#GameID", path: table.path, table: "rpg.item", field: "#PrimaryKey"}]},
        {id: "common.#NonNegative", kind: "constraint", path: publicSchema.path, usages: []},
        {id: "rpg.item._local.#Position", kind: "type", path: table.path, table: "rpg.item", usages: []},
    ], diagnostics: [
        {path: publicSchema.path, field: "common.#NonNegative", message: "Schema reference is not a valid constraint"},
    ]},
    foreignKeys: {diagnostics: [
        {path: table.path, table: "rpg.item", rowID: "sword", column: "effectID", message: "missing target row"},
    ]},
    compile: {mode: "manual", pending: false, compiledAt: "2026-08-13T12:00:00Z", diagnostics: []},
    canUndo: true,
    canRedo: true,
};

test.beforeEach(async ({page}) => {
    await page.addInitScript(({table, workspace, publicSchema}) => {
        const api = {
            SelectProject: () => Promise.resolve(workspace),
            NewProject: () => Promise.resolve(workspace),
            SaveProjectAs: () => Promise.resolve(workspace),
            GetSettings: () => Promise.resolve({language: "en", theme: "system"}),
            GetExportFormats: () => Promise.resolve([{id: "csv", labels: {en: "CSV", "zh-CN": "CSV"}, supportsColumns: true, extensions: ["csv"]}]),
            SetTableExportSettings: () => Promise.resolve({table, workspace}),
            GetProjectSettings: () => Promise.resolve({project: {path: "/project/game.dtproj", root: "/project", tableRoots: ["tables"], schemaRoots: ["schema"], exclude: []}, export: {directory: "exports", targets: ["csv"], tables: {}}, formats: [{id: "csv", labels: {en: "CSV", "zh-CN": "CSV"}, supportsColumns: true, extensions: ["csv"]}], tables: [{id: "rpg.item", namespace: "rpg", name: "item", path: "/project/tables/item.cue", selected: true, configured: true, columns: ["id", "name"]}], workspace}),
            SetProjectExportSettings: (settings) => {
                (window as Window & {projectExportSettings?: unknown}).projectExportSettings = settings;
                return Promise.resolve({workspace});
            },
            SetLanguage: () => Promise.resolve({language: "en", theme: "system"}),
            SetTheme: () => Promise.resolve({language: "en", theme: "system"}),
            GetTable: () => Promise.resolve(table),
            GetPublicSchema: () => Promise.resolve(publicSchema),
            UpdateCell: () => Promise.resolve({table, workspace}),
            SetCellFormula: () => Promise.resolve({table, workspace}),
            SetCellComment: () => Promise.resolve({table: {...table, cellComments: {sword: {name: "comment"}}}, workspace}),
            SetCellText: () => Promise.resolve({table, workspace}),
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
            Export: () => Promise.resolve({directory: "/project/exports", files: ["/project/exports/rpg.item.csv"]}),
            OpenPath: () => Promise.resolve(),
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
            Navigate: (request) => Promise.resolve({kind: request.kind, path: table.path, rowID: request.rowID, column: request.column, definition: request.reference}),
            PrintSchema: (node) => {
                if (!node) return Promise.resolve("string");
                if (node.kind === "enum") return Promise.resolve(`"N" | "R"`);
                let expression = node.kind === "reference" ? (node.reference ?? "string") : node.kind;
                for (const requirement of (node.requirements ?? [])) {
                    if (requirement.kind === "compare") expression += ` ${requirement.op} ${requirement.value}`;
                    if (requirement.kind === "reference") expression += ` & ${requirement.reference}`;
                }
                return Promise.resolve(expression);
            },
            RenameTable: (argument) => Promise.resolve({table: {...table, namespace: argument.namespace, name: argument.name, path: `/project/tables/${argument.namespace}/${argument.name}/${argument.name}.cue`}, workspace}),
            DeleteTable: () => Promise.resolve({workspace}),
            CreateSchemaFile: (namespace) => Promise.resolve({schema: {...publicSchema, package: namespace, definitions: []}, workspace}),
        };
        // Wails generated bindings resolve methods through this object.
        (window as Window & {go: unknown}).go = {main: {App: api}};
        (window as Window & {__mock: unknown}).__mock = {table, workspace};
    }, {table, workspace, publicSchema});
    await page.goto("/");
    await page.getByRole("button", {name: "File"}).click();
    await page.getByRole("navigation", {name: "Application menu"}).getByRole("button", {name: "Open project"}).click();
    await page.getByRole("button", {name: "item", exact: true}).click();
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
    await page.mouse.click(box.x + 1000, box.y + 200);

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

test("opens the field display name editor with its stable ID", async ({page}) => {
    await expect(page.locator("canvas:visible").first()).toBeVisible();
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;
    await page.mouse.click(box.x + 225, box.y + 15, {button: "right"});
    await page.getByRole("button", {name: "Set display name"}).click();
    await expect(page.getByRole("heading", {name: "Display name"})).toBeVisible();
    await expect(page.getByText("Field name")).toContainText("tags");
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

test("offers only existing IDs in the foreign-key selector", async ({page}) => {
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;

    await page.mouse.dblclick(box.x + 395, box.y + 50);

    const input = page.getByRole("combobox", {name: "Select rpg.effect ID"});
    await expect(input).toHaveValue("burn");
    expect(await input.locator("option").evaluateAll((options) => options.map((option) => (option as HTMLOptionElement).value))).toEqual(["", "burn", "freeze"]);
    await input.selectOption("freeze");
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

test("keeps structured and foreign-key cells out of the formula bar", async ({page}) => {
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;

    await page.mouse.click(box.x + 225, box.y + 50);
    const formulaBar = page.getByRole("textbox", {name: "Formula bar"});
    await expect(formulaBar).toHaveAttribute("readonly", "");
    await page.mouse.dblclick(box.x + 225, box.y + 50);
    await expect(page.getByRole("dialog", {name: "Edit tags"})).toBeVisible();
    await page.getByRole("button", {name: "Close array editor"}).click();

    await page.mouse.click(box.x + 395, box.y + 50);
    await expect(formulaBar).toHaveAttribute("readonly", "");
    await page.mouse.dblclick(box.x + 395, box.y + 50);
    await expect(page.getByRole("combobox", {name: "Select rpg.effect ID"})).toBeVisible();
});

test("shows table diagnostics and focuses their cells", async ({page}) => {
    await expect(page.getByRole("region", {name: "Table diagnostics"})).toBeVisible();
    const diagnostic = page.getByRole("button", {name: /sword\.effectID.*foreign key references missing row/});
    await diagnostic.click();
    await expect(page.locator(".table-diagnostics-title")).toContainText("2");
    await expect(page.getByRole("button", {name: /table.*invalid export configuration/})).toBeDisabled();
});

test("adds a multiline cell comment from the cell context menu", async ({page}) => {
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;
    await page.mouse.click(box.x + 225, box.y + 50, {button: "right"});
    await page.getByRole("button", {name: "Add comment"}).click();
    await expect(page.getByRole("heading", {name: "Cell comment"})).toBeVisible();
    await page.getByRole("textbox", {name: "Comment"}).fill("第一行\n第二行");
    await page.getByRole("button", {name: "Apply", exact: true}).click();
    await expect(page.locator(".table-document")).toBeVisible();
});

test("edits the selected column Schema", async ({page}) => {
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;
    await page.mouse.click(box.x + 565, box.y + 15, {button: "right"});
    await page.getByRole("button", {name: "Edit schema"}).click();
    await expect(page.getByRole("heading", {name: "Edit column schema"})).toBeVisible();
    await page.getByRole("button", {name: "Custom expression", exact: true}).click();
    const schema = page.getByRole("textbox", {name: "CUE Schema"});
    await schema.fill("int & >=20");
    await page.getByRole("button", {name: "Apply"}).click();
    await expect(page.getByRole("heading", {name: "Edit column schema"})).toBeHidden();
});

test("edits a column Schema with the structured editor", async ({page}) => {
    await page.evaluate(() => {
        const api = (window as Window & {go: {main: {App: Record<string, (...args: unknown[]) => unknown>}}}).go.main.App;
        const setColumnSchema = api.SetColumnSchema;
        (window as Window & {schemaCalls: {schemaExpression: string}[]}).schemaCalls = [];
        api.SetColumnSchema = (...args) => {
            (window as Window & {schemaCalls: {schemaExpression: string}[]}).schemaCalls.push(args[0] as {schemaExpression: string});
            return setColumnSchema(...args);
        };
    });
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;
    await page.mouse.click(box.x + 1075, box.y + 15, {button: "right"});
    await page.getByRole("button", {name: "Edit schema"}).click();
    await expect(page.getByRole("heading", {name: "Edit column schema"})).toBeVisible();
    await page.getByRole("combobox", {name: "Base type"}).selectOption("integer");
    await page.getByRole("button", {name: "Add condition"}).click();
    await page.getByRole("combobox", {name: "Comparison operator"}).selectOption(">=");
    await page.getByRole("textbox", {name: "Comparison value"}).fill("10");
    await page.getByRole("button", {name: "Apply"}).click();
    await expect(page.getByRole("heading", {name: "Edit column schema"})).toBeHidden();
    expect(await page.evaluate(() => (window as Window & {schemaCalls: {schemaExpression: string}[]}).schemaCalls)).toEqual([{path: "/project/tables/item.cue", name: "testAttack", schemaExpression: "integer >= 10"}]);
});

test("edits enum values in the structured editor", async ({page}) => {
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;
    await page.mouse.click(box.x + 1075, box.y + 15, {button: "right"});
    await page.getByRole("button", {name: "Edit schema"}).click();
    await page.getByRole("combobox", {name: "Base type"}).selectOption("enum");
    await expect(page.getByRole("button", {name: "Add enum value"})).toBeVisible();
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
        api.SetCellText = () => Promise.resolve({table: mock.table, workspace: {...mock.workspace, compile: {...mock.workspace.compile, pending: true}}});
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
        const setCellText = api.SetCellText;
        (window as Window & {formulaBarCalls: {rowID: string; column: string; expression: string}[]}).formulaBarCalls = [];
        api.SetCellText = (...args) => {
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

test("navigates from a foreign key cell with ctrl-click", async ({page}) => {
    await page.evaluate(() => {
        const api = (window as Window & {go: {main: {App: Record<string, (...args: unknown[]) => unknown>}}}).go.main.App;
        const navigate = api.Navigate;
        (window as Window & {navCalls: unknown[]}).navCalls = [];
        api.Navigate = (...args) => {
            (window as Window & {navCalls: unknown[]}).navCalls.push(args[0]);
            return navigate(...args);
        };
    });
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;
    await page.keyboard.down("Control");
    await page.mouse.click(box.x + 395, box.y + 50);
    await page.keyboard.up("Control");
    expect(await page.evaluate(() => (window as Window & {navCalls: unknown[]}).navCalls)).toEqual([{kind: "foreign-key", path: "/project/tables/item.cue", rowID: "burn", column: "effectID"}]);
});

test("navigates from a schema-reference column with ctrl-click", async ({page}) => {
    await page.evaluate(() => {
        const api = (window as Window & {go: {main: {App: Record<string, (...args: unknown[]) => unknown>}}}).go.main.App;
        const navigate = api.Navigate;
        (window as Window & {navCalls: unknown[]}).navCalls = [];
        api.Navigate = (...args) => {
            (window as Window & {navCalls: unknown[]}).navCalls.push(args[0]);
            return navigate(...args);
        };
    });
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;
    await page.keyboard.down("Control");
    await page.mouse.click(box.x + 1075, box.y + 50);
    await page.keyboard.up("Control");
    expect(await page.evaluate(() => (window as Window & {navCalls: unknown[]}).navCalls)).toEqual([{kind: "schema-reference", path: "/project/tables/item.cue", reference: "_local.#Position"}]);
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

test("exposes project file lifecycle actions under File", async ({page}) => {
    await page.getByRole("button", {name: "File"}).click();
    await expect(page.getByRole("button", {name: "New project"})).toBeVisible();
    await expect(page.getByRole("button", {name: "Save project as"})).toBeEnabled();
    await expect(page.getByRole("button", {name: "Export", exact: true})).toBeVisible();
});

test("switches theme and language from the settings panel", async ({page}) => {
    await page.getByRole("button", {name: "Settings", exact: true}).click();
    await expect(page.getByRole("heading", {name: "Settings"})).toBeVisible();
    await page.getByLabel("Theme").selectOption("dark");
    await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");
    await page.getByLabel("Language").selectOption("zh-CN");
    await expect(page.locator("html")).toHaveAttribute("lang", "zh-CN");
    await page.getByRole("button", {name: "应用"}).click();
    await page.getByRole("button", {name: "编辑"}).click();
    await expect(page.getByRole("button", {name: /撤销/})).toBeEnabled();
});

test("configures project export table selection", async ({page}) => {
    await page.getByRole("button", {name: "File"}).click();
    await page.getByRole("button", {name: "Project settings"}).click();
    await expect(page.getByRole("heading", {name: "Project settings"})).toBeVisible();
    await page.getByRole("button", {name: "Clear all"}).click();
    await expect(page.locator(".project-settings-table-list input[type=checkbox]")).not.toBeChecked();
    await page.getByRole("button", {name: "Select all"}).click();
    await expect(page.locator(".project-settings-table-list input[type=checkbox]")).toBeChecked();
    await page.getByRole("button", {name: "Apply"}).click();
    await expect(page.getByRole("heading", {name: "Project settings"})).toBeHidden();
    await page.getByRole("button", {name: "game.dtproj", exact: true}).click();
    await expect(page.getByRole("heading", {name: "Project settings"})).toBeVisible();
    expect(await page.evaluate(() => (window as Window & {projectExportSettings?: {tables: Record<string, boolean>}}).projectExportSettings)).toEqual({
        directory: "exports",
        targets: ["csv"],
        tables: {"rpg.item": true},
    });
});

test("configures table export formats and fields", async ({page}) => {
    await page.getByRole("button", {name: "Table export settings"}).click();
    await page.getByRole("checkbox", {name: "CSV"}).check();
    await expect(page.getByText("Export columns")).toBeVisible();
    await page.getByRole("button", {name: "Clear fields"}).click();
    await page.getByRole("button", {name: "Select all fields"}).click();
    await page.getByRole("button", {name: "Apply", exact: true}).click();
});

test("undoes and redoes edits from the Edit menu", async ({page}) => {
    await page.evaluate(() => {
        const mock = (window as Window & {__mock: {workspace: {canUndo: boolean; canRedo: boolean}}}).__mock;
        const api = (window as Window & {go: {main: {App: Record<string, (...args: unknown[]) => unknown>}}}).go.main.App;
        (window as Window & {undoCalls: string[]}).undoCalls = [];
        api.Undo = () => {
            (window as Window & {undoCalls: string[]}).undoCalls.push("undo");
            return Promise.resolve({workspace: {...mock.workspace, canUndo: false, canRedo: true}, canUndo: false, canRedo: true});
        };
        api.Redo = () => {
            (window as Window & {undoCalls: string[]}).undoCalls.push("redo");
            return Promise.resolve({workspace: {...mock.workspace, canUndo: true, canRedo: false}, canUndo: true, canRedo: false});
        };
    });
    await page.getByRole("button", {name: "Edit"}).click();
    await page.getByRole("button", {name: /Undo/}).click();
    await page.getByRole("button", {name: "Edit"}).click();
    await expect(page.getByRole("button", {name: /Undo/})).toBeDisabled();
    await expect(page.getByRole("button", {name: /Redo/})).toBeEnabled();
    await page.getByRole("button", {name: /Redo/}).click();
    await page.getByRole("button", {name: "Edit"}).click();
    await expect(page.getByRole("button", {name: /Undo/})).toBeEnabled();
    await expect(page.getByRole("button", {name: /Redo/})).toBeDisabled();
    expect(await page.evaluate(() => (window as Window & {undoCalls: string[]}).undoCalls)).toEqual(["undo", "redo"]);
});

test("keeps navigation visible after creating a draft row", async ({page}) => {
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;
    await page.mouse.dblclick(box.x + 55, box.y + 80);
    await page.locator("textarea").fill("new_row");
    await page.locator("textarea").press("Enter");
    await expect(page.getByRole("button", {name: "item", exact: true})).toBeVisible();
    await expect(page.locator(".schema-panel .panel-title")).toBeVisible();
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
    await page.getByRole("button", {name: "Custom expression", exact: true}).click();
    const expression = page.locator(".schema-editor-form textarea");
    await expect(expression).toHaveValue("string");
    await expression.fill(`string & =~"^[a-z]+$"`);
    await page.getByRole("button", {name: "Apply"}).click();
    await expect(page.getByText("Usages")).toBeVisible();
    await expect(page.getByText("#PrimaryKey")).toBeVisible();
});

test("creates and renames a public Schema definition", async ({page}) => {
    await page.getByRole("button", {name: "#GameID"}).first().click();
    await page.getByRole("button", {name: "New definition"}).click();
    await page.getByLabel("Name", {exact: true}).fill("Rarity");
    await page.getByLabel("Type", {exact: true}).selectOption("enum");
    await page.getByRole("button", {name: "Custom expression", exact: true}).click();
    await page.locator(".schema-editor-form textarea").fill(`"N" | "R"`);
    await page.getByRole("button", {name: "Create"}).click();
    await expect(page.getByRole("button", {name: "#Rarity"})).toBeVisible();
    await page.getByRole("button", {name: "Rename"}).click();
    await page.locator(".schema-rename-input").fill("RarityTier");
    await page.getByRole("button", {name: "Confirm rename"}).click();
    await expect(page.getByRole("button", {name: "#RarityTier"})).toBeVisible();
});

test("hides value options for constraint Schemas", async ({page}) => {
    await page.getByRole("button", {name: "#NonNegative"}).first().click();
    await expect(page.getByText("Conditions (all must hold)")).toBeVisible();
    await expect(page.getByText("Allow null")).toBeHidden();
    await expect(page.getByText("Default value")).toBeHidden();
});

test("does not save an unchanged formula on blur", async ({page}) => {
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

    await page.mouse.dblclick(box.x + 565, box.y + 50);
    await expect(page.locator("textarea")).toHaveValue("=20 + 5");
    await page.mouse.click(box.x + 1250, box.y + 200);
    await expect(page.locator(".table-document")).toHaveAttribute("data-inline-editor", "closed");
    expect(await page.evaluate(() => (window as Window & {clearCalls: number}).clearCalls)).toBe(0);
    expect(pageErrors).toEqual([]);
});

test("edits a named Schema reference through its effective type", async ({page}) => {
    const canvas = page.locator("canvas:visible").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    if (!box) return;

    await page.mouse.dblclick(box.x + 1075, box.y + 50);
    await expect(page.getByRole("dialog", {name: "Edit testAttack"})).toBeVisible();
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
    await expect(page.locator(".schema-panel .panel-title")).toBeVisible();
    await expect(page.locator(".schema-row")).toHaveCount(0);
});

test("shows project tree context menu actions", async ({page}) => {
    await page.getByRole("button", {name: "game.dtproj", exact: true}).click({button: "right"});
    await expect(page.getByRole("button", {name: "New table"})).toBeVisible();
    await expect(page.getByRole("button", {name: "New schema"})).toBeVisible();
    await page.keyboard.press("Escape");

    await page.getByRole("button", {name: "item", exact: true}).click({button: "right"});
    await expect(page.getByRole("button", {name: "Rename table"})).toBeVisible();
    await expect(page.getByRole("button", {name: "Delete table"})).toBeVisible();
    await page.keyboard.press("Escape");

    await page.getByRole("button", {name: "common", exact: true}).click({button: "right"});
    await expect(page.getByRole("button", {name: "New table"})).toBeVisible();
    await expect(page.getByRole("button", {name: "Open schema"})).toBeVisible();
    await expect(page.getByRole("button", {name: "New definition"})).toBeVisible();
    await page.keyboard.press("Escape");

    await page.getByRole("button", {name: "common.cue", exact: true}).click({button: "right"});
    await expect(page.getByRole("button", {name: "Open schema"})).toBeVisible();
    await expect(page.getByRole("button", {name: "New definition"})).toBeVisible();
    await page.keyboard.press("Escape");

    await page.getByRole("button", {name: "#GameID", exact: true}).click({button: "right"});
    await expect(page.getByRole("button", {name: "Rename definition"})).toBeVisible();
    await expect(page.getByRole("button", {name: "Delete definition"})).toBeVisible();
});

test("renames a table from the project tree", async ({page}) => {
    await page.getByRole("button", {name: "item", exact: true}).click({button: "right"});
    await page.getByRole("button", {name: "Rename table"}).click();
    await expect(page.getByRole("heading", {name: "Rename table"})).toBeVisible();
    await page.getByLabel("Namespace", {exact: true}).fill("combat");
    await page.getByRole("button", {name: "Apply"}).click();
    await expect(page.getByRole("heading", {name: "Rename table"})).toBeHidden();
});

test("does not show a context menu for group nodes", async ({page}) => {
    await page.getByRole("button", {name: "Tables", exact: true}).click({button: "right"});
    await expect(page.locator(".tree-context-menu")).toHaveCount(0);
});

test("collapses and expands a namespace in the project tree", async ({page}) => {
    const rpgRow = page.locator(".tree-row").filter({hasText: "rpg"}).first();
    await rpgRow.locator(".tree-toggle").click();
    await expect(page.getByRole("button", {name: "item", exact: true})).toBeHidden();
    await rpgRow.locator(".tree-toggle").click();
    await expect(page.getByRole("button", {name: "item", exact: true})).toBeVisible();
});

test("opens table and schema as tabs and switches between them", async ({page}) => {
    await page.getByRole("button", {name: "#GameID", exact: true}).click();
    await expect(page.locator(".tab-bar .tab")).toHaveCount(2);
    await expect(page.getByText("Public Schema")).toBeVisible();
    await page.locator(".tab-bar .tab").filter({hasText: "rpg.item"}).click();
    await expect(page.getByRole("heading", {name: "rpg.item"})).toBeVisible();
    await page.locator(".tab-bar .tab").filter({hasText: "common"}).click();
    await expect(page.getByText("Public Schema")).toBeVisible();
});

test("closes a tab from the tab bar", async ({page}) => {
    await page.getByRole("button", {name: "#GameID", exact: true}).click();
    await expect(page.locator(".tab-bar .tab")).toHaveCount(2);
    await page.locator(".tab-bar .tab").filter({hasText: "common"}).locator(".tab-close").click();
    await expect(page.locator(".tab-bar .tab")).toHaveCount(1);
    await expect(page.getByRole("heading", {name: "rpg.item"})).toBeVisible();
});

test("opens and closes the schema detail peek", async ({page}) => {
    await page.locator(".schema-row").filter({hasText: "common.#NonNegative"}).click();
    await expect(page.locator(".schema-detail")).toBeVisible();
    await expect(page.locator(".schema-detail h1")).toHaveText("common.#NonNegative");
    await page.locator(".peek-close").click();
    await expect(page.locator(".schema-detail")).toHaveCount(0);
});

test("jumps to a table cell from a foreign key diagnostic", async ({page}) => {
    await page.locator(".schema-diagnostic-action").filter({hasText: "sword.effectID"}).click();
    await expect(page.getByRole("heading", {name: "rpg.item"})).toBeVisible();
    await expect(page.locator(".error-toast")).toHaveCount(0);
});

test("jumps to a schema definition from a schema diagnostic", async ({page}) => {
    await page.locator(".schema-diagnostic-action").filter({hasText: "common.#NonNegative"}).click();
    await expect(page.locator(".tab-bar .tab").filter({hasText: "common"})).toBeVisible();
    await expect(page.getByText("Public Schema")).toBeVisible();
});
