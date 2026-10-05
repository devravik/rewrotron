# Authentication and session compatibility

Authentication/session behavior is a required contract dimension. Detect the actual password scheme, principal/role lookup, cookie name/domain/path/flags/lifetime, session storage/serialization, rotation, flash lifetime, CSRF behavior, remember-me, reset and logout flows. Never assume CI3 uses encrypted cookies or that a Laravel session can decode it.

## Passwords

Preserve verification of existing password hashes through an explicit target compatibility adapter. Record algorithm, salt/pepper references and observed success/failure behavior without persisting passwords. Rehash-on-login is a proposed target rule requiring unchanged external behavior and regression cases. Unknown hash schemes block authentication generation.

## Active session continuity

The default product requirement is transparent continuity for active source sessions. A plan must choose an evidenced strategy: compatible reading of a safe existing session store plus target issuance, or an explicit reviewed customer exception requiring re-login. A source/target parallel rollout also requires declared route ownership, shared state semantics and rollback behavior.

A bridge must validate source session integrity, expiry, principal/permissions, fixation resistance, rotation and logout/revocation. Keys are secret references and data stays in isolated fixture stores during verification. Insecure deserialization or unknown encryption cannot be accepted by inventing a decoder. Record `ACTIVE_SESSION_CONTINUITY_UNPROVEN` and block readiness.

Required cases: valid source session reaches target as same principal, expired/tampered/missing sessions fail, a session cannot switch users, fixation token rotates, flash survives the required number of requests, logout invalidates access on both sides according to the chosen strategy, and rollback restores the declared session behavior. Cookie names may be opaque values but flags/domain/path cannot be blanket masked.

This capability is a release requirement, not automatically solved by adding Laravel middleware. Customer-specific driver formats and secrets are inputs in [PROJECT-INPUTS](../10-development/PROJECT-INPUTS.md).

## Fixture identity versus authenticated identity

principal-fixture 1.0 names intended actor/subject/credential ref and approved initial session/DB relation; it contains no authentication verdict. Actual principal is a separate trusted capture in evaluation-context.actor.actual_principal, including side-local subject/session identity, authentication status, roles and evidence channel. A swapped jar must establish B rather than inheriting A's fixture label. Unknown principal cannot satisfy an authorization predicate. Response/session lowering is defined in BEHAVIOR-MODEL and response.schema, including separate repeated cookies, attributes, rotation, invalidation and replay of old tokens. Hashing/codec capability must be qualified and pinned; unsupported credential/session format blocks that scope.

## Bounded lifecycle acceptance contract (M4 Unit 4)

[session-continuity.schema.json](../../specs/verification/session-continuity.schema.json) records the executable projection of the approved X19 obligations. Its content address is sha256 of rwt-cjson-1 of the complete record; the address is external to its preimage. It is consumed through the existing plan unit contract_refs and immutable decision_refs, not a new confirmation or permission mechanism. Exact project/case/workflow, plan hash and contract bytes must match. The guarded caller supplies the current approved plan and engine-admitted qualifications; agent declarations cannot populate those trust inputs. Editing predicates, selectors, strategy or provider pins requires a successor contract/plan and affected verification.

Checks evaluate the existing bounded boolean expression vocabulary against exact SOURCE or TARGET PRE_OPERATION/POST_OPERATION occurrence captures. Actor fixture refs select requests; separate actual principal capture proves authentication. Required issuance, integrity/expiry, login/logout/revocation, old-token replay, principal/permission/session binding, CSRF acceptance/binding/regeneration, flash lifetime, reset/remember-me/throttle, concurrent sessions and declared rollback cases retain the paths and conditions established by qualified source evidence. No universal session-store field layout or token policy is inferred. Missing contexts, operands or principal capture are UNKNOWN, including attempts to use is_missing/exists as security acceptance. False established obligations are MISMATCH and remain so when another obligation is incomplete.

Relations compare exact typed captured paths across requests, phases or sides. Same-side equality/inequality establishes continuity and regeneration without replacing observed identifiers. Cross-side opaque scalar identities may use only an exact approved TOKEN_LIFECYCLE/BIJECTIVE_ID rule pinned to this contract/decision and actor/relation-group/connection bijection. This permission never applies to composite session state, principal binding assertions, expiry boundaries, replay acceptance or cookie scope/security/lifetime. Separate obligations remain mandatory; a matching opaque token is not proof of its lifecycle. Cookie occurrences retain ordering/repetition, name, flags, domain/path and lifetime; null attributes remain absent.

SAFE_BRIDGE requires a pinned codec provider and an engine-admitted SUPPORTED qualification covering the exact implementation, capability, profile and controls with no uncovered limitations. Missing/unsafe/unknown codecs produce ACTIVE_SESSION_CONTINUITY_UNPROVEN and block readiness; acceptance never decodes a session or writes either application store. REVIEWED_RELOGIN uses an explicit plan decision and still verifies its observed obligations, reporting that active continuity is excluded; it cannot silently claim transparent continuity. Unit 5 consumes these semantics for target lowering. Go X19 tests prove the bounded deterministic acceptance behavior with synthetic captures; they do not qualify a real legacy codec, runtime bridge or release rollback.
