# ADR 0005: Automated Differential Verification with Normalization

- **Status:** Accepted
- **Date:** 2026-10-03
- **Deciders:** Rewrotron Core Architecture Team

## Context
Proving that a modernized target application behaves identically to the legacy source application requires empirical evidence. Unit tests alone are insufficient because developers or agents can write tests that pass against buggy code while failing to capture legacy edge cases.

## Decision
Implement automated **Differential Verification**: replay identical HTTP payloads and preconditions against both the Source Application and Target Application, capturing responses, database mutations, and session changes. Validate equivalence using a 3-tier policy:
1. **Byte-Exact:** For deterministic status codes, redirect URLs, and static responses.
2. **Canonical Semantic:** For JSON/HTML payloads where formatting or key order varies innocently.
3. **Normalized Dynamic:** For timestamps, nonces, CSRF tokens, and generated UUIDs.

Any unresolvable discrepancy is classified as `UNKNOWN` or `MISMATCH`.

## Alternatives Considered
- **Target Unit Tests Only:** Writing tests only in Laravel. Rejected because it provides zero guarantee that legacy business logic was accurately ported.
- **Manual QA Testing:** Too slow, expensive, and non-reproducible.

## Consequences
- **Positive:** Reproducible compatibility evidence for an explicit finite corpus and scope; catches silent side-effect bugs and missed database updates early.
- **Negative:** Requires running both stacks simultaneously in isolated sandboxes and resetting test database state between runs.

## Implementation clarification (2026-10-03)

[ADR 0011](0011-development-baseline-and-contract-completion.md) defines the initial concrete package, hash, protocol, evidence and release baseline. This accepted design is not a claim that implementation/runtime validation has completed.
