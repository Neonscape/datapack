"""Independent literal oracle: CSV reader + JSON Decimal, never JS expected output."""
import csv,json,sys
from pathlib import Path
from decimal import Decimal as D
root=Path(sys.argv[1]);out=Path(sys.argv[2]);assert not out.exists()
initial={
 'reference':['旧值','two'],'conjunction':['旧值','two'],'default_expr':['旧值','two'],
 'comprehension':['旧值','two'],'ellipsis':['旧值','two'],'comment':['旧值','two'],
 'raw':['raw \\" comma, slash\\\\','two'],'multiline':['中文, first\nsecond "quote"','two'],
 'literal':['中文','引号"逗号,反斜杠\\','tail'],'empty':[],'defaulted':['默认','保留']}
def load(label):
 with (root/label/'export.csv').open(newline='') as f:
  reader=csv.DictReader(f);assert reader.fieldnames==['id','a','n','q'];rows=list(reader)
 assert len(rows)==len(initial)
 result={}
 for r in rows:
  result[r['id']]={k:json.loads(r[k],parse_int=D,parse_float=D) for k in ['a','n','q']}
 return result

mode=json.loads((root/'result.json').read_text())['mode']
labels=['00-initial','01-reference-apply','02-undo-reference','03-add-from-empty','04-reopened'] if mode=='before' else ['00-initial','01-reference-apply','03-complex-no-ops','04-reference-add','05-undo-reference','06-redo-reference','07-literal-add','08-literal-reorder','09-literal-delete','10-exact-number-add','11-quoted-struct-sibling','12-clear-default','13-explicit-empty','14-undo-empty-to-default','15-fk-reference-add','16-reopened']
assert [s['label'] for s in json.loads((root/'steps.json').read_text())]==labels
receipts=[]
for label in labels:
 got=load(label);stage=int(label.split('-')[0])
 for row,arr in initial.items():
  want=arr
  if row=='reference':
   if mode=='before' and stage==1:want=[]
   if mode=='before' and stage>=3:want=['新项']
   if mode=='after' and (stage==4 or stage>=6):want=['旧值','two','新项']
  if mode=='after' and row=='literal':
   new='新项, "quote"\\path'
   if stage==7:want=arr+[new]
   elif stage==8:want=[new]+arr
   elif stage>=9:want=[new]+arr[:2]
  if mode=='after' and row=='default_expr' and stage>=12:want=[] if stage==13 else ['默认','保留']
  assert got[row]['a']==want,(label,row,got[row]['a'],want)
  numbers=[D('9007199254740993'),D('1e-400'),D('1e400')]
  if mode=='after' and row=='literal' and stage>=10:numbers += [D('-9007199254740995.0'),D('0.12345678901234567890123456789')]
  assert got[row]['n']==numbers,(label,row,got[row]['n'],numbers)
  enabled=not(mode=='after' and row=='literal' and stage>=11)
  assert got[row]['q']=={'exact-value':D('9007199254740993.0'),'enabled':enabled},(label,row,got[row]['q'])
 fields=len(got)*4
 if (root/label/'fks.csv').exists():
  with (root/label/'fks.csv').open(newline='') as f:fk=list(csv.DictReader(f))
  assert len(fk)==1 and fk[0]['id']=='alpha'
  assert json.loads(fk[0]['ids'])==(['first','second','third'] if mode=='after' and stage>=15 else ['first','second'])
  with (root/label/'targets.csv').open(newline='') as f:targets=list(csv.DictReader(f))
  assert targets==[{'id':'first','label':'第一'},{'id':'second','label':'第二'},{'id':'third','label':'第三'}]
  fields+=8
 receipts.append({'label':label,'rows':len(got),'csv_fields':fields,'all_array_json_and_exact_numbers_checked':True})
out.write_text(json.dumps({'passed':True,'hardcoded_oracle':True,'receipts':receipts,'total_csv_fields':sum(x['csv_fields'] for x in receipts)},indent=2)+'\n');print(out.read_text())
