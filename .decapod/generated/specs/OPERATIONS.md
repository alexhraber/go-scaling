# Operations

## Operational Readiness Checklist
- [ ] On-call ownership defined.
- [ ] SLOs and alert thresholds defined.
- [ ] Dashboards for latency/errors/throughput are live.
- [ ] Runbooks linked for all Sev1/Sev2 alerts.
- [ ] Rollback plan validated.
- [ ] Capacity guardrails documented.

## Workspace Isolation
Git worktrees and optional Docker containers provide isolated workspaces scoped to specific todos, preventing interference with the main repository checkout. Key features:
- **Todo-scoped Worktrees**: Each todo gets an isolated git worktree with branch naming that includes todo IDs/hashes
- **Exclusive Agent Ownership**: Claiming mechanism ensures only one agent can work on a todo at a time
- **Event Journaling**: Todo state changes are journaled for deterministic rebuild
- **Health Subsystem Integration**: Proof events can be associated with todos via health claims

## Generated Artifact Operations
Generated artifacts are operational outputs, not static docs. Agents should expect Decapod to refresh `.decapod/generated/specs/*.md` during explicit refresh operations and validation-assisted refresh. The operation is bounded: product docs under `docs/` remain the human learning surface for Decapod itself, while generated specs carry repo-specific live architecture, interface, validation, semantic, operational, and security facts.

## Service Level Objectives
| SLI | SLO Target | Measurement Window | Owner |
|---|---|---|---|
| Availability | 99.9% | 30d | TBD |
| P95 latency | TBD | 7d | TBD |
| Error rate | < 1% | 7d | TBD |

## Monitoring
| Signal | Metric | Threshold | Alert |
|---|---|---|---|
| Traffic | requests/sec | baseline drift | warn |
| Latency | p95/p99 | threshold breach | page |
| Reliability | error ratio | threshold breach | page |
| Saturation | cpu/memory/queue depth | sustained high | page |

## Health Checks
- Liveness:
- Readiness:
- Dependency health:
- Synthetic transaction:

## Incident Response
- Detection:
- Triage:
- Mitigation:
- Communication:
- Post-mortem:

## Rollout Strategy
- Blue/green deployment:
- Canary release:
- Rolling update:
- Feature flags:

## Capacity Planning
- Traffic patterns:
- Resource utilization:
- Scaling triggers:

## Logging
Use `zap` or `zerolog` with structured fields and propagated context ids.

## Secrets Management
| Secret | Source | Rotation | Consumer |
|---|---|---|---|
| External service auth material | managed runtime configuration | periodic | runtime services |
| Artifact signing material | managed signing service/local secure store | periodic | release pipeline |

## Security Testing
| Test Type | Cadence | Tooling |
|---|---|---|
| SAST | each PR | language linters/scanners |
| Dependency scan | each PR + weekly | supply-chain tools |
| DAST/pentest | scheduled | external/internal |

## Compliance and Audit
- Regulatory scope:
- Audit evidence location:
- Exception process:

## Pre-Promotion Security Checklist
- [ ] Threat model updated for changed surfaces.
- [ ] Auth/authz tests pass.
- [ ] Dependency vulnerability scan reviewed.
- [ ] No unresolved critical/high security findings.

<!-- decapod:codebase-attestation:start -->
## Codebase Attestation

- Repository signal fingerprint: `dffa4e466e1010afa0c48b8b3e0f818d52daa0aaf8e42bb79f99a1c62045c7db`
- Significant implementation surfaces: `.github/` (1 files), `Dockerfile/` (1 files), `README.md/` (1 files), `go.mod/` (1 files), `module-001-hello-go/` (2 files), `module-002-variables-values-and-types/` (2 files), `module-003-functions-and-control-flow/` (2 files), `module-004-slices-maps-and-structs/` (2 files), `module-005-errors-and-return-values/` (2 files), `module-006-files-and-standard-streams/` (2 files), `module-007-first-http-server/` (2 files), `module-008-routes-handlers-and-responses/` (2 files), `module-009-json-request-response/` (2 files), `module-010-configuration-env-vars-flags/` (2 files), `module-011-logging-and-request-output/` (2 files), `module-012-graceful-shutdown/` (2 files), `module-013-request-lifetime-and-instrumentation/` (2 files), `module-014-server-timeouts-and-deadlines/` (2 files), `module-015-rate-limits-and-capacity-protection/` (2 files), `module-016-bearer-token-encrypted-persistence/` (2 files)
- Refreshed from the current codebase by `decapod specs.refresh`
<!-- decapod:codebase-attestation:end -->
