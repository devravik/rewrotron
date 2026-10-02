# ADR 0010: Mandatory Engine-Enforced User Confirmation Gates

- **Status:** Accepted
- **Date:** 2026-10-03
- **Deciders:** Rewrotron Core Architecture Team

## Context
When autonomous AI agents operate on legacy repositories, they often make premature assumptions: assuming a framework version without verification, proceeding with full-codebase generation without checking target preferences, or attempting destructive migrations. Relying solely on prompt instructions ("please ask the user") is fragile because LLMs can hallucinate confirmation or skip steps.

## Decision
Enforce three **Hard User Confirmation Gates** directly within the **Rewrotron Core Engine finite state machine**:
1. **Gate 1 (Source Detection):** Detect language, framework, version, runtime, and database. Present to user with confidence scores. Block deep analysis until explicit user confirmation is persisted in `user_confirmations`. If source files change, invalidate the `configuration_hash`.
2. **Gate 2 (Target & Migration Plan):** Propose target stack (language, framework, architecture style) and migration units. Block code generation until explicit user confirmation is persisted.
3. **Gate 3 (Production Cutover):** Require explicit human confirmation before any production deployment or destructive migration.

The engine returns `ERROR: USER_CONFIRMATION_REQUIRED` if any command attempts to bypass an unconfirmed gate. Agents CANNOT infer confirmation from silence or conversational context.

## Alternatives Considered
- **Prompt-Only Guidelines:** Instructing agents in system prompts to ask the user. Rejected because stochastic LLMs frequently bypass prompt guidelines when given complex tasks.
- **Single Confirmation at End:** Allowing full autonomous migration and asking only at the end. Rejected because operating on an incorrect initial source framework or version assumption invalidates all subsequent work.

## Consequences
- **Positive:** Absolute safety; user maintains full sovereignty over project baseline; eliminates catastrophic misunderstandings before expensive compute/token expenditure.
- **Negative:** Requires an interactive confirmation step in CLI or explicit tool call via MCP.
