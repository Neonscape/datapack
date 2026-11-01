import {expect, test, type Page} from "@playwright/test";

test.beforeEach(async ({page}) => {
    await page.addInitScript(() => {
        const table = (name: string) => ({path: `/project/${name}.cue`, namespace: "game", name, displayName: "相同显示名", columns: [{name: "label", schemaExpression: "string", schema: {kind: "string"}}], rows: [{id: "one", values: {label: {expression: '"文本"', display: "文本"}}}], localSchema: [], diagnostics: []});
        const tables = [table("item"), table("hero")];
        const schema = {path: "/project/common.cue", package: "common", displayName: "公共规则", definitions: ["ID", "Label"].map((name) => ({name, displayName: "相同定义名", kind: "type", type: "string", schema: {kind: "string"}, schemaExpression: "string"}))};
        const view = () => ({
            project: {path: "/project/game.dtproj", root: "/project"}, settings: {build: {mode: "manual"}},
            tree: {root: {kind: "project", name: "game.dtproj", children: [
                ...tables.map((table) => ({kind: "table", name: table.name, path: table.path, displayName: table.displayName})),
                {kind: "schemaFile", name: "common.cue", path: schema.path, displayName: schema.displayName, children: schema.definitions.map((def) => ({kind: "schemaDefinition", name: `#${def.name}`, path: schema.path, displayName: def.displayName}))},
            ]}},
            schemas: {definitions: schema.definitions.map((def) => ({id: `common.#${def.name}`, kind: "type", path: schema.path, displayName: def.displayName, usages: []})), diagnostics: []},
            foreignKeys: {diagnostics: []}, compile: {mode: "manual", pending: false, diagnostics: []}, canUndo: false, canRedo: false,
        });
        const calls: {path: string; definition?: string; displayName: string}[] = [];
        const api = {
            GetSettings: async () => ({language: "en", theme: "light"}), GetExportFormats: async () => [], SetLanguage: async () => ({}), SetTheme: async () => ({}),
            SelectProject: async () => view(), GetTable: async (path: string) => tables.find((table) => table.path === path), GetPublicSchema: async () => schema,
            SetDisplayName: async (update: {path: string; definition?: string; displayName: string}) => {
                calls.push(update);
                const target = update.path === schema.path ? (update.definition ? schema.definitions.find((def) => def.name === update.definition)! : schema) : tables.find((table) => table.path === update.path)!;
                target.displayName = update.displayName.trim();
                return {table: tables.find((table) => table.path === update.path), schema, workspace: view()};
            },
        };
        Object.assign(window, {go: {main: {App: api}}, __titles: {calls}});
    });
    await page.goto("/");
    await openProject(page);
});
async function openProject(page: Page) {
    await page.getByRole("button", {name: "File", exact: true}).click();
    await page.getByRole("navigation", {name: "Application menu"}).getByRole("button", {name: "Open project", exact: true}).click();
}
async function editTitle(page: Page, label: string, value: string) {
    await page.locator(".tree-label").filter({hasText: label}).click({button: "right"});
    await page.getByRole("button", {name: "Edit display name", exact: true}).click();
    await page.getByLabel("Display name", {exact: true}).fill(value);
    await page.getByRole("dialog").getByRole("button", {name: "Apply", exact: true}).click();
    await expect(page.getByRole("heading", {name: "Edit display name", exact: true})).toHaveCount(0);
}

test("duplicate titles keep distinct table tabs and stable internal identities", async ({page}) => {
    for (const name of ["item", "hero"]) await page.getByRole("button", {name: `相同显示名 (${name})`, exact: true}).click();
    await expect(page.locator(".tab-bar .tab")).toHaveCount(2);
    await expect(page.getByRole("heading", {name: "相同显示名 (game.hero)", exact: true})).toBeVisible();
    await page.getByRole("tab", {name: "相同显示名 (game.item) ×", exact: true}).click();
    await expect(page.getByRole("heading", {name: "相同显示名 (game.item)", exact: true})).toBeVisible();
});

test("table title saves, reopens and falls back when cleared without changing its path", async ({page}) => {
    await page.getByRole("button", {name: "相同显示名 (item)", exact: true}).click();
    await editTitle(page, "相同显示名 (item)", " 道具总表 ");
    await expect(page.getByRole("heading", {name: "道具总表 (game.item)", exact: true})).toBeVisible();
    await openProject(page);
    await expect(page.getByRole("button", {name: "道具总表 (item)", exact: true})).toBeVisible();
    await editTitle(page, "道具总表 (item)", " ");
    await expect(page.getByRole("button", {name: "item", exact: true})).toBeVisible();
    expect(await page.evaluate(() => (window as unknown as {__titles: {calls: unknown[]}}).__titles.calls)).toEqual([
        {path: "/project/item.cue", definition: "", displayName: " 道具总表 "},
        {path: "/project/item.cue", definition: "", displayName: " "},
    ]);
});

test("schema file and definition titles appear in the tree, editor, peek and tabs", async ({page}) => {
    await editTitle(page, "公共规则 (common.cue)", "共享校验规则");
    await editTitle(page, "相同定义名 (#ID)", "稳定标识");
    await expect(page.locator(".schema-detail-head h1")).toHaveText("稳定标识 (common.#ID)");
    await page.getByRole("button", {name: "稳定标识 (#ID)", exact: true}).click();
    await expect(page.locator(".tab-bar .tab")).toContainText(["共享校验规则 (common)"]);
    await expect(page.locator(".schema-definition-row").filter({hasText: "稳定标识 (#ID)"})).toBeVisible();
    await page.getByRole("button", {name: "相同定义名 (#Label)", exact: true}).first().click();
    await expect(page.locator(".schema-editor-form input").first()).toHaveValue("Label");
    await editTitle(page, "稳定标识 (#ID)", "");
    await expect(page.getByRole("button", {name: "#ID", exact: true}).first()).toBeVisible();
});

test("language changes preserve custom titles and expanded definition identities", async ({page}) => {
    await page.getByRole("button", {name: "Settings", exact: true}).click();
    await page.locator(".structure-dialog select").first().selectOption("zh-CN");
    await page.getByRole("button", {name: "应用", exact: true}).click();
    await expect(page.getByRole("button", {name: "公共规则 (common.cue)", exact: true})).toBeVisible();
    await expect(page.getByRole("button", {name: "相同定义名 (#ID)", exact: true})).toBeVisible();
    await page.getByRole("button", {name: "相同显示名 (item)", exact: true}).click({button: "right"});
    await expect(page.getByRole("button", {name: "编辑显示名", exact: true})).toBeVisible();
});
