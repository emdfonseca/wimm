---
name: storybook-svelte
description: How the pen.dev design system becomes Storybook stories in a SvelteKit project - which layers get stories (Atoms through Templates, never Pages or journeys), how design states become the story set, how Light/Dark and Compact/Medium/Wide map to globals and viewports instead of duplicated stories, and how accessibility contracts become play functions and a11y checks. Use this whenever someone writes, reviews, organizes, or names a .stories.svelte file; sets up or changes .storybook/main.ts or preview.ts; asks where stories live or what should have one; builds a component, template, or app shell that came out of a .pen design; wires Storybook into CI or testing; or asks how to show components, variants, states, themes, or responsive behavior in Storybook. Reach for it even when the ask is a single component ("add a story for the button", "show this in dark mode") - the answer is usually a global or an arg, not another story.
---

# Storybook for the Svelte design system

**Storybook is the implementation-side mirror of the `.lib.pen` design library.** The design library says what reusable UI exists and how it must behave; Storybook proves the built components actually do. If something is in the library, it has stories. If it is not in the library, it almost certainly does not belong in Storybook.

That mirror relationship gives every rule below its shape — including the most important one, which is inherited directly from the canvas: **do not multiply stories by theme × viewport × state.** A story is a *state*. Theme and viewport are context, selected at view time.

```text
.lib.pen library          Storybook
─────────────────────────────────────────────────────
Foundations         →     tokens as CSS custom properties, documented not storied
Atom                →     Atoms/<Name>
Molecule            →     Molecules/<Name>
Organism            →     Organisms/<Name>
Template            →     Templates/<Name>
Page                →     not a story — it is a SvelteKit route
Journey             →     not a story — it lives in the .pen file and in e2e tests
```

## Start here, every time

1. **Which layer is this?** That decides the story title and whether it belongs here at all. Pages and journeys do not get stories — putting them here duplicates routing, data, and flow that Storybook cannot honestly represent.
2. **What are its states?** The story set comes from the design: the local states beside the journey step, plus the stress fixtures. Not from imagination, and not one story per visual variant.
3. **Is this a new story, or an arg / global / play step?** Most additions are an arg (component API), a global (theme, viewport), or an assertion inside an existing story. Add a story only for a genuinely distinct state.

## Where to read next

The API below is current as of the Storybook 10.2 docs (checked 2026-09-13). Storybook moves quickly — if something does not behave as described, fetch the current docs rather than working around it.

| Read this | When |
|---|---|
| `references/setup.md` | Installing or changing `.storybook/main.ts`, `preview.ts`, framework and addon packages, SvelteKit mocking (`sveltekit_experimental`). |
| `references/story-format.md` | Writing a `.stories.svelte` file: `defineMeta`, snippets, args, tags, naming, file placement. |
| `references/taxonomy.md` | Deciding what gets a story, the title hierarchy, and where the Page/journey line falls. |
| `references/states-and-fixtures.md` | Choosing the story set for a component; stress fixtures; mapping design states to story names. |
| `references/theming-and-viewports.md` | Light/Dark via globals; Compact/Medium/Wide viewport presets matching the design frames; the one legitimate side-by-side case. |
| `references/accessibility.md` | a11y addon config, play functions asserting focus/keyboard/name contracts, what automated checks can and cannot prove. |
| `references/testing-and-ci.md` | `@storybook/addon-vitest`, running stories as tests, the `just` verbs, CI wiring. |
| `references/tokens.md` | Design tokens as CSS custom properties; keeping story styling honest. |
| `assets/` | Starter `main.ts`, `preview.ts`, and a `Component.stories.svelte` template. |

## Rules that are constantly needed

### One story per state, never per context

```text
✅ Default · Loading · ValidationError · Empty · LongContent · Disabled
❌ ButtonDark · ButtonMobile · ButtonPrimaryDarkMobile
```

Theme and viewport are **globals**, switchable from the toolbar and settable per story when a state only makes sense in one context. Duplicating stories across them produces the same unreviewable matrix the design standard exists to prevent, and it doubles every future edit.

The single exception mirrors zone 40 of a journey file: a deliberately labelled side-by-side comparison story for a component whose theming is genuinely risky (focus rings, elevation, error states). One per component at most, named `ThemeQA`, and understood as a test fixture rather than a second specification.

### Story names come from the design

The local states beside a journey step are the story set. Keep the names recognisably the same — `Validation error` on the canvas becomes `ValidationError` in Storybook — so a reviewer can hold the design and the implementation side by side without translating. A state that exists in the design and not in Storybook is the gap worth flagging.

### Stories live beside the component

```text
packages/ui/src/lib/components/Button/
├── Button.svelte
├── Button.stories.svelte
└── Button.test.ts
```

Co-location keeps the story in the same diff as the change that breaks it. A `stories/` directory far from the source guarantees the two drift.

### Args are the component's API

`args` are props. If a story needs to set something that is not a prop, that is usually a missing prop or a decorator, not a reason to hand-write markup inside the story. Keep the component's public surface honest — the design system's component contract and the Svelte component's props should describe the same thing.

### Accessibility contracts are executable here

The design library states contracts (focus moves into the dialog, the control has an accessible name, error is not colour alone). Storybook is where those stop being prose: the a11y addon catches the automatable subset, and a `play` function asserts the behavioural part.

```svelte
<Story name="Default" play={async ({ canvas, userEvent }) => {
  await userEvent.tab();
  await expect(canvas.getByRole('button', { name: 'Save' })).toHaveFocus();
}} />
```

Automated axe checks prove the absence of some failures, never the presence of accessibility — keyboard order, focus return, and announcement still need the play function and a human pass.

### Foundations are documented, not storied

Colour ramps, spacing scales, and type scales belong in an MDX docs page that reads the actual CSS custom properties. A story per token is noise, and worse, a story that hardcodes a token's value will keep rendering happily after the token changes.

## Checkpoints

- **A component is "done"** when every design state has a story, the a11y addon is clean, and its accessibility contract has a play function. Visual completeness alone is not done — see `references/accessibility.md`.
- **Before a library release** — the design standard's §10.3 checklist has counterparts here: every adopted theme renders, narrow and wide viewports hold, long-text and empty fixtures exist. `references/states-and-fixtures.md` lists them.
- **When a design changes** — the story set changes with it. A removed state leaves a stale story; a new state that never reaches Storybook is invisible to review.
- **When you see a story that is really a page** — data fetching, routing, multiple organisms wired to a flow — say so. That belongs in a route and an e2e test, and leaving it here makes Storybook slow and dishonest about what it verifies.
