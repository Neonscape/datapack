import {expect,test} from "@playwright/test";
import {cellWithinBounds,clipboardRange,rangeWithinBounds} from "../src/table/selection";

test("grid coordinates include ID and draft cells but reject stale deleted columns",()=>{
 expect(cellWithinBounds(0,2,3,2)).toBe(true);
 expect(cellWithinBounds(3,0,3,2)).toBe(false);
 expect(cellWithinBounds(1,3,3,2)).toBe(false);
 expect(cellWithinBounds(-1,0,3,2)).toBe(false);
});
test("range validation covers both ends and rejects empty rectangles",()=>{
 expect(rangeWithinBounds({x:1,y:0,width:2,height:3},3,2)).toBe(true);
 expect(rangeWithinBounds({x:1,y:0,width:3,height:1},3,2)).toBe(false);
 expect(rangeWithinBounds({x:1,y:0,width:0,height:1},3,2)).toBe(false);
});
test("clipboard range excludes ID and draft while retaining valid multi-cell coordinates",()=>{
 const range={x:1,y:0,width:2,height:2};
 expect(clipboardRange({cell:[1,0],range},3,2)).toEqual(range);
 expect(clipboardRange({cell:[0,0],range:{...range,x:0}},3,2)).toBeNull();
 expect(clipboardRange({cell:[1,2],range:{x:1,y:2,width:1,height:1}},3,2)).toBeNull();
});
test("clipboard ranges do not outlive a structure edit or mutate their input",()=>{
 const range=Object.freeze({x:2,y:0,width:1,height:1});
 const current=Object.freeze({cell:[2,0] as const,range});
 expect(clipboardRange(current,3,1)).toEqual(range);
 expect(clipboardRange(current,2,1)).toBeNull();
 expect(current.range).toEqual({x:2,y:0,width:1,height:1});
});
