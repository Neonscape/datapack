#!/usr/bin/env python3
"""Real Service and OS-node acceptance probe; no GUI or timing injection."""
import argparse,csv,decimal,hashlib,json,os,pathlib,stat,subprocess
P=pathlib.Path
ap=argparse.ArgumentParser();ap.add_argument('--binary',required=True);ap.add_argument('--out',required=True);ap.add_argument('--mode',choices=['core','full'],default='core');a=ap.parse_args()
out=P(a.out).resolve();out.mkdir(parents=True,exist_ok=False);binary=str(P(a.binary).resolve())
results=[];processes=[];completed=False
def dump(p,v):p.write_text(json.dumps(v,indent=2,ensure_ascii=False)+'\n')
def tree(root):
 r={}
 def walk(p):
  s=p.lstat();rel=str(p.relative_to(root));v={'type':stat.S_IFMT(s.st_mode),'mode':stat.S_IMODE(s.st_mode)}
  if p.is_symlink():v['link']=os.readlink(p)
  elif p.is_file():v['sha256']=hashlib.sha256(p.read_bytes()).hexdigest()
  r[rel]=v
  if stat.S_ISDIR(s.st_mode):
   for c in sorted(p.iterdir()):walk(c)
 walk(root);return r
class Service:
 def __init__(self,d):
  self.log=(d/'wire.jsonl').open('w');self.err=(d/'stderr.log').open('w')
  self.p=subprocess.Popen([binary,'-test.run=^TestSaveAsServicePipe$'],stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=self.err,text=True,env=dict(os.environ,SAVEAS_PIPE='1',SAVEAS_SETTINGS=str(d/'settings')));processes.append(self)
 def call(self,method,**kw):
  q=dict(Method=method,**kw);self.p.stdin.write(json.dumps(q)+'\n');self.p.stdin.flush()
  while True:
   line=self.p.stdout.readline()
   if not line:raise RuntimeError('Service pipe exited')
   if line.startswith('SAVEAS_JSON '):
    r=json.loads(line[12:]);self.log.write(json.dumps({'request':q,'response':r})+'\n');self.log.flush();return r
 def ok(self,m,**kw):
  r=self.call(m,**kw);assert not r['error'],r['error'];return r
 def close(self):
  if self.p.stdin.closed:return
  self.p.stdin.close();self.p.stdout.read();code=self.p.wait(timeout=20);self.log.close();self.err.close();assert code==0,code

def fixture(d):
 src=d/'source';(src/'tables').mkdir(parents=True);(src/'schema').mkdir();(src/'notes.txt').write_text('owned source sidecar\n')
 (src/'source.dtproj').write_text(json.dumps({'version':1,'tableRoots':['tables'],'schemaRoots':['schema'],'exclude':['exports','.git'],'export':{'directory':'exports','targets':['csv']}}))
 (src/'tables/item.cue').write_text('''package item
#PrimaryKey: string
#Row: {name: string, q: {"exact-value": *9007199254740993.0 | number}, xs: [...int]}
rows: {first: {name: "before", xs: [1, 2]}}
_tool: {table: {namespace: "probe", name: "item"}}
_export: {csv: {enabled: true, columns: ["id", "name", "q", "xs"]}}
''')
 return src

def consume(src):
 files=list((src/'exports').rglob('*.csv'));assert len(files)==1,files
 with files[0].open(newline='') as f:rows=list(csv.DictReader(f))
 assert len(rows)==1 and rows[0]['id']=='first' and rows[0]['name']=='saved',rows
 assert json.loads(rows[0]['q'],parse_float=decimal.Decimal,parse_int=decimal.Decimal)=={'exact-value':decimal.Decimal('9007199254740993')},rows
 assert json.loads(rows[0]['xs'])==[1,2],rows
 return files[0].read_bytes()

