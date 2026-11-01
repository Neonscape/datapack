import {expect, test, type Page} from "@playwright/test";

type SessionControl = {delay(path:string):void; release():void; fail(path:string):void; pending:number; completed:number; failed:number; removeOnUndo():void};
const control = (page:Page) => page.evaluateHandle(() => (window as unknown as {__session:SessionControl}).__session);

test.beforeEach(async({page})=>{
    await page.addInitScript(()=>{
        let project=0, hold="", fail="", names=["item","hero","enemy"], definitions=["Temporary","Other"], removeOnUndo=false, release: (()=>void)|null=null;
        const state={pending:0,completed:0,failed:0,removeOnUndo:()=>{removeOnUndo=true;},delay:(path:string)=>{hold=path;},release:()=>{release?.();},fail:(path:string)=>{fail=path;}};
        const document=(path:string)=>({path,namespace:"game",name:path.split("/").at(-1)!.replace(".cue",""),columns:[{name:"label",schemaExpression:"string",schema:{kind:"string"}}],rows:[{id:"one",values:{label:{expression:'"text"',display:"text"}}}],localSchema:[{name:"LocalID",kind:"type",type:"string",schema:{kind:"string"},schemaExpression:"string"}],diagnostics:[]});
        const view=()=>({project:{path:`/project${project}/game.dtproj`,root:`/project${project}`},settings:{build:{mode:"manual"}},tree:{root:{kind:"project",name:`Project${project}`,children:[...(project===1?names:["newitem"]).map(name=>({kind:"table",name,path:`/project${project}/${name}.cue`,children:[{kind:"localSchema",name:"local_schema",path:`/project${project}/${name}.cue`}]})),...(project===1?[{kind:"schemaFile",name:"common.cue",path:"/project1/common.cue"}]:[])]}},schemas:{definitions:project===1?definitions.map(name=>({id:`common.#${name}`,kind:"type",path:"/project1/common.cue",schema:{kind:"string"},usages:[]})):[],diagnostics:[]},foreignKeys:{diagnostics:[]},compile:{mode:"manual",pending:false,diagnostics:[]},canUndo:true,canRedo:true});
        const api={GetSettings:async()=>({language:"en",theme:"light"}),GetExportFormats:async()=>[],SetLanguage:async()=>({}),SetTheme:async()=>({}),SelectProject:async()=>{project++;return view();},Undo:async()=>{if(removeOnUndo){definitions=definitions.filter(name=>name!=="Temporary");removeOnUndo=false;}const snapshot=view();if(hold==="history"){hold="";state.pending++;await new Promise<void>(resolve=>{release=resolve;});state.pending--;state.completed++;}return {workspace:snapshot};},DeleteTable:async(update:{path:string})=>{names=names.filter(name=>!update.path.endsWith(`/${name}.cue`));return {workspace:view()};},RenameTable:async(update:{path:string;name:string})=>{names=names.map(name=>update.path.endsWith(`/${name}.cue`)?update.name:name);return {table:document(`/project${project}/${update.name}.cue`),workspace:view()};},GetPublicSchema:async(path:string)=>({path,package:"common",definitions:definitions.map(name=>({name,kind:"type",type:"string",schema:{kind:"string"},schemaExpression:"string"}))}),GetTable:async(path:string)=>{
            if(path===hold){hold="";state.pending++;await new Promise<void>(resolve=>{release=resolve;});state.pending--;state.completed++;}
            if(path===fail){fail="";state.failed++;throw new Error("temporary read failed");}
            return document(path);
        }};
        Object.assign(window,{go:{main:{App:api}},__session:state});
    });
    await page.goto("/");
    await openProject(page);
});
async function openProject(page:Page){
    await page.getByRole("button",{name:"File",exact:true}).click();
    await page.getByRole("navigation",{name:"Application menu"}).getByRole("button",{name:"Open project",exact:true}).click();
}
async function undo(page:Page){
    await page.getByRole("button",{name:"Edit",exact:true}).click();
    await page.getByRole("navigation",{name:"Application menu"}).getByRole("button",{name:/^Undo /}).click();
}

