import {test,expect,type Page} from "@playwright/test";
import type {TableDocument,SchemaNode} from "../src/types";

type Variant={expressions:string[];initial:string;nullable?:boolean;reference?:boolean;formula?:boolean};
type Probe={table:TableDocument;calls:{method:string;column?:string;expression?:string}[];draws:{text:string;color:string;font:string}[]};
declare global {interface Window{__enumLanguage:Probe}}
test.use({viewport:{width:1600,height:1000}});
async function cell(page:Page,column:number,row=0){const box=await page.locator(".tab-panel:visible canvas").first().boundingBox();expect(box).not.toBeNull();return{x:box!.x+(column===0?85:225+170*(column-1)),y:box!.y+50+30*row};}
async function language(page:Page,value:"en"|"zh-CN"){
 await page.getByRole("button",{name:/^(Settings|设置)$/}).click();
 const dialog=page.locator(".structure-dialog");await dialog.locator("select").first().selectOption(value);await dialog.getByRole("button",{name:/^(Apply|应用)$/}).click();await expect(page.locator("html")).toHaveAttribute("lang",value);
}
async function setup(page:Page,mode:"manual"|"auto",diagnostic:"field"|"row"|"enum"|"none",variant?:Variant){
 await page.addInitScript(({mode,diagnostic,variant})=>{
  const expressions=variant?.expressions??['"N"','"R"'];
  const schema:SchemaNode={kind:"enum",enumValues:expressions.map(expression=>({expression})),nullable:variant?.nullable};
  const initial=variant?.initial??(diagnostic==="enum"?'"BAD"':'"N"');
  let pending=false;
  const table:TableDocument={path:"/project/tables/probe.cue",namespace:"game",name:"probe",columns:[
   {name:"amount",schemaExpression:"int",schema:{kind:"integer"}},
   {name:"rarity",schemaExpression:variant?.reference?"common.#Rarity":expressions.join("|"),schema:variant?.reference?{kind:"reference",reference:"common.#Rarity"}:schema,effectiveSchema:variant?.reference?schema:undefined},
   {name:"total",schemaExpression:"int",schema:{kind:"integer"}}],
   rows:[{id:"one",values:{amount:{expression:'"invalid"',display:"invalid",formula:false},rarity:{expression:initial,display:variant?.formula?"<formula error>":(initial.startsWith('"')?JSON.parse(initial) as string:initial),formula:variant?.formula??false},total:{expression:"20 + 5",display:"25",formula:true}}}],
   localSchema:[],diagnostics:diagnostic==="none"?[]:[{rowID:"one",column:diagnostic==="row"?undefined:diagnostic==="enum"?"rarity":"amount",message:"existing row validation error"}]};
  const probe:Probe={table,calls:[],draws:[]};window.__enumLanguage=probe;
  const original=CanvasRenderingContext2D.prototype.fillText;
  CanvasRenderingContext2D.prototype.fillText=function(text,x,y,maxWidth){probe.draws.push({text,color:String(this.fillStyle),font:this.font});if(maxWidth===undefined)original.call(this,text,x,y);else original.call(this,text,x,y,maxWidth);};
  const view=()=>({project:{path:"/project/probe.dtproj",root:"/project"},settings:{build:{mode}},tree:{root:{kind:"project",name:"probe.dtproj",children:[{kind:"table",name:"probe",path:table.path}]}},schemas:{definitions:[],diagnostics:[]},foreignKeys:{diagnostics:[]},compile:{mode,pending,diagnostics:[]},canUndo:false,canRedo:false});
  const update=async(argument:{column:string;expression:string})=>{probe.calls.push({method:"UpdateCell",...argument});let display:string;try{display=String(JSON.parse(argument.expression));}catch{display=argument.expression;}
   table.rows[0].values[argument.column]={expression:argument.expression,display,formula:false};pending=mode==="manual";if(mode==="auto")table.diagnostics=table.diagnostics?.filter(d=>d.column!==argument.column);return structuredClone({table,workspace:view()});};
  (window as unknown as Window&{go:unknown}).go={main:{App:{GetSettings:async()=>({language:"en",theme:"light"}),GetExportFormats:async()=>[],SelectProject:async()=>view(),GetTable:async()=>structuredClone(table),UpdateCell:update,SetCellFormula:async(argument:{column:string;expression:string})=>{probe.calls.push({method:"SetCellFormula",...argument});table.rows[0].values[argument.column]={expression:argument.expression,display:"2",formula:true};pending=mode==="manual";return structuredClone({table,workspace:view()});},SetLanguage:async(lang:string)=>{probe.calls.push({method:"SetLanguage",expression:lang});return{language:lang,theme:"light"};},Compile:async()=>{pending=false;table.diagnostics=table.diagnostics?.filter(d=>d.column!=="rarity");return{workspace:view()};}}}};
 },{mode,diagnostic,variant});
 await page.goto("/");await page.getByRole("button",{name:"File",exact:true}).click();await page.getByLabel("Application menu").getByRole("button",{name:"Open project",exact:true}).click();await page.getByRole("button",{name:"probe",exact:true}).click();
}
for(const mode of ["manual","auto"] as const)for(const diagnostic of ["field","row","enum","none"] as const)test(`enum field remains editable with ${diagnostic} diagnostic in ${mode}`,async({page})=>{
 await setup(page,mode,diagnostic);const point=await cell(page,2);await page.mouse.dblclick(point.x,point.y);
 const select=page.locator("#portal select");await expect(select).toBeVisible();await select.selectOption({label:"R"});await page.mouse.click(point.x,point.y+120);
 await expect.poll(()=>page.evaluate(()=>window.__enumLanguage.calls.filter(c=>c.method==="UpdateCell"))).toEqual([{method:"UpdateCell",path:"/project/tables/probe.cue",rowID:"one",column:"rarity",expression:'"R"'}]);
 expect(await page.evaluate(()=>window.__enumLanguage.table.rows[0].values.total)).toEqual({expression:"20 + 5",display:"25",formula:true});
});
for(const mode of ["manual","auto"] as const)test(`language changes retain formula rendering in ${mode}`,async({page})=>{
 await setup(page,mode,"none");
 for(const lang of ["zh-CN","en","zh-CN"] as const){await language(page,lang);await page.evaluate(()=>window.__enumLanguage.draws=[]);const point=await cell(page,3);await page.mouse.click(point.x,point.y);await expect(page.locator(".formula-bar-input")).toHaveValue("=20 + 5");await expect.poll(()=>page.evaluate(()=>window.__enumLanguage.draws.filter(d=>d.text==="25").map(d=>d.color))).toContain("#1c5d73");}
 expect(await page.evaluate(()=>window.__enumLanguage.calls.filter(c=>c.method!=="SetLanguage"))).toEqual([]);
});

