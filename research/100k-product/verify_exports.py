#!/usr/bin/env python3
"""Verify actual Service CSVs against all 100k deterministic generated records."""
import csv,hashlib,importlib.util,json,sys
from decimal import Decimal
from pathlib import Path
root=Path(sys.argv[1]).resolve()
spec=importlib.util.spec_from_file_location('fixture_generator',root/'datapack-100k-stress-20261101/scripts/generate.py')
g=importlib.util.module_from_spec(spec);spec.loader.exec_module(g)
columns=g.COLUMNS
cases=[('local-perf-baseline-alpha',{'edit_manual'}),('local-perf-final',{'edit_manual'}),('local-perf-final-cold',{'edit_sequence','edit_before_export'}),('local-perf-final-delta',{'edit_sequence'})]
results=[]
for directory,stages in cases:
 folder=root/directory
 logs=[json.loads(line) for line in (folder/'measurements.jsonl').read_text().splitlines()]
 changes={}
 for record in logs:
  if record['stage'] not in stages:continue
  edit=record['edit'];value=edit['expression'] if edit['method']=='SetCellText' else json.loads(edit['expression'])
  changes.setdefault(edit['row_id'],{})[edit['column']]=value
 path=folder/'working-project/exports/stress.items.csv';seen=set();samples={}
 with path.open(newline='',encoding='utf8') as f:
  reader=csv.reader(f,strict=True);assert next(reader)==columns
  for count,(row_id,row) in enumerate(g.records(),1):
   row.update(changes.get(row_id,{}));actual=next(reader)
   assert len(actual)==10 and actual[0]==row_id and row_id not in seen
   seen.add(row_id)
   for col,value in zip(columns[1:],actual[1:]):
    expected=row[col]
    if col=='weight':assert Decimal(value)==Decimal(str(expected)),(directory,row_id,col,value,expected)
    else:
     expected=('true' if expected else 'false') if isinstance(expected,bool) else str(expected)
     assert value==expected,(directory,row_id,col,value,expected)
   if count in [1,2,3,50000,100000]:samples[str(count)]=dict(zip(columns,actual))
  assert count==100000 and next(reader,None) is None
 results.append({'case':directory,'actual_service_export':str(path.relative_to(root)),'rows':count,'unique_ids':len(seen),'columns':10,'verified_fields':1000000,'csv_bytes':path.stat().st_size,'csv_sha256':hashlib.sha256(path.read_bytes()).hexdigest(),'samples':samples})
original=root/'datapack-100k-stress-20261101/project/tables/items/items.cue'
source_hash=hashlib.sha256(original.read_bytes()).hexdigest()
for name in ['local-perf-baseline-saveas','local-perf-final-saveas']:
 for n in range(1,4):
  path=root/name/f'save-as-{n}/tables/items/items.cue'
  assert hashlib.sha256(path.read_bytes()).hexdigest()==source_hash
results.append({'save_as_copies_verified':6,'source_sha256':source_hash})
print(json.dumps(results,ensure_ascii=False,indent=2))
