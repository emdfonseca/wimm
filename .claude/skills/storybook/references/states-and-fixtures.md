# The story set

## Derive it, do not invent it

A component's stories come from two places in the design:

1. **The states the state plan names** — Default, Loading/Submitting, Validation error, Server error, Empty, Permission denied, and whatever else the change's `canvas.md` names.
2. **Library stress fixtures (library zone 90)** — long labels, maximum content, empty content, focus, disabled, nested components, localization stress, text growth.

Naming rule: `SKILL.md` ("Story names come from the design").

## A practical baseline per layer

**Atom** — `Default`, one story per state that is not reachable by an arg control (`Disabled`, `Loading`), plus `LongContent`. Visual variants (primary/secondary/ghost) are `argTypes` controls, not stories, unless a variant carries a distinct contract — `Destructive` usually does.

**Molecule** — the Atom set plus its composed states: `WithHint`, `Invalid`, `Required`, `ReadOnly`.

**Organism** — `Default`, `Loading`, `Empty`, `Error`, `LongContent`, and any permission-dependent rendering. Data arrives as args; nothing fetches.

**Template** — one story per genuinely distinct shell composition, with placeholder regions. Fluid adaptation is a viewport global, not extra stories.

**Page (pure screen)** — the states of that step of the flow:

```text
Pages/SignIn      Default · InvalidCredentials · Submitting · AccountLocked · ServiceUnavailable
Pages/Dashboard   Loading · Empty · Populated · PartialData · LoadError · PermissionDenied
```

Data arrives as props, actions leave as callbacks (`onsubmit`, `onretry`) spied with `fn()`. If a state cannot be reached by changing props, the screen is still connected (`SKILL.md` table).

## Stress fixtures are not optional decoration

`LongContent` and `Empty` are where layout actually breaks:

```svelte
<Story name="LongContent" args={{
  title: 'Quarterly revenue reconciliation and variance analysis for EMEA',
  description: 'A'.repeat(400),
}} />

<Story name="Empty" args={{ items: [] }} />
```

If localization is in scope, add a `Localized` fixture using a language that materially lengthens text (German) and, if RTL is in scope, one with `dir="rtl"`. A fixed-height control that clips enlarged text fails WCAG 1.4.4.

## Do not story the theme or the viewport

Those are globals. The one exception is `ThemeQA` — `references/theming-and-viewports.md`.
