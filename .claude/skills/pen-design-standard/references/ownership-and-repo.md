## 3. Repository and ownership model

The design organization has two distinct homes:

1. **Reusable design-system work** lives with the shared design-system package.
2. **Product journey work** lives with the deployable app it describes.

For a pnpm workspace, the recommended shape is:

```text
my-product/
│
├── apps/
│   └── web/
│       ├── src/
│       └── design/
│           ├── journeys/
│           │   ├── 01-onboarding.pen
│           │   ├── 02-authentication.pen
│           │   ├── 03-dashboard.pen
│           │   ├── 04-search.pen
│           │   ├── 05-checkout.pen
│           │   └── 06-account.pen
│           ├── explorations/
│           │   ├── navigation-exploration.pen
│           │   └── checkout-redesign.pen
│           ├── prototypes/
│           │   └── optional-prototype.pen
│           └── archive/
│               └── deprecated-work/
│
└── packages/
    └── design-system/
        ├── src/
        └── design/
            └── product-ui.lib.pen
```

This is the **canonical organizational model** used throughout this document. A small product may temporarily keep several closely related journeys in one `.pen` file, but that is a scaling choice—not a different naming model. Do not create one design file per route or page.

### What belongs where?

**Journey files answer:**

> How does this user accomplish this product goal, including success, failure, recovery, states, and responsive behavior?

**The design-system library answers:**

> What reusable Foundations, Atoms, Molecules, Organisms, Templates, interaction contracts, accessibility rules, and tokens should every journey use consistently?

Do not mix those responsibilities into one giant file.

The `apps/` and `packages/` names are team conventions. A pnpm workspace requires a root `pnpm-workspace.yaml`; actual packages need manifests and declared dependencies. The tree above omits ordinary code/build files. A design folder is not automatically a package. See [pnpm Workspaces](https://pnpm.io/workspaces).

Keep the reusable library beside its shared implementation. Create a separate tokens package only when tokens have independent consumers, a build pipeline, or ownership needs. Do not create packages per journey or force code folders to mirror Atomic Design labels. Repository boundaries follow deployability and code ownership; journey files follow goals and product areas.

For non-monorepos, preserve these ownership relationships within the existing repository structure instead of creating unnecessary packages. A journey file can cover a product area such as projects or billing; a family filename need not itself be phrased as a user goal. The journeys inside it must identify concrete goals. Splitting criteria are in Section 7.2.

---

### 3.1. Design, token, and product authority

> **The design system owns Foundations plus reusable Atoms, Molecules, Organisms, and Templates. Journey files own realized Pages, product-specific composition, content, and flow.**

| Item | Owner |
|---|---|
| Semantic color definitions | Design system |
| Button appearance and generic accessibility contract | Design system |
| Button used to submit checkout | Checkout journey |
| Generic error banner | Design system |
| Exact payment failure copy and recovery route | Checkout journey |
| Dialog shell and focus contract | Design system |
| Cancel-subscription confirmation content | Account journey |
| Mobile navigation structure | Design system if shared |
| Which navigation item is active | Journey/screen |
| Generic form validation treatment | Design system |
| Which field fails and what happens next | Journey |

---

#### 3.1.1. Authority, revisions, and shared screens

Record separate authorities for product decisions, reusable visual/interaction specifications, tokens, and executable behavior. “Source of truth” does not mean both design and code may overwrite each other independently. Choose one authoritative token source and a reviewed synchronization direction; generated counterparts are projections of that source.

When a Page participates in several journeys, name one owner and canonical frame for each shared composition/state. Other journeys reference that frame or use a clearly labeled contextual/QA instance. Link the canonical location, purpose of the copy, and reviewed revision. Promote generic mechanics, not an entire business Page merely to avoid a reference. Cross-app flows record one coordinating journey owner and links to app-owned portions and handoff contracts.

When a library, token, or shared Page changes, review its registered consumers, update affected journey references, and reconcile the code mapping. Local overrides must document intent; overrides to accessibility-critical properties need the same review as the underlying contract.

---
