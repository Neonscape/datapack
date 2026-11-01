export type GridRange={x:number;y:number;width:number;height:number};
type CurrentSelection={cell:readonly[number,number];range?:GridRange};

// Grid coordinates include the ID column and the extra unsaved draft row.
export function cellWithinBounds(column:number,row:number,columnCount:number,persistedRows:number) {
    return column>=0 && column<columnCount && row>=0 && row<=persistedRows;
}
export function rangeWithinBounds(range:GridRange,columnCount:number,persistedRows:number) {
    return range.width>0 && range.height>0 && cellWithinBounds(range.x,range.y,columnCount,persistedRows)
        && cellWithinBounds(range.x+range.width-1,range.y+range.height-1,columnCount,persistedRows);
}
// Range copy/paste excludes both identities and the draft row; their existing
// single-cell/native editor path remains owned by the grid library/editor.
export function clipboardRange(current:CurrentSelection|undefined,columnCount:number,persistedRows:number):GridRange|null {
    if(!current || current.cell[0]===0 || current.cell[1]>=persistedRows)return null;
    const range=current.range ?? {x:current.cell[0],y:current.cell[1],width:1,height:1};
    if(range.x===0 || !rangeWithinBounds(range,columnCount,persistedRows) || range.y+range.height>persistedRows)return null;
    return range;
}
