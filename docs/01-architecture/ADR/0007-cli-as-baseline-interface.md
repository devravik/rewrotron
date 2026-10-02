# ADR 0007: CLI as the Baseline Machine Interface

- **Status:** Accepted
- **Date:** 2026-10-03
- **Deciders:** Rewrotron Core Architecture Team

## Context
While MCP provides rich structured tooling for advanced agent clients, many environments (CI/CD pipelines, bash scripts, terminal-based AI tools like Aider) operate purely through command-line invocations. Rewrotron must be fully operable without an MCP client.

## Decision
Design the CLI (`cmd/rewrotron`) as the primary, fully capable baseline interface. Every major subsystem operation (`init`, `doctor`, `analyze`, `observe`, `discover`, `plan`, `migrate`, `verify`, `status`, `next`) must be executable via the CLI. When invoked with `--json`, commands MUST emit pure structured JSON to `stdout` and all logs/diagnostics to `stderr`.

## Alternatives Considered
- **MCP-Only Architecture:** Building Rewrotron purely as an MCP background daemon. Rejected because it precludes running in traditional CI/CD runners and simple shell scripts.
- **REST/gRPC Daemon Only:** Requiring an HTTP daemon to be running. Rejected because it introduces unnecessary process management overhead.

## Consequences
- **Positive:** Maximum scriptability; works seamlessly in headless GitHub Actions runners; allows terminal agents (Aider, Claude Code CLI) to drive the migration directly.
- **Negative:** CLI output payloads must be maintained and verified against JSON schemas alongside MCP tool schemas.
