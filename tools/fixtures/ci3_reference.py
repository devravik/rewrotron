#!/usr/bin/env python3
"""Pinned source-only fixture build, driver and read-only capture. No target inputs."""
import argparse
import base64
import hashlib
import json
import sys
from pathlib import Path
import shutil
import subprocess
import tempfile
import time
import urllib.parse
import uuid

REPO = Path(__file__).resolve().parents[2]
FIXTURE = REPO / 'fixtures/ci3-reference'


def canonical(value):
    return json.dumps(value, sort_keys=True, separators=(',', ':'), ensure_ascii=False).encode()


def ref(data):
    return 'sha256:' + hashlib.sha256(data).hexdigest()


def command(args, data=None, check=True, timeout=180):
    p = subprocess.run(args, input=data, stdout=subprocess.PIPE, stderr=subprocess.PIPE, timeout=timeout)
    if check and p.returncode:
        raise RuntimeError('fixed fixture operation failed: ' + p.stderr.decode(errors='replace')[-2500:])
    return p


def docker(*args, **kw):
    return command(['docker', *args], **kw)


def pins(root=FIXTURE):
    lock = json.loads((root / 'runtime.lock.json').read_bytes())
    if '.'.join(map(str, sys.version_info[:3])) != lock['python_version']:
        raise ValueError('observer Python runtime mismatch')
    actual_source = {p.relative_to(root).as_posix() for p in root.rglob('*') if p.is_file()
                     and p.name not in ('runtime.lock.json', 'README.md')
                     and 'dependencies' not in p.relative_to(root).parts and 'evidence' not in p.relative_to(root).parts}
    if actual_source != set(lock['source_files']):
        raise ValueError('undeclared fixture source file')
    for name, digest in lock['source_files'].items():
        if ref((root / name).read_bytes()) != 'sha256:' + digest:
            raise ValueError('fixture source/credential configuration changed: ' + name)
    for name, digest in lock['archives'].items():
        if ref((root / name).read_bytes()) != 'sha256:' + digest:
            raise ValueError('dependency archive changed: ' + name)
    if ref((root / 'composer.lock').read_bytes()) != 'sha256:' + lock['composer_lock_sha256']:
        raise ValueError('Composer lock changed')
    package = json.loads((root / 'composer.lock').read_bytes())['packages']
    if len(package) != 1 or package[0]['version'] != lock['ci3_version'] or package[0]['dist']['reference'] != lock['ci3_commit']:
        raise ValueError('CI3 dependency identity changed')
    for name in ('php_image', 'mysql_image'):
        if '@sha256:' not in lock[name] or len(lock[name].split('@sha256:')[1]) != 64:
            raise ValueError('floating runtime image')
        # No automatic network/pull or host PHP fallback during qualification.
        docker('image', 'inspect', lock[name])
    server = json.loads(docker('version', '--format', '{{json .Server}}').stdout)
    if server['Version'] != lock['docker_engine_version'] or server['Os'] != 'linux' or server['Arch'] != 'amd64':
        raise ValueError('unqualified container runtime')
    return lock


def source_manifest(root=FIXTURE):
    entries = []
    for p in sorted(root.rglob('*')):
        relative = p.relative_to(root).as_posix()
        if relative == 'README.md' or relative == 'evidence' or relative.startswith('evidence/'):
            continue
        if p.is_symlink():
            raise ValueError('fixture symlink forbidden')
        if p.is_dir():
            entries.append({'path': relative, 'kind': 'DIRECTORY'})
        else:
            entries.append({'path': relative, 'kind': 'FILE', 'sha256': ref(p.read_bytes())[7:], 'size_bytes': p.stat().st_size})
    return {'schema_version': '1.0', 'identity_algorithm': 'rwt-source-root-1', 'entries': entries,
            'exclusions': [{'path': 'README.md', 'reason': 'human documentation'}, {'path': 'evidence', 'reason': 'immutable derived observations; not runtime source'}],
            'completeness': 'COMPLETE', 'scanner_ref': ref(Path(__file__).read_bytes())}


def store(out, name, value):
    data = canonical(value)
    (out / name).write_bytes(data + b'\n')
    return ref(data)


