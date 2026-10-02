# DATABASE-COMPARISON.md — Relational State & Mutation Verification

> **Version:** 1.0  
> **Status:** Verification Reference Guide  

---

## 1. Comparing State Mutations

Comparing HTTP responses alone is insufficient; an endpoint might return HTTP 200 OK while failing to insert a row into an audit table or accidentally updating the wrong user record.

### Transaction Diffing Process
1. **Isolated Sandboxes:** Source and Target run against separate, identically-seeded database instances (or container schemas).
2. **Pre-Execution Snapshot:** Checksums of affected tables are recorded prior to request execution.
3. **Execution & Interception:** SQL queries executed during the request are intercepted and logged.
4. **Post-Execution State Comparison:**
   - Evaluates row count diffs for each table.
   - Evaluates exact column-value diffs for newly inserted or updated rows.
   - Ignores auto-increment ID variance by checking foreign key referential integrity rather than absolute integer values.
