# Architecture overview

[SYSTEM-DESIGN](SYSTEM-DESIGN.md) owns the canonical pipeline: sealed source/profiles → Source IR → Application IR → Behavior IR/evidence → executable contracts/scenarios → requested target → exact feasibility/resolution → target model/plan → Gate 2 → staged implementation → isolated scoped verification/repair → performance → release package/Gate 3.

Core defines neutral ports; adapter/storage implementations depend on them; cmd composes concrete services. Three IRs are not a universal guarantee: unknown/unsupported mechanisms retain provenance and explicit capability gaps. User intent and exact resolution differ; target ownership/lowering is inspectable before generation. CLI/MCP/Skills call identical guarded services. Background work requires explicit supervisor ownership; current state plus event window does not imply historical projections. See canonical owners in [docs index](../README.md).