def build(lock, tag, out, root=FIXTURE):
    print('BUILD ' + tag, flush=True)
    p = docker('build', '--network', 'none', '--no-cache', '--platform', lock['platform'], '-t', tag, str(root), timeout=600)
    (out / (tag.rsplit(':', 1)[1] + '-build.log')).write_bytes(p.stdout + p.stderr)
    image = json.loads(docker('image', 'inspect', tag).stdout)[0]['Id']
    # Content-based build identity excludes image creation timestamps and layer metadata.
    php = """$files=[]; foreach(['/app','/usr/local/lib/php/extensions'] as $root) {
    $it=new RecursiveIteratorIterator(new RecursiveDirectoryIterator($root,FilesystemIterator::SKIP_DOTS));
    foreach($it as $file) { if($file->isFile()) { $files[$file->getPathname()]=hash_file('sha256',$file->getPathname()); } }
    } $files['/usr/local/bin/php']=hash_file('sha256','/usr/local/bin/php'); ksort($files); echo json_encode($files);"""
    artifact = json.loads(docker('run', '--rm', '--network', 'none', '--read-only', image, 'php', '-r', php).stdout)
    runtime = docker('run', '--rm', '--network', 'none', image, 'php', '-r',
                     "echo json_encode(['php'=>PHP_VERSION,'extensions'=>get_loaded_extensions()]);").stdout
    observed = json.loads(runtime)
    if observed['php'] != lock['php_version'] or not set(lock['required_extensions']).issubset(observed['extensions']):
        raise ValueError('runtime/extension mismatch')
    return image, store(out, tag.rsplit(':', 1)[1] + '-artifact.json', artifact), observed


class SourceRuntime:
    """Owns only labeled disposable SOURCE resources. No caller endpoints/env/argv."""
    def __init__(self, image, lock, run_id, root=FIXTURE):
        self.image, self.lock, self.run_id, self.root = image, lock, run_id, root
        self.prefix = 'rwt-ci3-' + uuid.uuid4().hex[:12]
        self.network = self.prefix + '-net'
        self.db, self.app, self.sink, self.client = [self.prefix + '-' + s for s in ('db', 'source', 'sink', 'driver')]
        self.created = []
        self.network_created = False

    def __enter__(self):
        try:
            docker('network', 'create', '--internal', '--label', 'rewrotron.fixture=ci3-reference', self.network)
            self.network_created = True
            common = ['--detach', '--network', self.network, '--dns', '127.0.0.1', '--read-only', '--cap-drop', 'ALL',
                      '--security-opt', 'no-new-privileges', '--pids-limit', '128', '--cpus', '1',
                      '--label', 'rewrotron.fixture=ci3-reference', '--label', 'rewrotron.run=' + self.run_id]
            self.created.append(self.db)
            docker('run', *common, '--name', self.db, '--network-alias', 'db', '--user', '999:999', '--memory', '768m',
                   '--tmpfs', '/var/lib/mysql:rw,uid=999,gid=999,mode=0700,size=512m',
                   '--tmpfs', '/var/run/mysqld:rw,uid=999,gid=999', '--tmpfs', '/tmp:rw,uid=999,gid=999',
                   '--env', 'MYSQL_ROOT_PASSWORD=synthetic-root-only', self.lock['mysql_image'],
                   '--skip-log-bin', '--general-log=ON', '--log-output=TABLE', '--local-infile=OFF',
                   '--secure-file-priv=NULL', '--default-time-zone=+00:00', '--character-set-server=utf8mb4', '--collation-server=utf8mb4_bin')
            deadline = time.monotonic() + 90
            while True:
                p = docker('exec', self.db, 'mysql', '-h127.0.0.1', '-uroot', '-psynthetic-root-only', '--batch', '-e', 'SELECT 1', check=False)
                if p.returncode == 0:
                    break
                if time.monotonic() >= deadline:
                    raise RuntimeError('DB startup timeout: ' + docker('logs', self.db, check=False).stderr.decode()[-1500:])
                time.sleep(.25)
            docker('exec', '-i', self.db, 'mysql', '-uroot', '-psynthetic-root-only', data=(self.root / 'seed.sql').read_bytes())
            for name, alias, port, entry in [(self.app, 'source', '8080', 'router.php'), (self.sink, 'sink', '8081', 'sink.php')]:
                self.created.append(name)
                docker('run', *common, '--name', name, '--network-alias', alias, '--user', '1000:1000', '--memory', '128m',
                       '--tmpfs', '/tmp:rw,uid=1000,gid=1000,size=16m', '--tmpfs', '/state:rw,uid=1000,gid=1000,mode=0770,size=16m',
                       self.image, 'sh', '-c',
                       'mkdir -p /state/sessions /state/notes /state/sink; exec php -d opcache.enable=0 -S 0.0.0.0:' + port + ' ' + entry)
            inspect = json.loads(docker('inspect', self.app).stdout)[0]
            self.created.append(self.client)
            docker('run', *common, '--name', self.client, '--user', '1000:1000', '--memory', '64m', self.image, 'sleep', '600')
            self.launch = {k: inspect[k] for k in ('Id', 'Image', 'State', 'HostConfig', 'NetworkSettings')}
            self.launch['host_boot_id'] = Path('/proc/sys/kernel/random/boot_id').read_text().strip()
            self.prepared = True
            return self
        except BaseException:
            self.__exit__(None, None, None)
            raise

    def __exit__(self, *_):
        failures = []
        for name in reversed(self.created):
            p = docker('rm', '-f', name, check=False)
            if p.returncode and b'No such container' not in p.stderr:
                failures.append(name)
        if self.network_created and docker('network', 'rm', self.network, check=False).returncode:
            failures.append(self.network)
        if failures:
            raise RuntimeError('fixture cleanup resource leak: ' + ','.join(failures))