def scenario(slot,kind='dangling',operation='SaveProjectAs',alias=False):
 label=f'{operation}-{slot}-{kind}'+('-alias' if alias else '');d=out/label;d.mkdir();src=fixture(d)
 s=Service(d);s.ok('OpenProject',Path=str(src/'source.dtproj'))
 for value in ['saved','redo-value']:
  s.ok('UpdateCell',Update={'path':str(src/'tables/item.cue'),'rowID':'first','column':'name','expression':json.dumps(value)})
 s.ok('Undo');s.ok('Compile');s.ok('Export');source_csv=consume(src)
 target=d/'target';target.mkdir();outside=d/'outside';outside.mkdir();(outside/'sentinel').write_text('DO NOT CHANGE\n')
 link=target/slot
 if kind=='dangling':link.symlink_to(outside/'missing')
 elif kind=='live-file':link.symlink_to(outside/'sentinel')
 elif kind=='live-dir':(outside/'empty').mkdir();link.symlink_to(outside/'empty',target_is_directory=True)
 elif kind=='file':link.write_text('target owned data\n')
 elif kind=='directory':link.mkdir();(link/'owned').write_text('target owned directory\n')
 elif kind=='empty-dir':link.mkdir()
 elif kind!='absent':raise AssertionError(kind)
 dest=target/'copy.dtproj'
 if alias:
  (d/'alias').symlink_to(target,target_is_directory=True);dest=d/'alias/copy.dtproj'
 before=s.ok('Current');snap={'source':tree(src),'target':tree(target),'outside':tree(outside),'current':before};dump(d/'before.json',snap)
 r=s.call(operation,Path=str(dest));after={'source':tree(src),'target':tree(target),'outside':tree(outside),'response':r};dump(d/'after.json',after)
 expect_error=kind not in ['absent'] and not (operation=='CreateProject' and kind=='empty-dir')
 checks={'correct_error_presence':bool(r['error'])==expect_error,'source_unchanged':snap['source']==after['source'],'outside_unchanged':snap['outside']==after['outside']}
 if expect_error:
  checks.update(target_unchanged=snap['target']==after['target'],history_unchanged=before['historyState']==r['historyState'],current_unchanged=before['current']==r['current'],operation_released=not r['busy'])
 record={'case':label,'error':r['error'],'checks':checks,'passed':all(checks.values())};results.append(record);dump(d/'verdict.json',record)
 if not record['passed']:s.close();return
 record['lifecycle_complete']=False
 if expect_error:
  s.ok('Redo');s.ok('Undo')
  restored=s.ok('Current');assert restored['undo']==before['undo'] and restored['redo']==before['redo']
  assert tree(src)==snap['source']
  # Only this fixture-owned collision node is removed, after verifying preservation.
  if link.is_dir() and not link.is_symlink():
   for child in link.iterdir():child.unlink()
   link.rmdir()
  else:link.unlink()
  r=s.ok(operation,Path=str(dest));dump(d/'retry.json',r)
 assert r['current']['project']['path']==str(dest) and r['current']['projectEpoch']!=before['current']['projectEpoch']
 assert r['undo']==0 and r['redo']==0
 if operation=='SaveProjectAs':
  for rel in ['tables/item.cue','notes.txt']:
   assert (src/rel).read_bytes()==(target/rel).read_bytes()
   assert (src/rel).stat().st_ino!=(target/rel).stat().st_ino
  s.ok('Compile');s.ok('Export');assert consume(target)==source_csv
  dump(d/'consumer.json',{'source_csv_sha256':hashlib.sha256(source_csv).hexdigest(),'target_csv_sha256':hashlib.sha256(consume(target)).hexdigest(),'expected':{'id':'first','name':'saved','q_exact_value':'9007199254740993','xs':[1,2]},'numeric_json_type_checked':True})
  s.close();rd=d/'reopen';rd.mkdir();s=Service(rd);s.ok('OpenProject',Path=str(dest));s.ok('Compile');s.ok('Export');assert consume(target)==source_csv
  (target/'notes.txt').write_text('copy-only edit\n');assert (src/'notes.txt').read_text()=='owned source sidecar\n'
 assert tree(outside)==snap['outside']
 assert tree(src)==snap['source']
 dump(d/'lifecycle.json',{'same_service_retry':expect_error,'source_copy_independent':operation=='SaveProjectAs','new_process_reopen':operation=='SaveProjectAs','compile_export_consumer':operation=='SaveProjectAs','outside_unchanged':True,'source_unchanged':True})
 record['lifecycle_complete']=True
 dump(d/'verdict.json',record)
 assert not list(d.glob('.datapack-save-as-*'))
 s.close()

def source_links():
 for live in [False,True]:
  d=out/f'source-link-{live}';d.mkdir();src=fixture(d);outside=d/'sentinel';outside.write_text('KEEP')
  link=src/'linked.txt';link.symlink_to(outside if live else d/'missing');target=d/'target';target.mkdir()
  s=Service(d);s.ok('OpenProject',Path=str(src/'source.dtproj'));before=s.ok('Current');snap=tree(src)
  r=s.call('SaveProjectAs',Path=str(target/'copy.dtproj'));assert 'symbolic link' in r['error'] and tree(src)==snap and not list(target.iterdir()) and outside.read_text()=='KEEP' and before['historyState']==r['historyState']
  results.append({'case':f'source-link-{live}','error':r['error'],'passed':True});s.close()
try:
 for slot in ['copy.dtproj','notes.txt']:scenario(slot)
 if a.mode=='full':
  for slot in ['copy.dtproj','notes.txt']:
   for kind in ['live-file','live-dir','file','directory']:scenario(slot,kind)
  for slot in ['tables','schema','cue.mod']:scenario(slot)
  scenario('unused','absent',alias=True)
  for slot in ['copy.dtproj','tables','schema','cue.mod']:
   for kind in ['dangling','live-file','live-dir','file','directory']:scenario(slot,kind,'CreateProject')
  for slot in ['tables','schema','cue.mod']:scenario(slot,'empty-dir','CreateProject')
  scenario('unused','absent','CreateProject',alias=True)
  source_links()
 completed=True
finally:
 for p in processes:p.close()
 dump(out/'processes.json',[{'pid':p.p.pid,'exit_code':p.p.returncode,'waited':True} for p in processes])
 dump(out/'result.json',{'cases':results,'passed':completed and all(r['passed'] and r.get('lifecycle_complete',True) for r in results),'binary_sha256':hashlib.sha256(P(binary).read_bytes()).hexdigest()})
print(json.dumps({'passed':completed and all(r['passed'] and r.get('lifecycle_complete',True) for r in results),'cases':len(results),'failed':[r for r in results if not r['passed']]}))
raise SystemExit(0 if all(r['passed'] for r in results) else 1)
