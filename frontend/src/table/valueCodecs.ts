// This is a closed literal-list codec, not a CUE expression parser. Return null
// unless every token can be reconstructed without changing expression semantics.
export function arrayEntries(expression: string): string[] | null {
    const source = expression.trim();
    if (!source.startsWith("[") || !source.endsWith("]")) return null;
    const entries: string[] = [];
    const brackets: string[] = [];
    let normalized = "";
    let start = 1;
    for (let index = 0; index < source.length; index++) {
        const character = source[index];
        if (character === '"') {
            // JSON validation excludes CUE interpolation and non-JSON escapes.
            if (source.startsWith('"""', index)) return null;
            const begin = index++;
            for (; index < source.length; index++) {
                if (source[index] === "\\") index++;
                else if (source[index] === '"') break;
            }
            if (index >= source.length) return null;
            const token = source.slice(begin, index + 1);
            try { JSON.parse(token); } catch { return null; }
            normalized += token;
        } else if (/[-+0-9.]/.test(character)) {
            const token = /^-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?/.exec(source.slice(index))?.[0];
            if (!token) return null;
            normalized += token;
            index += token.length - 1;
        } else if (/[\p{L}_$#]/u.test(character)) {
            const token = /^[\p{L}_$#][\p{L}\p{Nd}_$#]*/u.exec(source.slice(index))![0];
            const after = source.slice(index + token.length);
            // Ordinary struct labels are the only non-JSON names we accept.
            // Definitions/hidden labels, references and comprehension clauses
            // need the original-expression editor, including inside containers.
            if (!token.startsWith("_") && !token.startsWith("#") && /^\s*:/.test(after)) normalized += JSON.stringify(token);
            else if (["true", "false", "null"].includes(token)) normalized += token;
            else return null;
            index += token.length - 1;
        } else if (character === "[" || character === "{") {
            brackets.push(character);
            normalized += character;
        } else if (character === "]" || character === "}") {
            if (brackets.pop() !== (character === "]" ? "[" : "{")) return null;
            if (brackets.length === 0 && index !== source.length - 1) return null;
            normalized += character;
        } else if (character === ",") {
            if (brackets.length === 1) {
                const entry = source.slice(start, index).trim();
                if (!entry) return null;
                entries.push(entry);
                start = index + 1;
            }
            // CUE permits trailing commas. Remove them only from validation
            // text, outside strings; returned entries always use source slices.
            if (!/^\s*[}\]]/.test(source.slice(index + 1))) normalized += character;
            else if (/[{[,]\s*$/.test(normalized)) return null;
        } else if (/[\s:]/.test(character)) {
            normalized += character;
        } else return null;
    }
    if (brackets.length !== 0) return null;
    // Use JSON only as a grammar check: never use its decoded numeric values.
    try { JSON.parse(normalized); } catch { return null; }
    const entry = source.slice(start, -1).trim();
    if (entry) entries.push(entry);
    return entries;
}

// Return null when this small editor codec cannot account for every entry.
// The caller can then edit the original CUE expression without dropping data.
export function structValues(expression: string, fieldNames: string[]): Record<string, string> | null {
    const source = expression.trim();
    const values: Record<string, string> = Object.create(null);
    // UI field names do not retain whether a schema label was quoted. CUE's
    // hidden/definition labels have different semantics from string labels.
    if (fieldNames.some(name => name.startsWith("_") || name.startsWith("#"))) return null;
    if (!source.startsWith("{") || !source.endsWith("}")) return null;
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
        // Comments, raw strings and multiline strings need the full CUE parser.
        if ((character === "/" && ["/", "*"].includes(source[index + 1])) ||
            (character === "#" && ["\"", "'"].includes(source[index + 1])) ||
            source.startsWith('"""', index) || source.startsWith("'''", index)) return null;
        if (character === `"` || character === `'`) quote = character;
        else if (character === "[" || character === "{") depth++;
        else if (character === "]" || character === "}") {
            if (--depth < 0) return null;
        }
        else if ((character === "," || character === "\n") && depth === 0) {
            entries.push(source.slice(start, index).trim());
            start = index + 1;
        }
    }
    if (depth !== 0 || quote) return null;
    entries.push(source.slice(start, -1).trim());
    for (const entry of entries) {
        if (!entry) continue;
        // Only ordinary identifiers can safely become quoted string labels.
        // Optional/required markers, aliases and other CUE label syntax must
        // retain the original expression, even if a schema field has that name.
        const match = /^("(?:[^"\\]|\\.)*"|[\p{L}$][\p{L}\p{Nd}_$]*)\s*:\s*([\s\S]+)$/u.exec(entry);
        if (!match) return null;
        let name = match[1];
        if (name.startsWith('"')) {
            try { name = JSON.parse(name); } catch { return null; }
        }
        if (!fieldNames.includes(name) || Object.prototype.hasOwnProperty.call(values, name)) return null;
        values[name] = match[2].trim();
    }
    return values;
}
