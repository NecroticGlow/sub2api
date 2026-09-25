# v2.8.11 integration audit (2026-09-25)

## Source and scope

- Source: `ranxi2001/sub2api`, tag `v2.8.11`, peeled commit `3bf31dedc335318238fbf29e376e10f9329d3eb5`.
- Integration branch: `codex/v2.8.11-custom`.
- The whole release is merged. Only the legacy standalone intelligence-test implementation is retired; unrelated local customizations remain.
- Code integration and local verification do **not** imply production deployment or a demonstrated improvement in model intelligence.

## Intelligence / alternate routing inventory

| Area | Implementation / connection |
| --- | --- |
| Excel/BPS adapter | `backend/internal/service/openai_excel_bps.go`, `basispoints/`; `Forward` checks `IsExcelBPSEnabled` before routing. Account tests use the same gateway for BPS accounts. |
| Codex tickets and native mint | `openai_codex_ticket*.go`, `codex_harvest_780.go`; gateway ticket binding, egress pinning, response observation/strict validation retained. |
| Harvest management | Admin harvest control/node/flow/manual endpoints and account UI retained. |
| Manual tests | Upstream `IQTestModal.vue`, `pelican_test_service.go`, candy and Pelican questions retained. |
| Scheduled tests / quality operations | `pelican_scheduled.go`, `account_quality.go`, `quality_judge.go`, corresponding repositories, routes and UI retained. |

Core BPS, ticket/harvest and judge implementation files were compared against the source tag. Request-wrapper differences preserve local 429 retries while keeping ticket pinning and upstream plugin/proxy handling.

## Local compatibility changes

- Added the supplied offline-knowledge question alongside candy/Pelican. Knowledge answers use text, not HTML; a completed response is not an automatic “not degraded” verdict. Knowledge-only completion cannot automatically recover a disabled account.
- Manual history is shared through admin-only `GET/POST /admin/accounts/:id/pelican-test-history`, not browser storage. Database appends are atomic, deduplicated by batch ID and bounded to eight batches (256 KiB each). Ordinary account edits preserve the current locked history. These admin-submitted display records are not authoritative billing or automated quality-action evidence.
- Quality plans require a configured judge. Previously a plan without a judge could repeatedly generate answers without ever reaching a verdict.
- Restored account-creation capabilities that an older local file had lost: MiniMax/OpenCode, expiry shortcuts, upstream request-ID options, capability sync and image URL conversion settings. Kept configurable 429 retries, unique account-device defaults and quota-overdraft settings.
- Retained custom Wanwu pages and restored upstream platform/effective-rate filters in their model catalog.
- Regenerated Ent and Wire integration to retain local schema fields and dependencies together with the upstream release.
- Local DeepSeek cache estimator (50% default confidence with jitter and LRU prefix history), native-cache precedence, time-dependent pricing, CCS/OpenCode import, group concurrency/fallback and import identity matching remain.
- Both build version inputs read `2.8.11`.

## Verification completed

- Frontend: 152 tests across nine suites passed, including account creation/editing, BPS toggle persistence, import endpoints, question/history UI, model catalog and locale completeness.
- `pnpm run build` passed (locale checks, Vue/TypeScript checking and production asset build).
- Focused Go unit-tag suites passed in service, admin handler and repository packages, covering DeepSeek pricing/cache, 429 retries, account fingerprints, fallback/concurrency policies, BPS/tickets and intelligence/quality tests.
- Additional repository tests cover atomic manual-history append and preserving the latest database history during a stale account edit.
- GPT-6 Luna performed read-only integration and shared-history safety reviews. No production test or inference request was sent.

## No-quota verification / deployment gate

- Local tests use mocks or loopback servers. Do not call production `pelican-test`, account `test`, manual harvest, quality trigger or model inference endpoints for acceptance.
- BPS is account opt-in; automatic ticket harvesting defaults off. Existing database/runtime settings can override defaults and must be inspected before deployment.
- Explicit manual harvesting can probe even with automatic harvesting disabled. Enabled scheduled plans and configured judges make real upstream requests. Never enable them as an acceptance check.
- Deployment requires a fresh database/data/config backup, migration validation on an isolated copy with no upstream egress, and a rollback plan. This audit does not change production settings or enable any account feature.
