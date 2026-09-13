<script module>
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import { expect, fn } from 'storybook/test';
  import Component from './Component.svelte';

  const { Story } = defineMeta({
    // Atoms | Molecules | Organisms | Templates | Pages — mirrors the design layers
    title: 'Atoms/Component',
    component: Component,
    tags: ['autodocs'],
    args: {
      label: 'Save',
      onclick: fn(),
    },
    argTypes: {
      variant: { control: 'select', options: ['primary', 'secondary', 'ghost', 'destructive'] },
      size: { control: 'radio', options: ['sm', 'md', 'lg'] },
    },
  });
</script>

<!-- One story per design state; theme and viewport are globals -->

<!-- Pin a regime only when the state exists in one: globals={{ viewport: { value: 'compact', isRotated: false } }} -->

<Story name="Default" />

<Story name="Disabled" args={{ disabled: true }} />

<Story name="LongContent" args={{ label: 'Save and continue to the next step' }} />

<!-- Contract (from the design library): keyboard focus reaches the control and is visible -->
<Story
  name="Focused"
  play={async ({ canvas, userEvent }) => {
    await userEvent.tab();
    await expect(canvas.getByRole('button', { name: 'Save' })).toHaveFocus();
  }}
/>

<!-- Custom markup uses a snippet, never a <Template> component (removed in Svelte 5) -->
<Story name="WithIcon">
  {#snippet template(args)}
    <Component {...args}>
      <span aria-hidden="true">↓</span>
      Download
    </Component>
  {/snippet}
</Story>
