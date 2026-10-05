# CI3 source qualification fixture — boundary ci3-reference-b1

This is the explicitly authorized M4 prerequisite (Checkpoint A). It is a synthetic,
source-only CI3 3.1.13 / PHP 7.4.33 / MySQL 8.0.36 fixture. It does not complete
M4 Unit 5, qualify Laravel, create customer approvals, or claim the full M2
50-route reference scope. [Fixture requirements](../../docs/10-development/FIXTURE-CORPUS.md)
and [runtime observation](../../docs/03-analysis/RUNTIME-OBSERVATION.md) retain ownership.

The eight routes in contracts.json exercise Alice/Bob login, principal continuity,
JSON sessions, fixation rotation, one-request flash, CSRF/session binding,
regeneration, single use, expiry, logout, owner checks, DB insertion and a server
trigger, direct PDO reads, a mode-preserving note file, and a queued application
worker delivering to a disposable acknowledged sink. All credentials and tokens
are synthetic fixture-only inputs. Configuration reads no ambient environment,
.env file or customer secret. The source has no target implementation dependency.

The custom session driver uses bounded opaque hexadecimal filenames and JSON;
it never deserializes objects. Sessions last 60 seconds; CSRF tokens last 10
seconds. Logout requires a valid CSRF token. The worker processes the explicitly
queued item job and records the sink receipt. No real payments/email occur. This
is a controlled fixture policy, not an inferred CI3 default or a customer policy.
The source and checked-in independent obligations derive from X19/X67 and the
M2 capture/reset/containment controls. Expected values never populate observations.

## Reproduce

Prerequisites: Linux amd64, Python 3.12.3, Docker Engine 29.8.2 (the executed
qualification profile), and the pinned Go 1.25.0 toolchain for the evidence reader.
The Docker daemon must support read-only roots, internal networks, non-root users,
tmpfs and resource limits. Host PHP, MySQL and Composer installations are unused.
Acquire the two exact images named in runtime.lock.json with `docker pull` before
qualification; setup acquisition may use the network. Neither build nor measured
execution downloads packages. Do not replace an unavailable digest with a tag.

```bash
python3 tools/fixtures/ci3_reference.py --output /tmp/ci3-reference-new-run
go run ./tools/fixtures/attest fixtures/ci3-reference /tmp/ci3-reference-new-run > /tmp/ci3-reference-new-run/attestation.json
python3 -m unittest discover -s tools/fixtures -p 'test_*.py' -v
```

Run from the repository root. Output must be a new directory. The harness seals
an exact source copy, builds twice with `--no-cache --network none`, installs the
Composer lock with plugins/scripts disabled using vendored checksum-pinned archives,
and verifies actual PHP/extension versions. Each of three runs gets fresh MySQL,
source, driver and sink containers and independent tmpfs state. Application and
service containers run as non-root, drop all capabilities, have no host mounts,
no privileged mode or Docker socket, and use an internal network with external
DNS disabled. No host ports are exposed. Only the application writes measured
state. The observer's DB account is SELECT-only; write-denial is tested. Every
owned resource is removed on success or failure; cleanup errors are failures.

Raw response headers preserve repetitions and cookie security/lifetime attributes.
Before/after DB, safe JSON session, note and sink snapshots come from execution,
not handwritten goldens. Passwords are excluded from DB projections and request
records. Recorded session/CSRF values are explicitly disposable fixture secrets;
keep their relationships and expiry decisions intact. Three fresh repetitions
compare bounded observed response contracts and independently test every required
lifecycle relation. Dynamic bytes and clocks remain in the raw captures; the
reproduction projection proposes variability only and authorizes no normalizer.

## Evidence and limits

The evidence directory contains raw capture fragments, source-manifest 1.0,
content-based build identities, real container/process launch records, runtime
settings, three source corpora, a bounded reproduction projection and the
qualification report. Checked-in build logs use lossless gzip storage. The evidence reader verifies the current source bytes and
locks using the existing rwt-source-root-1 domain contract, rejects altered,
incomplete or target-contaminated inputs, and derives attestation 2.1 through
DeriveObservationAttestation. A source manifest hash and content root hash are
separate identities. Image IDs can differ because image metadata contains build
times; all included application/PHP/extension file contents must match across
independent builds.

The DB attestation remains PARTIAL: committed rows, trigger effects and direct PDO
server diagnostics are captured, but read sets, locks, intermediate transaction
visibility and server outbound routines are not qualified. The overall attestation
therefore remains PARTIAL. Worker/external mechanisms also remain PARTIAL because
current causal-span/delivery-occurrence/completion ABI publication is unqualified. The qualification claim is only runtime execution and
observed obligations within ci3-reference-b1. Raw fragments are not automatically
upgraded to current verifier traces, paired reports or a compatibility PASS.
They support bounded source contract development, not customer execution admission.
No installed source adapter, target codec, cross-framework bridge/rollback,
Redis/concurrency, remember-me/reset/throttle, cron schedule, arbitrary native
extension or full-reference characterization is certified. Unit 5 must retain
these boundaries and qualify its own target lowerings and independent verification.

The CI3 archive includes the MIT license (also retained as CI3-LICENSE.txt).
Composer and Debian unzip distribution notices remain in their original pinned
archives. Archive bytes, upstream commit, Composer lock and image digests are
recorded in runtime.lock.json. Evidence under evidence/ is generated, immutable
checkpoint data, excluded from the runtime build and source identity.
