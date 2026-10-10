# Security Policy

**Owner:** Security vulnerability reporting & mediated trust boundaries  
**Canonical Specification:** [docs/09-security/SECURITY.md](docs/09-security/SECURITY.md)  
**Status:** Authoritative Vulnerability Reporting Guide  

---

## Supported Versions

Security fixes and patches are applied to the active development baseline:

| Version | Supported | Notes |
| :--- | :--- | :--- |
| `v1.0.x` (current `main`) | :white_check_mark: | Primary development and qualified baseline |
| `< 1.0.0` | :x: | Internal pre-release milestones |

---

## Reporting a Vulnerability

We take the security of Rewrotron, its sandbox isolation guarantees, and its secret-redaction mechanisms seriously.

If you believe you have found a security vulnerability—especially concerning:
- **Sandbox escape or path traversal:** Writing outside workspace boundaries (`../`, symlink swaps, null bytes).
- **Secret or credential leakage:** Raw passwords, bearer tokens, or private keys appearing in unscrubbed logs or public reports.
- **Unauthorized production effects:** Bypassing `ERR_PRODUCTION_EFFECT_REFUSED` or unapproved live external mutations.
- **Gate authentication bypass:** Impersonation of human authorization tokens (Gate 1, Gate 2, Gate 3).

Please **DO NOT** file a public GitHub issue.

### Private Reporting Channels

1. **GitHub Private Vulnerability Reporting:**  
   Use GitHub's built-in "Report a vulnerability" button on the repository's Security tab (when enabled).
2. **Email Security Contact:**  
   Send an encrypted or private report to: **`security@rewrotron.dev`**

Include:
- Type of vulnerability and affected component (`internal/domain`, sandbox, CLI, MCP).
- Exact step-by-step reproduction instructions or code snippet.
- Impact assessment (e.g., credential disclosure, filesystem containment breach).

---

## Disclosure Policy & Trust Boundaries

Rewrotron distinguishes three mediated trust boundaries (detailed in [docs/09-security/SECURITY.md](docs/09-security/SECURITY.md)):
1. **Mediated Engine Enforcement:** All operations via CLI, MCP, and Skills validate schemas, paths, gates, and runner policies.
2. **Trusted Same-Account Local Operation:** Users or agents with existing filesystem/shell access operate within their local account privileges. Rewrotron does not guarantee isolation from hostile OS users sharing the same local account.
3. **Hostile Code Execution:** Runtime source and target application code must always be executed in isolated disposable containers with production access denied.

Vulnerability reports will be acknowledged within 48 hours, and patches will be developed and verified against test suites before public release.
