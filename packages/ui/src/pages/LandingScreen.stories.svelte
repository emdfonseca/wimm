<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import LandingScreen from './LandingScreen.svelte';

	const { Story } = defineMeta({
		title: 'Pages/LandingScreen',
		component: LandingScreen,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen' },
		args: { firstName: 'Ana', lastName: 'Reis', onsignout: fn() }
	});
</script>

<!-- J01.A / 03 and J02.A / 02 are the same screen. Both journeys end by naming
     the member, which is how they know the enrolment or the sign-in worked. -->
<Story
	name="Default"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Signed in as Ana Reis')).toBeInTheDocument();
		await expect(canvas.getByText('There is nothing here yet.')).toBeInTheDocument();
		await expect(canvas.getByRole('link', { name: 'Overview' })).toHaveAttribute(
			'aria-current',
			'page'
		);
	}}
/>

<Story
	name="Signs out"
	play={async ({ canvasElement, args }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'Sign out' }));
		await expect(args.onsignout).toHaveBeenCalledOnce();
	}}
/>
