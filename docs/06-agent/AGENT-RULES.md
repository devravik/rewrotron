# AGENT-RULES.md — Operative Invariants for LLMs Working on Rewrotron

> **Version:** 1.0  
> **Status:** Operative Invariants  

---

These rules represent hard constraints that every LLM session must preserve:

1. **NEVER assume dead code without proof:** A piece of code is only dead if static reachability is zero AND runtime observation is zero.
2. **NEVER modify `specs/` or `docs/` schemas without code migration:** Schema changes must be backward compatible.
3. **NEVER introduce LLM prompt calls into Go internal packages:** `internal/` must remain 100% deterministic Go.
4. **ALWAYS use `--json` for automated script parsing:** And never log debug strings to `stdout` when `--json` is active.
5. **ALWAYS check database diffs after differential replays:** An HTTP 200 is not a pass if side-effect mutations diverged.
