"""Independent data contract assertions using Python CSV, JSON and Decimal."""
import csv
import json
import pathlib
import sys
from decimal import Decimal, InvalidOperation

root = pathlib.Path(sys.argv[1])
failures = []

def check(label, actual, expected):
    if actual != expected:
        failures.append(f"{label}: got {actual!r}, expected {expected!r}")

def read(name, header):
    with (root / name).open(encoding="utf-8", newline="") as stream:
        reader = csv.DictReader(stream, strict=True)
        check(name + " header", reader.fieldnames, header)
        rows = list(reader)
    check(name + " row width", all(None not in row and None not in row.values() for row in rows), True)
    check(name + " unique IDs", len(rows), len({row["id"] for row in rows}))
    return {row["id"]: row for row in rows}

def decimal_value(label, value):
    try:
        return Decimal(value)
    except InvalidOperation:
        failures.append(f"{label}: invalid decimal cell {value!r}")
        return None

def decode(value):
    return json.loads(value, parse_int=int, parse_float=Decimal)

rows = read("boundary.text.csv", ["id","文本","空值","空串","数组","结构","默认值","原值","计算值"])
check("text IDs", set(rows), {"中文行", "空集合"})
a = rows["中文行"]
check("quoted multiline text", a["文本"], '第一行,"引号"\n第二行\r\n第三行\t尾部')
check("nullable text", a["空值"], "")
check("empty text", a["空串"], "")
check("text array", decode(a["数组"]), ["中文", '逗号,引号"', "换\n行", ""])
check("flat structure", decode(a["结构"]), {"说明": '结构,"值"\n后行', "数量": 7, "开启": True})
check("CUE default", a["默认值"], "默认中文")
check("formula", int(a["计算值"]), 42)
b = rows["空集合"]
check("whitespace", b["文本"], "  首尾空白  ")
check("empty array", decode(b["数组"]), [])
check("zero structure", decode(b["结构"]), {"说明": "", "数量": 0, "开启": False})
check("zero formula", int(b["计算值"]), 0)

rows = read("boundary.numbers.csv", ["id","exactInteger","exactDecimal","integerArray","decimalStruct"])
expected = {
    "beyond_binary64": (9007199254740993, Decimal("0.12345678901234567890123456789"), [9007199254740993, -9007199254740993]),
    "beyond_int64": (18446744073709551615, Decimal("12345678901234567890.123456789"), [18446744073709551615]),
    "large_integer": (123456789012345678901234567890, Decimal("1e-100"), [123456789012345678901234567890]),
    "tiny_decimal": (1, Decimal("1e-400"), [1]),
    "huge_decimal": (1, Decimal("1e400"), [1]),
}
check("number IDs", set(rows), set(expected))
for row_id, (integer, decimal, array) in expected.items():
    row = rows[row_id]
    check(row_id + " integer", decimal_value(row_id + " integer", row["exactInteger"]), Decimal(integer))
    check(row_id + " decimal", decimal_value(row_id + " decimal", row["exactDecimal"]), decimal)
    check(row_id + " array", decode(row["integerArray"]), array)
    check(row_id + " struct", decode(row["decimalStruct"]), {"value": decimal})
if failures:
    print("FAIL: " + "\nFAIL: ".join(failures))
    sys.exit(1)
print("PASS: exact Unicode/CSV/null/empty/array/structure/default/formula/numeric data contracts")