class Driver:
    """Only delivers requests and invokes the owned application worker."""
    def __init__(self, runtime):
        self.runtime = runtime

    def request(self, path, fields=None, cookie=None):
        if not path.startswith('/') or '\r' in path or '\n' in path:
            raise ValueError('invalid fixed fixture path')
        script = """$p=json_decode(stream_get_contents(STDIN),true);$headers=[];
        $c=curl_init('http://source:8080'.$p['path']); curl_setopt($c,CURLOPT_RETURNTRANSFER,true);
        curl_setopt($c,CURLOPT_TIMEOUT,5); curl_setopt($c,CURLOPT_HEADERFUNCTION,function($c,$line) use (&$headers) {
            $i=strpos($line,':'); if($i!==false){$headers[]=[substr($line,0,$i),trim(substr($line,$i+1))];} return strlen($line); });
        if($p['cookie']){curl_setopt($c,CURLOPT_COOKIE,$p['cookie']);}
        if($p['fields']!==null){curl_setopt($c,CURLOPT_POST,true);curl_setopt($c,CURLOPT_POSTFIELDS,http_build_query($p['fields']));}
        $body=curl_exec($c); if($body===false){fwrite(STDERR,curl_error($c));exit(1);}
        echo json_encode(['status'=>curl_getinfo($c,CURLINFO_HTTP_CODE),'headers'=>$headers,'body_base64'=>base64_encode($body)]);"""
        result = json.loads(docker('exec', '-i', self.runtime.client, 'php', '-r', script,
                                  data=canonical({'path': path, 'fields': fields, 'cookie': cookie})).stdout)
        if len(base64.b64decode(result['body_base64'])) > 65536:
            raise RuntimeError('HTTP capture truncated')
        return result

    def worker(self):
        return docker('exec', self.runtime.app, 'php', '/app/worker.php').returncode


class Observer:
    """Reads actual store state through fixed SELECTs and file capture only."""
    def __init__(self, runtime):
        self.runtime = runtime

    def snapshot(self):
        # Password values never enter captured DB projections.
        query = "SELECT id,name,role FROM fixture.users ORDER BY id; SELECT id,owner,name FROM fixture.items ORDER BY id; SELECT id,item,owner FROM fixture.audit ORDER BY id; SELECT id,item,status,receipt FROM fixture.jobs ORDER BY id;"
        database = docker('exec', self.runtime.db, 'mysql', '-uobserver', '-psynthetic-observer-only', '--batch', '-e', query).stdout.decode()
        php = """$out=[]; $it=new RecursiveIteratorIterator(new RecursiveDirectoryIterator('/state',FilesystemIterator::SKIP_DOTS));
        foreach($it as $f) { if($f->isFile()) { $out[$f->getPathname()]=['content'=>file_get_contents($f->getPathname()),'mode'=>decoct($f->getPerms() & 0777)]; }} ksort($out); echo json_encode($out);"""
        source = json.loads(docker('exec', self.runtime.app, 'php', '-r', php).stdout)
        sink = json.loads(docker('exec', self.runtime.sink, 'php', '-r', php).stdout)
        return {'side': 'SOURCE', 'database': database, 'source_files': source, 'sink_files': sink}


def body(response):
    return json.loads(base64.b64decode(response['body_base64']))


def cookie(response):
    values = [v for k, v in response['headers'] if k.lower() == 'set-cookie' and v.startswith('fixture_session=')]
    if len(values) != 1:
        raise ValueError('missing/ambiguous session cookie')
    return values[0].split(';', 1)[0]


