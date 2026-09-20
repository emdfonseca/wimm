# Accessibility in Storybook

A change's `canvas.md` states accessibility contracts as prose, under Contracts for implementation. Storybook is where the automatable part becomes a check and the behavioural part becomes an assertion. Neither replaces a manual pass, and being honest about that gap is what keeps the checks worth running.

## The a11y addon

```ts
// preview.ts
parameters: { a11y: { test: 'error' } }
```

`'error'` fails the build in CI. `'todo'` reports in the UI without failing — useful while adopting, dangerous as a permanent setting, because a permanently yellow check is a check nobody reads. Per-story overrides handle genuine exceptions:

```svelte
<Story name="Decorative" parameters={{ a11y: { test: 'off' } }} />
```

Any `off` needs a comment saying why. "It was failing" is not why.

## What axe catches, and what it does not

Automated rules find missing names, contrast failures on rendered text, bad ARIA, and structural problems — perhaps a third of what matters. They cannot tell you whether focus goes somewhere sensible when a dialog opens, whether the error message is announced, whether the keyboard path through a widget makes sense, or whether a status change is perceivable. Those need play functions and people.

## Contracts as play functions

Take the contract from `canvas.md` and assert it:

```svelte
<script module>
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import { expect, fn, userEvent } from 'storybook/test';
  import Dialog from './Dialog.svelte';

  const { Story } = defineMeta({
    title: 'Organisms/Dialog',
    component: Dialog,
    args: { open: true, title: 'Delete project', onclose: fn() },
  });
</script>

<!-- Contract: focus moves into the dialog; Escape closes it; the dialog has an accessible name -->
<Story
  name="Default"
  play={async ({ args, canvas, userEvent }) => {
    const dialog = canvas.getByRole('dialog', { name: 'Delete project' });
    await expect(dialog).toContainElement(document.activeElement);

    await userEvent.keyboard('{Escape}');
    await expect(args.onclose).toHaveBeenCalled();
  }}
/>
```

Write the contract as a comment above the story, in the same words `canvas.md` uses. When the assertion fails later, the comment tells the next person whether the code or the contract is wrong.

## Contracts worth asserting

| Component kind | Assert |
|---|---|
| Button / icon button | Accessible name; disabled behaviour; focus visible after `tab()` |
| Form field | Label association (`getByLabelText` resolves); error text linked and present; error not colour-only |
| Dialog / drawer (modal) | Focus enters; Escape closes; focus returns to invoker; accessible name |
| Drawer (non-modal) | Background stays reachable; open/close is keyboard operable |
| Tabs | Arrow-key movement; correct `tab`/`tabpanel` roles; activation model is deliberate |
| Destructive confirmation | Names the affected object; destructive and cancel are distinguishable without colour |

Note what is *not* here: reflow, zoom, text spacing, and target size. Those are measured on the rendered implementation, not asserted in a story.

## Focus indicators

A focus ring that only appears on `:focus-visible` will not show in a story rendered without keyboard interaction. Give any component with a meaningful focus treatment a story whose play function tabs to it, so the ring is captured in whatever visual review or snapshot runs.
