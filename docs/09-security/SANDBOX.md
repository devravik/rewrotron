# Sandbox execution contract

## Reference container profile

Runtime source/target execution uses separate non-root containers and independent DB/cache/session/filesystem/sink state. Use digest-pinned images, no privileged mode, no Docker socket mount, dropped capabilities, `no-new-privileges`, read-only application image with only declared scratch/output volumes, CPU/memory/process limits, and a closed internal network. Host source is mounted read-only or copied into the disposable instrumented image.

Mock HTTP/SMTP/queue/payment receivers live on the internal network. A mock cannot fall through to a real service. Do not place public/production secrets in the container. Deny internet egress and DNS resolution beyond the internal allowlist; validate IP literals, alternate schemes and redirects. A proxy environment variable alone cannot stop raw sockets or programs that ignore it.

The runner creates labeled resources with project/operation IDs and records cleanup ownership. On timeout, cancellation or restart, terminate the process tree, remove temporary networks/volumes, and preserve redacted evidence. Failed cleanup produces a visible resource leak finding.

## Build and execution separation

Acquiring Composer packages and images is a separate controlled setup operation. Dependency manifests can execute scripts/plugins, so installation occurs in an isolated build container and scripts are disabled until required ones are reviewed. Preserve package archives/image digests/locks for offline replay. “Local-first” means no cloud service is required; initial dependency acquisition may require network access or a supplied cache.

`mode: local` is a trusted-fixture convenience for deterministic development only. It provides no general containment guarantee and is rejected for arbitrary customer runtime observation or cutover evidence. `mode: docker` is the initial customer execution profile; aliases such as `sandbox` are removed from the canonical config.

The doctor command verifies container availability, image cache, network denial, mount restrictions, fixture restore, sink health and runtime compatibility before replay. Missing containment is `ERR_UNSAFE_ENVIRONMENT`, never a silently weaker fallback.

## Effect replay matrix

| Class | Replay requirement | Refusal behavior |
| --- | --- | --- |
| PURE | Controlled input/clock; no effect surface | Unknown surfaces remain incomplete |
| READ_ONLY | Owned snapshots/read endpoints; no production secrets or mutation | Unverified read-only claim blocks |
| LOCAL_WRITE | Owned DB/files/cache with proven full reset/isolation | Missing reset produces UNKNOWN/block |
| EXTERNAL_WRITE | Denied real egress; validated sink/simulator with receipts/acks/idempotency | No safe simulator means UNSUPPORTED/UNKNOWN |
| IRREVERSIBLE | Never blind replay; approved faithful simulator/retained boundary only | Real payments/transactions/SMS/cloud actions are prohibited |

A mock cannot fall through. Async sink acceptance, acknowledgement, retries and duplicate delivery are distinct recorded events linked by causality; expected terminal predicate/quiescence bounds completion. If no safe sandbox or faithful comparison exists, keep UNKNOWN/UNSUPPORTED and disclose scope. Container guarantees require executed control tests; same-account hostile-agent OS isolation is not claimed.
