# Project Specs

Canonical path: `.decapod/generated/specs/`.
These files are the project-local contract for humans and agents.

## Snapshot
- Project: go-scaling
- Outcome: A learning workspace where each module is created as isolated, validated, proof-backed work that turns natural-language intent into runnable Go files and documented scaling knowledge.
- Detected languages: Go
- Detected surfaces: go

## How to use this folder
- [INTENT.md](./INTENT.md): what success means and what is explicitly out of scope.
- [ARCHITECTURE.md](./ARCHITECTURE.md): topology, runtime model, data boundaries, and ADR trail.
- [INTERFACES.md](./INTERFACES.md): API/CLI/events/storage contracts and failure behavior.
- [VALIDATION.md](./VALIDATION.md): proof commands, quality gates, and evidence artifacts.
- [SEMANTICS.md](./SEMANTICS.md): state machines, invariants, replay rules, and idempotency.
- [OPERATIONS.md](./OPERATIONS.md): SLOs, monitoring, incident response, and rollout strategy.
- [SECURITY.md](./SECURITY.md): threat model, trust boundaries, auth/authz, and supply-chain posture.

## Canonical `.decapod/` Layout
- `.decapod/data/`: canonical control-plane state (SQLite + ledgers).
- `.decapod/generated/specs/`: **Living project specs** for humans and agents.
- `.decapod/generated/context/`: deterministic context capsules.
- `.decapod/generated/policy/context_capsule_policy.json`: repo-native JIT context policy contract.
- `.decapod/generated/artifacts/provenance/`: promotion manifests and convergence checklist.
- `.decapod/generated/artifacts/custody/`: epistemic custody artifacts (assumptions, contradictions, deferred questions).
- `.decapod/generated/artifacts/inventory/`: deterministic release inventory.
- `.decapod/generated/artifacts/diagnostics/`: opt-in diagnostics artifacts.
- `.decapod/workspaces/`: isolated todo-scoped git worktrees.

## Day-0 Onboarding Checklist
- [ ] Replace all placeholders in all 8 spec files.
- [ ] Confirm primary user outcome and acceptance criteria in [INTENT.md](./INTENT.md).
- [ ] Confirm topology and runtime model in [ARCHITECTURE.md](./ARCHITECTURE.md).
- [ ] Document all inbound/outbound contracts in [INTERFACES.md](./INTERFACES.md).
- [ ] Define validation gates and CI proof surfaces in [VALIDATION.md](./VALIDATION.md).
- [ ] Define state machines and invariants in [SEMANTICS.md](./SEMANTICS.md).
- [ ] Define SLOs, alerting, and incident process in [OPERATIONS.md](./OPERATIONS.md).
- [ ] Define threat model and auth/authz decisions in [SECURITY.md](./SECURITY.md).
- [ ] Ensure architecture diagram, docs, changelog, and tests are mapped to promotion gates.
- [ ] Run all validation/test commands and attach evidence artifacts.

<!-- decapod:codebase-attestation:start -->
## Codebase Attestation

- Repository signal fingerprint: `dffa4e466e1010afa0c48b8b3e0f818d52daa0aaf8e42bb79f99a1c62045c7db`
- Significant implementation surfaces: `.github/` (1 files), `Dockerfile/` (1 files), `README.md/` (1 files), `go.mod/` (1 files), `module-001-hello-go/` (2 files), `module-002-variables-values-and-types/` (2 files), `module-003-functions-and-control-flow/` (2 files), `module-004-slices-maps-and-structs/` (2 files), `module-005-errors-and-return-values/` (2 files), `module-006-files-and-standard-streams/` (2 files), `module-007-first-http-server/` (2 files), `module-008-routes-handlers-and-responses/` (2 files), `module-009-json-request-response/` (2 files), `module-010-configuration-env-vars-flags/` (2 files), `module-011-logging-and-request-output/` (2 files), `module-012-graceful-shutdown/` (2 files), `module-013-request-lifetime-and-instrumentation/` (2 files), `module-014-server-timeouts-and-deadlines/` (2 files), `module-015-rate-limits-and-capacity-protection/` (2 files), `module-016-bearer-token-encrypted-persistence/` (2 files)
- Refreshed from the current codebase by `decapod specs.refresh`
<!-- decapod:codebase-attestation:end -->
