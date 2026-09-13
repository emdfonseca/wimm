## 10. Verification and readiness gates

### 10.1. Ready for build

#### Project alignment

- [ ] The Project Setup Record exists and reflects the current scope.
- [ ] Open decisions record question, assumption, owner, decision due date, and impact; none block the proposed implementation scope.
- [ ] Journey screens match the agreed navigation hierarchy and task-surface model.
- [ ] Responsive frames match the agreed layout regimes rather than an arbitrary device catalog.
- [ ] Theme axes/values used by the journey match the project setup decision.
- [ ] Design-system reuse/new-system-work decisions are still valid.
- [ ] Implementation constraints that materially affect the experience are reflected in the design.

#### Flow

- [ ] User goal is stated.
- [ ] Entry point is clear.
- [ ] Primary success path is complete.
- [ ] Success outcome is explicit.
- [ ] Important alternative success paths are shown.
- [ ] Important failure paths are shown.
- [ ] Recovery behavior is shown where relevant.
- [ ] Cancel/back/exit behavior is clear where relevant.

#### States

- [ ] Loading/submitting behavior is defined where material.
- [ ] Validation errors are defined.
- [ ] System/server errors are covered where material.
- [ ] Empty states are covered where material.
- [ ] Permission/authentication states are covered where material.
- [ ] Destructive actions have intentional confirmation/recovery behavior.

#### Responsive

- [ ] Each adopted structural regime has a composition or an explicit shared-structure mapping.
- [ ] Frame widths/heights and exact structural boundaries are recorded.
- [ ] All adopted Device values map to the intended shell/token behavior.
- [ ] Structural breakpoint changes are visible.
- [ ] Flex/Hug/Fill behavior is used intentionally.
- [ ] Device theme values are explicitly assigned where Device-aware variables are used.

#### Themes

- [ ] Semantic variables are used instead of journey-local theme colors.
- [ ] Representative screens were checked in every adopted Color mode.
- [ ] Focus, error, disabled, overlay, border, and elevation states remain understandable in every adopted mode.

#### Navigation and task surfaces

- [ ] Current product location and active navigation state are clear.
- [ ] Breadcrumbs, tabs, steppers, and Back each represent the correct relationship.
- [ ] Save, Cancel, Close, and Back destinations are explicit where relevant.
- [ ] Temporary forms are intentionally classified as inline, modal, drawer/side panel, or full page.
- [ ] Modal/drawer states are identified as ephemeral or route-backed when that distinction matters.
- [ ] Persistent sidebar navigation is not confused with a temporary drawer/side panel.
- [ ] Destructive actions use an appropriate confirmation/recovery model.

#### Accessibility

- [ ] Critical path can be completed without pointer-only interaction assumptions.
- [ ] Focus order and important focus transitions are specified/reviewed.
- [ ] Validation and error recovery do not rely only on color.
- [ ] Status/success/error changes that require assistive-technology announcement are identified for implementation.
- [ ] The measurable requirements in Section 5.1.5 and criterion-level applicability record in Section 5.1.6 are reflected in the design and implementation acceptance criteria.
- [ ] Responsive variants preserve access to essential actions/content.
- [ ] Motion-sensitive behavior has an appropriate reduced-motion approach where relevant.
- [ ] Every adopted presentation mode preserves the applicable accessibility baseline.

#### System integrity

- [ ] Shared controls use imported design-system instances.
- [ ] No accidental detached/local copies remain.
- [ ] Reusable new solutions were considered for promotion.
- [ ] Instance overrides are intentional.

---

### 10.2. Implementation verification and definition of done

`READY FOR BUILD` approves a specification; `IMPLEMENTED` requires evidence from working code. For the named scope, record:

- [ ] Approved design revision, library revision, token source, code commit/build, and review owners are linked.
- [ ] Required paths, local states, permission boundaries, and transitions work with realistic data and failures.
- [ ] Routing, deep links, direct entry, refresh, Back, cancellation, unsaved work, and resume behavior match the contract.
- [ ] Repeated submissions, delayed responses, unknown outcomes, and concurrent edits have defined handling where relevant.
- [ ] Every adopted shell and presentation mode has been verified; boundary widths, text growth, overflow, and Section 5.1.5 checks pass.
- [ ] Automated accessibility checks are supplemented by manual keyboard, focus, and supported browser/assistive-technology testing. Record versions and results.
- [ ] Applicable WCAG criteria are evaluated for the complete process, including third-party steps; defects and limitations are recorded with owners.
- [ ] Design/code differences are resolved or explicitly accepted within the team's authority; applicable external failures remain failures.
- [ ] The canonical design and documentation reflect implemented behavior. Deployment/release status is recorded separately.

Use a compact coverage record: `journey/path/step → scenario → required contexts → expected outcome → evidence → owner → result`. Mark results Pass, Fail, Not tested, or N/A with reason. Sampling controls canvas size; it does not turn Not tested into Pass. Select representative cases to cover every changed reusable asset, unique layout/interaction, and risky cross-axis combination, and expand coverage when failures or material differences warrant it.

---

### 10.3. Design-system release checklist

Before releasing changes to Foundations, Atoms, Molecules, Organisms, or Templates:

- [ ] The consumer impact is explicit, including additive changes and changes to existing defaults.
- [ ] Every adopted Color value resolves correctly.
- [ ] Relevant Device values resolve correctly.
- [ ] Relevant Atoms, Molecules, Organisms, and Templates were checked at narrow and wide widths where relevant.
- [ ] Focus, disabled, error, selected, and other important states remain coherent.
- [ ] Accessibility contract is still valid.
- [ ] Long text/localization stress cases were considered.
- [ ] Text growth and wrapping were considered.
- [ ] Consumer compatibility is verified, or an intentional breaking change has a coordinated migration under Section 9.5.
- [ ] Code token/reusable-asset impact was considered.
- [ ] The Atomic Design classification is still appropriate and does not duplicate a neighboring layer.
- [ ] Deprecated patterns are clearly marked.

---

### 10.4. Library stress-test fixtures

The library should contain dedicated stress-test frames across representative Atoms, Molecules, Organisms, and Templates.

Recommended coverage:

```text
All adopted Color values
All adopted Device token contexts
Long labels
Maximum content
Empty content
Error state
Focus state
Disabled state
Nested components
Localization stress
Text growth
High-content-density examples
```

For theme testing, set the relevant frame's theme explicitly rather than recoloring instances manually.

---

