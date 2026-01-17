# Archon Docker Compose Support - Roadmap

## Milestone 1: Docker Compose Deployment Support

**Goal:** Enable multi-service Docker Compose deployments through Archon while maintaining backwards compatibility with existing single-container workflows.

### Phases

#### Phase 1: Data Model Extension ✓
**Goal:** Add compose-related fields to Site and DeployRequest models to support compose deployments.

**Deliverables:**
- [x] Add `SiteType` field (container/compose) to Site model
- [x] Add `ComposeContent` field to DeployRequest
- [x] ~~Update database schema if needed~~ (N/A - file-based config)
- [x] Ensure backwards compatibility for existing sites

**Research:** No — straightforward model changes
**Completed:** 2026-01-17 | **Plans:** 1

---

#### Phase 1.1: Fix Merge Conflicts (INSERTED) ✓
**Goal:** Resolve merge conflicts after merging port mapping functionality into compose branch.

**Deliverables:**
- [x] Resolve conflicts in `archon/internal/models/site.go`
- [x] Keep both SiteType/Compose methods (from Phase 1) and port mapping functions (from merged branch)
- [x] Verify both packages build successfully

**Depends on:** Phase 1
**Completed:** 2026-01-17 | **Plans:** 0 (direct fix)

---

#### Phase 2: Compose Deployment Backend ✓
**Goal:** Implement compose CLI execution with temporary file handling on Archon Node.

**Deliverables:**
- [x] Create compose executor using `docker compose` CLI
- [x] Implement temp file creation for compose YAML
- [x] Add cleanup logic to remove temp files after deployment
- [x] Handle compose up/down lifecycle
- [ ] ~~Integrate with existing deployment progress reporting~~ (deferred to Phase 3 API integration)

**Research:** No — clear implementation path using exec
**Completed:** 2026-01-17 | **Plans:** 1

---

#### Phase 2.1: Implement Pipeline System for Node Deployments (INSERTED) ✓
**Goal:** Refactor the Node deployment architecture into a pipeline-based system for cleaner extensibility and easier future modifications.

**Deliverables:**
- [x] Create pipeline core (Stage interface, Pipeline executor, DeploymentState)
- [x] Implement stages (Validation, PortCheck, SSL, Deployment, Proxy)
- [x] DeploymentStage routes to container OR compose based on type
- [x] Automatic rollback on failure
- [x] Wire pipeline into handlers
- [x] Add compose executor to handlers for type-aware status/stop/delete

**Depends on:** Phase 2
**Research:** Completed — see `.planning/phases/2.1-implement-pipeline-system/RESEARCH.md`
**Completed:** 2026-01-17 | **Plans:** 3

---

#### Phase 3: API Extension ✓
**Goal:** Extend deployment API to accept and route compose content to the new backend.

**Deliverables:**
- [x] Update `/api/v1/sites/deploy` endpoint to accept compose content
- [x] Add compose-specific validation
- [x] Route to compose executor when compose content present
- [x] Maintain existing single-container path unchanged

**Research:** No — extending existing patterns
**Completed:** 2026-01-17 | **Plans:** 1

---

#### Phase 4: TUI Compose Input ✓
**Goal:** Add compose YAML input methods to the TUI for deployment.

**Deliverables:**
- [x] Add compose input screen (paste YAML or file path)
- [x] File path loading with error handling
- [ ] ~~Preview compose content before deployment~~ (deferred - not needed for MVP)
- [x] Update site type indicator in site list

**Research:** No — Bubbletea patterns established
**Completed:** 2026-01-17 | **Plans:** 1

---

#### Phase 5: Port Auto-Detection ✓
**Goal:** Parse compose YAML to automatically detect exposed ports for domain mapping.

**Deliverables:**
- [x] YAML parser for compose files (using gopkg.in/yaml.v3)
- [x] Extract exposed ports from service definitions (short + long form)
- [x] Auto-populate domain mapping port from detected ports
- [x] Allow manual override if auto-detection fails

**Research:** Completed — compose port syntax (short form, long form, ranges, protocols)
**Completed:** 2026-01-17 | **Plans:** 1

---

#### Phase 6: Integration Testing ✓
**Goal:** End-to-end validation of compose deployments with various configurations.

**Deliverables:**
- [x] Test compose deployment lifecycle (unit tests + manual checklist)
- [x] Test domain routing to compose services (checklist scenario)
- [x] Test SSL with compose services (checklist scenario)
- [x] Verify single-container sites still work (code review confirmed)
- [x] Document compose deployment workflow

**Research:** No — executing tests
**Completed:** 2026-01-17 | **Plans:** 1

---

#### Phase 6.1: Fix Site Edit Input Issues (INSERTED) ✓
**Goal:** Fix critical bugs in site edit page preventing user input and navigation.

**Deliverables:**
- [x] Fix text input not working on site edit page (only cycles ENV list)
- [x] Fix 'e' key on selected site showing "not implemented" message
- [x] Verify site edit form fields accept character input
- [x] Verify backspace/delete work in edit form

**Depends on:** Phase 6
**Research:** No — bug fix
**Completed:** 2026-01-17 | **Plans:** 1

---

#### Phase 6.2: Separate ENV Page (INSERTED) ✓
**Goal:** Move environment variables to a dedicated screen, fixing form reinitialization bugs and simplifying navigation.

**Deliverables:**
- [x] Create dedicated ENV vars screen (ScreenSiteEnvVars)
- [x] Add 'v' key on site edit form to navigate to ENV screen
- [x] Fix edit form to only initialize data ONCE (not every render)
- [x] Remove ENV section from main site forms
- [x] Update navigation flow

**Depends on:** Phase 6.1
**Research:** No — architectural simplification
**Completed:** 2026-01-17 | **Plans:** 1

---

#### Phase 6.3: Focused Field Visual Indicator (INSERTED) ✓
**Goal:** Add clear visual indication of which field is currently focused in edit forms (color change + arrow indicator).

**Deliverables:**
- [x] Add ">" arrow prefix to focused field label
- [x] Change focused field color (highlight style)
- [x] Apply to all form screens (create, edit, ENV vars)
- [x] Consistent styling across all field types

**Depends on:** Phase 6.2
**Research:** No — UI enhancement
**Completed:** 2026-01-17 | **Plans:** 1

---

## Dependencies

```
Phase 1 ──► Phase 2 ──► Phase 3 ──► Phase 4
                              │
                              ▼
                          Phase 5 ──► Phase 6
```

Phases 1-3 are sequential (model → backend → API).
Phase 4 (TUI) can start after Phase 3.
Phase 5 (port detection) can run parallel to Phase 4.
Phase 6 (testing) requires all previous phases.

---
*Created: 2026-01-17*
