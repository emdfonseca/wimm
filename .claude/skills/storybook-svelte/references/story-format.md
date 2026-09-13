# Story format

## The file

Stories are `.stories.svelte` files next to the component, written with the Svelte CSF addon.

```svelte
<script module>
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import { expect, fn } from 'storybook/test';
  import Button from './Button.svelte';

  const { Story } = defineMeta({
    title: 'Atoms/Button',
    component: Button,
    tags: ['autodocs'],
    args: {
      label: 'Save',
      variant: 'primary',
      onclick: fn(),
    },
    argTypes: {
      variant: { control: 'select', options: ['primary', 'secondary', 'ghost', 'destructive'] },
      size: { control: 'radio', options: ['sm', 'md', 'lg'] },
    },
  });
</script>

<Story name="Default" />

<Story name="Destructive" args={{ variant: 'destructive', label: 'Delete' }} />

<Story name="Disabled" args={{ disabled: true }} />

<Story name="LongContent" args={{ label: 'Save and continue to the next step' }} />
```

`defineMeta` goes in `<script module>` — the module script, not the instance script. Everything the stories share (component, default args, argTypes, tags, parameters) is declared once there.

## When a story needs custom markup

Use a snippet rather than a wrapper component:

```svelte
<Story name="WithIcon">
  {#snippet template(args)}
    <Button {...args}>
      <Icon name="download" slot="leading" />
      Download
    </Button>
  {/snippet}
</Story>
```

The older `<Template>` component is deprecated in the Svelte 5 addon; snippets replace it.

## Titles

`title` mirrors the design library's zones: `Atoms/…`, `Molecules/…`, `Organisms/…`, `Templates/…`, plus `Foundations/…` for docs pages. Set it explicitly rather than relying on path-derived titles — the file path follows code organization, and the sidebar should follow design taxonomy. See `references/taxonomy.md`.

## Tags

`tags: ['autodocs']` generates the docs page. Two other tags earn their keep:

- `'!test'` on a story that is a visual fixture and should not run as a test.
- `'dev'` / custom tags to filter noisy fixtures out of the default sidebar view.

## Args, not prose

Every knob a consumer can turn is an arg with an `argTypes` control, so the docs page is generated from the real API instead of described in Markdown that goes stale. `fn()` from `storybook/test` for callback args — it makes the action visible in the panel and assertable in a play function.

## What does not belong in a story file

Data fetching, router setup, and multi-organism flows. If a story needs those, it is a page (see `references/taxonomy.md`). Stories that mock half an application are slow, brittle, and prove less than the e2e test that should exist instead.
