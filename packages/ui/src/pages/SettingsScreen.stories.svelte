<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import SettingsScreen from './SettingsScreen.svelte';

	const { Story } = defineMeta({
		title: 'Pages/SettingsScreen',
		component: SettingsScreen,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen' },
		args: { ontheme: fn(), ondensity: fn() }
	});
</script>

<Story
	name="Default"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('heading', { name: 'Settings', level: 1 })).toBeInTheDocument();
		await expect(canvas.getByRole('heading', { name: 'Appearance', level: 2 })).toBeInTheDocument();
		await expect(canvas.getByRole('radiogroup', { name: 'Theme' })).toBeInTheDocument();
		await expect(canvas.getByRole('radiogroup', { name: 'Rows' })).toBeInTheDocument();
	}}
/>

<Story
	name="Choosing a theme"
	play={async ({ canvasElement, args }) => {
		const group = within(canvasElement).getByRole('radiogroup', { name: 'Theme' });
		await userEvent.click(within(group).getByRole('radio', { name: 'Dark' }));
		await expect(args.ontheme).toHaveBeenCalledWith('dark');
	}}
/>

<Story
	name="Choosing row height"
	play={async ({ canvasElement, args }) => {
		const group = within(canvasElement).getByRole('radiogroup', { name: 'Rows' });
		await userEvent.click(within(group).getByRole('radio', { name: 'Compact' }));
		await expect(args.ondensity).toHaveBeenCalledWith('compact');
	}}
/>

<!-- ADR 0004: on a touch screen the row-height control is hidden entirely,
     never disabled, and the helper says why rather than leaving a gap.
     `(any-pointer: coarse)` cannot be forced from Storybook's own controls, so
     this documents the state rather than asserting it with a play function. -->
<Story name="On a touch screen (see the rule in SettingsScreen.svelte)" tags={['!test']} />
