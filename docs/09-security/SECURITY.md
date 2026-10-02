# SECURITY.md — Safe Execution, Isolation, and Secret Protection

> **Authority Level:** Mandatory Security Specification  
> **Status:** Authoritative Core Specification  
> **Target Subsystems:** All Subsystems, Sandbox Runners, CLI & MCP  

---

## 1. Threat Model & Principles

Rewrotron executes, instruments, and analyzes arbitrary, unvetted legacy applications that may contain:
- Unpatched remote code execution vulnerabilities (e.g. legacy SQL injection, insecure deserialization).
- Hardcoded production credentials and API keys in source files.
- Destructive outbound integrations (e.g. real payment gateway calls, production email dispatches).

Rewrotron adheres to three non-negotiable security mandates:
1. **Zero Production Mutation:** Rewrotron is strictly forbidden from connecting directly to production databases, payment endpoints, or messaging queues.
2. **Local-First Secret Isolation:** No source code, passwords, database dumps, or traces are ever uploaded to cloud AI providers or external analytics servers without explicit configuration.
3. **Sandbox Execution Isolation:** Legacy and modern target applications execute in containerized or isolated loopback environments with mock external services.

---

## 2. Sandbox Execution Policies

When executing legacy or target code during runtime observation or differential verification:
- **Network Egress Block:** Outbound HTTP/TCP traffic to the public internet is disabled by default via sandbox proxy or iptables/Docker network restrictions.
- **External Side-Effect Mocking:** Outbound calls to SMTP, Twilio, Stripe, PayPal, or third-party webhooks are trapped by Rewrotron's local mock server (`127.0.0.1:8999`).
- **Filesystem Write Containment:** Application writes outside the project root or sandbox temporary directory are trapped and prevented.

---

## 3. Secret Handling & Redaction Invariants

1. **Detection:** All strings matching known secret patterns (AWS keys, Stripe keys, database passwords, private keys, JWTs) are detected during file parsing and trace recording.
2. **Redaction Prior to Storage:** Before writing any trace, log, or artifact to disk, sensitive keys are replaced with `[REDACTED_BY_REWROTRON]`.
3. **Environment Injection:** Target applications use modern `.env.example` configurations. Secrets are never hardcoded in generated PHP/Laravel classes.

---

## 4. Human Review Gates for High-Risk Actions

The following actions **unconditionally require manual human confirmation** and cannot be bypassed autonomously by an AI agent:
1. Disabling the local mock server to permit real external network egress.
2. Deleting any code classified as `DORMANT` or `UNKNOWN`.
3. Accepting an unclassified `UNKNOWN` verification failure as an intentional design change.
4. Executing live production database migrations or final traffic cutover.