const literals=[
 {name:"number",expressions:["1","2"],initial:"1",target:"2"},
 {name:"boolean",expressions:["false","true"],initial:"false",target:"true"},
 {name:"nullable",expressions:["1","2"],initial:"1",target:"null",nullable:true},
 {name:"escaped string",expressions:[JSON.stringify("plain"),JSON.stringify('a"b\\c')],initial:JSON.stringify("plain"),target:JSON.stringify('a"b\\c')},
 {name:"invalid current",expressions:["1","2"],initial:JSON.stringify("BAD"),target:"2"},
 {name:"reference",expressions:["1","2"],initial:"1",target:"2",reference:true},
];
for(const mode of ["manual","auto"] as const)for(const literal of literals)test(`repairs ${literal.name} enum with source expression intact in ${mode}`,async({page})=>{
 const diagnostic=literal.name==="invalid current"?"enum":"field";
 await setup(page,mode,diagnostic,literal);const p=await cell(page,2);await page.mouse.dblclick(p.x,p.y);const select=page.locator("#portal select");await expect(select).toBeVisible();
 await select.selectOption({label:literal.target.startsWith('"')?JSON.parse(literal.target) as string:literal.target});await page.mouse.click(p.x,p.y+120);
 await expect.poll(()=>page.evaluate(()=>window.__enumLanguage.calls.filter(c=>c.method==="UpdateCell").map(c=>c.expression))).toEqual([literal.target]);
 expect(await page.evaluate(()=>window.__enumLanguage.table.rows[0].values.total)).toEqual({expression:"20 + 5",display:"25",formula:true});
 if(mode==="manual"&&diagnostic==="enum"){await page.getByRole("button",{name:"Compile",exact:true}).click();await expect(page.getByRole("region",{name:"Table diagnostics"})).toHaveCount(0);}
 await language(page,"zh-CN");await language(page,"en");
 expect(await page.evaluate(()=>window.__enumLanguage.table.rows[0].values.rarity.expression)).toBe(literal.target);
});
for(const kind of ["bar","inline"] as const)test(`repairs an invalid formula-backed enum through ${kind}`,async({page})=>{
 await setup(page,"manual","enum",{expressions:["1","2"],initial:"missing + 1",formula:true});const p=await cell(page,2);
 if(kind==="bar"){await page.mouse.click(p.x,p.y);const input=page.locator(".formula-bar-input");await expect(input).toHaveValue("=missing + 1");await input.click();await expect(input).not.toHaveAttribute("readonly","");await input.fill("=1 + 1");await input.press("Enter");}
 else{await page.mouse.dblclick(p.x,p.y);const input=page.locator(".inline-cell-editor");await expect(input).toBeVisible();await input.fill("=1 + 1");await input.press("Enter");}
 await expect.poll(()=>page.evaluate(()=>window.__enumLanguage.calls.filter(c=>c.method==="SetCellFormula").map(c=>c.expression))).toEqual(["1 + 1"]);
 expect(await page.evaluate(()=>window.__enumLanguage.table.rows[0].values.rarity)).toEqual({expression:"1 + 1",display:"2",formula:true});
});
