# ADR 0006: Vendor-Neutral Agent Architecture

- **Status:** Accepted
- **Date:** 2026-10-03
- **Deciders:** Rewrotron Core Architecture Team

## Context
AI coding agents are evolving rapidly across vendors (Claude Code, Cursor, Codex, Antigravity, Aider, Hermes). Hardcoding Rewrotron to any single vendor's API or SDK would lock users in and render Rewrotron obsolete as agent paradigms evolve.

## Decision
Design Rewrotron's core engine, state machines, and analysis tools to be completely vendor-neutral. Rewrotron interacts with agents strictly through open, standardized interfaces:
1. **Model Context Protocol (MCP)** server emitting standardized tools and resources.
2. **Machine-readable JSON CLI** (`--json`).
3. **Agent Skill Packages** (markdown guidelines and workflow scripts).

## Alternatives Considered
- **Direct LLM API Integrations:** Embedding OpenAI or Anthropic SDKs directly into Rewrotron core. Rejected because it introduces API keys into core, creates billing complexity, and restricts agent autonomy.
- **Proprietary Agent Plugin Only:** Limiting to a single IDE extension. Rejected to maximize broad ecosystem adoption.

## Consequences
- **Positive:** Users can bring any LLM or agent of their choice; core remains lightweight, deterministic, and easily testable without mock LLM servers.
- **Negative:** Requires maintaining multiple interface surfaces (CLI JSON + MCP server) in lockstep.
