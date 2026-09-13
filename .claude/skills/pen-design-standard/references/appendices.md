## Appendix A — pen.dev translation cheat sheet

| If this document says… | In pen.dev, that means… |
|---|---|
| Journey file | Normal `.pen` document |
| Design-system library | Converted through Libraries and saved as `.lib.pen`; save source and reopen consumers to adopt updates |
| Foundation | Shared variables, constraints, and principles; not an Atom |
| Atom | Small reusable component such as Button, Checkbox, or Input Control |
| Molecule | Reusable component composed from Atoms, such as Form Field or Search Field |
| Organism | Larger reusable component/composition such as Navigation, Drawer, Filter Panel, or Form |
| Template | Reusable page-level composition/scaffold, often implemented with components + slots |
| Page | Product-specific screen Frame in a journey `.pen` file populated with real content/state |
| Journey | Ordered/branching collection of Pages, overlays, states, and outcomes in a journey `.pen` file |
| Screen | Usually a Frame |
| Responsive row | Group of explicitly sized screen Frames on the canvas |
| Responsive App Shell | Canonical reusable shell composition from the `.lib.pen` library for a meaningful layout regime such as Compact/Medium/Wide |
| Fluid layout | Frame flex layout + Hug/Fill/fixed sizing decisions |
| Variable/token | Variable created in the Variables panel |
| `Color` axis | A theme axis we create with values such as Light and Dark |
| `Device` axis | A separate theme axis we create with Mobile/Tablet/Desktop values |
| Theme axis on canvas | A selectable context on a Frame; one value per axis is active on that Frame at a time |
| Side-by-side Light/Dark QA | Two intentionally duplicated representative Frames/instances with different `Color` selections |
| Set theme on screen | Select Frame → properties → Theme → Add theme → choose axis/value |
| Theme inheritance | Child objects inherit parent frame's axis selection unless overridden |
| Light/Dark component behavior | Component properties reference Color-aware semantic variables |
| Responsive token behavior | Component properties reference Device-aware variables; exact value/shell mapping is in Section 5.3.1 |
| Structural mobile difference | Explicit mobile composition/component structure, not merely Device variables |
| Shared reusable asset | Atom/Molecule/Organism/Template implemented as a component/composition in the library and consumed as instances |
| Flexible reusable region | Empty frame marked as a slot in a component origin; instance content is customizable and suggestions are advisory |
| Persistent sidebar | Part of the app/page layout, typically navigation; represented with reusable shell/navigation components |
| Drawer / side panel | Temporary/contextual task surface; usually a reusable Organism; define modal vs non-modal behavior |
| Modal dialog | Short blocking task/decision with modal focus behavior |
| Route-backed task | A journey state with a stable destination/deep-link behavior, even if visually presented as an overlay |
| Navigation depth | Parent-relationship count from the documented root; route mapping is recorded separately |
| Accessibility | Requirements/annotations/contracts across system and journeys; not a pen.dev theme mechanism by itself |

---

## Appendix B — References

The technical review checked the following primary sources on **2026-09-13**. pen.dev capabilities can change; record the tested app/extension version in project setup and recheck relevant documentation after upgrades. The standard combines product-design conventions with documented capabilities and external guidance. Not every recommendation in this document is dictated by these sources; where the document makes an organizational recommendation, it is a team convention built on top of the referenced capabilities.

### pen.dev capabilities

- [pen.dev Variables and theme dimensions](https://docs.pen.dev/core-concepts/variables)
- [pen.dev Design Libraries](https://docs.pen.dev/core-concepts/design-libraries)
- [pen.dev Components](https://docs.pen.dev/core-concepts/components)
- [pen.dev Slots](https://docs.pen.dev/core-concepts/slots)
- [pen.dev Interface / Flex layout](https://docs.pen.dev/core-concepts/pencil-interface)
- [pen.dev `.pen` format](https://docs.pen.dev/for-developers/the-pen-format)
- [pen.dev Design ↔ Code](https://docs.pen.dev/design-and-code/design-to-code)

Additional capability details are cited where used in Section 2. Canvas/file-format support does not by itself establish runtime behavior.

### Design-system methodology

- [Brad Frost, *Atomic Design — Atomic Design Methodology*](https://atomicdesign.bradfrost.com/chapter-2/)

### Accessibility

- [W3C, Web Content Accessibility Guidelines (WCAG) 2.2](https://www.w3.org/TR/WCAG22/)
- [W3C, ARIA Authoring Practices Guide (APG)](https://www.w3.org/WAI/ARIA/apg/)

APG is practical, non-normative guidance for interaction patterns, including keyboard and focus behavior. WCAG remains the conformance baseline.

### Security-sensitive example

- [OWASP Authentication Cheat Sheet — Authentication and Error Messages](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#authentication-and-error-messages)

The individual W3C Understanding pages cited in Section 5.1.5 explain the relevant success criteria and exceptions. They and APG are informative guidance; the WCAG Recommendation is normative.

### Repository / workspace model

- [pnpm Workspaces](https://pnpm.io/workspaces)

---

**Document principle:** If a first-time reader cannot answer **“what must be decided before work starts, which file does this belong in, which pen.dev mechanism represents it, and who owns the rule?”**, this standard should be clarified rather than relying on team memory.
