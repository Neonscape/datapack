import {test,expect} from "@playwright/test";

test("mutation dispatch is ordered and a rejected request does not block retry",async({page})=>{
 await page.goto("/");
 const result=await page.evaluate(async()=>{
  // Browser imports use the same module loaded by the application.
  const modulePath="/src/bridge/ordered.ts";
  const {orderedCall}=await import(modulePath);
  const events:string[]=[];let release!:()=>void;
  const first=orderedCall("UpdateCell",async()=>{events.push("A start");await new Promise<void>(resolve=>{release=resolve;});events.push("A commit");return "A";});
  const second=orderedCall("UpdateCell",async()=>{events.push("B start");events.push("B commit");return "B";});
  await Promise.resolve();await Promise.resolve();
  const before=[...events];release();await Promise.all([first,second]);
  await orderedCall("UpdateCell",async()=>{throw new Error("injected");}).catch(()=>events.push("error"));
  await orderedCall("UpdateCell",async()=>{events.push("retry");});
  return {before,events};
 });
 expect(result.before).toEqual(["A start"]);
 expect(result.events).toEqual(["A start","A commit","B start","B commit","error","retry"]);
});

test("heavy operation reports true stages, blocks commands, and unlocks on failure",async({page})=>{
 await page.addInitScript(()=>{
  let finish!:()=>void;let compiles=0;
  const view=()=>({projectEpoch:"epoch1",revision:0,project:{path:"/project/game.dtproj",root:"/project"},settings:{build:{mode:"manual"}},tree:{root:{kind:"project",name:"Project",children:[]}},schemas:{definitions:[]},foreignKeys:{},compile:{mode:"manual",pending:true,inputRevision:-1,diagnostics:[]},canUndo:false,canRedo:false});
  const api={GetSettings:async()=>({language:"en",theme:"light"}),GetExportFormats:async()=>[],SelectProject:async()=>view(),GetOperationState:async()=>({id:1,kind:"compile",phase:"validate",completed:1000,total:100000,running:true}),Compile:async()=>{compiles++;await new Promise<void>(resolve=>{finish=resolve;});throw new Error("compile injected failure");}};
  Object.assign(window,{go:{main:{App:api}},__operation:{release:()=>finish(),calls:()=>compiles}});
 });
 await page.goto("/");
 await page.getByRole("button",{name:"Open project",exact:true}).click();
 await expect(page.getByText("Not validated yet.",{exact:false})).toBeVisible();
 await page.keyboard.press("Control+Shift+B");
 await expect(page.getByRole("status")).toContainText("1000 / 100000");
 await expect(page.locator("header")).toHaveAttribute("inert","");
 await page.keyboard.press("Control+Shift+B");
 expect(await page.evaluate(()=> (window as unknown as {__operation:{calls:()=>number}}).__operation.calls())).toBe(1);
 await page.evaluate(()=> (window as unknown as {__operation:{release:()=>void}}).__operation.release());
 await expect(page.getByRole("status")).toHaveCount(0);
 await expect(page.locator("header")).not.toHaveAttribute("inert","");
 await expect(page.getByText("compile injected failure")).toBeVisible();
});

test("cell deltas preserve identity and reject stale, reordered and cross-project data",async({page})=>{
 await page.goto("/");
 const result=await page.evaluate(async()=>{
  const path="/src/workspace/cellDelta.ts";const {patchCell}=await import(path);
  const initial={path:"/table",projectEpoch:"p1",workspaceRevision:2,columns:[{name:"price"}],rows:[{id:"first",values:{price:{expression:"1",display:"1"}}},{id:"last",values:{price:{expression:"2",display:"2"}}}]};
  const delta={path:"/table",projectEpoch:"p1",baseRevision:2,revision:3,rowIndex:1,rowID:"last",column:"price",cell:{expression:"99",display:"99"}};
  const patched=patchCell(initial,delta);
  return {value:patched.rows[1].values.price.display,original:initial.rows[1].values.price.display,wrongEpoch:patchCell(initial,{...delta,projectEpoch:"p2"}),wrongRow:patchCell(initial,{...delta,rowID:"first"}),stale:patchCell(patched,delta),future:patchCell(initial,{...delta,baseRevision:3,revision:4})};
 });
 expect(result).toEqual({value:"99",original:"2",wrongEpoch:null,wrongRow:null,stale:null,future:null});
});