test("refresh keeps tabs opened or closed while a read is pending",async({page})=>{
    for(const name of ["item","hero"])await page.getByRole("button",{name,exact:true}).click();
    const state=await control(page);await state.evaluate(s=>s.delay("/project1/item.cue"));
    await undo(page);
    await expect.poll(()=>state.evaluate(s=>s.pending)).toBe(1);
    await page.getByRole("tab",{name:"game.hero ×",exact:true}).locator(".tab-close").click();
    await page.getByRole("button",{name:"enemy",exact:true}).click();
    await state.evaluate(s=>s.release());
    await expect.poll(()=>state.evaluate(s=>s.completed)).toBe(1);
    await expect(page.getByRole("tab",{name:"game.enemy ×",exact:true})).toBeVisible();
    await expect(page.getByRole("tab",{name:"game.hero ×",exact:true})).toHaveCount(0);
});

test("late table read from previous project cannot open a tab in the new project",async({page})=>{
    const state=await control(page);await state.evaluate(s=>s.delay("/project1/item.cue"));
    await page.getByRole("button",{name:"item",exact:true}).click();
    await expect.poll(()=>state.evaluate(s=>s.pending)).toBe(1);
    await openProject(page);
    await expect(page.getByRole("button",{name:"newitem",exact:true})).toBeVisible();
    await state.evaluate(s=>s.release());
    await expect.poll(()=>state.evaluate(s=>s.completed)).toBe(1);
    await expect(page.locator(".tab-bar .tab")).toHaveCount(0);
});

test("temporary refresh read failure keeps the still-present table tab",async({page})=>{
    await page.getByRole("button",{name:"hero",exact:true}).click();
    const state=await control(page);await state.evaluate(s=>s.fail("/project1/hero.cue"));
    await undo(page);
    await expect.poll(()=>state.evaluate(s=>s.failed)).toBe(1);
    await expect(page.getByRole("tab",{name:"game.hero ×",exact:true})).toBeVisible();
});


test("closing a repeated open invalidates an older pending read",async({page})=>{
 const state=await control(page);await state.evaluate(s=>s.delay("/project1/item.cue"));
 await page.getByRole("button",{name:"item",exact:true}).click();
 await expect.poll(()=>state.evaluate(s=>s.pending)).toBe(1);
 await page.getByRole("button",{name:"item",exact:true}).click();
 await page.getByRole("tab",{name:"game.item ×",exact:true}).locator(".tab-close").click();
 await state.evaluate(s=>s.release());
 await expect.poll(()=>state.evaluate(s=>s.completed)).toBe(1);
 await expect(page.locator(".tab-bar .tab")).toHaveCount(0);
});

test("an earlier slow open does not steal focus from the latest selection",async({page})=>{
 const state=await control(page);await state.evaluate(s=>s.delay("/project1/item.cue"));
 await page.getByRole("button",{name:"item",exact:true}).click();
 await expect.poll(()=>state.evaluate(s=>s.pending)).toBe(1);
 await page.getByRole("button",{name:"hero",exact:true}).click();
 await state.evaluate(s=>s.release());
 await expect.poll(()=>state.evaluate(s=>s.completed)).toBe(1);
 await expect(page.getByRole("tab",{name:"game.item ×",exact:true})).toBeVisible();
 await expect(page.getByRole("tab",{name:"game.hero ×",exact:true})).toHaveAttribute("aria-selected","true");
});

test("a late history response cannot restore the previous project view",async({page})=>{
 await page.getByRole("button",{name:"hero",exact:true}).click();
 const state=await control(page);await state.evaluate(s=>s.delay("history"));
 await undo(page);await expect.poll(()=>state.evaluate(s=>s.pending)).toBe(1);
 await openProject(page);
 await expect(page.getByRole("button",{name:"newitem",exact:true})).toBeVisible();
 await state.evaluate(s=>s.release());
 await expect.poll(()=>state.evaluate(s=>s.completed)).toBe(1);
 await expect(page.getByRole("button",{name:"newitem",exact:true})).toBeVisible();
 await expect(page.getByRole("button",{name:"hero",exact:true})).toHaveCount(0);
 await expect(page.locator(".tab-bar .tab")).toHaveCount(0);
});


