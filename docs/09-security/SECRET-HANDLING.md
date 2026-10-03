# Secret handling and replay-safe evidence

## Credential sources

Project config stores secret references such as `env:RWT_SANDBOX_DB_PASSWORD`, never password values. Resolve only an explicit allowlist for the current environment; no secret is echoed in doctor errors, commands, SQL diagnostics, event payloads, snapshots or agent context. Synthetic test credentials are resolved at execution and kept in memory.

## Redaction before persistence

Scrub headers, cookies, URLs/query/form/JSON bodies, SQL parameters/literals, sessions, external payloads and subprocess stderr before writing artifacts or SQLite. Redaction combines structured field rules, configured application-specific paths and secret-value matching. Regex alone cannot guarantee every secret is detected. Malformed or unknown body formats are withheld or safely replaced; absence of evaluable evidence becomes UNKNOWN.

Sensitive opaque values use a project-scoped keyed HMAC for correlation when needed; plain hashing of a low-entropy password is not safe. Persist a redaction manifest indicating removed fields and lost assertion coverage. Do not persist raw secret-containing byte hashes as a correlation substitute. The local correlation key is stored separately with owner-only permissions and never included in exported artifacts.

Streaming captures must redact before staging on disk. If a format cannot be redacted incrementally, buffer within the configured size limit or decline persistence. Screenshots, DOM and binary attachments can contain PII even without credentials; restrict them by capture policy and review exports.

## Verification versus redaction

Replay cases reference synthetic principal/secret IDs and explicit token extraction bindings. They never use a literal `[REDACTED]` value as an authentication token. Compare sensitive required values in memory, preserving only permitted relation/assertion outcomes. If a discarded value is required to prove compatibility and cannot be safely compared, return UNKNOWN.

Workspace and artifact directories use owner-only permissions where the OS supports them. Supply `.gitignore` rules for `.rewrotron`, dumps, caches and secret files in generated repositories. Local storage is not automatically encrypted at rest; customer disk encryption and access control remain deployment inputs. Retention/export behavior follows [ARTIFACT-STORE](../08-storage/ARTIFACT-STORE.md).

## Execution explanations and developer views

Redact event titles, summaries, reasons, structured details, errors and evidence links before persistence, not only in the UI. Developer verbosity does not reveal credentials, raw sensitive captures, internal prompts or hidden reasoning. Escape source-controlled text as untrusted display data; strings cannot become executable commands or instructions. Hash/cursor/entity refs must preserve project ownership. Approvals never expose tokens in activity timelines.
