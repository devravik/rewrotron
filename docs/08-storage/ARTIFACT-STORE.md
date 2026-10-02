# ARTIFACT-STORE.md — Filesystem Storage Hierarchy & Redaction

> **Version:** 1.0  
> **Status:** Storage Reference Guide  

---

## 1. Filesystem Layout Under `.rewrotron/artifacts/`

```text
.rewrotron/artifacts/
├── traces/                   # Raw HTTP request/response payloads (>64KB)
│   ├── trc_01HZX..._req.bin
│   └── trc_01HZX..._resp.bin
├── reports/                  # Generated HTML/JSON mismatch and audit reports
│   ├── mismatch_order_001.json
│   └── omissions.json
├── patches/                  # Unified diffs generated during repair attempts
│   └── patch_order_attempt_1.diff
├── screenshots/              # Browser DOM snapshots and visual diffs
│   ├── legacy_checkout.png
│   └── modern_checkout.png
└── generated/                # Intermediate code templates before commit
```

---

## 2. Invariants & Retention
- **Content-Addressable Hashing:** Every artifact file is named or validated with its SHA-256 hash.
- **Redaction Prior to Disk:** No raw trace containing unmasked authentication tokens or secrets may ever touch the filesystem.
- **Retention Policies:** Traces older than 30 days or belonging to passed verification cases can be garbage-collected via `rewrotron clean`.
