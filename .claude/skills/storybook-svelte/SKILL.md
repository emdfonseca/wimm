---
name: storybook-svelte
paths:
  - "**/*.stories.svelte"
  - "**/.storybook/**"
  - "**/src/**/*.mdx"
description: "Storybook for the SvelteKit design system: which layers get stories (Atoms to pure-screen Pages, never routes or journeys), one story per state with theme and viewport as globals, a11y contracts as play functions. Use whenever writing, reviewing, or naming a .stories.svelte file, editing .storybook/main.ts or preview.ts, wiring it into tests or CI, or asked to show a state or theme in Storybook."
---

# Storybook for the Svelte design system

Storybook mirrors the `.lib.pen` design library: a story is a *state*; theme and viewport are globals selected at view time. Never multiply stories by theme × viewport × state.

## What gets a story

| Design layer | Storybook title | Notes |
|---|---|---|
| Foundations | `Foundations/*` MDX docs | Read live CSS custom properties; no story per token |
| Atom | `Atoms/<Name>` | Every variant an arg, every state a story |
| Molecule | `Molecules/<Name>` | Includes composed states (field + error + hint) |
| Organism | `Organisms/<Name>` | Mock data as args; no network |
| Template | `Templates/<Name>` | Placeholder regions, not real pages; one story per structurally distinct shell, fluid adaptation is the viewport global |
| Page | `Pages/<Name>` | The **pure** screen component in `src/lib/screens/<Name>/`: data as props, actions as callbacks |
| (route module) | — | `+page.svelte` / `+page.server.ts` wire data and navigation. Not a story. |
| Journey | — | Lives in the `.pen` file; verified by e2e tests |

If a Page story needs `$app/state`, a load function, or `fetch` mocked to render, the screen is still connected: split the pure screen out of the route instead of mocking.

## Start here, every time

1. **Which layer is this?** That decides the title.
2. **What are its states?** The local states beside the journey step plus the stress fixtures — not one story per visual variant.
3. **New story, or an arg / global / play step?** Add a story only for a genuinely distinct state.

## Where to read next

| Read this | When |
|---|---|
| `references/setup.md` | Packages, `.storybook/main.ts`, `preview.ts`, SvelteKit mocking (`sveltekit_experimental`). |
| `references/story-format.md` | Writing a `.stories.svelte` file: `defineMeta`, snippets, args, tags, titles. |
| `references/states-and-fixtures.md` | The story set per layer; stress fixtures. |
| `references/theming-and-viewports.md` | Light/Dark ↔ `light`/`dark` globals; Compact/Medium/Wide viewports; the one `ThemeQA` exception. |
| `references/accessibility.md` | a11y addon config; contracts as play functions. |
| `references/testing-and-ci.md` | `@storybook/addon-vitest`, stories as tests, CI. |
| `references/tokens.md` | Design variable → CSS custom property mapping; layer → code vocabulary. |
| `assets/` | Starter `preview.ts` and `Component.stories.svelte`. |

## Rules that are constantly needed

### One story per state, never per context

```text
✅ Default · Loading · ValidationError · Empty · LongContent · Disabled
❌ ButtonDark · ButtonCompact · ButtonPrimaryDarkCompact
```

Theme and viewport are **globals**, switchable from the toolbar and pinnable per story when a state only makes sense in one context. The single exception is one `ThemeQA` story per component at most — `references/theming-and-viewports.md`.

### Story names come from the design

Story name = pen state name with spaces stripped, PascalCase: `Validation error` → `ValidationError`. A state in the design with no story is the gap to flag.

### Stories live beside the component

```text
packages/ui/src/lib/components/Button/
├── Button.svelte
├── Button.stories.svelte
└── Button.test.ts
```

### Args are the component's API

`args` are props. A story that needs something that is not a prop is missing a prop or a decorator, not hand-written markup.

### Accessibility contracts are executable here

The a11y addon catches the automatable subset; a `play` function asserts the behavioural part.

```svelte
<Story name="Default" play={async ({ canvas, userEvent }) => {
  await userEvent.tab();
  await expect(canvas.getByRole('button', { name: 'Save' })).toHaveFocus();
}} />
```

Axe proves the absence of some failures, never the presence of accessibility — keyboard order, focus return, and announcement need the play function and a human pass.

### Foundations are documented, not storied

Colour ramps, spacing scales, and type scales belong in an MDX page that reads the actual CSS custom properties.

## Checkpoints

- **A component is done** when every design state has a story, the a11y addon is clean, and its accessibility contract has a play function.
- **Before a library release** — every adopted theme renders, Compact and Wide viewports hold, `LongContent` and `Empty` fixtures exist.
- **When a design changes** — the story set changes with it; a removed state leaves a stale story.
- **When a story needs heavy mocking** — split the pure screen out of the route.
