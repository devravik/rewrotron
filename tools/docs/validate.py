#!/usr/bin/env python3
"""Validate repository documentation and undeployed contract artifacts offline."""
from pathlib import Path
import json
import re
import sqlite3
import sys
from urllib.parse import unquote, urldefrag

try:
    from jsonschema import Draft202012Validator, FormatChecker, RefResolver
except ImportError:
    sys.exit("Install the documented prerequisite: python3 -m pip install 'jsonschema>=4.10,<5'")

ROOT = Path(__file__).resolve().parents[2]
BASE = 'https://rewrotron.dev/specs/'
errors = []
schemas = {}
json_files = {}


def reject_network(uri):
    raise ValueError(f'Reference is not in the local schema registry: {uri}')


for path in sorted((ROOT / 'specs').rglob('*.json')):
    try:
        document = json.loads(path.read_text())
        json_files[path] = document
        if isinstance(document, dict) and '$schema' in document:
            Draft202012Validator.check_schema(document)
            schema_id = document.get('$id')
            if schema_id != BASE + str(path.relative_to(ROOT / 'specs')):
                errors.append(f'{path.relative_to(ROOT)}: schema ID does not match its local path')
            if schema_id in schemas:
                errors.append(f'Duplicate schema ID: {schema_id}')
            schemas[schema_id] = document
    except Exception as exc:
        errors.append(f'{path.relative_to(ROOT)}: {exc}')


def references(value):
    if isinstance(value, dict):
        for key, child in value.items():
            if key == '$ref':
                yield child
            else:
                yield from references(child)
    elif isinstance(value, list):
        for child in value:
            yield from references(child)


def resolve(uri, current=None):
    base, fragment = urldefrag(uri)
    document = schemas.get(base or current)
    if document is None:
        raise ValueError(f'Unregistered schema reference: {uri}')
    if fragment:
        if not fragment.startswith('/'):
            raise ValueError(f'Unsupported schema anchor: {uri}')
        for segment in fragment[1:].split('/'):
            segment = unquote(segment).replace('~1', '/').replace('~0', '~')
            document = document[segment]
    return document


for schema_id, schema in schemas.items():
    for reference in references(schema):
        try:
            resolve(reference, schema_id)
        except Exception as exc:
            errors.append(str(exc))


def validator(uri):
    base = urldefrag(uri)[0]
    return Draft202012Validator(
        resolve(uri),
        resolver=RefResolver(base_uri=base, referrer=schemas[base], store=schemas,
                             handlers={'http': reject_network, 'https': reject_network}),
        format_checker=FormatChecker(),
    )


fixture_count = 0
fixture_file = ROOT / 'specs/examples/validation-cases.json'
for fixture in json_files.get(fixture_file, []):
    fixture_count += 1
    try:
        actual_errors = list(validator(fixture['schema']).iter_errors(fixture['instance']))
        valid = not actual_errors
        if valid != fixture['valid']:
            detail = '; '.join(err.message for err in actual_errors[:3])
            errors.append(f"Fixture {fixture['name']}: expected valid={fixture['valid']}; {detail}")
    except Exception as exc:
        errors.append(f"Fixture {fixture['name']}: {exc}")

for filename, schema_path in [('behavior-contract.json', 'behavior/behavior-contract.schema.json'),
                             ('behavior-case.json', 'behavior/behavior-case.schema.json'),
                             ('project.json', 'project/project.schema.json'),
                             ('trace.json', 'schemas/trace-schema.json')]:
    try:
        validator(BASE + schema_path).validate(json_files[ROOT / 'specs/examples' / filename])
    except Exception as exc:
        errors.append(f'Example {filename}: {exc}')

registry_count = 0
for filename, key, name_key in [('cli/commands.json', 'commands', 'command'),
                               ('mcp/tools.json', 'tools', 'name')]:
    path = ROOT / 'specs' / filename
    entries = json_files.get(path, {}).get(key, [])
    names = [entry[name_key] for entry in entries]
    if len(names) != len(set(names)) or not entries:
        errors.append(f'{filename}: empty or duplicate registry')
    for entry in entries:
        registry_count += 1
        try:
            resolve(entry['input_schema'])
            resolve(entry['output_schema'])
            for variant in entry.get('output_variants', []):
                resolve(variant)
            documentation = (ROOT / ('docs/06-agent/CLI-SPEC.md' if key == 'commands'
                                     else 'docs/06-agent/MCP-SPEC.md')).read_text()
            if entry[name_key] not in documentation:
                errors.append(f'{filename}: undocumented operation {entry[name_key]}')
        except Exception as exc:
            errors.append(f'{filename}: {exc}')

