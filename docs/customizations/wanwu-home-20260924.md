# Wanwu homepage and upstream integration (2026-09-24)

## Sources and scope

- Upstream: `Wei-Shaw/sub2api`, `upstream/main` at `a3eb7ef30` (version 0.2.8).
- Input: `wanwu-home-20260924.idiff`, 21,096,429 bytes.
- SHA256: `2a948dc222100c06c2a0fea36c2b63934b0d5bc8fc84664cc2c2440d8bc88b90`.
- The input is a full frontend snapshot, not an incremental homepage patch.
  It was restored into a separate directory before selecting homepage changes.
- All 853 manifest entries verified: 53 matched byte-for-byte; 800 matched
  after reversing Git's Windows CRLF conversion. No content mismatches remained.
- Imported `HomeView`, the Wanwu homepage component/copy/motion/tests,
  its catalog helper and eight responsive WebP assets (483,372 bytes total).
  Unrelated snapshot administration pages and dependency files were not overlaid.

## Compatibility decisions

- Administrator-provided HTML/URL home content and compact mode retain precedence.
- Homepage uses configured branding, auth state, documentation URL and API URL.
- Public model visibility follows the existing model-plaza permission flag.
  Exclusive groups are not shown. Backends without popularity ranks display the
  real public catalog without claiming it is a seven-day popularity ranking.
- CCS/OpenCode import retains configurable client/model selection, a fixed
  `https://wanwuplus.com` API host, and the current website as the homepage.
  The upstream Codex `/v1` endpoint correction is retained. The usage query uses
  the fixed `/v1/usage` URL to avoid a doubled `/v1` after that correction.
- Cache-estimation preparation on the Messages-to-Chat-Completions path is
  retained alongside upstream final reasoning-effort normalization.
- Codex quota-overdraft status and duplicate-refresh suppression remain, together
  with upstream usage display additions. API-key fallback routing fields remain.
- DeepSeek pricing policies, Beijing peak/weekend rules and the default 50%
  cache confidence remain. Legacy Flash price regression expectations now agree
  with the previously corrected catalog and constants (1.54/4.62/0.049 per MTok).
  Pro pricing tests use a fixed date before the existing September 14 cutoff.

## Verification

- Frontend production build: passed (existing chunk-size warnings only).
- Backend server build: passed with the official Go 1.27.0 Windows toolchain.
- Homepage/import/account usage/key-management regression suite: 114 passed.
- Progress bar/OpenCode Go/Codex model selection regression suite: 25 passed.
- Locale completeness: 3 passed during the build.
- Backend targeted unit suite passed for DeepSeek pricing/cache estimation,
  Codex fingerprint/overdraft, OpenCode Go, Ollama host restrictions and
  API-key fallback. Explicit upstream `cached_tokens: 0` remains authoritative.
- Restored upstream OpenCode Go connection-test routing and session headers
  that otherwise would have been lost while merging the customized test service.
- Browser inspection: desktop and 390px mobile homepage render correctly;
  no broken loaded images or horizontal document overflow observed.

The initial integration was local-only. A separately requested deployment audit
found additional September 19 production customizations and reconciled them:

- Restored the public model catalog's account/group mappings, seven-day top-ten
  ranking, five-minute popularity cache and private-group exclusion.
- Preserved production model-plaza CNY relay prices, cache-price cards, group
  filters and user-specific rates using the matching supplied frontend snapshot.
- Kept upstream 0.2.8 generic reasoning-effort multipliers and group-price
  precedence, rather than reintroducing the older implicit Max multiplier.
- Added the production catalog tests; homepage/plaza regression suite: 77 passed.
- Frontend production build and targeted service/handler Plaza and DeepSeek
  tests passed. Repository package compiled (no matching unit tests).
- Deployment target: existing `NecroticGlow/sub2api` main branch and server
  production service on port 8080. The separate 8888 installation is untouched.
- Release backups and isolated Docker validation are kept privately on the
  server under `/opt/sub2api-release-20260924`, never in Git.
