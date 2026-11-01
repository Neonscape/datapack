import {expect, test, type Page} from "@playwright/test";

type SessionControl = {delay(path:string):void; release():void; fail(path:string):void; pending:number; completed:number; failed:number};
const control = (page:Page) => page.evaluateHandle(() => (window as unknown as {__session:SessionControl}).__session);

test.beforeEach(async({page})=>{
    await page.addInitScript(()=>{
        let project=0, hold="", fail="", release: (()=>void)|null=null;
        const state={pending:0,completed:0,failed:0,delay:(path:string)=>{hold=path;},release:()=>{release?.();},fail:(path:string)=>{fail=path;}};
        const document=(path:string)=>({path,namespace:"game",name:path.split("/").at(-1)!.replace(".cue",""),columns:[{name:"label",schemaExpression:"string",schema:{kind:"string"}}],rows:[{id:"one",values:{label:{expression:'"text"',display:"text"}}}],localSchema:[],diagnostics:[]});
        const view=()=>({project:{path:`/project${project}/game.dtproj`,root:`/project${project}`},settings:{build:{mode:"manual"}},tree:{root:{kind:"project",name:`Project${project}`,children:(project===1?["item","hero","enemy"]:["newitem"]).map(name=>({kind:"table",name,path:`/project${project}/${name}.cue`}))}},schemas:{definitions:[],diagnostics:[]},foreignKeys:{diagnostics:[]},compile:{mode:"manual",pending:false,diagnostics:[]},canUndo:true,canRedo:true});
        const api={GetSettings:async()=>({language:"en",theme:"light"}),GetExportFormats:async()=>[],SetLanguage:async()=>({}),SetTheme:async()=>({}),SelectProject:async()=>{project++;return view();},Undo:async()=>{const snapshot=view();if(hold==="history"){hold="";state.pending++;await new Promise<void>(resolve=>{release=resolve;});state.pending--;state.completed++;}return {workspace:snapshot};},GetTable:async(path:string)=>{
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
