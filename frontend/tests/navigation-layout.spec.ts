import {expect, test, type Page} from "@playwright/test";

// Real App, DOM scrolling and Glide canvas; only native Wails calls are mocked.
async function fixture(page: Page, withDefault = true, defaultLocal = false) {
    await page.addInitScript(({withDefault, defaultLocal}) => {
        const path = "/project/schema/common.cue";
        const table = {
            path: "/project/item.cue", namespace: "game", name: "item", localSchema: [], diagnostics: [],
            columns: [{name: "name", schemaExpression: "string", schema: {kind: "string"}}],
            rows: Array.from({length: 300}, (_, i) => ({id: `row_${i}`, values: {name: {expression: '"item"', display: `Item ${i}`}}})),
        };
        const workspace = {
            project: {path: "/project/game.dtproj", root: "/project"}, settings: {build: {mode: "manual"}},
            tree: {root: {kind: "project", name: "game.dtproj", children: [
                {kind: "namespace", name: "game", children: [{kind: "group", name: "Tables", children: [
                    {kind: "table", name: "item", path: table.path, children: [{kind: "localSchema", name: "local_schema", path: table.path, children: [{kind: "localSchema", name: "#Position", path: table.path}]}]},
                    ...Array.from({length: 160}, (_, i) => ({kind: "table", name: `long_table_${i}_${"中文显示说明".repeat(10)}`, path: `/project/${i}.cue`})),
                ]}]},
                {kind: "namespace", name: "common", children: [{kind: "group", name: "Schemas", children: [{kind: "schemaFile", name: "common.cue", path, children: [{kind: "schemaDefinition", name: "#ID", path}]}]}]},
            ]}},
            schemas: {definitions: withDefault ? [{id: defaultLocal ? "game.item._local.#Position" : "common.#ID", kind: "type", path: defaultLocal ? table.path : path, usages: []}] : [], diagnostics: []},
            foreignKeys: {diagnostics: []}, compile: {mode: "manual", pending: true, diagnostics: []}, canUndo: true, canRedo: true,
        };
        const api = {
            GetSettings: async () => ({language: "en", theme: "light"}), GetExportFormats: async () => [],
            SetLanguage: async () => ({}), SetTheme: async () => ({}),
            SelectProject: async () => {
                const view = structuredClone(workspace);
                if (state.opens++ > 0) { view.project.path = "/other/other.dtproj"; view.tree.root.name = "other.dtproj"; }
                return view;
            }, GetTable: async () => table,
            Compile: async () => ({workspace: structuredClone(workspace)}),
            Undo: async () => { state.undo++; return {table, workspace}; }, Redo: async () => { state.undo++; return {table, workspace}; },
            GetProjectSettings: async () => ({project: {...workspace.project, tableRoots: ["tables"], schemaRoots: ["schema"]}, export: {directory: "exports", targets: []}, formats: [], workspace,
                tables: Array.from({length: 160}, (_, i) => ({id: `game.item_${i}`, path: `/project/${i}.cue`, selected: true, configured: true}))}),
        };
        const state = {undo: 0, opens: 0};
        Object.assign(window, {go: {main: {App: api}}, __layout: state});
    }, {withDefault, defaultLocal});
    await page.goto("/");
    await openProject(page);
}
async function openProject(page: Page) {
    await page.getByRole("button", {name: "File", exact: true}).click();
    await page.getByRole("navigation", {name: "Application menu"}).getByRole("button", {name: "Open project", exact: true}).click();
}
function treeRow(page: Page, label: string) {
    return page.locator(".tree-row").filter({has: page.locator(".tree-label").filter({hasText: new RegExp(`^${label.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}$`)})});
}
async function toggle(page: Page, label: string) { await treeRow(page, label).locator(".tree-toggle").click(); }
async function openGrid(page: Page) {
    await toggle(page, "game");
    await toggle(page, "Tables");
    await page.locator(".tree-label").getByText("item", {exact: true}).click();
    await expect(page.locator(".dvn-scroller")).toBeVisible();
}

test("opens only the default file ancestry and preserves manual toggles on refresh", async ({page}) => {
    await fixture(page);
    await expect(treeRow(page, "common.cue")).toBeVisible();
    await expect(treeRow(page, "#ID")).toBeVisible();
    await expect(treeRow(page, "item")).toHaveCount(0);
    await toggle(page, "common");
    await toggle(page, "game");
    await toggle(page, "Tables");
    await page.getByRole("button", {name: "Compile", exact: true}).click();
    await expect(treeRow(page, "item")).toBeVisible();
    await expect(treeRow(page, "common.cue")).toHaveCount(0);
    await openProject(page);
    await expect(treeRow(page, "item")).toHaveCount(0);
    await expect(treeRow(page, "#ID")).toBeVisible();
});

test("without a default file starts with only the project expanded", async ({page}) => {
    await fixture(page, false);
    await expect(treeRow(page, "game")).toBeVisible();
    await expect(treeRow(page, "common")).toBeVisible();
    await expect(treeRow(page, "Tables")).toHaveCount(0);
    await expect(treeRow(page, "Schemas")).toHaveCount(0);
});

