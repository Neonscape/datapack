import {expect,test} from "@playwright/test";
import {document,workspace} from "../wailsjs/go/models";
import {documentResult,projectSettings,publicSchema,schemaResult,tableDocument,workspaceView} from "../src/bridge/models";

function viewDTO() {
 return new workspace.View({project:{path:"/project/game.dtproj",root:"/project"},settings:{build:{mode:"manual",autoCompileDebounceSeconds:0}},tree:{root:{kind:"project",name:"game",children:[{kind:"table",name:"item",path:"/project/item.cue",displayName:"道具"}]}},schemas:{definitions:[{id:"common.#ID",kind:"type",path:"/project/common.cue",displayName:"标识",schema:{kind:"string"},usages:[]}]},foreignKeys:{diagnostics:[]},compile:{mode:"manual",pending:true,diagnostics:[]},canUndo:true,canRedo:false});
}
function freeze<T>(value:T):T {
 if(value&&typeof value==="object"){
  for(const nested of Object.values(value))freeze(nested);
  Object.freeze(value);
 }
 return value;
}

test("generated table DTO maps without changing IDs, values or source expressions",()=>{
 const dto=freeze(new document.Table({path:"/project/item.cue",namespace:"game",name:"item",displayName:"道具",columns:[{name:"price",displayName:"价格",schema:{kind:"reference",reference:"shared.#Price",default:{expression:"0"}},effectiveSchema:{kind:"integer"},schemaExpression:"shared.#Price"}],rows:[{id:"one",values:{price:{expression:"0",display:"0",formula:false}}}],localSchema:[{name:"ID",kind:"type",type:"string",schema:{kind:"string"},schemaExpression:"string"}],cellComments:{one:{price:"中文备注"}}}));
 const before=JSON.stringify(dto);const result=tableDocument(dto);
 expect(result.path).toBe(dto.path);expect(result.columns[0].schema.reference).toBe("shared.#Price");
 expect(result.rows[0].values.price).toEqual({expression:"0",display:"0",formula:false});
 expect(result.cellComments?.one.price).toBe("中文备注");expect(JSON.stringify(result)).toBe(before);expect(JSON.stringify(dto)).toBe(before);
});

test("generated view preserves navigation identity and manual pending diagnostics",()=>{
 const dto=freeze(viewDTO());const before=JSON.stringify(dto);const result=workspaceView(dto);
 expect(result.tree.root.children?.[0]).toMatchObject({name:"item",displayName:"道具",path:"/project/item.cue"});
 expect(result.schemas.definitions[0]).toMatchObject({id:"common.#ID",kind:"type",displayName:"标识"});
 expect(result.compile).toMatchObject({mode:"manual",pending:true,diagnostics:[]});expect(result.canUndo).toBe(true);
 expect(JSON.stringify(result)).toBe(before);expect(JSON.stringify(dto)).toBe(before);
});

test("canceled project and zero schema/table halves remain safe sentinels",()=>{
 expect(workspaceView(new workspace.View({})).project.path).toBe("");
 const result=schemaResult(new workspace.SchemaEditResult({schema:{path:"/project/common.cue",package:"common",definitions:[]},table:{},workspace:viewDTO()}));
 expect(result.schema.path).toBe("/project/common.cue");expect(result.table.path).toBe("");expect(result.table.rows).toEqual([]);expect(result.table.localSchema).toEqual([]);
 expect(publicSchema(new document.PublicSchema({})).definitions).toEqual([]);
});

test("document result retains history flags and maps the common payload once",()=>{
 const result=documentResult(new workspace.HistoryResult({table:{path:"/project/item.cue",namespace:"game",name:"item",columns:[],rows:[],localSchema:[]},workspace:viewDTO(),canUndo:true,canRedo:false}));
 expect(result).toMatchObject({canUndo:true,canRedo:false,table:{path:"/project/item.cue"},workspace:{canUndo:true}});
});

test("project settings retain export selection and normalize empty lists",()=>{
 const result=projectSettings(new workspace.ProjectSettings({project:{path:"/project/game.dtproj",root:"/project",tableRoots:["tables"],schemaRoots:["schema"]},export:{directory:"exports",targets:["csv"],tables:{"game.item":false}},formats:null,tables:null,workspace:viewDTO()}));
 expect(result.export.tables).toEqual({"game.item":false});expect(result.project.tableRoots).toEqual(["tables"]);expect(result.formats).toEqual([]);expect(result.tables).toEqual([]);
});

test("unknown tagged DTO values fail at the bridge instead of entering UI state",()=>{
 const unknownTree=viewDTO();unknownTree.tree.root.kind="future-kind";
 expect(()=>workspaceView(unknownTree)).toThrow("Unsupported project tree kind");
 const unknownMode=viewDTO();unknownMode.compile.mode="future-mode";
 expect(()=>workspaceView(unknownMode)).toThrow("Unsupported compile mode");
 expect(()=>publicSchema(new document.PublicSchema({path:"x",definitions:[{name:"X",kind:"future-kind"}]}))).toThrow("Unsupported Schema kind");
});

test("PrintSchema request adapter preserves lexical input through generated models",async()=>{
 const {PrintSchema}=await import("../src/bridge");
 const descriptor=Object.getOwnPropertyDescriptor(globalThis,"window");let captured:unknown;
 Object.defineProperty(globalThis,"window",{configurable:true,value:{go:{main:{App:{PrintSchema:async(value:unknown)=>{captured=value;return "preserved";}}}}}});
 try {
  const source={kind:"array",nullable:false,element:{kind:"reference",reference:"shared.#Value"},requirements:[{kind:"compare",op:">=",value:"0"}]};
  expect(await PrintSchema(source)).toBe("preserved");
  expect(JSON.parse(JSON.stringify(captured))).toEqual(source);
 }finally{
  if(descriptor)Object.defineProperty(globalThis,"window",descriptor);else Reflect.deleteProperty(globalThis,"window");
 }
});
