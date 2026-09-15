<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import EmptyState from './EmptyState.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/EmptyState',
		component: EmptyState,
		tags: ['autodocs'],
		argTypes: { elevated: { control: 'boolean' } },
		args: { title: 'No accounts yet', elevated: false }
	});
</script>

<Story name="Default">
	{#snippet template(args)}
		<div style="inline-size: 480px;">
			<EmptyState {...args}>Add one to start seeing where your money goes.</EmptyState>
		</div>
	{/snippet}
</Story>

<!-- Content sits one step above the chrome, so a card inside a shell takes
     bg-elevated and a border rather than the library's bg-subtle (ADR 0014). -->
<Story name="Elevated in a shell" args={{ elevated: true }}>
	{#snippet template(args)}
		<div style="inline-size: 480px; padding: 32px; background: var(--color-bg-surface);">
			<EmptyState {...args}>There is nothing here yet.</EmptyState>
		</div>
	{/snippet}
</Story>
