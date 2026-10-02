# CODE-CLASSIFICATION.md — Detailed Classification Matrix & Proof Rules

> **Version:** 1.0  
> **Status:** Standard Reference Specification  

---

This document establishes the exact algorithmic rules for classifying legacy code artifacts.

## 1. Classification Status Matrix

| Classification | Reachable Statically? | Observed Dynamically? | Dynamic Dispatch Risk? | Action Taken by Rewrotron |
| :--- | :--- | :--- | :--- | :--- |
| **`ACTIVE`** | Yes or Unknown | Yes (>= 1 hit) | Any | **Reconstruct in Target** |
| **`DORMANT`** | Yes (>= 1 path) | No (0 hits) | Low / None | **Reconstruct or Human Review Gate** |
| **`DEAD`** | No (0 paths) | No (0 hits) | None (Proven) | **Omit from Target + Emit Proof** |
| **`UNKNOWN`** | Indeterminate | No (0 hits) | High (`eval`, variable methods) | **Must Preserve or Require Gate** |

---

## 2. Generating Cryptographic Proof of Dead Code

When a symbol, method, or class is classified as `DEAD`, Rewrotron generates an unforgeable proof bundle:

```json
{
  "symbol": "LegacyExportHelper::exportToDbase3",
  "file": "application/helpers/export_helper.php",
  "line": 142,
  "file_hash": "a1b2c3d4e5f6...7890",
  "classification": "DEAD",
  "proof": {
    "reachability_scan": {
      "inbound_static_calls": 0,
      "inbound_view_references": 0,
      "routes_mapped": 0
    },
    "dynamic_scan": {
      "variable_function_risk": false,
      "eval_usage_detected": false
    },
    "runtime_traces_checked": 1482,
    "runtime_hits": 0
  },
  "signed_timestamp": "2026-10-03T04:45:00Z"
}
```

This proof guarantees that developers and compliance auditors have clear empirical evidence of why legacy baggage was omitted.
