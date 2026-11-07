import {expect, test} from "@playwright/test";
import {arrayEntries, structValues} from "../src/table/valueCodecs";

test("struct editing retains quoted exact numeric fields beside an edited sibling", () => {
    const values = structValues('{"exact-value": 9007199254740993.0, enabled: true}', ["exact-value", "enabled"]);
    expect(values).toEqual({"exact-value": "9007199254740993.0", enabled: "true"});
    values!.enabled = "false";
    expect(values!["exact-value"]).toBe("9007199254740993.0");
});

test("quoted labels containing separators and escaped quotes keep nested numeric expressions", () => {
    expect(structValues('{"a:b": [1e-400, {n: 0.12345678901234567890123456789}], "a\\\"b": 1e400}', ["a:b", 'a"b']))
        .toEqual({"a:b": "[1e-400, {n: 0.12345678901234567890123456789}]", 'a"b': "1e400"});
    expect(arrayEntries('[9007199254740993, 1e-400, {n: 1e400}]'))
        .toEqual(["9007199254740993", "1e-400", "{n: 1e400}"]);
});

test("struct values that cannot be reconstructed completely request raw expression editing", () => {
    for (const source of [
        '{known: 1, extra: 9007199254740993}',
        '{known: 1, known: >=0}',
        '{known: 1} & {known: >=0}',
        '{// retain this comment\nknown: 1}',
        '{#"known"#: 1}',
        '{known: """\nmultiline\n"""}',
        '{"\\U00000061": 1}',
        'null',
    ]) expect(structValues(source, ["known", "a"])).toBeNull();
    expect(structValues("{}", ["known"])).toEqual({});
});

test("labels with hidden or definition semantics keep the original CUE expression", () => {
    for (const label of ["_exact", "#Exact", "__proto__"]) {
        expect(structValues(`{${label}: 9007199254740993}`, [label])).toBeNull();
        expect(structValues(`{${JSON.stringify(label)}: 9007199254740993}`, [label])).toBeNull();
        // Missing values must not expose an editor that guesses label semantics.
        expect(structValues("{}", [label])).toBeNull();
    }
});

test("special bare labels cannot collide with ordinary quoted field names", () => {
    for (const label of ["amount?", "amount!", "A=amount", 'A="amount"', '"amount"?', '"amount"!', "[string]", "(amount)", "123", "a-b"]) {
        expect(structValues(`{${label}: 9007199254740993.0, enabled: true}`, ["amount", label, "enabled"])).toBeNull();
    }
    // The same text is safe when it is explicitly a string label.
    expect(structValues('{"amount?": 9007199254740993.0, enabled: true}', ["amount?", "enabled"]))
        .toEqual({"amount?": "9007199254740993.0", enabled: "true"});
});

test("ordinary identifiers and quoted labels preserve exact values and boolean siblings", () => {
    for (const label of ["amount", "amount2", "a_b", "$amount", "数值", "数值２", '"exact-value"', '"a:b"', '"a\\\"b"', '"a\\\\b"', '"中文"']) {
        const name = label.startsWith('"') ? JSON.parse(label) : label;
        const values = structValues(`{${label}: -9007199254740995.0, untouched: true, enabled: true}`, [name, "untouched", "enabled"]);
        expect(values).toEqual({[name]: "-9007199254740995.0", untouched: "true", enabled: "true"});
        values!.enabled = "false";
        expect(values![name]).toBe("-9007199254740995.0");
        expect(values!.untouched).toBe("true");
    }
});

test("incomplete and unsupported label parses retain the original expression", () => {
    for (const source of ['{amount?: 1', '{amount: [1}', '{"amount: 1}', '{amount: }', '{amount: 1, ...}', '{amount: 1, let A = 2}']) {
        expect(structValues(source, ["amount", "amount?", "A"])).toBeNull();
    }
});

test("array editing distinguishes an explicit empty list from unsupported expressions", () => {
    expect(arrayEntries("[]")).toEqual([]);
    for (const source of [
        "", "items", '({let seed = ["old", "two"], value: seed}).value',
        '*["old"] | [...string]', '["old"] & [...string]',
        '[for v in ["old", "two"] {v}]', '["old", ...string]',
        '["old", // keep\n "two"]', '[#"comma, quote"#, "two"]',
        '["""\ntext\n""", "two"]', '["\\(value)", "two"]',
        '[1,,2]', '[1', '[{n: 1]]', '["unterminated]', '[1] + [2]',
        '[items[0], "two"]', '[a=1, a]',
    ]) expect(arrayEntries(source), source).toBeNull();
});

test("literal array splitting preserves exact source tokens and nested ordinary literals", () => {
    expect(arrayEntries('[9007199254740993, -9007199254740995.0, 1e-400, 1e400,]'))
        .toEqual(['9007199254740993', '-9007199254740995.0', '1e-400', '1e400']);
    expect(arrayEntries('["中文", "引号\\\"逗号,反斜杠\\\\", [1, 2], {n: 1e400, "a:b": true}, null]'))
        .toEqual(['"中文"', '"引号\\\"逗号,反斜杠\\\\"', '[1, 2]', '{n: 1e400, "a:b": true}', 'null']);
    expect(arrayEntries('[{数值: 1, $value: false, "__proto__": 2,}, []]'))
        .toEqual(['{数值: 1, $value: false, "__proto__": 2,}', '[]']);
});
