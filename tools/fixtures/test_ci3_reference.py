"""Deterministic harness refusal tests; Docker runtime qualification is separate."""
import importlib.util
import json
from pathlib import Path
import shutil
import tempfile
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location('ci3_reference', Path(__file__).with_name('ci3_reference.py'))
harness = importlib.util.module_from_spec(spec)
spec.loader.exec_module(harness)


class FixtureRefusalTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name) / 'fixture'
        shutil.copytree(harness.FIXTURE, self.root, ignore=shutil.ignore_patterns('evidence', 'README.md'))

    def test_missing_runtime_never_falls_back_to_host(self):
        with patch.object(harness, 'docker', side_effect=RuntimeError('missing digest image')):
            with self.assertRaises(RuntimeError):
                harness.pins(self.root)

    def test_changed_lock_refuses_before_launch(self):
        with (self.root / 'composer.lock').open('ab') as f:
            f.write(b' ')
        with patch.object(harness, 'docker') as d:
            with self.assertRaisesRegex(ValueError, 'composer.lock'):
                harness.pins(self.root)
            d.assert_not_called()

    def test_changed_archive_refuses_before_launch(self):
        with (self.root / 'dependencies/ci3-3.1.13.zip').open('ab') as f:
            f.write(b'x')
        with patch.object(harness, 'docker') as d:
            with self.assertRaisesRegex(ValueError, 'archive changed'):
                harness.pins(self.root)
            d.assert_not_called()

    def test_source_change_invalidates_snapshot(self):
        before = harness.source_manifest(self.root)
        p = self.root / 'application/controllers/Reference.php'
        p.write_text(p.read_text() + '\n// changed source\n')
        after = harness.source_manifest(self.root)
        self.assertNotEqual(harness.ref(harness.canonical(before)), harness.ref(harness.canonical(after)))

    def test_missing_php_extension_blocks_build_qualification(self):
        result = lambda data: type('Result', (), {'returncode': 0, 'stdout': data, 'stderr': b''})()
        replies = [result(b''), result(b'[{"Id":"image"}]'), result(b'{}'),
                   result(b'{"php":"7.4.33","extensions":["json"]}')]
        lock = json.loads((self.root / 'runtime.lock.json').read_bytes())
        with patch.object(harness, 'docker', side_effect=replies):
            with self.assertRaisesRegex(ValueError, 'extension mismatch'):
                harness.build(lock, 'fixture:missing-ext', Path(self.tmp.name), self.root)

    def test_production_credentials_refused_before_execution(self):
        p = self.root / 'application/config/database.php'
        p.write_text(p.read_text().replace('synthetic-fixture-only', 'production-secret-canary'))
        with patch.object(harness, 'docker') as d:
            with self.assertRaisesRegex(ValueError, 'credential configuration changed'):
                harness.pins(self.root)
            d.assert_not_called()

    def test_undeclared_source_file_refuses_execution(self):
        (self.root / 'application/controllers/Unknown.php').write_text('<?php echo "unreviewed";')
        with patch.object(harness, 'docker') as d:
            with self.assertRaisesRegex(ValueError, 'undeclared'):
                harness.pins(self.root)
            d.assert_not_called()

    def test_source_symlink_is_not_followed(self):
        (self.root / 'application/escape').symlink_to('/etc/passwd')
        with self.assertRaisesRegex(ValueError, 'symlink'):
            harness.source_manifest(self.root)

    def test_cleanup_on_failed_start(self):
        calls = []
        def fake(*args, **kw):
            calls.append(args)
            if args[0] == 'run':
                raise RuntimeError('start failed')
            return type('Result', (), {'returncode': 0, 'stdout': b'', 'stderr': b''})()
        with patch.object(harness, 'docker', side_effect=fake):
            with self.assertRaises(RuntimeError):
                with harness.SourceRuntime('image', {'mysql_image': 'mysql'}, 'run'):
                    self.fail('start cannot succeed')
        self.assertTrue(any(c[:2] == ('rm', '-f') for c in calls))
        self.assertTrue(any(c[:2] == ('network', 'rm') for c in calls))


if __name__ == '__main__':
    unittest.main()
