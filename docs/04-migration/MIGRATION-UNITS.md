# Migration units and dependency planning

A unit is an externally verifiable route, operation group or bounded workflow. Individual helper files are supporting implementation, not canonical units. Include authentication, authorization, validation, state/effect changes, output and compatibility cases in the unit's scope.

Unit records identify contracts/cases, supported capabilities, prerequisites, shared components, generation ownership and evidence. Shared tables do not automatically imply order; a unit depends on another when it needs that unit's generated behavior or shared implementation. Use read/write/effect conflicts to decide concurrent work safety.

Construct dependency edges, compute strongly connected components, and propose merging mutually dependent units into one verifiable slice. If that slice is too broad, surface a redesign decision before Gate 2. Never break a cycle by arbitrarily deleting an edge. Condense to a DAG; stable topological ordering uses priority then unit ID for ties.

Default execution is serial. A dependent unit cannot generate until prerequisite units are current VERIFIED/COMPLETED. Independent units can proceed after another unit blocks, but project completion still requires the entire approved scope. Target file overlap requires a shared component owner or serial publication; mission fencing and before hashes protect edits.

A plan includes omission proofs, dormant/unknown disposition, scope exclusions, unresolved questions and a requirement-to-unit/case mapping. It cannot silently drop an unsupported entrypoint from the denominator.