def observe(runtime, out, run_id):
    driver, observer = Driver(runtime), Observer(runtime)
    observations, controls = [], []

    def step(name, path, expected, fields=None, jar=None):
        before = observer.snapshot()
        response = driver.request(path, fields, jar)
        after = observer.snapshot()
        # Do not persist synthetic password values in requests. Captured tokens are
        # fixture-only; preserving their exact relationships is part of this test.
        public_fields = None if fields is None else {k: ('[fixture-secret]' if k == 'password' else v) for k, v in fields.items()}
        observations.append({'side': 'SOURCE', 'scenario_id': 'ci3-reference', 'step_id': name, 'case_id': None,
                             'request': {'path': path, 'fields': public_fields, 'cookie': jar},
                             'before': before, 'response': response, 'after': after})
        store(out, run_id + '-partial.json', {'status': 'INCOMPLETE', 'side': 'SOURCE', 'observations': observations})
        controls.append({'id': name, 'expected_status': expected, 'observed_status': response['status'], 'ok': response['status'] == expected})
        if response['status'] != expected:
            raise RuntimeError(name + ': expected ' + str(expected) + '; captured ' + str(response))
        return response

    def check(name, outcome):
        controls.append({'id': name, 'ok': bool(outcome)})
        if not outcome:
            raise RuntimeError('runtime control failed: ' + name)

    step('health', '/health', 200)
    initial = observer.snapshot()
    check('fresh-state-reset', initial['database'] == 'id\tname\trole\n1\talice\towner\n2\tbob\tother\n' and not initial['source_files'] and not initial['sink_files'])
    step('missing-session', '/whoami', 401)
    step('failed-login', '/login', 401, {'name': 'alice', 'password': 'wrong'})
    alice_login = step('alice-login', '/login', 303, {'name': 'alice', 'password': 'fixture-password'}, 'fixture_session=' + 'a' * 64)
    alice = cookie(alice_login)
    check('fixation-rotation', alice != 'fixture_session=' + 'a' * 64)
    check('ordered-cookies-and-redirect', [k.lower() for k, _ in alice_login['headers']].count('set-cookie') == 2 and ['Location', '/whoami'] in [list(v) for v in alice_login['headers']])
    who = step('alice-continuity', '/whoami', 200, jar=alice)
    observed_session = json.loads(observations[-1]['after']['source_files']['/state/sessions/' + alice.split('=', 1)[1] + '.json']['content'])
    check('principal-continuity', body(who)['principal'] == 1 and body(who)['name'] == 'alice' and observed_session['principal'] == 1 and '\n1\talice\towner\n' in observations[-1]['after']['database'])
    check('flash-once', body(who)['flash'] == 'logged-in' and body(step('flash-consumed', '/whoami', 200, jar=alice))['flash'] is None)
    bob = cookie(step('bob-login', '/login', 303, {'name': 'bob', 'password': 'fixture-password'}))
    bob_csrf = body(step('bob-csrf', '/csrf', 200, jar=bob))['token']
    first = body(step('alice-csrf', '/csrf', 200, jar=alice))['token']
    token = body(step('alice-csrf-regenerate', '/csrf', 200, jar=alice))['token']
    check('csrf-regeneration', first != token)
    step('regenerated-token-rejected', '/create', 403, {'name': 'bad', 'csrf': first}, alice)
    step('wrong-csrf', '/create', 403, {'name': 'bad', 'csrf': 'wrong'}, alice)
    step('cross-session-csrf', '/create', 403, {'name': 'bad', 'csrf': bob_csrf}, alice)
    create = step('alice-create', '/create', 202, {'name': 'fixture-item', 'csrf': token}, alice)
    item_id = body(create)['id']
    check('insert-and-trigger-audit', '1\t1\tfixture-item\n' in observations[-1]['after']['database'] and '1\t1\t1\n' in observations[-1]['after']['database'])
    step('csrf-single-use', '/create', 403, {'name': 'bad', 'csrf': token}, alice)
    read = step('direct-pdo-read', '/item?id=' + str(item_id), 200, jar=alice)
    check('pdo-returned-item', body(read) == {'id': 1, 'owner': 1, 'name': 'fixture-item'})
    step('bob-denied-alice-item', '/item?id=' + str(item_id), 403, jar=bob)
    token = body(step('csrf-note', '/csrf', 200, jar=alice))['token']
    step('file-write', '/note', 200, {'text': 'fixture-note', 'csrf': token}, alice)
    check('file-content-and-mode', observer.snapshot()['source_files']['/state/notes/note.txt'] == {'content': 'fixture-note', 'mode': '640'})
    worker_before = observer.snapshot()
    check('pending-job-before-worker', 'pending' in worker_before['database'] and not worker_before['sink_files'])
    check('application-worker-exit', driver.worker() == 0)
    worker_after = observer.snapshot()
    observations.append({'side': 'SOURCE', 'step_id': 'worker', 'before': worker_before, 'after': worker_after})
    check('queued-sink-ack', len(worker_after['sink_files']) == 1 and 'acked' in worker_after['database'])
    driver.worker()
    check('no-duplicate-business-effect', observer.snapshot()['sink_files'] == worker_after['sink_files'])
    # Real time expiry; no verifier edits of session/token stores or simulated clocks.
    token = body(step('csrf-to-expire', '/csrf', 200, jar=alice))['token']
    time.sleep(11)
    step('expired-csrf', '/create', 403, {'name': 'bad', 'csrf': token}, alice)
    step('tampered-session', '/whoami', 401, jar='fixture_session=' + 'b' * 64)
    old_alice = alice
    alice = cookie(step('rotate-existing', '/login', 303, {'name': 'alice', 'password': 'fixture-password'}, alice))
    check('existing-session-rotated', old_alice != alice)
    step('old-session-invalidated', '/whoami', 401, jar=old_alice)
    token = body(step('csrf-logout', '/csrf', 200, jar=alice))['token']
    step('logout', '/logout', 200, {'csrf': token}, alice)
    step('logout-invalidates', '/whoami', 401, jar=alice)
    expiring = cookie(step('session-to-expire', '/login', 303, {'name': 'alice', 'password': 'fixture-password'}))
    print('OBSERVE real session expiry ' + run_id, flush=True)
    time.sleep(61)
    step('expired-session', '/whoami', 401, jar=expiring)
    # Whole-service egress and mount controls, outside measured business requests.
    for container in (runtime.app, runtime.sink):
        script = "$s=@fsockopen('1.1.1.1',80,$e,$m,1); echo $s?'OPEN':'DENIED'; if($s){fclose($s);exit(1);}"
        check('egress-denied-' + container.rsplit('-', 1)[1], docker('exec', container, 'php', '-r', script).stdout == b'DENIED')
    result = docker('exec', runtime.app, 'php', '-r', "exit(@file_put_contents('/app/unowned','bad')===false?0:1);", check=False)
    check('read-only-root', result.returncode == 0)
    check('db-egress-denied', docker('exec', runtime.db, 'timeout', '2', 'bash', '-c', 'echo >/dev/tcp/1.1.1.1/80', check=False).returncode != 0)
    check('external-dns-denied', docker('exec', runtime.app, 'php', '-r', "exit(gethostbyname('example.com')==='example.com'?0:1);").returncode == 0)
    denied = docker('exec', runtime.db, 'mysql', '-uobserver', '-psynthetic-observer-only', '-e', "INSERT INTO fixture.items(owner,name) VALUES(1,'checker-write')", check=False)
    check('observer-write-denied', denied.returncode != 0)
    check('no-checker-write', 'checker-write' not in observer.snapshot()['database'])
    # Server general-log diagnostics prove both mysqli and direct PDO reached the DB.
    queries = docker('exec', runtime.db, 'mysql', '-uroot', '-psynthetic-root-only', '--batch', '-e', "SELECT command_type,argument FROM mysql.general_log WHERE user_host LIKE 'fixture[%';").stdout.decode()
    check('driver-and-pdo-server-evidence', 'SELECT id,owner,name FROM items WHERE id=' in queries and 'INSERT INTO `items`' in queries)
    server = docker('exec', runtime.db, 'mysql', '-uobserver', '-psynthetic-observer-only', '--batch', '-e', 'SELECT VERSION(),@@sql_mode,@@collation_server,@@time_zone,@@transaction_isolation,@@autocommit; SHOW CREATE TABLE fixture.items; SHOW CREATE TABLE fixture.audit;').stdout.decode()
    # No password-bearing general log is persisted; retain only matching read/write diagnostics.
    queries = '\n'.join(line for line in queries.splitlines() if 'items' in line and not line.startswith('Connect'))
    projection = []
    for o in observations:
        if 'response' not in o:
            continue
        value = body(o['response'])
        if 'token' in value:
            value = {'token_type': 'opaque-hex-64', 'issuance_and_binding': 'tested-by-lifecycle-controls'}
        headers = []
        for k, v in o['response']['headers']:
            if k.lower() == 'date':
                continue
            if k.lower() == 'set-cookie' and v.startswith('fixture_session='):
                v = 'fixture_session=[observed-opaque-id];' + v.split(';', 1)[1]
            headers.append([k, v])
        projection.append({'step_id': o['step_id'], 'status': o['response']['status'], 'body': value, 'headers': headers})
    return {'side': 'SOURCE', 'scenario_id': 'ci3-reference', 'case_id': None,
            'run_id': run_id, 'observations': observations, 'controls': controls,
            'server_profile': server, 'server_diagnostics': queries, 'reproduction_projection': projection}


