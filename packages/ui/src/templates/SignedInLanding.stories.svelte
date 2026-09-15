<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, within } from 'storybook/test';
	import SignedInLanding from './SignedInLanding.svelte';
	import EmptyState from '../molecules/EmptyState.svelte';

	const { Story } = defineMeta({
		title: 'Templates/SignedInLanding',
		component: SignedInLanding,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen' },
		args: { memberName: 'Ana Reis', onsignout: fn() }
	});
</script>

<!-- No page header, deliberately: App header is a stack of breadcrumbs, a title
     row and tabs, and here all three are empty but the title, which repeats what
     the sidebar already marks as current. -->
<Story
	name="Default"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('navigation', { name: 'Sections' })).toBeInTheDocument();
		await expect(canvas.queryByRole('heading', { level: 1 })).toBeNull();
	}}
>
	{#snippet template(args)}
		<SignedInLanding {...args}>
			<EmptyState title="Signed in as Ana Reis" elevated>There is nothing here yet.</EmptyState>
		</SignedInLanding>
	{/snippet}
</Story>

<!-- At Compact the sidebar is gone, so the brand and the account move into a
     top bar. Select the Compact viewport to see it. A screen showing neither
     does not say which app it is or who is signed in. -->
<Story
	name="Compact carries the brand and the account"
	globals={{ viewport: { value: 'compact' } }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('banner')).toBeVisible();
		// display:none takes the sidebar out of the accessibility tree, so it is
		// absent rather than hidden — which is the point: nothing announces a
		// navigation that cannot be reached.
		await expect(canvas.queryByRole('navigation', { name: 'Sections' })).toBeNull();
		// Scoped to the bar: the hidden sidebar still holds a lockup in the DOM.
		await expect(within(canvas.getByRole('banner')).getByText('wimm')).toBeVisible();
	}}
>
	{#snippet template(args)}
		<SignedInLanding {...args}>
			<EmptyState title="Signed in as Ana Reis" elevated>There is nothing here yet.</EmptyState>
		</SignedInLanding>
	{/snippet}
</Story>

<Story name="Empty body" tags={['!test']}>
	{#snippet template(args)}
		<SignedInLanding {...args} />
	{/snippet}
</Story>
