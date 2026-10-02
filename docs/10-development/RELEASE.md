# RELEASE.md — Semantic Versioning & Binary Distribution

> **Version:** 1.0  
> **Status:** Release Process Reference  

---

## 1. Versioning Protocol

Rewrotron strictly adheres to Semantic Versioning 2.0.0 (`MAJOR.MINOR.PATCH`):
- **MAJOR:** Breaking changes to CLI machine output schema, MCP tool contracts, or SQLite schema requiring manual intervention.
- **MINOR:** New framework adapters, new commands, non-breaking contract additions.
- **PATCH:** Bug fixes, performance improvements, documentation updates.

---

## 2. Release Artifacts

For each release tag:
- Statically linked standalone binaries compiled via Go for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, and `windows/amd64`.
- SHA-256 checksum file signed with release keys.
- Standalone Agent Skill bundle (`rewrotron-skill-vX.Y.Z.tar.gz`).
