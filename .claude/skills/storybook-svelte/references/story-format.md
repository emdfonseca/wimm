# Story format

## The file

Stories are `.stories.svelte` files next to the component, written with the Svelte CSF addon. Start from `assets/Component.stories.svelte`.

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

`title` mirrors the design library's layers: `Atoms/…`, `Molecules/…`, `Organisms/…`, `Templates/…`, `Pages/…`, plus `Foundations/…` for docs pages. Set it explicitly rather than relying on path-derived titles — the file path follows code organization, the sidebar follows design taxonomy (table in `SKILL.md`).

## Tags

`tags: ['autodocs']` generates the docs page. Two other tags earn their keep:

- `'!test'` on a story that is a visual fixture and should not run as a test.
- `'dev'` / custom tags to filter noisy fixtures out of the default sidebar view.

## Args, not prose

Every knob a consumer can turn is an arg with an `argTypes` control, so the docs page is generated from the real API. `fn()` from `storybook/test` for callback args — it makes the action visible in the panel and assertable in a play function.

## What does not belong in a story file

Data fetching, router setup, and multi-organism flows. If a story needs those, it is a connected route or a journey (`SKILL.md` table); the e2e test covers it.
