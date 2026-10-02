# GOLDEN-MASTER.md — Golden Master Generation & Baseline Locking

> **Version:** 1.0  
> **Status:** Verification Reference Guide  

---

## 1. Golden Master Lifecycle

To prevent target developers from aiming at a moving target, Rewrotron generates a locked **Golden Master** for every behavior case:

1. **Multiple Replay Capture:** The legacy application executes the scenario multiple times under identical database seeds.
2. **Determinism Verification:** Response hashes and database mutations across runs are compared.
3. **Variance Extraction:** Any field that changes between legacy runs (e.g. timestamps, random tokens) is flagged as dynamic and mapped to an extraction pattern.
4. **Baseline Lock:** The final normalized legacy response and mutation set is cryptographically hashed and saved in `.rewrotron/contracts/golden_master_{case_id}.json`.

Once locked, the Golden Master cannot be mutated without a recorded engineering justification.
