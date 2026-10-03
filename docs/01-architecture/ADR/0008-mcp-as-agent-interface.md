# ADR 0008: Model Context Protocol (MCP) as Primary Agent Interface

- **Status:** Accepted
- **Date:** 2026-10-03
- **Deciders:** Rewrotron Core Architecture Team

## Context
AI agents need to discover project state, request the next migration mission, inspect execution traces, run verification cases, and record diagnostic lessons without invoking clunky bash wrappers or scraping terminal text. The Model Context Protocol (MCP) is the emerging open standard for agent-to-tool connectivity.

## Decision
Provide a first-class, dedicated MCP server binary (`cmd/rewrotron-mcp`) compliant with the pinned MCP revision 2025-11-25 (JSON-RPC over stdio initially; future Streamable HTTP is a separate transport decision). The MCP server exposes high-level domain operations (`project.get_snapshot`, `research.get_next_mission`, `behavior.get_case`, `verification.run_case`, `knowledge.record_lesson`) rather than low-level database queries or unrestricted bash execution.

## Alternatives Considered
- **Custom JSON-RPC Protocol:** Inventing an internal socket protocol. Rejected because modern agent clients already natively speak MCP.
- **Exposing Raw SQL or Shell Execution:** Providing generic `execute_sql` or `exec_bash` tools. Strongly rejected due to severe safety risks and lack of domain semantics.

## Consequences
- **Positive:** Zero-friction integration with Claude Code, Cursor, Antigravity, and other MCP-native clients; structured JSON validation for all tool calls; strict capability security boundaries.
- **Negative:** Requires running `rewrotron-mcp` as a subprocess or daemon during agent sessions.

## Implementation clarification (2026-10-03)

[ADR 0011](0011-development-baseline-and-contract-completion.md) defines the initial concrete package, hash, protocol, evidence and release baseline. This accepted design is not a claim that implementation/runtime validation has completed.