test("large tree and long labels scroll inside a small window above the footer", async ({page}, info) => {
    await page.setViewportSize({width: 980, height: 580});
    await fixture(page, false);
    // Use the current expanded state so the regression also demonstrates the old layout.
    if (await treeRow(page, "Tables").count() === 0) await toggle(page, "game");
    if (await treeRow(page, "item").count() === 0) await toggle(page, "Tables");
    const dimensions = await page.evaluate(() => {
        const tree = document.querySelector(".tree")!;
        const footer = document.querySelector(".log-panel")!;
        return {treeBottom: tree.getBoundingClientRect().bottom, footerTop: footer.getBoundingClientRect().top, footerBottom: footer.getBoundingClientRect().bottom,
            viewport: innerHeight, pageWidth: document.documentElement.scrollWidth, width: innerWidth, height: tree.clientHeight, content: tree.scrollHeight};
    });
    expect(dimensions.footerBottom).toBeLessThanOrEqual(dimensions.viewport);
    expect(dimensions.treeBottom).toBeLessThanOrEqual(dimensions.footerTop);
    expect(dimensions.pageWidth).toBeLessThanOrEqual(dimensions.width);
    expect(dimensions.content).toBeGreaterThan(dimensions.height);
    await page.locator(".tree").hover();
    await page.mouse.wheel(0, 1600);
    await expect.poll(() => page.locator(".tree").evaluate((el) => el.scrollTop)).toBeGreaterThan(0);
    await page.screenshot({path: info.outputPath("large-tree.png")});
});

test("settings consume modal wheel events, block background history and restore scrolling and focus", async ({page}, info) => {
    await page.setViewportSize({width: 980, height: 580});
    await fixture(page);
    await openGrid(page);
    const grid = page.locator(".dvn-scroller");
    await grid.hover();
    await page.mouse.wheel(0, 300);
    await expect.poll(() => grid.evaluate((el) => el.scrollTop)).toBeGreaterThan(0);
    const canvas = await page.locator("canvas:visible").first().boundingBox();
    expect(canvas).not.toBeNull();
    await page.mouse.click(canvas!.x + 225, canvas!.y + 50);
    await page.getByRole("button", {name: "game.dtproj", exact: true}).click();
    await expect(page.locator(".project-settings-dialog")).toBeVisible();
    const dialogBounds = await page.locator(".project-settings-dialog").boundingBox();
    expect(dialogBounds!.y).toBeGreaterThanOrEqual(20);
    expect(dialogBounds!.y + dialogBounds!.height).toBeLessThanOrEqual(560);
    const before = await grid.evaluate((el) => el.scrollTop);
    const backdrop = page.locator(".project-settings-backdrop");
    await backdrop.hover({position: {x: 5, y: 5}});
    await page.mouse.wheel(0, 500);
    await expect(grid).toHaveJSProperty("scrollTop", before);
    const content = page.locator(".project-settings-table-list");
    await content.hover();
    await page.mouse.wheel(0, 500);
    await expect.poll(() => content.evaluate((el) => el.scrollTop)).toBeGreaterThan(0);
    await page.getByRole("button", {name: "Select all", exact: true}).focus();
    await page.keyboard.press("Control+z");
    expect(await page.evaluate(() => (window as unknown as {__layout: {undo: number}}).__layout.undo)).toBe(0);
    const ownsNativeEvents = await page.getByRole("button", {name: "Select all", exact: true}).evaluate((button) => {
        const f12 = new KeyboardEvent("keydown", {key: "F12", bubbles: true, cancelable: true});
        button.dispatchEvent(f12);
        const input = document.querySelector<HTMLInputElement>(".project-settings-directory-input input")!;
        input.focus();
        const transfer = new DataTransfer(); transfer.setData("text/plain", "native text");
        const paste = new ClipboardEvent("paste", {clipboardData: transfer, bubbles: true, cancelable: true});
        input.dispatchEvent(paste);
        return {f12: f12.defaultPrevented, paste: paste.defaultPrevented};
    });
    expect(ownsNativeEvents).toEqual({f12: false, paste: false});
    await page.getByRole("button", {name: "Apply", exact: true}).focus();
    await page.keyboard.press("Tab");
    await expect(page.getByRole("button", {name: "Close", exact: true})).toBeFocused();
    await page.keyboard.press("Shift+Tab");
    await expect(page.getByRole("button", {name: "Apply", exact: true})).toBeFocused();
    await page.screenshot({path: info.outputPath("project-settings.png")});
    await page.getByRole("button", {name: "Close", exact: true}).click();
    await grid.hover();
    await page.mouse.wheel(0, 400);
    await expect.poll(() => grid.evaluate((el) => el.scrollTop)).toBeGreaterThan(before);
    await expect(page.getByRole("button", {name: "game.dtproj", exact: true})).toBeFocused();
});

test("fixed navigation labels switch language without changing expansion", async ({page}) => {
    await fixture(page);
    await toggle(page, "game");
    await expect(treeRow(page, "Tables")).toBeVisible();
    await page.getByRole("button", {name: "Settings", exact: true}).click();
    await page.locator(".structure-dialog select").first().selectOption("zh-CN");
    await page.getByRole("button", {name: "应用", exact: true}).click();
    await expect(treeRow(page, "表")).toBeVisible();
    await expect(treeRow(page, "校验规则")).toBeVisible();
    await expect(page.locator(".schema-panel > .panel-title")).toHaveText("校验规则");
    await expect(treeRow(page, "#ID")).toBeVisible();
    await expect(treeRow(page, "item")).toHaveCount(0);
    await page.getByRole("button", {name: "设置", exact: true}).click();
    await page.locator(".structure-dialog select").first().selectOption("en");
    await page.getByRole("button", {name: "Apply", exact: true}).click();
    await expect(treeRow(page, "Tables")).toBeVisible();
    await expect(treeRow(page, "Schemas")).toBeVisible();
    await expect(treeRow(page, "#ID")).toBeVisible();
});


test("a default local definition expands its ancestors without translating its identity", async ({page}) => {
    await fixture(page, true, true);
    await expect(treeRow(page, "item")).toBeVisible();
    await expect(treeRow(page, "Local schemas")).toBeVisible();
    await expect(treeRow(page, "#Position")).toBeVisible();
    await expect(treeRow(page, "common.cue")).toHaveCount(0);
});