test("deleting a document invalidates its pending open without canceling another tab",async({page})=>{
 const state=await control(page);await state.evaluate(s=>s.delay("/project1/item.cue"));
 await page.getByRole("button",{name:"item",exact:true}).click();await expect.poll(()=>state.evaluate(s=>s.pending)).toBe(1);
 await page.getByRole("button",{name:"item",exact:true}).click({button:"right"});
 await page.getByRole("menu",{name:"Project tree actions"}).getByRole("button",{name:"Delete table",exact:true}).click();
 await page.locator(".structure-dialog").getByRole("button",{name:"Delete table",exact:true}).click();
 await page.getByRole("button",{name:"hero",exact:true}).click();
 await state.evaluate(s=>s.release());await expect.poll(()=>state.evaluate(s=>s.completed)).toBe(1);
 await expect(page.getByRole("tab",{name:"game.item ×",exact:true})).toHaveCount(0);
 await expect(page.getByRole("tab",{name:"game.hero ×",exact:true})).toHaveAttribute("aria-selected","true");
});

test("table rename invalidates a pending local-schema read at the old path",async({page})=>{
 await page.getByRole("button",{name:"item",exact:true}).click();
 const state=await control(page);await state.evaluate(s=>s.delay("/project1/item.cue"));
 await page.getByRole("button",{name:"Expand item",exact:true}).click();
 await page.getByRole("button",{name:"Local schemas",exact:true}).click();
 await expect.poll(()=>state.evaluate(s=>s.pending)).toBe(1);
 await page.getByRole("button",{name:"item",exact:true}).click({button:"right"});
 await page.getByRole("button",{name:"Rename table",exact:true}).click();
 await page.locator(".structure-dialog input").nth(1).fill("renamed");
 await page.locator(".structure-dialog").getByRole("button",{name:"Apply",exact:true}).click();
 await expect(page.getByRole("tab",{name:"game.renamed ×",exact:true})).toBeVisible();
 await state.evaluate(s=>s.release());await expect.poll(()=>state.evaluate(s=>s.completed)).toBe(1);
 await expect(page.getByRole("tab",{name:"game.item (local) ×",exact:true})).toHaveCount(0);
 await expect(page.locator(".tab-bar .tab")).toHaveCount(1);
});

test("Undo clears a removed public definition peek when no tabs remain",async({page})=>{
 await page.getByRole("button",{name:"hero",exact:true}).click();
 await page.getByRole("tab",{name:"game.hero ×",exact:true}).locator(".tab-close").click();
 await page.locator(".schema-row").filter({hasText:"common.#Temporary"}).click();
 const state=await control(page);await state.evaluate(s=>s.removeOnUndo());
 await undo(page);
 await expect(page.locator(".schema-row").filter({hasText:"common.#Temporary"})).toHaveCount(0);
 await expect(page.locator(".schema-detail")).toHaveCount(0);
});

test("a pending history refresh retains a later Schema peek selection",async({page})=>{
 await page.getByRole("button",{name:"hero",exact:true}).click();
 const state=await control(page);await state.evaluate(s=>s.delay("/project1/hero.cue"));await undo(page);
 await expect.poll(()=>state.evaluate(s=>s.pending)).toBe(1);
 await page.locator(".schema-row").filter({hasText:"common.#Other"}).click();
 await state.evaluate(s=>s.release());await expect.poll(()=>state.evaluate(s=>s.completed)).toBe(1);
 await expect(page.locator(".schema-row-selected")).toContainText("common.#Other");
 await expect(page.locator(".schema-detail")).toContainText("common.#Other");
});


test("a late history view cannot overwrite a newer document command",async({page})=>{
 await page.getByRole("button",{name:"hero",exact:true}).click();
 const state=await control(page);await state.evaluate(s=>s.delay("history"));await undo(page);await expect.poll(()=>state.evaluate(s=>s.pending)).toBe(1);
 await page.getByRole("button",{name:"item",exact:true}).click({button:"right"});
 await page.getByRole("menu",{name:"Project tree actions"}).getByRole("button",{name:"Delete table",exact:true}).click();
 await page.locator(".structure-dialog").getByRole("button",{name:"Delete table",exact:true}).click();
 await expect(page.getByRole("button",{name:"item",exact:true})).toHaveCount(0);
 await state.evaluate(s=>s.release());await expect.poll(()=>state.evaluate(s=>s.completed)).toBe(1);
 await expect(page.getByRole("button",{name:"item",exact:true})).toHaveCount(0);
 await expect(page.getByRole("tab",{name:"game.hero ×",exact:true})).toHaveAttribute("aria-selected","true");
});
