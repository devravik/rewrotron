# EXPERIMENT-LEDGER.md — Experiment Fingerprinting & Rollback Ledger

> **Version:** 1.0  
> **Status:** Learning Reference Guide  

---

## 1. Fingerprint Calculation

To ensure experiment reproducibility and avoid repeating failed repair attempts, every experiment calculates a cryptographic fingerprint:

$$\text{Fingerprint} = \text{SHA256}(\text{SourceASTHash} + \text{ContractID} + \text{TargetAdapterVersion} + \text{PromptHash} + \text{PreconditionDBHash})$$

If an incoming mission matches a known fingerprint with a recorded failure, the agent is forced to select an alternative strategy or escalate to human review.

---

## 2. Rollback Ledger

Every code modification executed during a migration or repair attempt is backed by a transactional Git commit or filesystem stash referenced in the `repair_attempts` table. If verification fails, Rewrotron restores the exact prior state in < 50ms.
