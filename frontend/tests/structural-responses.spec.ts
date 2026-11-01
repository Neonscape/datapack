/// <reference path="./node-fs.d.ts" />
/// <reference path="./node-runtime.d.ts" />
import {test,expect,type Page} from "@playwright/test";
import {mkdtempSync,readFileSync,rmSync} from "node:fs";
import {tmpdir} from "node:os";
import {resolve,join} from "node:path";
import {execFileSync} from "node:child_process";
import type {TableDocument,WorkspaceView,PublicSchemaDocument} from "../src/types";
type Result={table:TableDocument;workspace:WorkspaceView};
type History={workspace:WorkspaceView};
type Fixture={Mode:"manual"|"auto";Deleted:string;Before:TableDocument;RawBefore:TableDocument;View:WorkspaceView;Schema:PublicSchemaDocument;Result:Result;Fresh:TableDocument;Undo:History;UndoTable:TableDocument;Redo:History;RedoTable:TableDocument;Width:Result;SourceAfter:string};
type State={calls:{method:string;rowID?:string;column?:string;expression?:string}[];resolveWidth?:(value:Result)=>void;fixture:Fixture};
declare global{interface Window{__structuralResponse:State}}
declare const process:{cwd():string;env:Record<string,string|undefined>};
let fixtures:Fixture[];
test.use({viewport:{width:1600,height:1000}});
test.beforeAll(()=>{
 const directory=mkdtempSync(join(tmpdir(),"datapack-structural-ui-"));const output=join(directory,"response.json");
 try{execFileSync("go",["test","./internal/workspace","-run","^TestExportStructuralUIFixtures$","-count=1"],{cwd:resolve(process.cwd(),".."),env:{...process.env,GOTOOLCHAIN:"local",GOPROXY:"off",GOSUMDB:"off",GOFLAGS:"-mod=readonly",DATAPACK_STRUCTURAL_UI_PATH:output},timeout:90000});fixtures=JSON.parse(readFileSync(output,"utf8")) as Fixture[];}
 finally{rmSync(directory,{recursive:true,force:true});}
});
async function point(page:Page,column:number,row=0){const box=await page.locator(".tab-panel:visible canvas").first().boundingBox();expect(box).not.toBeNull();return{x:box!.x+(column===0?85:225+170*(column-1)),y:box!.y+50+30*row};}
async function boot(page:Page,fixture:Fixture,raw=false){
 await page.addInitScript(({fixture,raw})=>{
  let table=raw?fixture.RawBefore:fixture.Before;const state:State={calls:[],fixture};window.__structuralResponse=state;
  (window as unknown as Window&{go:unknown}).go={main:{App:{GetSettings:async()=>({language:"en",theme:"light"}),GetExportFormats:async()=>[],SelectProject:async()=>fixture.View,GetTable:async()=>structuredClone(table),GetPublicSchema:async()=>fixture.Schema,
   DeleteRow:async({rowID}:{rowID:string})=>{state.calls.push({method:"DeleteRow",rowID});table=fixture.Result.table;return structuredClone(fixture.Result);},
   Undo:async()=>{state.calls.push({method:"Undo"});table=fixture.UndoTable;return structuredClone(fixture.Undo);},Redo:async()=>{state.calls.push({method:"Redo"});table=fixture.RedoTable;return structuredClone(fixture.Redo);},
   SetColumnWidth:async()=>{state.calls.push({method:"Width"});return await new Promise<Result>(resolve=>{state.resolveWidth=resolve;});},
   ClearCell:async(argument:{rowID:string;column:string;expression:string})=>{state.calls.push({method:"ClearCell",...argument});return structuredClone({table,workspace:fixture.View});},
   SetLanguage:async(language:string)=>({language,theme:"light"}),
  }}};
 },{fixture,raw});
 await page.goto("/");await page.getByRole("button",{name:"File",exact:true}).click();await page.getByLabel("Application menu").getByRole("button",{name:"Open project",exact:true}).click();
 await page.getByRole("button",{name:"Expand game",exact:true}).click();await page.getByRole("button",{name:"Expand Tables",exact:true}).click();await page.getByRole("button",{name:"items",exact:true}).click();
}
async function expectFormula(page:Page){const p=await point(page,2);await page.mouse.click(p.x,p.y);await expect(page.locator(".formula-bar-input")).toHaveValue("=amount * 2");}
for(const mode of ["manual","auto"] as const)for(const deleted of ["a","b","c"])test(`deleting ${deleted} retains actual Service formula cells in ${mode}`,async({page})=>{
 const fixture=fixtures.find(f=>f.Mode===mode&&f.Deleted===deleted)!;await boot(page,fixture);
 const row=fixture.Before.rows.findIndex(r=>r.id===deleted);const p=await point(page,0,row);const canvas=await page.locator(".tab-panel:visible canvas").first().boundingBox();await page.mouse.click(canvas!.x+16,p.y,{button:"right"});await page.getByRole("button",{name:"Delete row",exact:true}).click();
 await expect(page.locator(".table-document-header")).toContainText("2 rows");await expectFormula(page);
 await page.getByRole("button",{name:"Settings",exact:true}).click();await page.locator(".structure-dialog select").first().selectOption("zh-CN");await page.getByRole("button",{name:"应用",exact:true}).click();await expectFormula(page);
 for(const action of ["撤销","重做"]){await page.getByRole("button",{name:"编辑",exact:true}).click();await page.getByRole("button",{name:new RegExp(`^${action}`)}).click();await expectFormula(page);}
 expect(await page.evaluate(()=>window.__structuralResponse.calls.some(c=>c.method==="ClearCell"))).toBe(false);
 expect(fixture.SourceAfter).toMatch(/total:\s*"amount \* 2"/);
});
test("an untouched blank formula editor does not clear an expression arriving from refresh",async({page})=>{
 const fixture=fixtures.find(f=>f.Mode==="manual"&&f.Deleted==="b")!;await boot(page,fixture,true);
 const canvas=await page.locator(".tab-panel:visible canvas").first().boundingBox();await page.mouse.move(canvas!.x+361,canvas!.y+17);await page.mouse.down();await page.mouse.move(canvas!.x+381,canvas!.y+17,{steps:5});await page.mouse.up();
 await expect.poll(()=>page.evaluate(()=>typeof window.__structuralResponse.resolveWidth)).toBe("function");
 const p=await point(page,2);await page.mouse.click(p.x,p.y);const input=page.locator(".formula-bar-input");await expect(input).toHaveValue("");await input.click();
 await page.evaluate(()=>window.__structuralResponse.resolveWidth!(window.__structuralResponse.fixture.Width));
 await expect(page.getByRole("region",{name:"Table diagnostics"})).toBeVisible();
 await expect(input).toBeFocused();await expect(input).toHaveValue("");await page.mouse.click(100,50);
 expect(await page.evaluate(()=>window.__structuralResponse.calls.some(c=>c.method==="ClearCell"))).toBe(false);
 await expect(input).toHaveValue("=amount * 2");
});