link_count = 0
for path in sorted([ROOT / 'README.md', ROOT / 'REWROTRON_PROJECT_PLAN.md',
                    *ROOT.glob('docs/**/*.md')]):
    content = re.sub(r'```[^\n]*\n[\s\S]*?```', '', path.read_text())
    for target in re.findall(r'(?<!!)\[[^\]]*\]\(([^\s)]+)(?:\s+[^)]*)?\)', content):
        target = target.strip('<>')
        if target.startswith(('https://', 'http://', 'mailto:', '#')):
            continue
        link_count += 1
        local = unquote(target.split('#', 1)[0])
        if local.startswith('file://') or not (path.parent / local).exists():
            errors.append(f'{path.relative_to(ROOT)}: broken/nonportable link {target}')

# This validates DDL and invariant constraints, not the future Go migration runner.
sql = (ROOT / 'migrations/0001_initial.sql').read_text()
doc = (ROOT / 'docs/08-storage/SQLITE-SCHEMA.md').read_text()
match = re.search(r'```sql\n([\s\S]*?)\n```', doc)
if not match or match.group(1).strip() != sql.strip():
    errors.append('SQLITE-SCHEMA DDL does not match migrations/0001_initial.sql')
connection = sqlite3.connect(':memory:')
connection.execute('PRAGMA foreign_keys=ON')
try:
    connection.executescript(sql)
    expected_tables = set(re.findall(r'CREATE TABLE (\w+)', sql))
    actual_tables = {row[0] for row in connection.execute("SELECT name FROM sqlite_master WHERE type='table'")}
    if not expected_tables.issubset(actual_tables):
        errors.append('Initial migration did not create its declared tables')
    timestamp = '2026-10-03T00:00:00Z'
    connection.execute('INSERT INTO projects(id,name,root_path,status,created_at,updated_at) VALUES(?,?,?,?,?,?)',
                       ('fixture', 'Fixture', '/fixture', 'DETECTED', timestamp, timestamp))
    connection.execute('INSERT INTO artifacts VALUES(?,?,?,?,?,?,?,?)',
                       ('artifact', 'fixture', '0' * 64, 'sha256/00/' + '0' * 62,
                        'application/json', 2, '{}', timestamp))
    connection.execute('INSERT INTO source_detections VALUES(?,?,?,?,?)',
                       ('det', 'fixture', '0' * 64, 'artifact', timestamp))
    connection.execute('INSERT INTO baselines VALUES(?,?,?,?,?,?,?,?)',
                       ('0' * 64, 'fixture', 'det', 'artifact', '{}', '{}', '0' * 64, timestamp))
    connection.execute('INSERT INTO user_confirmations VALUES(?,?,?,?,?,?,?,?,?,?,?)',
                       ('confirm', 'fixture', 'GATE_1_SOURCE_DETECTION', '0' * 64,
                        '0' * 64, 'local-user', 'CLI', 'action', 'artifact', '{}', timestamp))
    connection.execute('INSERT INTO events(event_id,project_id,event_type,entity_type,entity_id,payload_json,request_id,timestamp) VALUES(?,?,?,?,?,?,?,?)',
                       ('evt', 'fixture', 'GATE_CONFIRMED', 'confirmation', 'confirm', '{}', 'req', timestamp))
    connection.commit()
    for table in ['source_detections', 'baselines', 'user_confirmations', 'events']:
        for statement in [f'UPDATE {table} SET project_id=project_id', f'DELETE FROM {table}']:
            try:
                connection.execute(statement)
                errors.append(f'{table}: append-only rule was not enforced')
            except sqlite3.IntegrityError:
                pass
    try:
        connection.execute('UPDATE projects SET status=? WHERE id=?', ('PARITY_ACHIEVED', 'fixture'))
        errors.append('Project lifecycle CHECK allowed an obsolete state')
    except sqlite3.IntegrityError:
        pass
    try:
        connection.execute('INSERT INTO artifacts VALUES(?,?,?,?,?,?,?,?)',
                           ('bad', 'unknown', '1' * 64, 'bad', 'application/json', 0, '{}', timestamp))
        errors.append('Foreign key allowed an artifact without a project')
    except sqlite3.IntegrityError:
        pass
    if list(connection.execute('PRAGMA foreign_key_check')):
        errors.append('Foreign key check failed')
    connection.rollback()
    connection.execute('BEGIN')
    connection.execute("UPDATE projects SET revision=revision+1 WHERE id='fixture'")
    connection.rollback()
    if connection.execute("SELECT revision FROM projects WHERE id='fixture'").fetchone()[0] != 0:
        errors.append('Transaction rollback did not preserve project revision')
except Exception as exc:
    errors.append(f'Initial SQL migration: {exc}')
finally:
    connection.close()

if errors:
    for error in errors:
        print('ERROR:', error, file=sys.stderr)
    sys.exit(1)
print(f'PASS: {len(schemas)} schemas; {fixture_count} positive/negative fixtures; '
      f'{registry_count} command/tool entries; {link_count} local links; '
      f'{len(expected_tables)} initial SQL tables and ownership/immutability/rollback checks.')
print('Scope: documentation and initial contract artifacts only; no Go/race/framework/runtime tests executed.')
