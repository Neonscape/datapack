// Narrow contracts for the Node helpers used by the real-Service browser fixture.
// The test runner already provides Node; no runtime package or compiler change.
declare module "node:fs" {
 export function readFileSync(path:string,encoding:"utf8"):string;
 export function mkdtempSync(prefix:string):string;
 export function rmSync(path:string,options:{recursive:boolean;force:boolean}):void;
}
declare module "node:os" {export function tmpdir():string;}
declare module "node:path" {export function resolve(...paths:string[]):string;export function join(...paths:string[]):string;}
declare module "node:child_process" {
 export function execFileSync(file:string,args:string[],options:{cwd:string;env:Record<string,string|undefined>;timeout:number}):Uint8Array;
}
