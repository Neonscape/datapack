import {test,expect,type Page} from "@playwright/test";
type Direction="Undo"|"Redo";
type HistoryProbe={calls:number;readPending:boolean;pending:boolean;releaseHistory():void;releaseRead():void;recover():void;value():string};
declare global {interface Window {__historyFlight:HistoryProbe}}

async function boot(page:Page,direction:Direction,mode:"delta"|"full") {
 await page.addInitScript(({direction,mode})=>{
  let project=0,revision=2,value=direction==="Undo"?"2":"1",first=true,recovering=false,tableReads=0;
  let historyRelease:(()=>void)|undefined,readRelease:(()=>void)|undefined;
  const state={calls:0,readPending:false,pending:false,releaseHistory:()=>historyRelease?.(),releaseRead:()=>readRelease?.(),recover:()=>{recovering=true;},value:()=>value};
  const path=()=>`/project${project}/items.cue`;
  // Availability deliberately remains advertised after success to inject a
  // subsequent exhausted-history error; it cannot erase the committed view.
  const view=()=>({projectEpoch:`epoch-${project}`,revision,project:{path:`/project${project}/data.dtproj`,root:`/project${project}`},settings:{build:{mode:"manual"}},tree:{root:{kind:"project",name:`Project${project}`,children:[{kind:"table",name:"items",path:path()}]}},schemas:{definitions:[]},foreignKeys:{},compile:{mode:"manual",pending:true,inputRevision:-1,diagnostics:[]},canUndo:true,canRedo:true});
  const table=()=>({projectEpoch:`epoch-${project}`,workspaceRevision:revision,path:path(),namespace:"game",name:"items",columns:[{name:"price",schemaExpression:"int",schema:{kind:"integer"}}],rows:[{id:"one",values:{price:{expression:value,display:value}}}],localSchema:[],diagnostics:[]});
  async function history(){
   state.calls++;
   if(!first&&!recovering)throw new Error(`nothing to ${direction.toLowerCase()}`);
   if(first){state.pending=true;await new Promise<void>(resolve=>{historyRelease=resolve;});state.pending=false;first=false;}
   const base=revision;value=recovering?"7":direction==="Undo"?"1":"2";recovering=false;revision++;
   return mode==="delta"?{delta:{path:path(),projectEpoch:`epoch-${project}`,baseRevision:base,revision,rowIndex:0,rowID:"one",column:"price",cell:{expression:value,display:value}},workspace:view()}:{workspace:view()};
  }
  Object.assign(window,{__historyFlight:state,go:{main:{App:{
   GetSettings:async()=>({language:"en",theme:"light"}),GetExportFormats:async()=>[],
   SelectProject:async()=>{project++;if(project>1){value="99";revision=0;}return view();},
   GetTable:async()=>{const snapshot=table();tableReads++;if(mode==="full"&&tableReads===2){state.readPending=true;await new Promise<void>(resolve=>{readRelease=resolve;});state.readPending=false;}return snapshot;},
   UndoDelta:history,RedoDelta:history,
  }}}});
 },{direction,mode});
 await page.goto("/");await page.getByRole("button",{name:"Open project",exact:true}).click();await page.getByRole("button",{name:"items",exact:true}).click();
 const bounds=await page.locator(".tab-panel:visible canvas").first().boundingBox();await page.mouse.click(bounds!.x+225,bounds!.y+50);
 await expect(page.locator(".formula-bar-input")).toHaveValue(direction==="Undo"?"2":"1");
}
async function keys(page:Page,direction:Direction,count:number){
 await page.evaluate(({direction,count})=>{for(let i=0;i<count;i++)document.body.dispatchEvent(new KeyboardEvent("keydown",{key:direction==="Undo"?"z":"y",ctrlKey:true,bubbles:true,cancelable:true}));},{direction,count});
}
async function menuDisabled(page:Page){
 await page.getByRole("button",{name:"Edit",exact:true}).click();
 const menu=page.getByRole("navigation",{name:"Application menu"});
 const result={undo:await menu.getByRole("button",{name:/^Undo /}).isDisabled(),redo:await menu.getByRole("button",{name:/^Redo /}).isDisabled()};
 await page.getByRole("button",{name:"Edit",exact:true}).click();return result;
}
async function historyMenu(page:Page,direction:Direction){await page.getByRole("button",{name:"Edit",exact:true}).click();await page.getByRole("navigation",{name:"Application menu"}).getByRole("button",{name:new RegExp(`^${direction} `)}).click();}
for(const mode of ["delta","full"] as const)for(const direction of ["Undo","Redo"] as const)test(`single flight ${direction} ${mode} preserves success before a later exhausted error`,async({page})=>{
 await boot(page,direction,mode);await keys(page,direction,2);await expect.poll(()=>page.evaluate(()=>window.__historyFlight.pending)).toBe(true);
 const duringRPC=await menuDisabled(page);await page.evaluate(()=>window.__historyFlight.releaseHistory());
 if(mode==="full") {await expect.poll(()=>page.evaluate(()=>window.__historyFlight.readPending)).toBe(true);const duringRead=await menuDisabled(page);await keys(page,direction,2);expect(duringRead).toEqual({undo:true,redo:true});await page.evaluate(()=>window.__historyFlight.releaseRead());}
 const expected=direction==="Undo"?"1":"2";
 await expect.poll(()=>page.evaluate(()=>window.__historyFlight.value())).toBe(expected);
 await expect(page.locator(".formula-bar-input")).toHaveValue(expected);
 expect(await page.evaluate(()=>window.__historyFlight.calls)).toBe(1);
 expect(duringRPC).toEqual({undo:true,redo:true});
 await expect(page.locator(".error-toast")).toHaveCount(0);
 await historyMenu(page,direction);await expect(page.locator(".error-toast")).toContainText(`nothing to ${direction.toLowerCase()}`);
 await expect(page.locator(".formula-bar-input")).toHaveValue(expected);
 expect(await page.evaluate(()=>window.__historyFlight.calls)).toBe(2);
 await page.evaluate(()=>window.__historyFlight.recover());await historyMenu(page,direction);
 await expect(page.locator(".formula-bar-input")).toHaveValue("7");await expect(page.locator(".error-toast")).toHaveCount(0);
 expect(await page.evaluate(()=>window.__historyFlight.calls)).toBe(3);
});

test("history read from a previous project cannot replace the new project",async({page})=>{
 await boot(page,"Undo","full");await keys(page,"Undo",1);await expect.poll(()=>page.evaluate(()=>window.__historyFlight.pending)).toBe(true);await page.evaluate(()=>window.__historyFlight.releaseHistory());await expect.poll(()=>page.evaluate(()=>window.__historyFlight.readPending)).toBe(true);
 await page.getByRole("button",{name:"File",exact:true}).click();await page.getByRole("navigation",{name:"Application menu"}).getByRole("button",{name:"Open project",exact:true}).click();await expect(page.locator(".project-name")).toContainText("Project2");await page.getByRole("button",{name:"items",exact:true}).click();
 await page.evaluate(()=>window.__historyFlight.releaseRead());const bounds=await page.locator(".tab-panel:visible canvas").first().boundingBox();await page.mouse.click(bounds!.x+225,bounds!.y+50);await expect(page.locator(".formula-bar-input")).toHaveValue("99");
 await expect.poll(()=>page.evaluate(()=>window.__historyFlight.value())).toBe("99");await expect(page.locator(".project-name")).toContainText("Project2");
});
