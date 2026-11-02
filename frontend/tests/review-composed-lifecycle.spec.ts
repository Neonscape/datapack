import {test,expect,type Page} from "@playwright/test";
async function boot(page:Page,scenario:"queued"|"export-first"|"export-repaired",mode:"delta"|"full"="delta") {
 await page.addInitScript(({scenario,mode})=>{
  let revision=1,price=scenario==="export-repaired"?"-1":"1",compiled=scenario==="export-repaired",diagnostics=scenario==="export-repaired"?[{path:"/project/items.cue",rowID:"one",column:"price",message:"old validation error"}]:[];
  let release:(()=>void)|undefined;
  const state={pendingEdit:false,historyCalls:0,exportCalls:0,compileCalls:0,getTableCalls:0,displayPrice:()=>price,revision:()=>revision,releaseEdit:()=>release?.()};
  const view=()=>({projectEpoch:"epoch-1",revision,project:{path:"/project/data.dtproj",root:"/project"},settings:{build:{mode:"manual"}},tree:{root:{kind:"project",name:"Project",children:[{kind:"table",name:"items",path:"/project/items.cue"}]}},schemas:{definitions:[]},foreignKeys:{},compile:{mode:"manual",pending:!compiled,inputRevision:compiled?revision:0,compiledAt:compiled?"2026-11-01T00:00:00Z":undefined,diagnostics},canUndo:true,canRedo:true});
  const hasFormula=scenario!=="queued" || mode==="full";
  const table=()=>({projectEpoch:"epoch-1",workspaceRevision:revision,path:"/project/items.cue",namespace:"game",name:"items",columns:[{name:"price",schemaExpression:"int",schema:{kind:"integer"}},...(hasFormula?[{name:"total",schemaExpression:"int",schema:{kind:"integer"}}]:[])],rows:[{id:"one",values:{price:{expression:price,display:price},...(hasFormula?{total:{expression:"price * 2",display:compiled?String(Number(price)*2):"",formula:true}}:{})}}],localSchema:[],diagnostics:diagnostics.map(d=>({rowID:d.rowID,column:d.column,message:d.message}))});
  const delta=(baseRevision:number)=>({path:"/project/items.cue",projectEpoch:"epoch-1",baseRevision,revision,rowIndex:0,rowID:"one",column:"price",cell:{expression:price,display:price}});
  const api={GetSettings:async()=>({language:"en",theme:"light"}),GetExportFormats:async()=>[],SelectProject:async()=>view(),GetTable:async()=>{state.getTableCalls++;return table();},
   UpdateCellDelta:async(update:{expression:string})=>{if(scenario==="queued"){state.pendingEdit=true;await new Promise<void>(resolve=>{release=resolve;});state.pendingEdit=false;}const base=revision;price=update.expression;revision++;compiled=false;return !hasFormula ? {delta:delta(base),workspace:view()} : {table:table(),workspace:view()};},
   UndoDelta:async()=>{state.historyCalls++;const base=revision;price="1";revision++;compiled=false;return mode==="delta" ? {delta:delta(base),workspace:view()} : {workspace:view()};},
   RedoDelta:async()=>{state.historyCalls++;throw new Error("nothing to redo after edit");},
   Export:async()=>{state.exportCalls++;compiled=true;diagnostics=[];return {directory:"/export",files:["items.csv"],workspace:view()};},
   Compile:async()=>{state.compileCalls++;return {workspace:view()};},OpenPath:async()=>{},
  };
  Object.assign(window,{go:{main:{App:api}},__review:state});
 },{scenario,mode});
 await page.goto("/");await page.getByRole("button",{name:"Open project",exact:true}).click();await page.getByRole("button",{name:"items",exact:true}).click();
 const canvas=await page.locator(".tab-panel:visible canvas").first().boundingBox();await page.mouse.click(canvas!.x+225,canvas!.y+50);
 await expect(page.locator(".formula-bar-input")).toHaveValue(scenario==="export-repaired"?"-1":"1");
}
async function edit(page:Page){const field=page.locator(".formula-bar-input");await field.click();await field.fill("2");await field.press("Enter");}
async function history(page:Page,name:"Undo"|"Redo"){await page.getByRole("button",{name:"Edit",exact:true}).click();await page.getByRole("navigation",{name:"Application menu"}).getByRole("button",{name:new RegExp(`^${name} `)}).click();}
async function exportProject(page:Page){await page.getByRole("button",{name:"File",exact:true}).click();await page.getByRole("navigation",{name:"Application menu"}).getByRole("button",{name:"Export",exact:true}).click();}
declare global {interface Window {__review:{pendingEdit:boolean;historyCalls:number;exportCalls:number;compileCalls:number;getTableCalls:number;displayPrice():string;revision():number;releaseEdit():void}}}
for(const mode of ["delta","full"] as const)for(const operation of ["Undo","Redo"] as const)test(`App installs authoritative ${operation} ${mode} result queued behind a pending edit`,async({page})=>{
 await boot(page,"queued",mode);await edit(page);await expect.poll(()=>page.evaluate(()=>window.__review.pendingEdit)).toBe(true);await history(page,operation);expect(await page.evaluate(()=>window.__review.historyCalls)).toBe(0);await page.evaluate(()=>window.__review.releaseEdit());await expect.poll(()=>page.evaluate(()=>window.__review.historyCalls)).toBe(1);
 const expected=operation==="Undo"?"1":"2";await expect.poll(()=>page.evaluate(()=>window.__review.displayPrice())).toBe(expected);await expect(page.locator(".formula-bar-input")).toHaveValue(expected);
 if(operation==="Redo")await expect(page.locator(".error-toast")).toContainText("nothing to redo after edit");
});
for(const scenario of ["export-first","export-repaired"] as const)test(`App publishes current validation after ${scenario}`,async({page})=>{
 await boot(page,scenario);if(scenario==="export-repaired") {await edit(page);await expect(page.locator(".formula-bar-input")).toHaveValue("2");await expect(page.locator(".compile-errors")).toContainText("1");}
 await exportProject(page);await expect(page.locator(".build-mode-toggle")).not.toContainText("pending");await expect(page.locator(".compile-freshness")).toHaveCount(0);await expect(page.locator(".compile-errors")).toHaveCount(0);await expect(page.getByRole("region",{name:"Table diagnostics"})).toHaveCount(0);await expect.poll(()=>page.evaluate(()=>window.__review.getTableCalls)).toBe(2);
 expect(await page.evaluate(()=>({exports:window.__review.exportCalls,compiles:window.__review.compileCalls}))).toEqual({exports:1,compiles:0});
});

test("Export queued after an edit publishes current input without repeating compilation",async({page})=>{
 await boot(page,"queued");await edit(page);await expect.poll(()=>page.evaluate(()=>window.__review.pendingEdit)).toBe(true);await exportProject(page);await page.evaluate(()=>window.__review.releaseEdit());
 await expect.poll(()=>page.evaluate(()=>window.__review.exportCalls)).toBe(1);await expect(page.locator(".build-mode-toggle")).not.toContainText("pending");await expect(page.locator(".formula-bar-input")).toHaveValue("2");
 expect(await page.evaluate(()=>({compiles:window.__review.compileCalls,reads:window.__review.getTableCalls}))).toEqual({compiles:0,reads:1});
});
