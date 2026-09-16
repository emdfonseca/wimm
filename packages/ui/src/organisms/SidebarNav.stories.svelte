<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import SidebarNav from './SidebarNav.svelte';

	const { Story } = defineMeta({
		title: 'Organisms/SidebarNav',
		component: SidebarNav,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen' },
		args: {
			memberName: 'Ana Reis',
			onsignout: fn(),
			destinations: [
				{ label: 'Overview', href: '/', current: true },
				{ label: 'Accounts' },
				{ label: 'Transactions' },
				{ label: 'Budgets' },
				{ label: 'Reports' },
				{ label: 'Settings' }
			]
		}
	});
</script>

<!-- Only Overview exists, so it is the only one enabled, and it is marked
     current. A page whose nav item is not marked is a page that does not say
     where you are. -->
<Story
	name="Only Overview exists"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const overview = canvas.getByRole('link', { name: 'Overview' });
		await expect(overview).toHaveAttribute('aria-current', 'page');
		// A destination that does not exist is absent, not greyed.
		await expect(canvas.queryByText('Accounts')).toBeNull();
	}}
/>

<Story
	name="Signs out"
	play={async ({ canvasElement, args }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'Sign out' }));
		await expect(args.onsignout).toHaveBeenCalledOnce();
	}}
/>

<Story
	name="Everything enabled"
	tags={['!test']}
	args={{
		destinations: [
			{ label: 'Overview', href: '/', current: true },
			{ label: 'Accounts', href: '/accounts' },
			{ label: 'Transactions', href: '/transactions' },
			{ label: 'Budgets', href: '/budgets' },
			{ label: 'Reports', href: '/reports' },
			{ label: 'Settings', href: '/settings' }
		]
	}}
/>