def main():
    p = argparse.ArgumentParser()
    p.add_argument('--output', type=Path, required=True)
    args = p.parse_args()
    out = args.output.resolve()
    out.mkdir(parents=True, exist_ok=False)
    lock = pins()
    manifest = source_manifest()
    store(out, 'source-manifest.json', manifest)
    contract = json.loads((FIXTURE / 'contracts.json').read_bytes())
    profile_ref = store(out, 'input-profile.json', {'runtime': lock, 'boundary': contract, 'mode': 'docker', 'side': 'SOURCE', 'fixture_only': True})
    summary = {'profile_ref': profile_ref, 'side': 'SOURCE', 'target_involved': False, 'runs': [], 'unqualified': contract['unqualified']}
    projections = []
    with tempfile.TemporaryDirectory(prefix='rwt-ci3-sealed-') as staging:
        sealed = Path(staging) / 'source'
        shutil.copytree(FIXTURE, sealed, ignore=shutil.ignore_patterns('README.md', 'evidence'))
        if source_manifest(sealed) != manifest:
            raise RuntimeError('incoherent source acquisition')
        pins(sealed)
        for i in range(3):
            tag = 'rewrotron-ci3-reference:qualified-' + str(i)
            if i < 2:
                image, artifact, observed_runtime = build(lock, tag, out, sealed)
            with SourceRuntime(image, lock, 'run-' + str(i), sealed) as runtime:
                result = observe(runtime, out, 'run-' + str(i))
                launch_ref = store(out, 'launch-' + str(i) + '.json', runtime.launch)
                corpus_ref = store(out, 'corpus-' + str(i) + '.json', result)
                projections.append(result['reproduction_projection'])
                summary['runs'].append({'corpus_ref': corpus_ref, 'launch_ref': launch_ref, 'build_ref': artifact, 'runtime': observed_runtime,
                                        'control_ids': [c['id'] for c in result['controls']], 'controls_passed': all(c['ok'] for c in result['controls'])})
            print('CLEANED run-' + str(i), flush=True)
    if any(p != projections[0] for p in projections):
        raise RuntimeError('observed bounded contracts differ across repetitions')
    summary['reproduction_projection_ref'] = store(out, 'reproduction-projection.json', projections[0])
    summary['variability_candidates'] = ['HTTP Date', 'session IDs', 'CSRF tokens', 'session/token expiry timestamps']
    summary['normalization_authorized'] = False
    summary['source_content_root_hash'] = ref(canonical({'identity': 'rwt-source-root-1', 'entries': manifest['entries']}))[7:]
    if summary['runs'][0]['build_ref'] != summary['runs'][1]['build_ref']:
        raise RuntimeError('independent clean builds changed content identity')
    if any(r['control_ids'] != summary['runs'][0]['control_ids'] or not r['controls_passed'] for r in summary['runs']):
        raise RuntimeError('repetition qualification differs')
    if source_manifest() != manifest:
        raise RuntimeError('source changed during qualification')
    summary['source_manifest_ref'] = ref(canonical(manifest))
    summary['dependency_lock_ref'] = ref((FIXTURE / 'composer.lock').read_bytes())
    summary['fixture_ref'] = ref(canonical(manifest))
    summary['runtime_ref'] = ref((FIXTURE / 'runtime.lock.json').read_bytes())
    summary['implementation_ref'] = ref(Path(__file__).read_bytes())
    summary['qualification_boundary'] = 'Executed ci3-reference-b1 control suite only; no target or customer qualification'
    summary['status'] = 'RUNTIME_QUALIFIED_WITHIN_BOUNDARY'
    store(out, 'qualification-report.json', summary)
    print('QUALIFIED ci3-reference-b1; three fresh runs; two clean builds', flush=True)


if __name__ == '__main__':
    main()
