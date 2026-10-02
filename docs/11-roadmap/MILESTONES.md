# MILESTONES.md — Checkpoints & Acceptance Gates

> **Version:** 1.0  
> **Status:** Project Milestones Reference  

---

## 1. Major Release Milestones

- **M0: Specification Lock & Schema Verification (Current Milestone)**  
  *Deliverables:* All 18 mandatory specifications approved; JSON schemas validated; SQLite schema frozen.  
  *Acceptance Gate:* Zero ambiguous contracts.

- **M1: CI3 Static Analyzer Complete**  
  *Deliverables:* AST parsing capable of indexing 100% of symbols in reference CI3 test suite. Reachability graph and code classifier active.

- **M2: Observer & Trace Replay Complete**  
  *Deliverables:* Runtime hooks intercepting requests, queries, and sessions; deterministic replay harness active.

- **M3: End-to-End Reference Migration (First Vertical Slice)**  
  *Deliverables:* Successful reconstruction and differential verification pass of CI3 authentication & user profile slice to Laravel 13.

- **M4: Public Open-Source 1.0 Release**  
  *Deliverables:* Standalone CLI binary, production MCP server, full CI3->Laravel migration capability, complete documentation.
