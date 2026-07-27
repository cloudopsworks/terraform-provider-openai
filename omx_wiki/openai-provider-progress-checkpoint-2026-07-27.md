---
title: "OpenAI Provider Progress Checkpoint 2026-07-27"
tags: ["terraform-provider", "openai", "admin-api", "service-accounts", "api-keys", "documentation", "lessons-learned"]
created: 2026-07-27T17:44:39.545Z
updated: 2026-07-27T17:44:39.545Z
sources: []
links: ["openai-provider-administration-round-1.md"]
category: session-log
confidence: medium
schemaVersion: 1
---

# OpenAI Provider Progress Checkpoint 2026-07-27

## Progress since [[openai-provider-administration-round-1]]

- Added admin API key hardening in prior commits: destroy/replacement now requires OpenAI delete confirmation, and expiration can be declared as `expires_in_seconds`, `expire_in_hours`, or `expire_in_days`.
- Updated OpenAI client initialization so OpenAI SDK debug logging is disabled with `WithDebugLog(nil)` unless Terraform logging is trace-level (`TF_LOG >= TRACE`).
- Removed unsupported admin-key `scopes` after confirming that API key scopes belong to service-account project API key creation, not admin API key creation.
- Generalized service-account bootstrap key handling with standalone `openai_project_api_key` creation: unscoped service-account creation captures OpenAI's default returned key; scoped service-account creation uses the create-only service-account path followed by the shared service-account API-key creation path.
- Updated provider docs, examples, `README.yaml`, generated `README.md`, and `docs/index.md` to describe the default-key and scoped-key workflows accurately.
- Kept the current admin API key example change small and sanitized: `expire_in_hours = 10` demonstrates the new hour-based duration input without adding credentials or state.

## Validation evidence from the latest completed implementation pass

- `go test ./internal/client ./internal/provider`
- `go test ./...`
- `go vet ./...`
- `go build ./...`
- `git diff --check`
- staged content safety scan before commit `7bafaaa feat(provider): support service account bootstrap keys`

## Lessons learned

1. Verify OpenAI API shape before schema design. The service-account create endpoint returns a default API key by default, while `scopes` are accepted on the service-account API-key create endpoint.
2. Prefer one shared API-key creation/mapping path for both embedded service-account bootstrap keys and standalone `openai_project_api_key` resources. This reduces drift in secret handling, scope handling, and tests.
3. Preserve create-only secrets across refresh and update. OpenAI does not return unredacted key values after create, so Terraform state must retain them as Sensitive computed attributes.
4. Delete known child keys before deleting parent service accounts. This makes destroy behavior explicit and avoids relying on implicit cascading behavior.
5. Keep README maintenance YAML-first: update `README.yaml`, run `make readme`, then sync `docs/index.md` and trim generator whitespace before diff checks.
6. Treat local Terraform examples and state as high-risk. Stage only sanitized example files, never local CLI config or Terraform state artifacts.
7. Keep examples type-correct. `expire_in_hours` and `expire_in_days` are numeric Terraform attributes, so examples should use numbers, not quoted strings.
8. The next requested admin APIs (data retention, invites, spend limits/alerts, certificates, and project sub-API evaluation) should start with a read-only discovery pass over official OpenAI docs and the current Go SDK before adding resources.

## Next work candidates

- Evaluate organization/project data-retention APIs and whether they share request/response models.
- Evaluate invite lifecycle APIs and data-source lookup patterns.
- Evaluate spend limit and spend alert APIs for reusable threshold/notification models.
- Evaluate certificate APIs at organization and project scope for shared certificate mapping/import/delete behavior.
- Revisit provider architecture after discovery for reusable CRUD/list scaffolding across nested admin sub-APIs.
