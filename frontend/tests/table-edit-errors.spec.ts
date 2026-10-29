import {expect, test, type Page} from "@playwright/test";
import type {TableDocument} from "../src/types";

// Exercise the real App/grid; only the Wails backend boundary is mocked.
test.use({viewport: {width: 1600, height: 1000}});

type CreateCall = {path: string; rowID: string; values: Record<string, string>; formulas: Record<string, string>};
type EditCall = {path: string; rowID?: string; column?: string; expression?: string; oldID?: string; newID?: string};
type EditState = {
    creates: CreateCall[];
    edits: {method: string; argument: EditCall}[];
    deletes: string[];
    createFailures: number;
    failEdit: string | null;
    unhandled: string[];
};

declare global {
    interface Window {
        __editErrorsTest: EditState;
    }
}

const path = "/project/tables/items.cue";
const formula = (page: Page) => page.getByRole("textbox", {name: "Formula bar"});

async function selectCell(page: Page, column: number, row: number) {
    const canvas = page.locator(".tab-panel:visible canvas").first();
    await expect(canvas).toBeVisible();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    await page.mouse.click(box!.x + (column === 0 ? 85 : 225 + (column - 1) * 170), box!.y + 50 + row * 30);
    await expect(formula(page)).toBeEnabled();
}

async function editCell(page: Page, column: number, row: number, value: string) {
    await selectCell(page, column, row);
    await formula(page).click();
    await formula(page).fill(value);
    await formula(page).press("Enter");
}

async function deleteColumn(page: Page, column: number, name: string) {
    const canvas = page.locator(".tab-panel:visible canvas").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    await page.mouse.click(box!.x + 225 + (column - 1) * 170, box!.y + 15, {button: "right"});
    await page.getByRole("button", {name: "Delete column", exact: true}).click();
    const dialog = page.locator(".structure-dialog");
    await expect(dialog.locator("input")).toHaveValue(name);
    await dialog.getByRole("button", {name: "Delete column", exact: true}).click();
    await expect(dialog).toHaveCount(0);
}

async function expectNoUnhandled(page: Page, pageErrors: string[]) {
    expect(await page.evaluate(() => window.__editErrorsTest.unhandled)).toEqual([]);
    expect(pageErrors).toEqual([]);
}

