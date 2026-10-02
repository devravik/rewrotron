# PERMISSION-MODEL.md — Agent Tool Permissions & Review Gates

> **Version:** 1.0  
> **Status:** Security Reference Guide  

---

## 1. Tool Permission Matrix

| Interface / Tool | Autonomous Agent Allowed? | Human Gate Required? | Risk Tier |
| :--- | :--- | :--- | :--- |
| `project.get_snapshot` | Yes | No | Low (Read-Only) |
| `research.get_next_mission`| Yes | No | Low (Read-Only) |
| `behavior.get_case` | Yes | No | Low (Read-Only) |
| `verification.run_case` | Yes | No | Medium (Sandbox Replay) |
| `migration.generate_code` | Yes | No | Medium (Target Workspace) |
| `code.delete_dead_symbol` | Yes (If proven DEAD) | No | Medium (Verified Safe) |
| `code.delete_dormant_symbol`| **NO** | **YES** | **High** (Potential Regress) |
| `config.enable_production`| **NO** | **YES** | **Critical** (Destructive) |
| `database.live_cutover` | **NO** | **YES** | **Critical** (Production Cutover) |
