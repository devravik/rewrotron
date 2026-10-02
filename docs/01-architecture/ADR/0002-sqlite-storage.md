# ADR 0002: SQLite as Local Relational Persistence

- **Status:** Accepted
- **Date:** 2026-10-03
- **Deciders:** Rewrotron Core Architecture Team

## Context
Rewrotron tracks deep relational data: projects, symbols, call graphs, routes, workflows, behavior cases, contracts, differential verification runs, mismatches, and lessons. The storage layer must be local-first, zero-configuration, robust under crashes, and easily inspectable by human engineers and agent tooling.

## Decision
Use **SQLite** with Write-Ahead Logging (WAL) mode as the canonical local operational database (`.rewrotron/project.db` for project state; `~/.rewrotron/global.db` for global learning). Large binary payloads, full request/response bodies, and raw traces reside in the filesystem artifact store (`.rewrotron/artifacts/`) with relative path references stored in SQLite.

## Alternatives Considered
- **PostgreSQL / MySQL:** Requires users to spin up external daemon containers, credentials, and network setup.
- **Embedded Key-Value Store (Badger / BoltDB):** Lacks expressive relational querying, JOINs, and foreign key integrity needed for complex behavior graphs.
- **Pure JSON / YAML files:** Inefficient for querying tens of thousands of symbol/trace relationships; vulnerable to partial write corruption.

## Consequences
- **Positive:** Zero configuration; embedded directly into the Go binary; full relational SQL and ACID guarantees; easily backed up or inspected via standard `sqlite3` CLI.
- **Negative:** Limited write concurrency (mitigated via WAL mode and single-writer coordinator).