test.beforeEach(async ({page}) => {
    await page.addInitScript(() => {
        let table: TableDocument = {
            path: "/project/tables/items.cue", namespace: "game", name: "items",
            columns: [
                {name: "name", schemaExpression: "string", schema: {kind: "string"}},
                {name: "points", schemaExpression: "int", schema: {kind: "integer"}},
                {name: "total", schemaExpression: "int", schema: {kind: "integer"}},
            ],
            rows: [{id: "row_one", values: {
                name: {expression: '"existing"', display: "existing", formula: false},
                points: {expression: "10", display: "10", formula: false},
                total: {expression: "20", display: "20", formula: false},
            }}],
            diagnostics: [], localSchema: [],
        };
        const workspace = {
            project: {path: "/project/game.dtproj", root: "/project"},
            settings: {build: {mode: "manual"}},
            tree: {root: {kind: "project", name: "game.dtproj", children: [{kind: "table", name: "items", path: table.path}]}},
            schemas: {definitions: [], diagnostics: []}, foreignKeys: {diagnostics: []},
            compile: {mode: "manual", pending: false, diagnostics: []}, canUndo: false, canRedo: false,
        };
        const state: EditState = {creates: [], edits: [], deletes: [], createFailures: 1, failEdit: null, unhandled: []};
        window.__editErrorsTest = state;
        window.addEventListener("unhandledrejection", (event) => state.unhandled.push(String(event.reason)));
        const result = () => structuredClone({table, workspace});
        const edit = (method: string) => async (argument: EditCall) => {
            state.edits.push({method, argument});
            if (state.failEdit === method) throw new Error(`${method} rejected`);
            table = {...table, rows: table.rows.map((row) => {
                if (method === "RenameRow") return row.id === argument.oldID ? {...row, id: argument.newID!} : row;
                if (row.id !== argument.rowID) return row;
                const values = {...row.values};
                if (method === "ClearCell") delete values[argument.column!];
                else values[argument.column!] = {
                    expression: method === "SetCellText" ? JSON.stringify(argument.expression) : argument.expression!,
                    display: argument.expression!, formula: method === "SetCellFormula",
                };
                return {...row, values};
            })};
            return result();
        };
        (window as unknown as Window & {go: unknown}).go = {main: {App: {
            GetSettings: async () => ({language: "en", theme: "light"}), GetExportFormats: async () => [],
            SelectProject: async () => structuredClone(workspace), GetTable: async () => structuredClone(table),
            CreateRow: async (argument: CreateCall) => {
                state.creates.push(argument);
                if (state.createFailures-- > 0) throw "CreateRow rejected";
                const values: TableDocument["rows"][number]["values"] = {};
                for (const [name, expression] of Object.entries(argument.values)) values[name] = {expression, display: expression, formula: false};
                for (const [name, expression] of Object.entries(argument.formulas)) values[name] = {expression, display: expression, formula: true};
                table = {...table, rows: [...table.rows, {id: argument.rowID, values}]};
                return result();
            },
            SetCellText: edit("SetCellText"), UpdateCell: edit("UpdateCell"), SetCellFormula: edit("SetCellFormula"),
            ClearCell: edit("ClearCell"), RenameRow: edit("RenameRow"),
            DeleteColumn: async ({name}: {name: string}) => {
                state.deletes.push(name);
                table = {...table, columns: table.columns.filter((column) => column.name !== name), rows: table.rows.map((row) => {
                    const values = {...row.values};
                    delete values[name];
                    return {...row, values};
                })};
                return result();
            },
        }}};
    });
    await page.goto("/");
    await page.getByRole("button", {name: "File", exact: true}).click();
    await page.getByLabel("Application menu").getByRole("button", {name: "Open project", exact: true}).click();
    await page.getByRole("button", {name: "items", exact: true}).click();
});

test("retains a rejected new row's ID, values and formulas, then retries without duplicate errors", async ({page}) => {
    const pageErrors: string[] = [];
    page.on("pageerror", (error) => pageErrors.push(error.message));
    // Values first exercises the entire draft payload, not only the triggering cell.
    await editCell(page, 1, 1, '"keep me"');
    await editCell(page, 2, 1, "=1 + 2");
    await editCell(page, 0, 1, "new_row");
    await expect(page.locator(".table-validation-error")).toContainText("CreateRow rejected");
    await expect(page.getByText("CreateRow rejected", {exact: true})).toHaveCount(1);
    await expect(page.locator(".error-toast")).toHaveCount(0);
    for (const [column, value] of [[0, "new_row"], [1, '"keep me"'], [2, "=1 + 2"]] as const) {
        await selectCell(page, column, 1);
        await expect(formula(page)).toHaveValue(value);
    }
    await editCell(page, 1, 1, '"fixed"');
    await expect.poll(() => page.evaluate(() => window.__editErrorsTest.creates)).toEqual([
        {path, rowID: "new_row", values: {name: '"keep me"'}, formulas: {points: "1 + 2"}},
        {path, rowID: "new_row", values: {name: '"fixed"'}, formulas: {points: "1 + 2"}},
    ]);
    await expect(page.locator(".table-validation-error")).toHaveCount(0);
    await selectCell(page, 0, 1);
    await expect(formula(page)).toHaveValue("new_row");
    for (const column of [0, 1, 2]) {
        await selectCell(page, column, 2);
        await expect(formula(page)).toHaveValue("");
    }
    await expectNoUnhandled(page, pageErrors);
});

