export function arrayEntries(expression: string): string[] {
    const source = expression.trim();
    if (!source.startsWith("[") || !source.endsWith("]")) return [];
    const entries: string[] = [];
    let start = 1;
    let depth = 0;
    let quote = "";
    let escaped = false;
    for (let index = 1; index < source.length - 1; index++) {
        const character = source[index];
        if (quote) {
            if (!escaped && character === quote) quote = "";
            escaped = !escaped && character === "\\";
            continue;
        }
        if (character === `"` || character === `'`) quote = character;
        else if (character === "[" || character === "{") depth++;
        else if (character === "]" || character === "}") depth--;
        else if (character === "," && depth === 0) {
            const entry = source.slice(start, index).trim();
            if (entry) entries.push(entry);
            start = index + 1;
        }
    }
    const entry = source.slice(start, -1).trim();
    if (entry) entries.push(entry);
    return entries;
}

export function structValues(expression: string, fieldNames: string[]): Record<string, string> {
    const source = expression.trim();
    const values: Record<string, string> = {};
    if (!source.startsWith("{") || !source.endsWith("}")) return values;
    let start = 1;
    let depth = 0;
    let quote = "";
    let escaped = false;
    const entries: string[] = [];
    for (let index = 1; index < source.length - 1; index++) {
        const character = source[index];
        if (quote) {
            if (!escaped && character === quote) quote = "";
            escaped = !escaped && character === "\\";
            continue;
        }
        if (character === `"` || character === `'`) quote = character;
        else if (character === "[" || character === "{") depth++;
        else if (character === "]" || character === "}") depth--;
        else if ((character === "," || character === "\n") && depth === 0) {
            entries.push(source.slice(start, index).trim());
            start = index + 1;
        }
    }
    entries.push(source.slice(start, -1).trim());
    for (const entry of entries) {
        const separator = entry.indexOf(":");
        if (separator < 0) continue;
        const name = entry.slice(0, separator).trim();
        if (fieldNames.includes(name)) values[name] = entry.slice(separator + 1).trim();
    }
    return values;
}
