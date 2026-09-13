# The story set

## Derive it, do not invent it

A component's stories come from two places in the design standard:

1. **Local states beside the journey step** — Default, Loading/Submitting, Validation error, Server error, Empty, Permission denied, and whatever else the design actually shows.
2. **Library stress fixtures (§10.4)** — long labels, maximum content, empty content, focus, disabled, nested components, localization stress, text growth.

Keep names recognisable across both artifacts. `Validation error` on the canvas is `ValidationError` here. A reviewer holding the design next to the sidebar should not have to translate, and a missing story should be obvious at a glance.

## A practical baseline per layer

**Atom** — `Default`, one story per state that is not reachable by an arg control (`Disabled`, `Loading`), plus `LongContent`. Visual variants (primary/secondary/ghost) are `argTypes` controls, not stories, unless a specific variant carries a distinct contract — `Destructive` usually does, because its confirmation behaviour differs.

**Molecule** — the Atom set plus its composed states: `WithHint`, `Invalid`, `Required`, `ReadOnly`. A form field that cannot show its error state in Storybook cannot have that state reviewed.

**Organism** — `Default`, `Loading`, `Empty`, `Error`, `LongContent`, and any permission-dependent rendering. Data arrives as args; nothing fetches.

**Template** — one story per genuinely distinct shell composition, with placeholder regions. Fluid adaptation is a viewport global, not extra stories.

**Page (pure screen)** — the journey step's local states, which is usually where the story set pays for itself:

```text
Pages/SignIn      Default · InvalidCredentials · Submitting · AccountLocked · ServiceUnavailable
Pages/Dashboard   Loading · Empty · Populated · PartialData · LoadError · PermissionDenied
```

Data arrives as props, actions leave as callbacks (`onsubmit`, `onretry`) spied with `fn()`. If a state cannot be reached by changing props, the screen is still connected to something it should not be — see `references/taxonomy.md`. The journey that strings these screens together is still e2e territory.

## Stress fixtures are not optional decoration

`LongContent` and `Empty` are where layout actually breaks, and they are cheap:

```svelte
<Story name="LongContent" args={{
  title: 'Quarterly revenue reconciliation and variance analysis for EMEA',
  description: 'A'.repeat(400),
}} />

<Story name="Empty" args={{ items: [] }} />
```

If localization is in scope, add a `Localized` fixture using a language that materially lengthens text (German is the usual choice) and, if RTL is in scope, one with `dir="rtl"`. Text growth breaks fixed-height controls, and a fixed-height control that clips enlarged text fails WCAG 1.4.4 — the fixture is how that surfaces before a user finds it.

## Do not story the theme or the viewport

Those are globals. A `ButtonDark` story is a maintenance liability: it will drift from `Button`, and nobody will notice because both render fine. The one deliberate exception is `ThemeQA` — see `references/theming-and-viewports.md`.
