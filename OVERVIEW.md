# Templated Secret Controller – Comprehensive Overview

<!-- TOC will be generated manually to keep repository tooling simple -->
1. [Executive Summary](#executive-summary)
2. [Problem Statement & Use Cases](#problem-statement--use-cases)
3. [High-Level Architecture](#high-level-architecture)
4. [Custom Resource Definition (CRD)](#custom-resource-definition-crd)
5. [Reconciliation & Control Loop Flow](#reconciliation--control-loop-flow)
6. [Templating & Data Extraction Mechanics](#templating--data-extraction-mechanics)
7. [Core Packages & Responsibilities](#core-packages--responsibilities)
8. [ServiceAccount Impersonation Model](#serviceaccount-impersonation-model)
9. [Tracking & Event Propagation](#tracking--event-propagation)
10. [Status & Conditions Semantics](#status--conditions-semantics)
11. [Security Considerations](#security-considerations)
12. [Performance Characteristics](#performance-characteristics)
13. [Operational Concerns](#operational-concerns)
14. [Testing Strategy](#testing-strategy)
15. [Deployment (Helm & Kustomize)](#deployment-helm--kustomize)
16. [Observability & Metrics](#observability--metrics)
17. [Error Handling & Failure Modes](#error-handling--failure-modes)
18. [Edge Cases & Limitations](#edge-cases--limitations)
19. [Recommended Improvements](#recommended-improvements)
20. [Roadmap Ideas / Stretch Goals](#roadmap-ideas--stretch-goals)
21. [Attribution & Upstream Inspirations](#attribution--upstream-inspirations)

> This document was generated via an automated repository analysis (Sep 2025). It aims to act as: architectural brief, onboarding aid, and improvement backlog seed.

## Executive Summary

templated-secret-controller introduces a Kubernetes Custom Resource (SecretTemplate) that declaratively composes a Secret from fields of other Kubernetes objects (Secrets, ConfigMaps, Services, etc.) via JSONPath-based inline templating. It continuously maintains the generated Secret, re‑rendering when source objects change, when the template definition changes, or when a configured max age threshold is exceeded. The controller focuses on:

- Deterministic Secret materialization from heterogeneous sources
- Minimal RBAC surface by default (only Secrets) with opt‑in ServiceAccount based elevation for multi‑resource inputs
- A simple, explicit JSONPath templating DSL wrapping expressions in $( ... ) for predictable evaluation
- Event driven plus periodic reconciliation (for SA / max age scenarios)

The design favors clarity, low operational complexity, and testability (broad unit coverage and integration tests) while leaving room for future enhancements (webhooks, metrics enrichment, secret rotation policies, cross-namespace restrictions, etc.).

## Problem Statement & Use Cases

Many workloads need composite Secrets built from multiple upstream configuration sources (e.g., combine DB credentials from one Secret with service endpoint from a ConfigMap; reformat or join values). Hand‑rolled scripts or CI jobs create drift risks and increase coupling. This controller supplies a native declarative pattern:

Primary use cases:

1. Aggregating credentials or values from multiple Secrets into a single runtime bundle.
2. Deriving stringData with interpolations (prefix/suffix, concatenations) using non‑sensitive inputs while pulling sensitive binary values from upstream Secrets.
3. Referencing dynamically named resources (second input name resolved from first input's data) using nested JSONPath evaluation.
4. Enforcing consistent labeling/annotations across generated Secrets.
5. Periodically forcing regeneration to integrate upstream rotation outside native events (e.g., external secret managers updating source secrets via replace instead of update, or TTL compliance).

Out of scope (today): direct integration with external secret stores, cross‑cluster sourcing, encryption-at-rest customization, fine-grained field transformations beyond simple concatenation and lookup.

## High-Level Architecture

Core loop: Kubernetes controller-runtime manager runs a single controller (SecretTemplateReconciler). For each SecretTemplate:

1. Fetch CR instance and guard deletion.
2. Resolve ordered InputResources (possibly dynamic names via JSONPath evaluated against earlier inputs).
3. Build a value map (with decodedData convenience for Secrets) and evaluate JSONPathTemplate into a desired corev1.Secret (Data, StringData, Type, metadata).
4. Create or Update the Secret (name matches SecretTemplate name) establishing controller owner reference.
5. Update status (conditions, referenced secret name, observed fields) using status subresource with conflict retries.
6. Requeue logic: based on serviceAccount usage or maxSecretAge or rely solely on event tracking of dependent Secrets.

Supporting components:

- Tracker: In-memory reverse index mapping dependent Secrets to owning SecretTemplates when no ServiceAccountName is used (only Secret inputs permitted). Enables targeted reconciles on upstream secret change events.
- Service Account Loader + Token Manager: Constructs scoped dynamic client impersonating a user-provided ServiceAccount to broaden readable resource kinds.
- JSONPath evaluation utilities: Translate $( ... ) expressions to k8s jsonpath parser format and execute with miss tracking.
- Expansion (secondary templating for legacy style $(VAR) variable interpolation in reconciler/secret.go path) used in earlier code paths; primary path now relies on JSONPathTemplate.

## Custom Resource Definition (CRD)

Group: templatedsecret.starstreak.dev, Version: v1alpha1, Kind: SecretTemplate (Namespaced)

Key spec fields:

- spec.inputResources[]: Ordered list. Each has `name` (local identifier) and `ref` (apiVersion, kind, name). `ref.name` itself can be a JSONPath expression referencing prior inputs enabling dynamic chaining.
- spec.template (JSONPathTemplate):
  - stringData: Plain strings with embedded $(jsonpath) expressions (supports concatenation in a single string).
  - data: Map of key -> $(jsonpath) referencing base64-encoded source (auto decoded for .data.* convenience when possible).
  - metadata.labels / metadata.annotations: $(jsonpath) supported per value.
  - type: Secret type templatable via expression.
- spec.serviceAccountName: Optional elevation enabling reading non-Secret resources. Without it only core/v1 Secrets allowed (enforced in code).

Status fields:

- status.secret.name: Reference to managed Secret (defaults to CR name).
- status.conditions: Reconciling | ReconcileSucceeded | ReconcileFailed | Invalid.
- status.observedGeneration & observedSecretResourceVersion: Drift / replay safety.
- status.friendlyDescription: Human oriented summary (currently rarely populated by reconciler—potential improvement area).

OpenAPI schema enforces structure; printer columns show description & age.

## Reconciliation & Control Loop Flow

Sequence (simplified):

1. Entry via watch event or periodic requeue.
1. Load SecretTemplate; if deleted or not found, cleanup tracking and exit.
1. Initialize status, set Reconciling condition.
1. Determine if current managed Secret exceeds maxSecretAge (forceRegeneration flag clears data to drive full rewrite).
1. Resolve input resources:

- Build dynamic client (impersonated if serviceAccountName set).
- For each InputResource sequentially: evaluate dynamic name (JSONPath) using already resolved resources; fetch unstructured object.
- If object is of kind Secret, compute decodedData map (base64-decoded values) merged for more ergonomic template references.
- Accumulate list for tracking (only if no ServiceAccountName).

1. Evaluate JSONPathTemplate into temporary Secret object (evaluateBytes for data; evaluate for stringData/labels/annotations/type with fallback to decoding logic for Secrets).
1. controllerutil.CreateOrUpdate reconciles the real Secret (owner reference → garbage collection on CR delete).
1. Set status.secret.name, adjust conditions (success/failure) via helper status module.
1. Decide requeue:

- If serviceAccountName present OR maxSecretAge>0 → schedule RequeueAfter = reconciliationInterval (default 1h; unit tests show 30s default inside reconciler constant for SA-only path before CLI override).
- Otherwise rely on tracker events of underlying Secrets.

Error handling: early returns propagate error; status helper sets ReconcileFailed with message (e.g., fetch failure). Partial failure leaves prior Secret intact (integration test validates secret remains after source deletion).

## Templating & Data Extraction Mechanics

Two layers appear historically; current primary is JSONPath-based.

1. JSONPath Expression Syntax: Wrap K8s jsonpath in `$(...)`. Example: `$(.creds.data.username)`; internal translation replaces `$( ... )` with `{ ... }` for the k8s jsonpath library. Allows concatenation inside stringData and metadata fields (e.g., `prefix-$(.map.data.key)-suffix`).
1. Data vs StringData:

- data: values expected to come from base64 (Secret.data). Controller attempts optimized decodedData path to avoid double encoding/decoding. If not found, decodes manually.
- stringData: plain strings with interpolation; final controller writes them as StringData then Kubernetes API server encodes to Data.

1. Decoded Secret Convenience: For Secret inputs, decodedData map is injected; templates referencing `.resource.data.key` are transparently translated to decodedData if available to simplify base64 plumbing.
1. Dynamic Input Names: For second+ inputs, `ref.name` can be expression referencing earlier `.first.data.field` or decodedData. Implementation ensures evaluation order is preserved.
1. Type templating: Type field accepts an expression; if existing Secret's type differs, controller logs a warning (K8s immutability) and does not forcibly recreate (improvement opportunity: optional replace mode).
1. Legacy expansion (reconciler/secret.go): Presents a $(VAR) substitution used when applying non-JSONPath style templates (used in earlier code path; retained for backward compatibility tests). Could be unified or deprecated over time.

## Core Packages & Responsibilities

Package overview (key directories and roles):

- cmd/controller: Binary entrypoint, flag parsing, manager setup, CRD readiness polling, rate limiter config, watch namespace scoping, leader election, metrics server binding.
- pkg/apis/templatedsecret/v1alpha1: Type definitions, status & conditions, deepcopy generation.
- pkg/generator: Core reconciliation (secret_template_reconciler.go), JSONPath utilities, input generation tracking annotation, dynamic name resolution.
- pkg/reconciler: Abstraction to assemble Secrets from templates (legacy variable expansion). Status helper (not fully shown above) manages condition transitions.
- pkg/satoken: ServiceAccount token manager & cache (adapted from kubelet logic) controlling token TTL and refresh jitter.
- pkg/tracker: In-memory map enabling reverse lookup from changed Secret → owning SecretTemplate(s) when using default permissions.
- test/ci: Integration test harness (kubectl wrapper, environment bootstrap, waiting functions).
- charts/...: Helm deployment including optional ServiceMonitor, leader election, namespace scoping, HA knobs.

## ServiceAccount Impersonation Model

When spec.serviceAccountName is set:

- Controller uses ServiceAccountLoader (not fully shown here) with TokenManager to request a token for the SA.
- Builds a new client using that token; all input resource GET operations then use the scoped client.
- This broadens allowed input kinds beyond Secrets (ConfigMaps, Services, arbitrary CRDs) subject to the SA's RBAC.
- Because tracker only manages plain Secret inputs without SA, periodic RequeueAfter is enforced for SA-based templates to detect updates to non-Secret inputs (since they are not being tracked via the simplified tracker mechanism).

Security tradeoff: dynamic client per template request introduces token churn; caching mitigates but watch-based propagation for non-Secret inputs is absent (future improvement: generic informer with fine-grained watch aggregator + RBAC guard checks).

## Tracking & Event Propagation

Tracker maps: SecretTemplate (tracking) → set of Secrets (tracked). On reconcile, after successful input resolution (and only when no ServiceAccountName) the tracker is reset then repopulated with currently referenced Secrets. A Secret watch lists owner secrets and uses reverse lookup to enqueue associated SecretTemplates. This enables low-latency propagation for rotated source Secrets without polling.

Limitations:

- Non-Secret resources are never tracked; rely on periodic reconcile if SA used.
- Changing serviceAccountName from empty → non-empty silently shifts from event-driven to interval-based; doc note could help avoid surprise.

## Status & Conditions Semantics

Conditions used:

- Reconciling (set at start, cleared/replaced)
- ReconcileSucceeded (ConditionTrue indicates last attempt succeeded)
- ReconcileFailed (ConditionTrue with message on last failure)
- Invalid (spec structural issues; not observed in current code path but reserved)

Behavioral notes (from tests):

- On missing input, status.secret.name stays empty and ReconcileFailed set.
- Once inputs appear, controller produces Secret and sets ReconcileSucceeded.
- If inputs later deleted, Secret persists (no garbage collection of generated Secret on failure) and condition flips back to ReconcileFailed (design choice prioritizing continuity of last known good secret—documented in tests). Potential improvement: optional pruning/invalidating annotation.

ObservedSecretResourceVersion retained but presently not heavily leveraged (could support diff-free idempotence or stale write detection).

## Security Considerations

Strengths:

- Principle of least privilege default: only allows reading Secrets unless explicitly provided a ServiceAccountName.
- OwnerReferences ensure cleanup on CR deletion (prevents orphaned generated secrets).
- ServiceAccount token manager refresh logic validates tokens (TokenReview) and applies jitter to avoid thundering herd.

Risks / Areas to monitor:

- Generated Secret name fixed to SecretTemplate name; name collisions with pre-existing user-managed Secret could cause silent merge/update (should perhaps refuse adoption unless annotated or owned already).
- Type change warning (not enforced) might leave mismatched Type vs consumer expectation.
- Lack of explicit validation webhook (webhooks.enabled flag present in Helm but disabled by default and no webhook implementation in code) -> invalid JSONPath or circular/dynamic names resolved late (only at reconcile) rather than rejected early.
- No per-field entropy / randomness injection (out of scope but some use cases may expect secret generation rather than synthesis).
- Decoded secret data inserted into evaluation map; ensure no unintended logging (current code does not log raw values; safe).
- Potential privilege escalation if a user can cause controller to impersonate high-privilege ServiceAccounts via spec.serviceAccountName—RBAC should restrict who can create SecretTemplates.
- Cross-namespace Secret sourcing (feature gated). When `--enable-cross-namespace-secret-inputs` is set the controller can read Secrets in other namespaces only if the source Secret declares `templatedsecret.starstreak.dev/export-to-namespaces` listing the consumer namespace or `*`. Wildcard usage broadens exposure and should be audited. A warning condition `CrossNamespaceInputDegraded` is set if a source namespace is not watched, since updates may not trigger reconciliation.

## Performance Characteristics

Expected low resource footprint:

- Single reconciliation does linear O(n) API GETs for n inputResources.
- Default periodic reconcile only when SA used or maxSecretAge>0; else event-driven.
- Workqueue rate limiter: exponential backoff min 100ms → max 120s for failing reconciles reduces API pressure on persistent errors.
- Secret decode overhead minimal (base64 decode per key) with caching only at per-reconcile scope.

Potential bottlenecks / scaling considerations:

- Many templates each using SA will all wake every reconciliationInterval (default 1h) — acceptable; lower intervals could cause bursty load.
- Input chain depth (dynamic names) introduces serial dependency; no parallel fetch optimization.
- Absence of shared informer caching for SA clients (each fetch uses non-cached client) may increase API calls vs possible aggregated multi-Kind watch approach.

## Operational Concerns

Flags (from main.go):

- --watch-namespaces (comma list) or legacy --namespace
- --metrics-bind-address (":8080" or 0 to disable)
- --leader-elect / --leader-election-id
- --reconciliation-interval (default 1h) applies to SA or max age requeues
- --max-secret-age (default 720h / 30d) age-based forced regeneration (set 0 to disable)
- --log-level (debug/info)

Helm values map cleanly to flags; multi-namespace uses controller-runtime cache defaults.

Upgrade concerns:

- Changing maxSecretAge could trigger mass regeneration cycle if many secrets aged beyond new threshold.
- Reducing reconciliationInterval below token refresh jitter might slightly increase token requests; still safe.

Disaster recovery:

- Recreating controller does not lose state; outputs re-derived from inputs.
- Tracker in-memory state rebuilt on first reconcile; no persistence required.

## Testing Strategy

Coverage Layers:

- Unit tests for generator (1k+ lines) covering templating matrix: data extraction, dynamic names, annotations, labels, secret type, multi-input ordering, service account path requeue expectations.
- Unit tests for reconciler/secret builder verifying label/annotation merging and template precedence.
- JSONPath translation tests (jsonpath.go) likely exist (not shown but typical) plus expansion tests.
- Integration tests (test/ci) validate lifecycle, failure recovery, SA permission denial, and persistence of previously created Secret after input removal.
- Status condition marshalling tests ensure API stability for clients.

Gaps / Suggestions:

- Add fuzz tests for JSONPathTemplate evaluation (malformed expressions, large nested maps).
- Add benchmark(s) for reconcile path with varying inputResource counts.
- Add test ensuring maxSecretAge triggers regeneration.
- Add security test verifying that non-Secret input without serviceAccountName is rejected.

## Deployment (Helm & Kustomize)

Helm Chart provides:

- CRD installation toggle, metrics toggle, ServiceMonitor, leader election, PDB, autoscaling scaffold, namespace watch scoping, reconciliation & age tuning.
- Image tag default "latest" (recommend pin by appVersion for reproducibility).

Kustomize overlays reference external config paths (prod/dev). Release manifest bundling supported (single YAML apply path).

Recommendation: add example values for enabling maxSecretAge=0 to show disabling, and for multi-namespace watch.

## Observability & Metrics

Metrics server is controller-runtime default (bind address configurable). Current repo does not expose custom Prometheus counters/histograms (no direct instrumentation in code). ServiceMonitor optionally creates scrape config.

Suggested metrics to add:

- secret_template_reconcile_total{result="success|error"}
- secret_template_inputs_resolved{count}
- secret_template_force_regeneration_total
- secret_template_reconcile_duration_seconds histogram
- secret_template_tracked_sources gauge

Logging: zap logger; log level debug toggles Development true (structured logs). No sensitive value logging observed.

## Error Handling & Failure Modes

Failure Classes:

- Input fetch not found → ReconcileFailed, no Secret creation (or Secret retained if previously created).
- JSONPath evaluation errors → ReconcileFailed.
- Status update conflict → retried; fallback to full object update for fake client/testing scenario.
- ServiceAccount token retrieval failure → surfaces as reconcile error; backoff via rate limiter.
- Secret type mismatch on update → logged warning only (potential drift hazard).

Resiliency: Exponential backoff plus eventual periodic requeue mitigate transient API outages.

## Edge Cases & Limitations

Identified:

- Changing template type after initial creation not enforced; clients may misinterpret secret.
- Dynamic name resolution chain failure mid-list aborts entire reconcile (no partial evaluation fallback).
- No guard against very large inputResources list (memory/time unbounded except by cluster policy).
- No concurrency >1 (MaxConcurrentReconciles left at default 1) may limit throughput under heavy write load (could be intentionally conservative to avoid API bursts).
- Reconcile interval for SA templates identical irrespective of number/age of inputs; adaptive scheduling could lower churn.
- Lack of webhook validation may allow obviously invalid specs to persist until first reconcile.
- FriendlyDescription rarely populated (unused user-facing field).

Security Limitations:

- No explicit support for secret rotation triggers (outside age or source change).
- If a source Secret key is removed, old value persists until template re-renders (which will drop key only if template no longer references it; current logic overwrites Data entirely so removed keys disappear—OK, but needs explicit doc).

## Recommended Improvements

Prioritized (P1 highest):

P1 Reliability & Safety:

1. Add validation webhook: ensure at least one inputResource, template not nil when required, disallow duplicate names, early JSONPath parse validation, enforce secret name collision policy.
2. Enforce or option-gate secret type change (e.g., annotation `templatedsecret.starstreak.dev/allow-type-recreate: "true"`).
3. Add metrics instrumentation for observability; expose readiness/health endpoints (controller-runtime supports healthz and readyz wiring—currently not configured in main.go).
4. Gracefully handle missing template (nil JSONPathTemplate) with clear Invalid condition.

P2 Performance / Scalability:
5. Support configurable MaxConcurrentReconciles for high object counts.
6. Introduce shared informer based generic watch for non-Secret inputs when SA used (with RBAC caveats) to reduce periodic polling.
7. Cache decoded Secret data per reconcile across multiple key references (currently re-evaluates expression each time, though fast).

P3 Security / Policy:
8. Optional annotation to force failing reconciliation to clear or mark generated Secret (e.g., set label stale=true) for downstream consumers.
9. Add RBAC recommendation doc snippet restricting who can create SecretTemplates vs read Secrets.
10. Allow opt-in namespace allow/deny list even when cluster-scope watch (restrict referencing to same namespace already enforced—but watch scoping could reduce cache memory).

P4 Developer Experience / UX:
11. CLI plugin (kubectl sts render) to locally dry-run template resolution.
12. Improve FriendlyDescription population (e.g., summarizing keys created, input count).
13. Add examples demonstrating dynamic name chaining and maxSecretAge usage.
14. Provide structured events (recorder) for success/failure for kubectl describe visibility (currently not using event recorder).

P5 Documentation & Community:
15. Document difference between .data and decodedData and best practice referencing.
16. Clarify retention policy: secrets are not deleted when inputs vanish—highlight for auditors.
17. Add architecture diagram (mermaid) to README or this overview.

P6 Future Features / Stretch:
18. Field transformation library (hashing, HMAC, base64 decode/encode toggles) via extension functions.
19. Cross-namespace safe referencing via explicit allowlist annotation & aggregated ClusterRole (careful security review needed).
20. Rotation hooks: annotation specifying an external trigger ConfigMap/Secret to bump to force regeneration.
21. Support generating multiple Secrets from one template (subresource or list output) for sharded credentials.
22. Add optional encryption KMS wrap (store ciphertext in Secret, sidecar decrypt at mount) – advanced.

## Roadmap Ideas / Stretch Goals

Summarized Vision Extensions:

- Rich validation + admission control
- Observable metrics & dashboards
- Advanced watch strategy replacing periodic poll
- Pluggable transformation functions
- Multi-output templates / composition sets
- Dry-run & preview tooling
- Policy integration (OPA/Gatekeeper samples)

## Attribution & Upstream Inspirations

Code headers attribute Carvel authors for several foundational files (secret reconciliation patterns, token manager adaptation, expansion engine). Design draws on familiar controller-runtime conventions. Token manager adapted from Kubernetes kubelet token management. JSONPath handling leverages k8s client-go jsonpath implementation.

This document intentionally consolidates dispersed knowledge for onboarding and strategic planning.

---
Generated: 2025-09-30
Tooling: Automated static analysis of repository contents.
