<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, within } from 'storybook/test';
	import LinkUnusableScreen from './LinkUnusableScreen.svelte';

	const { Story } = defineMeta({
		title: 'Pages/LinkUnusableScreen',
		component: LinkUnusableScreen,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen' }
	});
</script>

<!-- J01.B · the only branch in either journey. Expired, already used, replaced
     and never issued all render exactly this, and nothing here says which.
     A dead end offers no control, because a control that cannot work is worse
     than none. -->
<Story
	name="Default"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('heading', { level: 1 })).toHaveFocus();
		await expect(canvas.queryByRole('button')).toBeNull();
		await expect(canvas.queryByRole('link')).toBeNull();
		// No notice: the card is the message.
		await expect(canvas.queryByRole('alert')).toBeNull();
		await expect(canvas.getByText(/ask whoever set up your account/i)).toBeInTheDocument();
	}}
/>
