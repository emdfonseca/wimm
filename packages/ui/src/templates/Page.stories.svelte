<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, within } from 'storybook/test';
	import Page from './Page.svelte';

	const { Story } = defineMeta({
		title: 'Templates/Page',
		component: Page,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen' },
		args: { title: 'Transactions' }
	});
</script>

<Story
	name="Default"
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).getByRole('heading', { name: 'Transactions', level: 1 })
		).toBeInTheDocument();
	}}
>
	{#snippet template(args)}
		<Page title={args.title}>
			<p>Content fills this slot: a table region, a list of banks, a form panel, an empty state.</p>
		</Page>
	{/snippet}
</Story>

<!-- Compact: the subsection list stacks above the content rather than beside
     it. Select the Compact viewport to see it. -->
<Story name="With a subsection list" globals={{ viewport: { value: 'compact' } }}>
	{#snippet template(args)}
		<Page title={args.title}>
			{#snippet sections()}
				<div style="padding: 12px; background: var(--color-bg-surface); border-radius: 8px;">
					Profile · Connections · Categories
				</div>
			{/snippet}
			<p>Content.</p>
		</Page>
	{/snippet}
</Story>

<Story name="With an action" tags={['!test']}>
	{#snippet template(args)}
		<Page title={args.title}>
			{#snippet action()}
				<button type="button">Add account</button>
			{/snippet}
			<p>Content.</p>
		</Page>
	{/snippet}
</Story>

<!-- The inspector region, disabled by every screen today: nothing in wimm
     opens a drawer yet, so nothing is promoted into it. Shown here only to
     prove the slot exists. -->
<Story name="With an inspector" tags={['!test']}>
	{#snippet template(args)}
		<Page title={args.title}>
			{#snippet inspector()}
				<div style="padding: 12px; background: var(--color-bg-elevated); block-size: 100%;">
					Inspector
				</div>
			{/snippet}
			<p>Content.</p>
		</Page>
	{/snippet}
</Story>

<Story name="Compact" globals={{ viewport: { value: 'compact' } }} tags={['!test']}>
	{#snippet template(args)}
		<Page title={args.title}><p>Content.</p></Page>
	{/snippet}
</Story>

<Story name="Medium" globals={{ viewport: { value: 'medium' } }} tags={['!test']}>
	{#snippet template(args)}
		<Page title={args.title}><p>Content.</p></Page>
	{/snippet}
</Story>

<Story name="Wide" globals={{ viewport: { value: 'wide' } }} tags={['!test']}>
	{#snippet template(args)}
		<Page title={args.title}><p>Content.</p></Page>
	{/snippet}
</Story>

<Story name="Ultra" globals={{ viewport: { value: 'ultra' } }} tags={['!test']}>
	{#snippet template(args)}
		<Page title={args.title}><p>Content.</p></Page>
	{/snippet}
</Story>
