<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, within } from 'storybook/test';
	import SignedInLanding from './SignedInLanding.svelte';
	import Page from './Page.svelte';
	import EmptyState from '../molecules/EmptyState.svelte';
	import { destinations } from '../destinations.js';

	const { Story } = defineMeta({
		title: 'Templates/SignedInLanding',
		component: SignedInLanding,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen' },
		args: { memberName: 'Ana Reis', onsignout: fn() }
	});
</script>

<!-- Where the member is, marked. The shell marks nothing unless the route says
     which page this is, so the bug this catches is a caller that forgets: a
     sidebar that always says Overview navigates correctly and lies about where
     you are. -->
<Story
	name="Marks the page the member is on"
	args={{ destinations: destinations('/transactions') }}
	play={async ({ canvasElement }) => {
		const nav = within(canvasElement).getByRole('navigation', { name: 'Sections' });
		await expect(within(nav).getByRole('link', { name: 'Transactions' })).toHaveAttribute(
			'aria-current',
			'page'
		);
		await expect(within(nav).getByRole('link', { name: 'Overview' })).not.toHaveAttribute(
			'aria-current'
		);
	}}
>
	{#snippet template(args)}
		<SignedInLanding {...args}>
			<EmptyState title="Transactions" elevated>There is nothing here yet.</EmptyState>
		</SignedInLanding>
	{/snippet}
</Story>

<!-- Told nothing, it marks nothing: no marker is better than a wrong one. -->
<Story
	name="Marks nothing when the route did not say"
	play={async ({ canvasElement }) => {
		const nav = within(canvasElement).getByRole('navigation', { name: 'Sections' });
		for (const label of ['Overview', 'Transactions']) {
			await expect(within(nav).getByRole('link', { name: label })).not.toHaveAttribute(
				'aria-current'
			);
		}
	}}
>
	{#snippet template(args)}
		<SignedInLanding {...args}>
			<EmptyState title="Somewhere" elevated>There is nothing here yet.</EmptyState>
		</SignedInLanding>
	{/snippet}
</Story>

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
     top bar and the destinations move into a bar at the bottom. Select the
     Compact viewport to see it. A screen showing neither does not say which app
     it is or who is signed in, and one showing no navigation is a product with
     one destination. -->
<Story
	name="Compact carries the brand and the account"
	globals={{ viewport: { value: 'compact' } }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('banner')).toBeVisible();
		// Scoped to the bar: the hidden sidebar still holds a lockup in the DOM.
		await expect(within(canvas.getByRole('banner')).getByText('wimm')).toBeVisible();

		// The same destinations, in the bar at the bottom. display:none takes
		// the sidebar out of the accessibility tree, so the one navigation that
		// remains is the reachable one.
		const nav = canvas.getByRole('navigation', { name: 'Sections' });
		await expect(within(nav).getByRole('link', { name: 'Overview' })).toBeVisible();
		await expect(within(nav).getByRole('link', { name: 'Transactions' })).toBeVisible();
	}}
>
	{#snippet template(args)}
		<SignedInLanding {...args}>
			<EmptyState title="Signed in as Ana Reis" elevated>There is nothing here yet.</EmptyState>
		</SignedInLanding>
	{/snippet}
</Story>

<!-- Medium: the rail. The same destinations, icon-only, because labels do not
     fit beside a content column at this width (`WMlvF`). -->
<Story
	name="Medium is the rail"
	globals={{ viewport: { value: 'medium' } }}
	args={{ destinations: destinations('/transactions') }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const nav = canvas.getByRole('navigation', { name: 'Sections' });
		const transactions = within(nav).getByRole('link', { name: 'Transactions' });
		await expect(transactions).toBeVisible();
		await expect(transactions).toHaveAttribute('aria-current', 'page');
		await expect(within(nav).queryByText('Transactions', { selector: 'span' })).toBeNull();
	}}
>
	{#snippet template(args)}
		<SignedInLanding {...args}>
			<EmptyState title="Transactions" elevated>There is nothing here yet.</EmptyState>
		</SignedInLanding>
	{/snippet}
</Story>

<!-- Wide: the labelled sidebar at 264. -->
<Story
	name="Wide is the labelled sidebar"
	globals={{ viewport: { value: 'wide' } }}
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).getByRole('navigation', { name: 'Sections' })
		).toBeVisible();
	}}
>
	{#snippet template(args)}
		<SignedInLanding {...args}>
			<EmptyState title="Signed in as Ana Reis" elevated>There is nothing here yet.</EmptyState>
		</SignedInLanding>
	{/snippet}
</Story>

<!-- Ultra: the same labelled sidebar, at 288 (`NEIet`). No inspector region is
     filled — nothing in wimm opens a drawer yet, so the extra width goes to
     the page. -->
<Story
	name="Ultra widens the sidebar"
	globals={{ viewport: { value: 'ultra' } }}
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).getByRole('navigation', { name: 'Sections' })
		).toBeVisible();
	}}
>
	{#snippet template(args)}
		<SignedInLanding {...args}>
			<EmptyState title="Signed in as Ana Reis" elevated>There is nothing here yet.</EmptyState>
		</SignedInLanding>
	{/snippet}
</Story>

<!-- A page taller than the viewport scrolls inside the body, to both its ends,
     and the document itself never scrolls. Three ways this broke: Page sizing
     itself to the viewport and spilling the rest past its own bottom; the body
     centring an overflowing page so its top sat above the scroll origin; and a
     visually hidden label, absolutely positioned deep in the page, placed
     against the document and stretching it. scrollTop is clamped, so it proves
     nothing; the page's own edges against the body's are what is measured. -->
<Story
	name="A page taller than the viewport scrolls to both its ends"
	play={async ({ canvasElement }) => {
		const body = within(canvasElement).getByRole('main');
		const page = body.firstElementChild;
		if (!page) throw new Error('.body rendered no page to measure');
		const doc = canvasElement.ownerDocument.scrollingElement;
		if (!doc) throw new Error('no scrolling element to measure');
		await expect(doc.scrollHeight).toBeLessThanOrEqual(doc.clientHeight + 1);
		body.scrollTop = 0;
		const top = page.getBoundingClientRect().top - body.getBoundingClientRect().top;
		await expect(Math.abs(top)).toBeLessThan(2);
		body.scrollTop = body.scrollHeight;
		const bottom = body.getBoundingClientRect().bottom - page.getBoundingClientRect().bottom;
		await expect(Math.abs(bottom)).toBeLessThan(2);
	}}
>
	{#snippet template(args)}
		<SignedInLanding {...args}>
			<Page title="Overview">
				<!-- A block, not a flex item: an absolute child of a flex container
				     sits at the container's start, never deep in the page. -->
				<div style:inline-size="100%" style:flex-shrink="0">
					<div style:block-size="2000px"></div>
					<button type="button">Last row, at the very end of the content</button>
					<span
						style:position="absolute"
						style:inline-size="1px"
						style:block-size="1px"
						style:overflow="hidden"
						style:clip-path="inset(50%)">A label only a screen reader hears</span
					>
				</div>
			</Page>
		</SignedInLanding>
	{/snippet}
</Story>

<Story name="Empty body" tags={['!test']}>
	{#snippet template(args)}
		<SignedInLanding {...args} />
	{/snippet}
</Story>