test("retains an ID-first draft when an inline edit fails to create it", async ({page}) => {
    const pageErrors: string[] = [];
    page.on("pageerror", (error) => pageErrors.push(error.message));
    await editCell(page, 0, 1, "new_row");
    const canvas = page.locator(".tab-panel:visible canvas").first();
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    await page.mouse.dblclick(box!.x + 395, box!.y + 80);
    const inline = page.locator(".inline-cell-editor");
    await inline.fill("123");
    await inline.press("Enter");
    await expect(page.locator(".table-validation-error")).toContainText("CreateRow rejected");
    await expect(page.locator(".error-toast")).toHaveCount(0);
    await selectCell(page, 0, 1);
    await expect(formula(page)).toHaveValue("new_row");
    await selectCell(page, 2, 1);
    await expect(formula(page)).toHaveValue("123");
    await editCell(page, 2, 1, "124");
    await expect(page.locator(".table-validation-error")).toHaveCount(0);
    await expect.poll(() => page.evaluate(() => window.__editErrorsTest.creates.length)).toBe(2);
    await expectNoUnhandled(page, pageErrors);
});

for (const entry of [
    {method: "SetCellText", column: 1, input: "changed", original: "existing"},
    {method: "UpdateCell", column: 2, input: "99", original: "10"},
    {method: "SetCellFormula", column: 2, input: "=2 + 3", original: "10"},
    {method: "ClearCell", column: 2, input: "", original: "10"},
    {method: "RenameRow", column: 0, input: "renamed_row", original: "row_one"},
]) {
    test(`reports a normal ${entry.method} failure once and allows a later retry`, async ({page}) => {
        const pageErrors: string[] = [];
        page.on("pageerror", (error) => pageErrors.push(error.message));
        await page.evaluate((method) => { window.__editErrorsTest.failEdit = method; }, entry.method);
        await editCell(page, entry.column, 0, entry.input);
        await expect(page.locator(".error-toast")).toHaveText(`${entry.method} rejected`);
        await expect(page.getByText(`${entry.method} rejected`, {exact: true})).toHaveCount(1);
        await expect(page.locator(".table-validation-error")).toHaveCount(0);
        await expect(formula(page)).toHaveValue(entry.original);
        await page.evaluate(() => { window.__editErrorsTest.failEdit = null; });
        await editCell(page, entry.column, 0, entry.input);
        await expect(page.locator(".error-toast")).toHaveCount(0);
        await expect(formula(page)).toHaveValue(entry.input);
        await expect.poll(() => page.evaluate(() => window.__editErrorsTest.edits.length)).toBe(2);
        await expectNoUnhandled(page, pageErrors);
    });
}

test("clears a deleted last-column selection and keeps editing after deleting all data columns", async ({page}) => {
    const pageErrors: string[] = [];
    page.on("pageerror", (error) => pageErrors.push(error.message));
    for (const [column, name] of [[3, "total"], [2, "points"], [1, "name"]] as const) {
        await selectCell(page, column, 0);
        await expect(page.locator(".formula-bar-cell")).toContainText(name);
        await deleteColumn(page, column, name);
        await expect(formula(page)).toBeDisabled();
        await expect(formula(page)).toHaveValue("");
        await expect(page.locator(".formula-bar-cell")).toHaveText("");
        if (column === 3) {
            await editCell(page, 1, 0, "still editable");
            await expect(formula(page)).toHaveValue("still editable");
        }
    }
    await editCell(page, 0, 0, "renamed_after_deletion");
    await expect(formula(page)).toHaveValue("renamed_after_deletion");
    expect(await page.evaluate(() => window.__editErrorsTest.deletes)).toEqual(["total", "points", "name"]);
    await expectNoUnhandled(page, pageErrors);
});

test("preserves an unaffected valid cell selection when deleting a later column", async ({page}) => {
    const pageErrors: string[] = [];
    page.on("pageerror", (error) => pageErrors.push(error.message));
    await selectCell(page, 1, 0);
    await deleteColumn(page, 3, "total");
    await expect(formula(page)).toBeEnabled();
    await expect(page.locator(".formula-bar-cell")).toHaveText("name · row_one");
    await expect(formula(page)).toHaveValue("existing");
    await formula(page).click();
    await formula(page).fill("edited after deletion");
    await formula(page).press("Enter");
    await expect(formula(page)).toHaveValue("edited after deletion");
    await expectNoUnhandled(page, pageErrors);
});
