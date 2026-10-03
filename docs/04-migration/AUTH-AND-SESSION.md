# Authentication and session compatibility

Authentication/session behavior is a required contract dimension. Detect the actual password scheme, principal/role lookup, cookie name/domain/path/flags/lifetime, session storage/serialization, rotation, flash lifetime, CSRF behavior, remember-me, reset and logout flows. Never assume CI3 uses encrypted cookies or that a Laravel session can decode it.

## Passwords

Preserve verification of existing password hashes through an explicit target compatibility adapter. Record algorithm, salt/pepper references and observed success/failure behavior without persisting passwords. Rehash-on-login is a proposed target rule requiring unchanged external behavior and regression cases. Unknown hash schemes block authentication generation.

## Active session continuity

The default product requirement is transparent continuity for active source sessions. A plan must choose an evidenced strategy: compatible reading of a safe existing session store plus target issuance, or an explicit reviewed customer exception requiring re-login. A source/target parallel rollout also requires declared route ownership, shared state semantics and rollback behavior.

A bridge must validate source session integrity, expiry, principal/permissions, fixation resistance, rotation and logout/revocation. Keys are secret references and data stays in isolated fixture stores during verification. Insecure deserialization or unknown encryption cannot be accepted by inventing a decoder. Record `ACTIVE_SESSION_CONTINUITY_UNPROVEN` and block readiness.

Required cases: valid source session reaches target as same principal, expired/tampered/missing sessions fail, a session cannot switch users, fixation token rotates, flash survives the required number of requests, logout invalidates access on both sides according to the chosen strategy, and rollback restores the declared session behavior. Cookie names may be opaque values but flags/domain/path cannot be blanket masked.

This capability is a release requirement, not automatically solved by adding Laravel middleware. Customer-specific driver formats and secrets are inputs in [PROJECT-INPUTS](../10-development/PROJECT-INPUTS.md).
