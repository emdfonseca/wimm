## 9. Lifecycle and change management

### 9.1. Artifact status

Names should make the canvas understandable without opening every frame.

Use deterministic labels rather than names such as `Final`, `Latest`, or `Frame 423`.

Use this status vocabulary for reviewed artifacts:

```text
EXPLORING
IN REVIEW
READY FOR BUILD
IMPLEMENTED
DEPRECATED
```

---

#### 9.1.1. Status meaning and evidence

| Status | Meaning / entry evidence |
|---|---|
| EXPLORING | Alternatives or incomplete contracts; not a build specification. |
| IN REVIEW | A named revision is ready for product, design-system, accessibility, and engineering review as applicable. |
| READY FOR BUILD | Section 10.1 passes or records justified N/A items; implementation-blocking decisions are resolved; accountable reviewers and the approved design/library revision are recorded. |
| IMPLEMENTED | The linked code/build implements the approved scope and passes Section 10.2; material deviations are reconciled. It does not automatically mean deployed to users. |
| DEPRECATED | A replacement or retirement rationale, affected consumers, owner, and removal condition are documented. |

Record status, owner, revision, review date, decision/evidence links, and open issues. Status applies to the named artifact or scoped journey, not automatically to every frame in a file. A material design change returns affected work to IN REVIEW; adding an exploratory alternative does not change the approved baseline. Only the accountable owners identified in setup approve readiness and convention exceptions.

---

### 9.2. Reuse and local exceptions

Avoid local copies such as:

```text
checkout.pen
├── local Button
├── local Input
├── local Dialog
└── local Toast
```

when the shared library already provides them.

During exploration, local UI is acceptable. Before a journey becomes implementation-ready:

1. replace duplicated shared controls with imported library instances;
2. promote a genuinely reusable solution into the library; or
3. retain a named, product-specific local composition with an owner and documented rationale, while reusing shared controls inside it.

---

### 9.3. Promotion and variant decisions

Promote a local solution when it is reusable enough to become an Atom, Molecule, Organism, Template, or Foundation rule. Promote when it is:

- repeated across journeys;
- likely to be reused;
- behaviorally consistent across contexts;
- important to standardize;
- risky or expensive to implement differently.

Useful test:

> If this changes later, should all usages update together?

If yes, it is a strong design-system candidate. Then classify it at the **lowest Atomic Design level that fully describes its responsibility**; do not inflate a Molecule into an Organism merely because it has many visual variants.

Before adding a component or frame, ask whether the difference is a variable value, an allowed content override, slot content, a local state, or a structural/behavioral change. Use a new reusable composition when existing mechanisms cannot express the contract clearly. Do not multiply components by every color, device, emphasis, and interaction-state combination.

---

### 9.4. Exploration and archive

Archive superseded approved artifacts with their revision, replacement reference, owner, and reason. Keep stable references resolvable or update their consumers. Delete disposable exploration only when it carries no required decision history; version history should retain approved baselines. Temporary canvas notes are not the permanent decision log.

Use:

```text
apps/<app>/design/explorations/
```

for unconstrained concept work.

Once a direction is accepted:

1. move/refine the accepted solution into the appropriate journey;
2. replace temporary UI with design-system instances;
3. connect semantic variables;
4. define responsive behavior;
5. add important error/recovery paths;
6. perform theme and accessibility review;
7. archive or delete obsolete alternatives.

Do not leave multiple unlabeled “final” versions beside implementation-ready screens.

---

### 9.5. Library release and migration

1. Name the library owner, proposed revision, affected assets, and consumer list. Record changed defaults, overrides, slots, themes, and tokens.
2. Review the source change and perform the Section 10.3 checks before consumer adoption. Preserve stable component and variable identities where practical.
3. Save the source library and reopen representative consuming documents (Section 2.2). Inspect both inherited properties and intentional overrides.
4. Record design/library and code/token revisions together. Coordinate incompatible changes with affected applications; publish replacement and migration instructions before removal.
5. Mark deprecated assets with replacement, owner, and removal condition. Retain them until affected consumers migrate or explicitly retire their usage.
6. Verify references after moving/renaming files. Preserve required fonts and image assets. Copy/paste from an ordinary `.pen` file does not preserve a cross-document source link; resolve variable conflicts deliberately. See [pen.dev Design Libraries](https://docs.pen.dev/core-concepts/design-libraries).
7. Commit reviewed design and code changes with reproducible evidence. Treat `.pen` JSON merges as structural changes: reopen and inspect the result. Record a known-good revision and coordinated rollback procedure.

This release record is a team workflow. A library filename, file conversion, or pnpm package version does not by itself guarantee compatible design imports.

---
