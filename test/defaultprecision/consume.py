"""Independent exact consumer of actual exported CSV/JSON, no Go/JS expected data."""
import csv, json, sys
from decimal import Decimal as D
from pathlib import Path

root, receipt = map(Path, sys.argv[1:3])
assert not receipt.exists()
expected = {
    'largeI': D('9007199254740993'), 'exact-value': D('9007199254740993.0'),
    'fraction': D('9007199254740993.125'), 'negative': D('-9007199254740995.0'),
    'long': D('0.12345678901234567890123456789'), 'huge': D('1e400'), 'tiny': D('1e-400'),
    'small': D('0.125'), 'whole': D('2'), 'enabled': True,
    'a:b': '中文, "引号"\\路径\n下一行', 'nullable': None,
}
stages = ['00-initial', '01-saved', '02-cleared', '03-undo-clear', '04-redo-clear', '05-reopened']
checks = 0
for stage in stages:
    with (root/stage/'export.csv').open(newline='') as f:
        rows = {row['id']: row for row in csv.DictReader(f)}
    assert set(rows) == {'missing', 'empty', 'explicit'}
    for key, row in rows.items():
        q = json.loads(row['q'], parse_int=D, parse_float=D)
        want = dict(expected)
        if key == 'explicit':
            want.update({'exact-value': D('9007199254740995.0'), 'enabled': False})
        assert q == want, (stage, key, q, want)
        for name, value in want.items():
            assert type(q[name]) is type(value), (stage, key, name, type(q[name]))
            checks += 1
        # Pin established small-number spelling in the actual JSON text.
        assert '"small":0.125,' in row['q'] and '"whole":2}' in row['q'], row['q']
        branch = json.loads(row['branch'], parse_int=D, parse_float=D)
        assert branch == {'n': D('9007199254740993.0'), 'enabled': True}
        editable = json.loads(row['editable'], parse_int=D, parse_float=D)
        edited = key == 'missing' and stage in ['01-saved', '03-undo-clear']
        assert editable == {'n': D('9007199254740997.125' if edited else '9007199254740993.0'), 'enabled': not edited}, (stage, key, editable)
        assert isinstance(branch['n'], D) and isinstance(editable['n'], D)
        assert row['normal'] == '0.125' and row['flag'] == 'false' and row['text'] == '普通文本' and row['emptyList'] == '[]'
        checks += 8
assert (root/'04-redo-clear/export.csv').read_bytes() == (root/'05-reopened/export.csv').read_bytes()
result = {'passed': True, 'stages': len(stages), 'field_checks': checks, 'numeric_json_types_checked': True,
    'method': 'Python csv and json parse_int/parse_float=Decimal; hardcoded requirements, actual exported fields',
    'reopened_csv_byte_equal': True}
receipt.write_text(json.dumps(result, indent=2)+'\n')
print(json.dumps(result))
