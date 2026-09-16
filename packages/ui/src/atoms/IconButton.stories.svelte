<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import IconButton from './IconButton.svelte';

	const { Story } = defineMeta({
		title: 'Atoms/IconButton',
		component: IconButton,
		tags: ['autodocs'],
		argTypes: {
			icon: { control: 'select', options: ['log-out', 'panel-left-close'] },
			chrome: { control: 'boolean' }
		},
		args: { icon: 'log-out', label: 'Sign out', onclick: fn() }
	});
</script>

<!-- The label is the whole point: an icon with no accessible name announces
     itself as "button", which is how an icon-only control becomes unusable. -->
<Story
	name="Default"
	play={async ({ canvasElement, args }) => {
		const button = within(canvasElement).getByRole('button', { name: 'Sign out' });
		await userEvent.click(button);
		await expect(args.onclick).toHaveBeenCalledOnce();
	}}
/>

<!-- With its own fill and border, as on a specimen tile. In a shell footer it
     is drawn bare so the chrome does not compete with the account beside it. -->
<Story name="With chrome" args={{ chrome: true, icon: 'panel-left-close', label: 'Collapse' }} />
