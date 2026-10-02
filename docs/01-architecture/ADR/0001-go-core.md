# ADR 0001: Go as Primary Implementation Language for Rewrotron Core

- **Status:** Accepted
- **Date:** 2026-10-03
- **Deciders:** Rewrotron Core Architecture Team

## Context
Rewrotron must operate as a fast, cross-platform, standalone command-line tool, background daemon, and MCP server. It must perform heavy static AST parsing, fast concurrent differential HTTP replays, process management, and local database access without imposing complex runtime dependencies (like Node.js or a heavy Python environment) on the user's workstation.

## Decision
Implement the entire Rewrotron core engine, CLI, analyzer, observer harnesses, verification runner, and MCP server in **Go** (stable modern baseline, Go 1.22+).

## Alternatives Considered
- **Rust:** Excellent performance and safety, but slower compile times and higher barrier to entry for ecosystem contributors.
- **Python:** Rich ecosystem, but slow execution speeds for AST parsing/diffing, distribution headaches with virtualenvs, and high memory footprint.
- **TypeScript / Node.js:** Convenient for MCP, but single-threaded concurrency limitations for high-throughput HTTP replay and heavy runtime footprint.

## Consequences
- **Positive:** Single, statically-linked binary distribution; exceptional native concurrency for differential testing; low memory footprint; straightforward cross-compilation.
- **Negative:** AST parsing for PHP must either use robust Go bindings (e.g. tree-sitter or custom lexer) or invoke a small companion PHP CLI script for edge-case reflection.
