"""Real App/Service JSON pipe. No GUI, mounted React, or native Wails IPC."""
import csv, json, os, shutil, subprocess, sys, traceback
from decimal import Decimal
from pathlib import Path

binary, out = map(lambda s: Path(s).resolve(), sys.argv[1:3])
mode = sys.argv[3] if len(sys.argv) > 3 else 'minimal'
assert not out.exists(), 'new evidence directory required'
out.mkdir(parents=True)
processes = []

def write(name, value):
    (out/name).write_text(json.dumps(value, ensure_ascii=False, indent=2)+'\n')

class Pipe:
    def __init__(self, label):
        self.label, self.counter = label, 0
        self.stderr = (out/(label+'-stderr.log')).open('w')
        self.proc = subprocess.Popen([str(binary), '-test.run=^TestCompactConsistencyPipe$'],
            cwd=out, env=dict(os.environ, COMPACT_PIPE='1', COMPACT_SETTINGS=str(out/('settings-'+label)), DATAPACK_FAST_EDIT='1'),
            stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=self.stderr, text=True)
        self.record = {'label': label, 'pid': self.proc.pid}
        processes.append(self.record)

    def call(self, method, *args):
        self.counter += 1
        request = {'id': self.counter, 'method': method, 'args': args}
        self.proc.stdin.write(json.dumps(request)+'\n')
        self.proc.stdin.flush()
        while True:
            line = self.proc.stdout.readline()
            if not line:
                raise RuntimeError('App exited without response')
            with (out/(self.label+'-stdout.log')).open('a') as f:
                f.write(line)
            if line.startswith('COMPACT_JSON '):
                response = json.loads(line[13:])
                with (out/'wire.jsonl').open('a') as f:
                    f.write(json.dumps({'process': self.label, 'request': request, 'response': response}, ensure_ascii=False)+'\n')
                assert response['id'] == self.counter
                assert 'error' not in response, response
                return response['value']

    def close(self):
        self.proc.stdin.close()
        try:
            self.record['exit_code'] = self.proc.wait(timeout=10)
        except subprocess.TimeoutExpired:
            self.proc.kill()
            self.record['exit_code'] = self.proc.wait()
            raise
        finally:
            self.stderr.close()
            write('processes.json', processes)
        assert self.record['exit_code'] == 0

project = out/'project'
table = project/'tables/defaults.cue'
table.parent.mkdir(parents=True)
(project/'schema').mkdir()
project_file = project/'defaults.dtproj'
project_file.write_text(json.dumps({'version': 1, 'tableRoots': ['tables'], 'schemaRoots': ['schema'], 'export': {'directory': 'exports', 'targets': ['csv']}}))
fixture = Path(__file__).with_name('defaults.cue' if mode == 'lifecycle' else 'minimal.cue')
if mode == 'fixture':
    fixture = Path(sys.argv[4])
shutil.copy2(fixture, table)

def snapshot(label):
    directory = out/label
    directory.mkdir()
    shutil.copy2(table, directory/'source.cue')
    dto = pipe.call('GetTable', str(table))
    write(label+'/table.json', dto)
    write(label+'/source-table.json', pipe.call('ReadSourceTable', str(table)))
    compiled = pipe.call('Compile')
    write(label+'/compile.json', compiled)
    assert compiled['workspace']['compile']['diagnostics'] == []
    exported = pipe.call('Export')
    write(label+'/export-result.json', exported)
    assert exported['workspace']['compile']['diagnostics'] == []
    assert len(exported['files']) == 1
    shutil.copy2(exported['files'][0], directory/'export.csv')
    return dto

pipe = None
try:
    pipe = Pipe('first')
    write('open.json', pipe.call('OpenProject', str(project_file)))
    if mode == 'lifecycle':
        initial = snapshot('00-initial')
        cells = {row['id']: row['values'] for row in initial['rows']}
        assert 'q' not in cells['missing']
        assert cells['empty']['q']['expression'] == '{}'
        assert '9007199254740995.0' in cells['explicit']['q']['expression']
        update = {'path': str(table), 'rowID': 'missing', 'column': 'editable', 'expression': '{n: 9007199254740997.125, enabled: false}'}
        write('save-response.json', pipe.call('UpdateCellDelta', update))
        saved = snapshot('01-saved')
        saved_source = table.read_bytes()
        write('clear-response.json', pipe.call('ClearCell', update))
        cleared = snapshot('02-cleared')
        cleared_source = table.read_bytes()
        expression = next(r for r in cleared['rows'] if r['id'] == 'missing')['values']['editable']['expression']
        assert '9007199254740993.0' in expression
        write('undo-response.json', pipe.call('UndoDelta'))
        snapshot('03-undo-clear')
        assert table.read_bytes() == saved_source
        write('redo-response.json', pipe.call('RedoDelta'))
        snapshot('04-redo-clear')
        assert table.read_bytes() == cleared_source
        pipe.close()
        pipe = Pipe('reopened')
        write('reopen.json', pipe.call('OpenProject', str(project_file)))
        snapshot('05-reopened')
        assert table.read_bytes() == cleared_source
        assert processes[0]['pid'] != processes[1]['pid']
        assert (out/'04-redo-clear/export.csv').read_bytes() == (out/'05-reopened/export.csv').read_bytes()
        pipe.close()
        pipe = None
        write('result.json', {'passed': True, 'mode': mode, 'stages': 6, 'processes': processes,
            'scope': 'Real App/Service Compile/Export and persisted files; numeric verdict is in separate consume.py receipt. No GUI/React/native Wails IPC.'})
        sys.exit(0)
    shutil.copy2(table, out/'source.cue')
    write('table.json', pipe.call('GetTable', str(table)))
    compiled = pipe.call('Compile')
    write('compile.json', compiled)
    assert compiled['workspace']['compile']['diagnostics'] == []
    exported = pipe.call('Export')
    write('export-result.json', exported)
    assert exported['workspace']['compile']['diagnostics'] == []
    assert len(exported['files']) == 1
    shutil.copy2(exported['files'][0], out/'export.csv')
    pipe.close()
    pipe = None
    rows = list(csv.DictReader((out/'export.csv').open()))
    for row in rows:
        got = json.loads(row['q'], parse_int=Decimal, parse_float=Decimal)
        assert got['exact-value'] == Decimal('9007199254740993.0'), got
    write('result.json', {'passed': True, 'checked_rows': len(rows), 'processes': processes})
except Exception:
    (out/'failure.txt').write_text(traceback.format_exc())
    write('result.json', {'passed': False, 'error': traceback.format_exc(), 'processes': processes})
    raise
finally:
    if pipe is not None:
        pipe.close()
