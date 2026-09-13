# pen.dev Product Design Organization Standard

> **Purpose:** Define how we organize product design in pen.dev so multiple user journeys, success and failure paths, responsive layouts, light/dark themes, accessibility requirements, reusable Atomic Design assets, and design-to-code work stay understandable as the product grows.
>
> **Audience:** Product designers, design-system designers, engineers, product managers, and reviewers working with `.pen` files.
>
> **Accessibility baseline:** WCAG 2.2 Level AA for web experiences. Record additional platform, contractual, and legal requirements in project setup; they may require separate assessment.
>
> **Core rule:** **Journeys own product flow. The design system owns reusable rules. Accessibility applies to both.**

---

**Review date:** 2026-09-13. **Revision:** Technical and consistency review. **Adoption status:** Proposed standard; organizational approval is separate from this review.

This is a reference standard for organizing design work, with setup and review procedures. It covers design artifacts, reusable assets, interaction contracts, and implementation verification. It does not prescribe a frontend framework or replace a product specification, security review, or full accessibility evaluation.

**Requirement language:** “must” and “do not” express requirements of this team standard; “should” and “prefer” express recommendations that permit a documented reason to depart; “may” and “optional” identify choices. These are team rules unless explicitly attributed to an external specification. Examples illustrate the rules and do not independently create requirements.

**Applicability:** Project setup selects supported platforms, layout regimes, and presentation modes. References to Light/Dark, Mobile/Tablet/Desktop, or other optional capabilities apply only when adopted. Mark an inapplicable checklist item `N/A` with a reason. A narrower representative canvas does not reduce the implementation's accessibility obligations.

**Authority:** The detailed clauses define the standard; examples and checklists apply them. Record approved team-convention exceptions with scope, rationale, owner, impact, and review date. An internal exception does not waive an external requirement.

---

## 1. Scope and organizing model

A product design contains several different kinds of information. The biggest source of canvas disorder is representing all of them by duplicating screens.

We separate them instead.

| Concern | What it means | Where we represent it |
|---|---|---|
| **Journey** | A user goal, such as creating an account | Journey `.pen` file or major journey area |
| **Path** | Primary success, alternative success, failure, recovery, exit | Branches inside the journey |
| **Step** | A stage in the journey | Frames arranged left → right |
| **Navigation hierarchy** | Where the user is in the product | App shell + primary/secondary navigation + breadcrumbs/tabs where appropriate |
| **Task surface** | Where a local action happens without confusing it with product navigation | Inline edit, modal dialog, drawer/side panel, or full page |
| **Responsive structure** | Layout changes between mobile/tablet/desktop | Explicit frames/compositions + pen.dev flex layout |
| **Responsive token values** | Systematic spacing/type/size changes by layout regime | Optional `Device` theme axis + variables |
| **Light/Dark** | Appearance mode | `Color` theme axis + semantic variables |
| **Local UI state** | Loading, validation error, empty, disabled, etc. | Beside the relevant journey step |
| **Reusable UI hierarchy** | Atoms, molecules, organisms, and templates | Shared `.lib.pen` design library |
| **Accessibility** | Requirements that make all of the above usable | Cross-cutting rules in Foundations, reusable Atomic Design layers, Pages/Journeys, implementation, and QA |

The target architecture is therefore:

```text
PROJECT SETUP RECORD
Scope → actors → journeys → navigation → surfaces → responsive → themes → accessibility → implementation

JOURNEY FILES
User goal → paths → steps → responsive compositions → contextual states

DESIGN SYSTEM LIBRARY
Foundations → atoms → molecules → organisms → templates

CROSS-CUTTING STANDARDS
Accessibility + navigation rules + naming + responsive rules + theme rules + implementation contracts
```

### The rule that prevents most duplication

Do **not** build this matrix on the canvas:

```text
journey × path × breakpoint × theme × state × component variant
```

Instead:

```text
Canvas        = journey × path × meaningful responsive composition
Themes        = variable-driven context
Local states  = examples beside the relevant step
System rules  = shared design library across foundations/atoms/molecules/organisms/templates
Accessibility = requirements applied across all layers
```

---

