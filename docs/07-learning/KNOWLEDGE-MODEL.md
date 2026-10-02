# KNOWLEDGE-MODEL.md — Lessons, Patterns, and Hypotheses Schemas

> **Version:** 1.0  
> **Status:** Learning Reference Guide  

---

## 1. Schema Definitions

### 1.1 Project Lesson Schema
```json
{
  "lesson_id": "lsn_01HZX89AB",
  "project_id": "customer-portal",
  "unit_id": "auth-login",
  "mismatch_category": "SESSION_ENCRYPTION_MISMATCH",
  "legacy_behavior": "CI3 stores session cookie encrypted using legacy mcrypt/openssl AES-128-CBC",
  "modern_solution": "Laravel uses modern APP_KEY with AES-256-CBC; injected bridge decryptor for active session migration",
  "success_rate": 1.0,
  "created_at": "2026-10-03T04:55:00Z"
}
```

### 1.2 Global Pattern Schema
```json
{
  "pattern_id": "pat_ci3_session_bridge",
  "title": "CodeIgniter 3 to Laravel Session Interoperability",
  "source_framework": "codeigniter3",
  "target_framework": "laravel13",
  "confidence_score": 0.98,
  "occurrences_count": 7,
  "rule_type": "ADAPTER_TEMPLATE",
  "template_ref": "templates/ci3_laravel_session_bridge.php"
}
```
