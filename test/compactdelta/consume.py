#!/usr/bin/env python3
"""Independent stdlib CSV/Decimal consumer; never uses binary floating point.
Expected values are the declared operation inputs. Reviewer uses a separately
hardcoded phase oracle to test this plan as well as the real export/DTO output.
"""
import csv,decimal,json,pathlib,sys
root=pathlib.Path(sys.argv[1]).resolve();report=pathlib.Path(sys.argv[2]).resolve()
assert not report.exists(),'never overwrite evidence'
checks=[]
def equal_value(text,want):
 if 'number' in want:return decimal.Decimal(text)==decimal.Decimal(want['number'])
 if 'string' in want:return text==want['string']
 if 'bool' in want:return text==str(want['bool']).lower()
 if 'null' in want:return text=='' # existing CSV null/empty ambiguity, not resolved here
 raise AssertionError(want)
try:
 for export in json.loads((root/'exports.json').read_text()):
  label=export['label'];folder=root/label
  with (folder/'export.csv').open(newline='') as stream: rows=list(csv.DictReader(stream))
  expected=export['expected'];assert {r['id'] for r in rows}==set(expected)
  for row in rows:
   assert set(row)=={'id',*expected[row['id']]}
   for column,want in expected[row['id']].items():
    assert equal_value(row[column],want),(label,row['id'],column,row[column],want)
    checks.append({'label':label,'row':row['id'],'column':column,'surface':'csv','passed':True})
  state=json.loads((folder/'state.json').read_text())
  for cell in state['rawCells']:
   want=expected[cell['row']][cell['column']]
   if cell.get('formula'):
    assert cell['raw']=='='+cell['expression']
    assert equal_value(cell['display'],want)
    checks.append({'label':label,'row':cell['row'],'column':cell['column'],'surface':'formula-display-only','passed':True});continue
   expression=cell.get('expression','');display=cell.get('display','')
   if not expression:
    # Before compile, a cleared optional/default cell may be absent. CSV is authority.
    assert label in ('24-app-clear-default','26-redo-clear','27-reopened'),(label,cell)
    continue
   if 'number' in want:
    for key in ('expression','display','raw'):assert equal_value(cell[key],want),(label,key,cell,want)
   elif 'string' in want:
    assert json.loads(expression)==want['string'],(label,cell,want)
    assert display==want['string'];assert cell['raw'] in (expression,display)
   elif 'null' in want:assert expression==display==cell['raw']=='null'
   else:assert equal_value(expression,want) and display==expression and cell['raw']==expression
   checks.append({'label':label,'row':cell['row'],'column':cell['column'],'surface':'expression-display-raw','passed':True})
 result={'passed':True,'exports':len(json.loads((root/'exports.json').read_text())),'checks':checks,'decimal_context_used_for_arithmetic':False,'csv_null_empty_ambiguity_unchanged':True}
except Exception as e:
 result={'passed':False,'error':repr(e),'checks':checks}
report.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items() if k!='checks'},ensure_ascii=False));sys.exit(0 if result['passed'] else 1)
